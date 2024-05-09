package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/s6thgehr/sf-challenge/api"
	"github.com/s6thgehr/sf-challenge/internal/handlers"
)

func main() {
	rpc := os.Getenv("RPC_ENDPOINT")
	client, err := api.NewEthereumClient(rpc)
	if err != nil {
		log.Fatalf("Failed to connect to node: %v", err)
	}
	defer client.Close()

	r := gin.Default()
	r.GET("/blockreward/:slot", handlers.BlockRewardHandler(client, rpc))
	r.GET("/syncduties/:slot", handlers.SyncDutiesHandler(client, rpc))

	r.Run(":8080")
}
