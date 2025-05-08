package app

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/codytheroux96/go-mq/internal/broker"
	"github.com/codytheroux96/go-mq/internal/repository"
)

type Application struct {
	Logger *slog.Logger
	Client *http.Client
	Repo   repository.Repository
	Broker *broker.Manager
}

func NewApplication() *Application {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := repository.NewInMemoryRepo()
	broker := broker.NewManager(repo)

	app := &Application{
		Logger: logger,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Repo:   repo,
		Broker: broker,
	}

	return app
}
