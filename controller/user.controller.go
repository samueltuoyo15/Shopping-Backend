package controller

import (
	"context"
	"time"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"cloud.google.com/go/firestore"
)


func Me(firestoreClient *firestore.Client, redisClient *redis.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userId").(string)
		if !ok || userID == ""{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: missing userId",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cacheKey := "user:" + userID
		cachedUser, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var userData map[string]interface{}
			if err := json.Unmarshal([]byte(cachedUser), &userData); err == nil {
				return c.JSON(userData)
			}
		}
		userDoc, err := firestoreClient.Collection("users").Doc(userID).Get(ctx)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Usernot found",
			})
		}

		userData := userDoc.Data()

		userJSON, err := json.Marshal(userData)
		if err == nil {
			_ = redisClient.Set(ctx, cacheKey, userJSON, 10*time.Minute).Err()		}
		return c.JSON(userData)
	}
}
