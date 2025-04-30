package api

import (
	"net/http"

	"github.com/codytheroux96/go-mq/internal/app"
)

func Routes(app *app.Application) http.Handler {
	mux := http.NewServeMux()

	return mux
}
