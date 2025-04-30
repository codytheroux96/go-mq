package app

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/codytheroux96/go-mq/internal/repository"
)

type Application struct {
	Logger *slog.Logger
	Client *http.Client
	Repo   repository.Repository
}

func NewApplication() *Application {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := repository.NewInMemoryRepo()

	app := &Application{
		Logger: logger,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Repo: repo,
	}

	return app
}
