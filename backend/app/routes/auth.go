package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type authPayload struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}

func RegisterAuth(e *echo.Echo, database *mongo.Database) {
	e.POST("/api/auth/login", loginHandler(database))
	e.POST("/api/auth/register", registerHandler(database))
}

func loginHandler(database *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in authPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "Missing userId or password",
			})
		}

		in.UserID = strings.TrimSpace(in.UserID)
		if in.UserID == "" || in.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "Missing userId or password",
			})
		}

		enc, err := encrypt(in.Password, 3, 1)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": err.Error(),
			})
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		var user bson.M
		err = database.Collection("users").FindOne(ctx, bson.M{
			"userId":   in.UserID,
			"password": enc,
		}).Decode(&user)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"ok":    false,
				"error": "Invalid credentials",
			})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "Database error",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":      true,
			"message": "Login successful",
			"user": map[string]interface{}{
				"userId": in.UserID,
			},
		})
	}
}

func registerHandler(database *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in map[string]interface{}
		if err := c.Bind(&in); err != nil || len(in) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Expected JSON object body",
			})
		}

		required := []string{"userId", "password"}
		missing := make([]string, 0, 2)
		for _, field := range required {
			if _, ok := in[field]; !ok {
				missing = append(missing, field)
			}
		}
		if len(missing) > 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Missing fields: " + strings.Join(missing, ", "),
			})
		}

		userID, okUser := in["userId"].(string)
		password, okPass := in["password"].(string)
		userID = strings.TrimSpace(userID)
		if !okUser || !okPass || userID == "" || password == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "Missing fields: userId, password",
			})
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		users := database.Collection("users")

		var existing bson.M
		err := users.FindOne(ctx, bson.M{"userId": userID}).Decode(&existing)
		if err == nil {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error": "userId already exists",
			})
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Database error",
			})
		}

		hashPassword, err := encrypt(password, 3, 1)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		doc := bson.M{
			"userId":   userID,
			"password": hashPassword,
		}

		res, err := users.InsertOne(ctx, doc)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to create user",
			})
		}

		if res.InsertedID == nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to create user",
			})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"user": map[string]interface{}{
				"userId": userID,
			},
		})
	}
}

func encrypt(inputText string, numShift int, dirShift int) (string, error) {
	if !isASCII(inputText) {
		return "", fmt.Errorf("input must be ASCII")
	}
	if numShift < 1 {
		return "", fmt.Errorf("NUM shift N must be >=1")
	}
	if dirShift != 1 && dirShift != -1 {
		return "", fmt.Errorf("Direction shift D must be either +1 or -1")
	}
	if strings.Contains(inputText, " ") || strings.Contains(inputText, "!") {
		return "", fmt.Errorf("Input contains forbidden ASCII codes 32 or 33 (!/SPACE)")
	}

	reversed := reverseASCII(inputText)

	shifted := make([]byte, 0, len(reversed))
	for i := 0; i < len(reversed); i++ {
		oldASCII := int(reversed[i])
		newASCII := oldASCII + (numShift * dirShift)

		if newASCII > 127 {
			newASCII -= 128
		} else if newASCII < 34 {
			newASCII += 128
		}

		shifted = append(shifted, byte(newASCII))
	}

	return string(shifted), nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func reverseASCII(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
