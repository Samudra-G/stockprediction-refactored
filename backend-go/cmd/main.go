package main

import (
	"log"

	"github.com/Samudra-G/stockprediction-refactored/api"
	"github.com/gin-gonic/gin"
)

func main() {
	h := api.NewHandler()

	router := gin.Default()

	router.GET("/health", h.Health)
	router.POST("/metric", h.Metric)
	router.GET("/poll", h.Poll)

	log.Println("Go backend listening on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}