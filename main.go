package main

import (
	"gozurite/blobclient"
	"gozurite/expiryhelper"
	"gozurite/routes"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// periodically check for expired pins and remove them
	go CleanupExpiredPins()

	// Configure CORS middleware with allowed origins, methods, and headers
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "https://downloader98.apaluchdev.com"}, // Add your allowed origins here
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                                                          // Specify allowed HTTP methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},                                               // Allowed headers in requests
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},                                                        // Headers exposed to the browser
		AllowCredentials: true,                                                                                              // Allow credentials (cookies, etc.)
		MaxAge:           12 * time.Hour,                                                                                    // Preflight request cache duration
	}))

	routes.RegisterFileRoutes(router)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Get the port from the environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	router.Run(":" + port) // Listen and serve on the specified port
}

func CleanupExpiredPins() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		log.Println("Checking for expired pins...")
		for pin, expiryTime := range expiryhelper.GetPinExpiryMap() {
			if time.Now().After(expiryTime) {
				log.Printf("Pin %s expired, removing it", pin)

				// Delete the folder in the container and remove the pin from the expiry map
				err := blobclient.DeleteFolderInContainer(blobclient.FILES_CONTAINER_NAME, pin)
				if err != nil {
					log.Printf("Error deleting folder %s: %v", pin, err)
				}

				expiryhelper.RemovePinExpiry(pin)
			}
		}
	}
}
