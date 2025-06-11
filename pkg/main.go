package main

import (
	"log"
	"os"
	"runtime"
	"net/http"
	"time"
	"github.com/gofiber/fiber/v2"
	"shopping-backend/controller"
	"shopping-backend/database"
	"shopping-backend/utils"
	"github.com/robfig/cron/v3"
	"shopping-backend/middleware"
)

var startTime = time.Now()
func main() {
	utils.LoadEnv()
	firebaseApp, err := database.InitFirebase()
	if err != nil {
		log.Fatalf("failed to connect to firebase: %v", err)
	}

	log.Println("Successfully initialized Firebase app")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Shopping backend api running")
	})
	
	app.Post("/api/auth/register", controller.RegisterUser(firebaseApp.Auth, firebaseApp.Client))
	app.Post("/api/auth/login", controller.LoginUser(firebaseApp.Auth, firebaseApp.Client))
	app.Get("/api/user/me", middleware.AuthenticateRequest(firebaseApp.Auth), controller.Me(firebaseApp.Client))
	app.Get("/api/keep-alive", func(c *fiber.Ctx) error {
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

	cronJob := cron.New()
	_, err = cronJob.AddFunc("@every 14m", func(){
		backendDomain := os.Getenv("BACKEND_DOMAIN")
	if backendDomain == "" {
		log.Fatal("Environment variable BACKEND_DOMAIN is not set")
	}

		log.Println("Cron job running and keeping render service up every 14 minutes")
		
		resp, err := http.Get(backendDomain + "/api/keep-alive")
		if err != nil {
			log.Printf("Keep-alive request failed: %v", err)
			return
		}
		defer resp.Body.Close()
		log.Printf("keep alive request completed with status: %s", resp.Status)
	})
	
	if err != nil {
		log.Fatalf("Failed to schedule cron job: %v", err)
	}

	cronJob.Start()
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Fatal(app.Listen(":" + port))
}

