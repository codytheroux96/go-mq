package main

import (
	"fmt"

	"github.com/codytheroux96/go-mq/internal/core"
	"github.com/codytheroux96/go-mq/internal/repository"
)

func main() {
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
}