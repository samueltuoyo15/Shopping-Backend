package internal

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
	cacheTtl = 30 * time.Minute
	queryTimeout = 1 * time.Second
)

func GetCategories(firestoreClient *firestore.Client, redisClient *redis.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
		log.Println("Get categories endpoint hit")
		cachedCategories, err := redisClient.Get(c.Context(), cacheKey).Result()
		if err == nil {
			var result fiber.Map
			if err := json.Unmarshal([]byte(cachedCategories), &result); err == nil {
				log.Println("serving from redis cache", result)
				return c.JSON(result)
				log.Println("returned categories from redis cache")
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
			data, err := json.Marshal(fiber.Map{
			"success": true,
			"count": len(categoryNames),
			"list": categoryNames,	
			"source": "redis_cache",	
			})
			
			if err != nil {
				log.Println("Json marshal error", nil)
			}
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