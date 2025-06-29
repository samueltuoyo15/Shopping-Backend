package internal

import (
	"context"
	"encoding/json"
	"log"
	"io"
	"net/http"
	"github.com/redis/go-redis/v9"
	"time"
	"github.com/gofiber/fiber/v2"
	"cloud.google.com/go/firestore"
)

const (
	categoriesCacheKey = "categories:list"
	productsCacheKey = "products:list"
	cacheTtl = 30 * time.Minute
	queryTimeout = 1 * time.Second
)


type Product struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Price float64 `json:"price"`
	Description string `json:"description"`
	Category string `json:"category"`
	Image string `json:"image"`
	Rating ProductRating `json:"rating"`
}

type ProductRating struct {
	Rate float64 `json:"rate"`
	Count int `json:"count"`
}

func GetCategories(firestoreClient *firestore.Client, redisClient *redis.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
		log.Println("Get categories endpoint hit")
		cachedCategories, err := redisClient.Get(c.Context(), categoriesCacheKey).Result()
		if err == nil {
			c.Set("Content-Type", "application/json")
			log.Println("Serving categories from Redis cache")
			return c.SendString(cachedCategories)
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
			cachedData := fiber.Map{
			"success": true,
			"count": len(categoryNames),
			"list": categoryNames,	
			"source": "redis_cache",	
			}
			data, err := json.Marshal(cachedData)
			
			if err != nil {
				log.Println("Json marshal error", nil)
			}
			if err := redisClient.Set(context.Background(), categoriesCacheKey, data, cacheTtl).Err(); err != nil {
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

func GetProducts(firestoreClient *firestore.Client, redisClient *redis.Client) fiber.Handler{
	return func(c *fiber.Ctx) error {
			log.Println("Get products endpoint hit")
			start := time.Now()

			cachedProducts, err := redisClient.Get(context.Background(), productsCacheKey).Result()
			log.Println("redis read took:", time.Since(start))
			if err == nil {
				log.Println("Returned Products from redis cache and serving", time.Since(start))
				c.Set("Content-Type", "application/json")
		  		return c.SendString(cachedProducts)
				} 

			ctx, cancel := context.WithTimeout(c.Context(), queryTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://fakestoreapi.com/products/", nil)
			if err != nil {
				log.Println("Failed to fetch products", err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to fetch products",
				})
			}

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("failed to fetch products:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to fetch products",
				})
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				log.Println("Failed to read response", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to fetch products",
				})
			}
			
			var products []Product
			if err := json.Unmarshal(body, &products); err != nil {
					log.Println("Failed to parse json", err)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Internal Server Error",
					})
				}

			go func(){
				if err := redisClient.Set(context.Background(), productsCacheKey, body, cacheTtl).Err(); err != nil {
					log.Printf("failed to update redis cache with the products: %v", err)
				}
		  }()
		  c.Set("Content-Type", "application/json")
		  return c.Send(body)
	}
}