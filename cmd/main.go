package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	_ "github.com/sosshik/rest-user-management/cmd/docs"
	"github.com/sosshik/rest-user-management/cmd/internal/api"
	"github.com/sosshik/rest-user-management/cmd/internal/cache"
	"github.com/sosshik/rest-user-management/cmd/internal/database"
	"github.com/sosshik/rest-user-management/cmd/internal/rating"
	"github.com/sosshik/rest-user-management/pkg/config"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title User Managment API
// @version 1.0
// @description This is User Managment API

// @host localhost:8080

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env file")
	}

	cfg := config.GetConfig()

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Error parsing log level: %v, setting log level to info\n", err)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
		fmt.Printf("log level was set to %s\n", cfg.LogLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	fmt.Printf("config initialized\n")
}

func main() {

	cfg := config.GetConfig()

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Warn(err)
	}
	defer db.DB.Close()

	rating, err := rating.NewClickHouse(cfg)
	if err != nil {
		log.Warn(err)
	}

	api := api.API{DB: db, Cache: cache.NewRedis(cfg.Redis.Addr, cfg.Redis.DBIndex, cfg.Redis.ExpTimeSeconds), Rating: rating}

	e := echo.New()

	auth := e.Group("", middleware.BasicAuth(api.BasicAuth))

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.POST("/api/users", api.HandleCreateUserProfile)
	auth.POST("/api/users/login", api.HandleLogIn)
	e.PUT("/api/users/:id", api.HandleUpdateUserProfile, api.JWTMiddleware)
	e.PUT("/api/users/:id/password", api.HandleUpdateUserPassword, api.JWTMiddleware)
	e.GET("/api/users/:id", api.HandleGetUserById)
	e.GET("/api/users", api.HandleGetUsersList)
	e.DELETE("/api/users/:id", api.HandleDeleteUser, api.JWTMiddleware)
	e.POST("/api/vote", api.HandleVote, api.JWTMiddleware)
	e.PUT("/api/vote", api.HandleChangeVote, api.JWTMiddleware)

	e.Logger.Fatal(e.Start(":" + cfg.Port))

}
