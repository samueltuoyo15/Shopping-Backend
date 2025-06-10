package controller

import (
	"context"
	"os"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/gofiber/fiber/v2"
	"shopping-backend/utils"
	"firebase.google.com/go/auth"
	"cloud.google.com/go/firestore"
)

type RegisterUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Fullname string `json:"full_name" validate:"required,min=2,max=40"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type LoginUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type firebaseLoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
	ReturnSecureToken bool `json:"returnSecureToken"`
}

type firebaseLoginResponse struct {
	IDToken string `json:"idToken"`
	Email string `json:"email"`
	LocalID string `json:"localId"`
}


func RegisterUser(authClient *auth.Client, firestoreClient *firestore.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ "error": "Invalid Request Body"})
		}

	if err := utils.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ "error": err.Error() })
	}

	_, err := authClient.GetUserByEmail(context.Background(), req.Email)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{ "error": "User with this email already exists"})
	}

	userParams := (&auth.UserToCreate{}).Email(req.Email).Password(req.Password)

	user, err := authClient.CreateUser(context.Background(), userParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ "error": "Failed to create user: " + err.Error()})
	}

	_, err = firestoreClient.Collection("users").Doc(user.UID).Set(context.Background(), map[string]interface{}{
		"email": req.Email,
		"fullname": req.Fullname,
		"createdAt": firestore.ServerTimestamp,
	})
	if err != nil {
		_ = authClient.DeleteUser(context.Background(), user.UID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user record: " + err.Error(),
		})
	}

	
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully. Go ahead and login into your account",
	})
  }
}

func LoginUser(authClient *auth.Client, firestoreClient *firestore.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ "error": "Invalid Body Request"})
		}

		if err := utils.Validate.Struct(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ "error": err.Error()})
		}

		user, err := authClient.GetUserByEmail(context.Background(), req.Email)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{ "error": "User not found"})
		}


		payloadBytes, _ := json.Marshal(firebaseLoginRequest{
			Email: req.Email,
			Password: req.Password,
			ReturnSecureToken: true,
		})
		resp, err := http.Post(
			fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", os.Getenv("FIREBASE_API_KEY")),
			"application/json", bytes.NewBuffer(payloadBytes))
		
		
			if err != nil || resp.StatusCode != http.StatusOK{
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{ "error": "invalid credentials"})
			}
	
		var firebaseResp firebaseLoginResponse
		json.NewDecoder(resp.Body).Decode(&firebaseResp)
		c.Cookie(&fiber.Cookie{
			Name: "accessToken",
			Value: firebaseResp.IDToken,
			HTTPOnly: true,
			Secure: os.Getenv("GOLANG_ENV") == "production",
			SameSite: "Lax",
			Path: "/",
			Expires: time.Now().Add(5 * time.Minute),
		})
		return c.Status(fiber.StatusOK).JSON(fiber.Map {
			"message": "Login succesful",
			"uid": user.UID,
			"accessToken": firebaseResp.IDToken,
		}) 
	}
}