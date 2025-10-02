package main

import (
	"fmt"
	"log"
	"main/features"
	"main/utils"
	"net/http"
	"os"

	_ "main/docs"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Workflow Backend API
// @version 1.0
// @description This is a workflow management backend server API
// @host localhost:7000
// @BasePath /
func main() {
	e := echo.New()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with environment variables")
	}

	// 서버 초기화
	if err := utils.InitServer(); err != nil {
		fmt.Printf("서버 초기화 에러 : %v", err.Error())
		return
	}

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Custom CORS middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get CORS origin from environment variable
			origin := os.Getenv("CORS_ORIGIN")
			if origin == "" {
				origin = "http://13.203.37.93:4000"
			}

			c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	})

	// Handler 초기화
	if err := features.InitHandler(e); err != nil {
		fmt.Printf("handler 초기화 에러 : %v", err.Error())
		return
	}

	// Swagger 초기화
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Get server configuration from environment variables
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "7000"
	}

	// Start server
	address := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Server starting on %s", address)
	if err := e.Start(address); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
