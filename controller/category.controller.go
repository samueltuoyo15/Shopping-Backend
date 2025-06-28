package controller

import (
	"context"
	"encoding/json"
	"log"
	"github.com/redis/go-redis/v9"
	"time"
	"github.com/gofiber/fiber/v2"
	"cloud.google.com/go/firestore"
)

const (
	cacheKey = "categories:list"
	cacheTtl = 5 * time.Minute
	queryTimeout = 2 * time.Second
)

func GetCategories(firestoreClient *firestore.Client, redisClient *redis.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
		cachedCategories, err := redisClient.Get(c.Context(), cacheKey).Result()
		if err == nil {
			var result fiber.Map
			if err := json.Unmarshal([]byte(cachedCategories), &result); err == nil {
				return c.JSON(result)
			}
			log.Printf("Error decoding cache: %v", err)
		}

		ctx, cancel := context.WithTimeout(c.Context(), queryTimeout)
		defer cancel()

		iter := firestoreClient.Collection("categories").OrderBy("name", firestore.Asc).Documents(ctx)
		defer iter.Stop()

		var categoryNames []string 

		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}
			if name, ok := doc.Data()["name"].(string); ok {
				categoryNames = append(categoryNames, name)
			}
		}

		go func(){
			data, _ := json.Marshal(fiber.Map{
			"success": true,
			"count": len(categoryNames),
			"list": categoryNames,	
			"source": "database",	
			})

			if err := redisClient.Set(context.Background(), cacheKey, data, cacheTtl).Err(); err != nil {
				log.Printf("failed to update redis cache: %v", err)
			}
		}()

		return c.JSON(fiber.Map{
			"success": true,
			"count": len(categoryNames),
			"list": categoryNames,
			"source": "firestore",
		})
	}
}