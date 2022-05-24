package handlers

import (
	"context"
	"crypto/sha256"
	"go-api/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func getJwtSecret() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv("JWT_SECRET")
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	h := sha256.New()

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})

	if user.Username != "admin" || user.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})

		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)

	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(getJwtSecret()))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}

	c.JSON(http.StatusOK, jwtOutput)
}

// func (handler *AuthHandler) RefreshTokenHandler(c *gin.Context) {
// 	tokenValue := c.GetHeader("Authorization")
// 	claims := &Claims{}
// 	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(getJwtSecret()), nil
// 	})
// 	fmt.Println(tkn.Valid)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
// 		return
// 	}
// 	if tkn == nil || !tkn.Valid {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 		return
// 	}
// 	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Token is not expired yet",
// 		})
// 		return
// 	}
// 	expirationTime := time.Now().Add(5 * time.Minute)
// 	claims.ExpiresAt = expirationTime.Unix()
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(getJwtSecret())
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// 	jwtOutput := JWTOutput{
// 		Token:   tokenString,
// 		Expires: expirationTime,
// 	}
// 	c.JSON(http.StatusOK, jwtOutput)
// }

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(getJwtSecret()), nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if tkn == nil || !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Next()
	}
}

// For api key in header
// func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if c.GetHeader("X-API-KEY") != getApiKey() {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "API Key not provided or invalid",
// 			})

// 			c.AbortWithStatus(401)
// 		}

// 		c.Next()
// 	}
// }
