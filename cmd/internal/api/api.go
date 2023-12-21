package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/redis/go-redis/v9"

	"git.foxminded.ua/foxstudent106264/task-3.5/cmd/internal/domain"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	DB     domain.UserProfileManager
	Cache  domain.CacheInterface
	Rating domain.StatsManager
}

type CustomClaims struct {
	OID  uuid.UUID   `json:"oid"`
	Role domain.Role `json:"user_role"`
	jwt.StandardClaims
}

func (a *API) BasicAuth(username, password string, c echo.Context) (bool, error) {

	passwordHash, err := a.DB.GetPassword(username)
	if err != nil {
		log.Warn(err)
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		log.Warnf("auth: wrong password: %s", err)
		return false, err
	}
	return true, nil
}

func CheckPassword(psw string) error {

	if len(psw) < 8 {

		return errors.New("password is too short, should be at least 8 symbols")

	}

	var lower, upper, number, symbol bool

	for _, letter := range psw {

		if unicode.IsLower(letter) {
			lower = true
		}
		if unicode.IsUpper(letter) {
			upper = true
		}
		if unicode.IsNumber(letter) {
			number = true
		}
		if unicode.IsSymbol(letter) || unicode.IsPunct(letter) {
			symbol = true
		}
	}

	if lower && upper && number && symbol {
		return nil
	}
	return errors.New("wrong password format: password must contatin at least 1 upper case letter, 1 lower case letter, 1 number and 1 symbol")
}

func (a *API) JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString, _ := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing token"})
		}

		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_KEY")), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		claims, ok := token.Claims.(*CustomClaims)

		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error with authentication, please re-login."})
		}

		state, err := a.DB.GetUserState(claims.OID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error with authentication, please re-login."})
		}
		if state != 1 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Your profile is banned or deleted"})
		}

		c.Set("oid", claims.OID)
		c.Set("role", claims.Role)

		return next(c)
	}
}

func (a *API) CreateToken(nickname string) (string, error) {

	user, err := a.DB.GetUserForToken(nickname)
	if err != nil {
		return "", err
	}

	if user.State != domain.Active {
		return "", errors.New("unable to create JWT token: user is not in active status")
	}

	claims := &CustomClaims{
		OID:  user.OID,
		Role: user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", fmt.Errorf("unable to create JWT token: %w", err)
	}

	return tokenString, nil
}

// @Summary Create a user profile
// @Description Create a new user profile with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.CreateUserReq true "User profile details"
// @Success 201 {object} domain.CreateUserResp
// @Failure 400 {object} domain.ErrorResp "Invalid request payload"
// @Failure 500 {object} domain.ErrorResp "Failed to create user profile"
// @Router /users [post]
func (a *API) HandleCreateUserProfile(c echo.Context) error {

	var user domain.UserProfileDTO
	if err := c.Bind(&user); err != nil {
		log.Warnf("HandleCreateUserProfile - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	err := CheckPassword(user.Password)
	if err != nil {
		log.Warnf("HandleCreateUserProfile - user provided wrong password: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warnf("HandleCreateUserProfile - unable to generate hash for password: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unable to generate hash for password"})
	}
	user.Password = string(hash)
	user.OID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	user.State = domain.Active

	err = a.DB.CreateUserProfile(user)
	if err != nil {
		log.Warnf("HandleCreateUserProfile: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user profile"})
	}

	log.Infof("Successfully created user profile for user %s with oid %s", user.Nickname, user.OID.String())
	return c.JSON(http.StatusCreated, map[string]string{
		"oid":     user.OID.String(),
		"message": "User profile created successfully.",
	})
}

// @Summary Log in and generate JWT token
// @Description Log in with the provided credentials and generate a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.LoginReq true "User credentials"
// @Success 200 {object} domain.LoginResp
// @Failure 400 {object} domain.ErrorResp "Invalid request payload"
// @Failure 401 {object} domain.ErrorResp "Failed to log in"
// @Router /users/login [post]
func (a *API) HandleLogIn(c echo.Context) error {

	var user domain.UserProfileDTO
	if err := c.Bind(&user); err != nil {
		log.Warnf("HandleLogIn - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	token, err := a.CreateToken(user.Nickname)
	if err != nil {
		log.Warnf("HandleLogIn: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to log in"})
	}

	log.Infof("JWT token for user %s with oid %s", user.Nickname, user.OID.String())

	c.Response().Header().Set("x-auth-token", "Bearer "+token)

	return c.JSON(http.StatusOK, map[string]string{
		"token":   token,
		"message": "Successfully logged in",
	})

}

// @Summary Update user profile
// @Description Update an existing user profile with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body domain.UpdateUserReq true "User credentials"
// @Success 200 {object} domain.MessageResp
// @Failure 400 {object} domain.ErrorResp "Invalid request payload"
// @Failure 500 {object} domain.ErrorResp "Failed to update user profile"
// @Router /users/{id} [put]
func (a *API) HandleUpdateUserProfile(c echo.Context) error {
	userRoleFromAuth := c.Get("role").(domain.Role)
	userIDFromAuth := c.Get("oid")

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warnf("HandleUpdateUserProfile - unable to convert string to uuid: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if userID != userIDFromAuth && userRoleFromAuth == domain.Usr {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User is not permitted to change other profiles except his own."})
	}

	var updateUser domain.UserProfileDTO
	if err := c.Bind(&updateUser); err != nil {
		log.Warnf("HandleUpdateUserProfile - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	updateUser.UpdatedAt = time.Now().UTC()

	err = a.DB.UpdateUserProfile(updateUser, userID)
	if err != nil {
		log.Warnf("HandleUpdateUserProfile: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user profile"})
	}

	log.Infof("Successfully updated user profile for user %s with oid %s", updateUser.Nickname, userID)
	return c.JSON(http.StatusOK, map[string]string{"message": "User profile updated successfully."})
}

// @Summary Update user password
// @Description Update the password for the authenticated user or admin
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body domain.UpdatePasswordReq true "User credentials"
// @Success 200 {object} domain.MessageResp
// @Failure 400 {object} domain.ErrorResp "Invalid request payload"
// @Failure 500 {object} domain.ErrorResp "Failed to update user password"
// @Router /users/{id}/password [put]
func (a *API) HandleUpdateUserPassword(c echo.Context) error {

	userRoleFromAuth := c.Get("role").(domain.Role)
	userIDFromAuth := c.Get("oid").(uuid.UUID)

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warnf("HandleUpdateUserPassword - unable to convert string to uuid: %s", err)
		return err
	}

	if userID != userIDFromAuth && userRoleFromAuth == domain.Usr {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User is not permitted to change other profiles except his own."})
	}

	var updatePass domain.UserProfileDTO
	if err := c.Bind(&updatePass); err != nil {
		log.Warnf("HandleUpdateUserPassword - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	err = CheckPassword(updatePass.Password)
	if err != nil {
		log.Warnf("HandleUpdateUserPassword - user provided wrong password: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	newPass, err := bcrypt.GenerateFromPassword([]byte(updatePass.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warnf("HandleUpdateUserPassword - unable to generate hash for password: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unable to generate hash for password"})
	}

	err = a.DB.UpdatePassword(string(newPass), userID)
	if err != nil {
		log.Warnf("HandleUpdateUserPassword: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user profile"})
	}

	log.Infof("Successfully updated user password for user oid %s", userID.String())
	return c.JSON(http.StatusOK, map[string]string{"message": "User password updated successfully."})
}

// @Summary Get user by ID
// @Description Retrieve user details by the provided user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} domain.GetUserResp "User profile details"
// @Failure 400 {object} domain.ErrorResp "Wrong UserId"
// @Failure 500 {object} domain.ErrorResp "Failed to get user profile"
// @Router /users/{id} [get]
func (a *API) HandleGetUserById(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warnf("HandleGetUserById: unable to parse uuid: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Wrong UserId"})
	}

	user, err := a.Cache.GetUser(userID.String())
	if err == nil {
		return c.JSON(http.StatusOK, user)
	}
	if err != nil && err != redis.Nil {
		log.Warnf("HandleGetUserById: %s", err)
	}

	user, err = a.DB.GetUserById(userID)
	if err != nil {
		log.Warnf("HandleGetUserById: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user profile"})
	}

	rating, err := a.Rating.GetRatingSeparately(user.OID)
	if err != nil {
		log.Warnf("HandleGetUserById: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user profile"})
	}

	err = a.Cache.Set(userID.String(), user)
	if err != nil {
		log.Warnf("HandleGetUserById: unable to save cache: %s", err)
	}

	return c.JSON(http.StatusOK, domain.GetProfileDTO{
		OID:       user.OID,
		Nickname:  user.Nickname,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		State:     user.State,
		Role:      user.Role,
		Rating:    rating,
	})
}

// @Summary Get a paginated list of users
// @Description Retrieve a paginated list of user profiles
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} domain.GetUserListResp "Paginated list of user profiles"
// @Failure 500 {object} domain.ErrorResp "Failed to get users list"
// @Router /users [get]
func (a *API) HandleGetUsersList(c echo.Context) error {

	pageNumber, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	pageSize, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || pageSize < 1 {
		pageSize = defaultPageSize
	}

	offset := (pageNumber - 1) * pageSize

	usersList, err := a.Cache.GetUsersList(a.Cache.MakeKey(pageSize, offset))
	if err == nil {
		return c.JSON(http.StatusOK, usersList)
	}
	if err != nil && err != redis.Nil {
		log.Warnf("HandleGetUsersList: %s", err)
	}

	users, err := a.DB.GetUsersList(pageSize, offset)
	if err != nil {
		log.Warnf("HandleGetUsersList: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users list"})
	}

	var oids []uuid.UUID

	for _, user := range users {
		oids = append(oids, user.OID)
	}

	ratings, err := a.Rating.GetRatingForList(oids)
	if err != nil {
		log.Warnf("HandleGetUsersList: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users list"})
	}

	fmt.Println(len(users), ratings)

	for i, user := range users {

		if rating, ok := ratings[user.OID]; ok {
			users[i].Rating = rating
		}

	}

	totalUsers := len(users)

	if len(users) == pageSize || pageNumber > 0 {
		totalUsers, err = a.DB.GetUsersCount()
		if err != nil {
			log.Warnf("HandleGetUsersList: %s", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users amount"})
		}
	}

	usersList.CurrentPage = pageNumber
	usersList.TotalItems = totalUsers
	usersList.Users = users

	err = a.Cache.Set(a.Cache.MakeKey(pageSize, pageNumber), usersList)
	if err != nil {
		log.Warnf("HandleGetUsersList: unable to save cache: %s", err)
	}

	return c.JSON(http.StatusOK, usersList)

}

// @Summary Delete user by ID
// @Description Delete a user profile by the provided user ID
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {string} string "Profile successfully deleted"
// @Failure 400 {object} domain.ErrorResp
// @Failure 500 {object} domain.ErrorResp
// @Router /users/{id} [delete]
func (a *API) HandleDeleteUser(c echo.Context) error {
	userRoleFromAuth := c.Get("role").(domain.Role)
	userIDFromAuth := c.Get("oid").(uuid.UUID)

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Warnf("HandleUpdateUserProfile - unable to convert string to uuid: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if userID != userIDFromAuth && userRoleFromAuth != domain.Admin {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User is not permitted to change other profiles except his own."})
	}

	err = a.DB.DeleteUser(userID)
	if err != nil {
		log.Warnf("HandleUpdateUserProfile: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error happaned, unable to delete profile"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Profile successfully deleted"})
}

// @Summary Vote
// @Description Vote for a user by id
// @Tags vote
// @Accept json
// @Produce json
// @Param vote body domain.VoteReq true "Vote credentials"
// @Success 200 {object} domain.MessageResp
// @Failure 400 {object} domain.ErrorResp
// @Failure 500 {object} domain.ErrorResp
// @Router /vote [post]
func (a *API) HandleVote(c echo.Context) error {
	userIDFromAuth := c.Get("oid").(uuid.UUID)
	var vote domain.VoteDTO
	if err := c.Bind(&vote); err != nil {
		log.Warnf("HandleVote - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	vote.FromOID = userIDFromAuth
	vote.VotedAt = time.Now().UTC()
	if vote.FromOID == vote.ToOID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "You can't rate yourself"})
	} else if vote.EmojiId > 5 || vote.EmojiId < 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Wrong value"})
	}

	_, voteExists, err := a.Rating.GetVote(vote)
	if err != nil && err != sql.ErrNoRows {
		log.Warnf("HandleVote: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to check existance of vote"})
	}
	if voteExists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "You already rated this user"})
	}

	lastVoted, err := a.Rating.LastVotedAt(vote)
	if err != nil {
		log.Warnf("HandleVote: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to existance of vote"})
	}
	if lastVoted.Add(1 * time.Hour).After(time.Now().UTC()) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "You already voted in last hour, users can vote only one time per hour"})
	}

	err = a.Rating.RateProfile(vote)
	if err != nil {
		log.Warnf("HandleVote: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to change the rating"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Your vote has been submitted"})
}

// @Summary Change Vote
// @Description Change Vote for a user by id
// @Tags vote
// @Accept json
// @Produce json
// @Param vote body domain.VoteReq true "Vote credentials"
// @Success 200 {object} domain.MessageResp
// @Failure 400 {object} domain.ErrorResp
// @Failure 500 {object} domain.ErrorResp
// @Router /vote [put]
func (a *API) HandleChangeVote(c echo.Context) error {
	userIDFromAuth := c.Get("oid").(uuid.UUID)
	var vote domain.VoteDTO
	if err := c.Bind(&vote); err != nil {
		log.Warnf("HandleChangeVote - unable to decode JSON: %s", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	vote.FromOID = userIDFromAuth
	vote.VotedAt = time.Now().UTC()

	if vote.EmojiId > 5 || vote.EmojiId < 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Wrong value"})
	}

	dbVote, _, err := a.Rating.GetVote(vote)
	if err != nil {
		log.Warnf("HandleChangeVote: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to get the vote"})
	}

	if dbVote.EmojiId == vote.EmojiId {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Vote is the same as before"})
	}
	err = a.Rating.UpdateProfileRating(vote, dbVote.EmojiId)
	if err != nil {
		log.Warnf("HandleChangeVote: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to change the vote"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Your vote has been changed"})
}
