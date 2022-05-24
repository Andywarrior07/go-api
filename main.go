package main

import (
	"context"
	"go-api/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func getMongoUri() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv("MONGO_URI")
}

var authHandler *handlers.AuthHandler
var recipesHandler *handlers.RecipesHandler

func init() {
	// recipes = make([]Recipe, 0)
	// file, _ := ioutil.ReadFile("recipes.json")
	// _ = json.Unmarshal([]byte(file), &recipes)

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(getMongoUri()))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	collection := client.Database("recipesdb").Collection("recipes")

	log.Println("Conntected to MongoDB")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
		DB:   0,
	})

	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	authHandler = &handlers.AuthHandler{}

	// var listOfRecipes []interface{}
	// for _, recipe := range recipes {
	// 	recipe.ID = primitive.NewObjectID()
	// 	listOfRecipes = append(listOfRecipes, recipe)
	// }
	// collection := client.Database("recipesdb").Collection("recipes")
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

func main() {
	router := gin.Default()

	router.GET("/recipes", recipesHandler.GetRecipesHandler)

	router.POST("/signin", authHandler.SignInHandler)
	// router.POST("/refresh", authHandler.RefreshTokenHandler)

	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.GET("/recipes/:id", recipesHandler.GetRecipeByIdHandler)
		authorized.POST("/recipes", recipesHandler.CreateRecipeHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	}

	router.Run()
}
