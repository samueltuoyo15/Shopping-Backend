package controller

import (
	"context"
	"time"
	"github.com/gofiber/fiber/v2"
	"cloud.google.com/go/firestore"
)


func Me(firestoreClient *firestore.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userId").(string)
		if !ok || userID == ""{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: missing userId",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userDoc, err := firestoreClient.Collection("users").Doc(userID).Get(ctx)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Usernot found",
			})
		}

		return c.JSON(userDoc.Data())
	}
}
