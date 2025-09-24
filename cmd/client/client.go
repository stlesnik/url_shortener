package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stlesnik/url_shortener/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close grpc connection")
		}
	}()

	client := api.NewURLShortenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Test SaveURL
	fmt.Println("Testing SaveURL...")
	resp, err := client.SaveURL(ctx, &api.SaveURLRequest{
		LongUrl: "https://www.google.com",
		UserId:  "test-user-123",
	})
	if err != nil {
		log.Printf("SaveURL failed: %v", err)
	} else {
		fmt.Printf("Short URL: %s, IsDouble: %v\n", resp.ShortUrl, resp.IsDouble)
	}

	// Test APIPrepareShortURL
	fmt.Println("\nTesting APIPrepareShortURL...")
	apiResp, err := client.APIPrepareShortURL(ctx, &api.APIPrepareShortURLRequest{
		LongUrl: "https://www.github.com",
	})
	if err != nil {
		log.Printf("APIPrepareShortURL failed: %v", err)
	} else {
		fmt.Printf("API Short URL: %s, IsDouble: %v\n", apiResp.ShortUrl, apiResp.IsDouble)
	}

	// Test PingDB
	fmt.Println("\nTesting PingDB...")
	pingResp, err := client.PingDB(ctx, &api.PingDBRequest{})
	if err != nil {
		log.Printf("PingDB failed: %v", err)
	} else {
		fmt.Printf("PingDB success: %v\n", pingResp.Success)
	}

	// Test APIGetStats
	fmt.Println("\nTesting APIGetStats...")
	statsResp, err := client.APIGetStats(ctx, &api.APIGetStatsRequest{})
	if err != nil {
		log.Printf("APIGetStats failed: %v", err)
	} else {
		fmt.Printf("Stats - URLs: %d, Users: %d\n", statsResp.UrlCount, statsResp.UserCount)
	}

	fmt.Println("\nAll tests completed!")
}
