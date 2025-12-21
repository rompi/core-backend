package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rompi/core-backend/pkg/httpclient"
)

func main() {
	// Create a client with default configuration
	client := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Example 1: Simple GET request
	fmt.Println("=== Example 1: Simple GET ===")
	getExample(client)

	// Example 2: POST request with JSON
	fmt.Println("\n=== Example 2: POST with JSON ===")
	postExample(client)

	// Example 3: Query parameters
	fmt.Println("\n=== Example 3: Query Parameters ===")
	queryExample(client)
}

func getExample(client *httpclient.Client) {
	ctx := context.Background()

	resp, err := client.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	var post struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}

	if err := resp.JSON(&post); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	fmt.Printf("Post ID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Title)
}

func postExample(client *httpclient.Client) {
	ctx := context.Background()

	newPost := map[string]interface{}{
		"title":  "Example Post",
		"body":   "This is an example post created via httpclient",
		"userId": 1,
	}

	resp, err := client.Post(ctx, "/posts").
		JSON(newPost).
		Do()

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	var created struct {
		ID int `json:"id"`
	}

	if err := resp.JSON(&created); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	fmt.Printf("Created post with ID: %d\n", created.ID)
	fmt.Printf("Status: %d\n", resp.StatusCode)
}

func queryExample(client *httpclient.Client) {
	ctx := context.Background()

	resp, err := client.Get(ctx, "/posts").
		Query("userId", "1").
		Query("_limit", "3").
		Do()

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	var posts []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	if err := resp.JSON(&posts); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	fmt.Printf("Retrieved %d posts:\n", len(posts))
	for _, post := range posts {
		fmt.Printf("  - [%d] %s\n", post.ID, post.Title)
	}
}
