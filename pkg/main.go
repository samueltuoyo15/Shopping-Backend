package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"shopping-backend/controller"
	"shopping-backend/database"
	"shopping-backend/utils"
	"shopping-backend/middleware"
)

func main() {
	utils.LoadEnv()
	firebaseApp, err := database.InitFirebase()
	if err != nil {
		log.Fatalf("failed to connect to firebase: %v", err)
	}

	log.Println("Successfully connected to Firebase", firebaseApp)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to shopping backend api")
	})
	
	app.Post("/api/auth/register", controller.RegisterUser(firebaseApp.Auth, firebaseApp.Client))
	app.Post("/api/auth/login", controller.LoginUser(firebaseApp.Auth, firebaseApp.Client))
	app.Get("/api/user/me", middleware.AuthenticateRequest(firebaseApp.Auth), controller.Me(firebaseApp.Client))

	log.Fatal(app.Listen(":5000"))
}
