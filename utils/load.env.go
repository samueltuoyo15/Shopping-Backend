package utils

import (
	"github.com/joho/godotenv"
	"fmt"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Failed to load .env file")
	}
}
