package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codytheroux96/go-mq/internal/api"
	"github.com/codytheroux96/go-mq/internal/app"
)

// func main() {
// 	// ---- Fetch/Commit Offset Flow Test ----
// 	fmt.Println("\n--- Live Fetch/Commit Offset Flow Test ---")

// 	repo := repository.NewInMemoryRepo()

// 	err := repo.CreateTopic("test-topic")
// 	if err != nil {
// 		panic(err)
// 	}

// 	for i := 1; i <= 5; i++ {
// 		msg := &core.Message{
// 			Body: []byte(fmt.Sprintf("Message %d", i)),
// 		}
// 		if err := repo.Publish("test-topic", msg); err != nil {
// 			panic(err)
// 		}
// 	}

// 	messages, err := repo.Fetch("test-topic", "consumer-1", 3)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Fetched messages:")
// 	for _, msg := range messages {
// 		fmt.Println(string(msg.Body))
// 	}

// 	if err := repo.CommitOffset("test-topic", "consumer-1", 3); err != nil {
// 		panic(err)
// 	}

// 	messages, err = repo.Fetch("test-topic", "consumer-1", 3)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("\nFetched messages after committing offset:")
// 	for _, msg := range messages {
// 		fmt.Println(string(msg.Body))
// 	}

// 	// ---- Pub/Sub Test ----
// 	fmt.Println("\n--- Live Subscriber Test ---")

// 	inbox, err := repo.Subscribe("test-topic", "consumer-2")
// 	if err != nil {
// 		panic(err)
// 	}

// 	go func() {
// 		for msg := range inbox {
// 			fmt.Printf("[consumer-2] Received: %s\n", string(msg.Body))
// 		}
// 	}()

// 	for i := 6; i <= 10; i++ {
// 		msg := &core.Message{
// 			Body: []byte(fmt.Sprintf("Live Message %d", i)),
// 		}
// 		if err := repo.Publish("test-topic", msg); err != nil {
// 			panic(err)
// 		}
// 	}

// 	time.Sleep(200 * time.Millisecond)
// }

func main() {
	app := app.NewApplication()

	handler := api.Routes(app)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		app.Logger.Info("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	app.Logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Error("Graceful shutdown failed", "error", err)
	} else {
		app.Logger.Info("Server shut down cleanly")
	}
}
