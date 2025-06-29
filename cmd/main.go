package main

import (
	"log"
	"runtime"
	"github.com/redis/go-redis/v9"
	"time"
	"os"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"shopping-backend/internal"
	"shopping-backend/internal/middleware"
	"shopping-backend/database"
	"shopping-backend/utils"
)

var startTime = time.Now()
func main() {
	utils.LoadEnv()
	firebaseApp, err := database.InitFirebase()
	if err != nil {
		log.Fatalf("failed to connect to firebase: %v", err)
	}

	log.Println("Successfully initialized Firebase app")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis")
	defer redisClient.Close()
	
	app := fiber.New()
	app.Use(helmet.New())
	

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Shopping backend api running")
	})
	
	app.Post("/api/auth/register", internal.RegisterUser(firebaseApp.Auth, firebaseApp.Client))
	app.Post("/api/auth/login", internal.LoginUser(firebaseApp.Auth, firebaseApp.Client))
	app.Get("/api/categories/getCategories", internal.GetCategories(firebaseApp.Client, redisClient))
	app.Get("/api/categories/getProducts", internal.GetProducts(firebaseApp.Client, redisClient))
	app.Get("/api/user/me", middleware.AuthenticateRequest(firebaseApp.Auth), internal.Me(firebaseApp.Client, redisClient))
	app.Get("/api/health-check", func(c *fiber.Ctx) error {
		var memoryUsg runtime.MemStats
		runtime.ReadMemStats(&memoryUsg)

		return c.JSON(fiber.Map{
			"status": "Ok",
			"uptime": time.Since(startTime).String(),
			"memoryUsage": fiber.Map{
				"alloc": memoryUsg.Alloc,
				"totalAlloc": memoryUsg.TotalAlloc,
				"sys": memoryUsg.Sys,
				"numGC": memoryUsg.NumGC,
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Fatal(app.Listen(":" + port))
}

