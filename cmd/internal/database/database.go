package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"git.foxminded.ua/foxstudent106264/task-3.5/cmd/internal/domain"
	"git.foxminded.ua/foxstudent106264/task-3.5/pkg/config"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type DBConfig struct {
	DbUrl       string `env:"DATABASE_URL"`
	ReconnTime  int    `env:"RECONN_TIME" envDefault:"5"`
	ConnCheck   bool   `env:"CONN_CHECK" envDefault:"true"`
	ReconnTries int    `env:"RECONN_TRIES" envDefault:"5"`
}

type Database struct {
	config *DBConfig
	DB     *sql.DB
}

var once sync.Once

var dbinstance *Database

func NewDatabase(cfg *config.Config) (*Database, error) {

	if dbinstance == nil {
		once.Do(func() {
			db, err := sql.Open("postgres", cfg.DbUrl)
			if err != nil {
				log.Warnf("unable to create db instance: %s", err)
			}

			dbinstance = &Database{&DBConfig{cfg.DbUrl, cfg.ReconnTime, cfg.ConnCheck, cfg.ReconnTries}, db}

			if cfg.ConnCheck {
				go dbinstance.connectionCheck(cfg.DbUrl)
			}
		})

	}

	return dbinstance, nil
}

func (d *Database) connectionCheck(conn string) {
	log.Info("Connection check started")
	var i int
	for {
		time.Sleep(time.Duration(d.config.ReconnTime) * time.Second)
		if err := d.DB.Ping(); err != nil {
			log.Warnf("Lost connection to Database. Attempting to reconnect.")
			if err := d.DB.Close(); err != nil {
				log.Warnf("Error while disconecting: %s", err)
				continue
			}
			if i <= d.config.ReconnTries {
				d.DB, err = sql.Open("postgres", conn)
				if err != nil {
					log.Warnf("Failed to reconnect: %s", err)
					i++
				} else {
					log.Infof("Reconnected to PostgreSQL!")
					i = 0
				}
			} else {
				break
			}

		}
	}
}

func (d *Database) GetPassword(username string) (string, error) {
	var passwordHash string
	err := d.DB.QueryRow(`
		CALL  public.get_user_password($1,$2)
	`, username, &passwordHash).Scan(&passwordHash)
	if err != nil {
		return "", fmt.Errorf("auth: unable to execute query to DB: %w", err)
	}
	return passwordHash, nil
}

func (d *Database) CreateUserProfile(user domain.UserProfileDTO) error {

	_, err := d.DB.Exec(`
		CALL public.create_profile($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, user.OID, user.Nickname, user.FirstName, user.LastName, user.Password, user.CreatedAt, user.UpdatedAt, user.State, user.Role)
	if err != nil {
		return fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return nil
}

func (d *Database) UpdateUserProfile(user domain.UserProfileDTO, userID uuid.UUID) error {

	_, err := d.DB.Exec(`
		CALL public.update_profile($1,$2,$3,$4,$5)
	`, user.Nickname, user.FirstName, user.LastName, user.UpdatedAt, userID)
	if err != nil {
		return fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return nil
}

func (d *Database) UpdatePassword(newPass string, userID uuid.UUID) error {

	_, err := d.DB.Exec(`
		CALL public.update_password($1,$2,$3)
	`, newPass, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return nil

}

func (d *Database) GetUserById(userID uuid.UUID) (domain.UserProfileDTO, error) {
	var user UserProfile
	err := d.DB.QueryRow(`
		CALL public.get_user($1, $2, $3, $4, $5, $6, $7,$8);
	`, userID, &user.Nickname, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.State, &user.Role).Scan(&user.Nickname, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.State, &user.Role)
	if err != nil {
		return domain.UserProfileDTO{}, fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return domain.UserProfileDTO{
		OID:       userID,
		Nickname:  user.Nickname,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		State:     user.State,
		Role:      user.Role,
		Rating:    user.Rating,
	}, nil
}

func (d *Database) GetUsersList(pageSize int, offset int) ([]domain.UserProfileDTO, error) {
	rows, err := d.DB.Query(`
		SELECT * FROM public.get_all_users($1,$2);
	`, pageSize, offset)
	if err != nil {
		return []domain.UserProfileDTO{}, fmt.Errorf("unable to execute query to DB: %w", err)
	}
	defer rows.Close()

	var users []domain.UserProfileDTO

	for rows.Next() {
		var user UserProfile
		err := rows.Scan(&user.OID, &user.Nickname, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt, &user.State, &user.Role, &user.Rating)
		if err != nil {
			return []domain.UserProfileDTO{}, fmt.Errorf("unable to scan row from DB: %w", err)
		}

		users = append(users, domain.UserProfileDTO{
			OID:       user.OID,
			Nickname:  user.Nickname,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			State:     user.State,
			Role:      user.Role,
			Rating:    user.Rating,
		})
	}
	return users, nil
}

func (d *Database) GetUsersCount() (int, error) {
	var totalUsers int
	err := d.DB.QueryRow(`CALL public.get_count($1);`, &totalUsers).Scan(&totalUsers)
	if err != nil {
		return 0, fmt.Errorf("GetUsersCount: unable to execute query to DB: %w", err)
	}
	return totalUsers, nil
}

func (d *Database) GetUserForToken(nickname string) (domain.UserProfileDTO, error) {

	var user UserProfile

	err := d.DB.QueryRow(`
	CALL public.get_user_for_token($1,$2,$3,$4,$5)
	`, nickname, &user.OID, &user.Nickname, &user.Role, &user.State).Scan(&user.OID, &user.Nickname, &user.Role, &user.State)
	if err != nil {
		return domain.UserProfileDTO{}, fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return domain.UserProfileDTO{
		OID:      user.OID,
		Nickname: user.Nickname,
		Role:     user.Role,
		State:    user.State,
	}, nil
}

func (d *Database) GetUserState(oid uuid.UUID) (int, error) {
	var state int
	err := d.DB.QueryRow(`
	CALL public.get_user_state($1,$2)
	`, oid, &state).Scan(&state)
	if err != nil {
		return 0, fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return state, nil
}

func (d *Database) DeleteUser(oid uuid.UUID) error {
	_, err := d.DB.Exec(`
	CALL public.delete_user($1)
	`, oid)
	if err != nil {
		return fmt.Errorf("unable to execute query to DB: %w", err)
	}
	return nil
}
