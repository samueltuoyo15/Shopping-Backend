package middleware

import (
	"context"
	"time"
	"fmt"
	"strings"
	"github.com/gofiber/fiber/v2"
	"firebase.google.com/go/auth"
)


func AuthenticateRequest(authClient *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error { 
		var token string
		if strings.HasPrefix(c.Get("Authorization"), "Bearer ") {
			token = strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		}

		if token == "" {
			token = c.Cookies("accessToken")
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized pls provide an accessToken",
			})
		}
		fmt.Println("Token Received:", token)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		verifiedToken, err := authClient.VerifyIDToken(ctx, token)
		if err != nil {
			fmt.Println("verifyIDToken:", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or expired token",
			})
		}

		c.Locals("user", verifiedToken.UID)

		return c.Next()
	}
}