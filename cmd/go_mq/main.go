package main

import (
	"fmt"
	"time"

	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/codytheroux96/go-mq/internal/repository"
)

func main() {
	// ---- Fetch/Commit Offset Flow Test ----
	fmt.Println("\n--- Live Fetch/Commit Offset Flow Test ---")

	repo := repository.NewInMemoryRepo()

	err := repo.CreateTopic("test-topic")
	if err != nil {
		panic(err)
	}

	for i := 1; i <= 5; i++ {
		msg := &core.Message{
			Body: []byte(fmt.Sprintf("Message %d", i)),
		}
		if err := repo.Publish("test-topic", msg); err != nil {
			panic(err)
		}
	}

	messages, err := repo.Fetch("test-topic", "consumer-1", 3)
	if err != nil {
		panic(err)
	}
	fmt.Println("Fetched messages:")
	for _, msg := range messages {
		fmt.Println(string(msg.Body))
	}

	if err := repo.CommitOffset("test-topic", "consumer-1", 3); err != nil {
		panic(err)
	}

	messages, err = repo.Fetch("test-topic", "consumer-1", 3)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nFetched messages after committing offset:")
	for _, msg := range messages {
		fmt.Println(string(msg.Body))
	}

	// ---- Pub/Sub Test ----
	fmt.Println("\n--- Live Subscriber Test ---")

	inbox, err := repo.Subscribe("test-topic", "consumer-2")
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range inbox {
			fmt.Printf("[consumer-2] Received: %s\n", string(msg.Body))
		}
	}()

	for i := 6; i <= 10; i++ {
		msg := &core.Message{
			Body: []byte(fmt.Sprintf("Live Message %d", i)),
		}
		if err := repo.Publish("test-topic", msg); err != nil {
			panic(err)
		}
	}

	time.Sleep(200 * time.Millisecond)
}
