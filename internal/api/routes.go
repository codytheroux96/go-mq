package api

import (
	"net/http"

	"github.com/codytheroux96/go-mq/internal/app"
)

func Routes(app *app.Application) http.Handler {
	mux := http.NewServeMux()
	handler := &Handler{App: app}

	mux.HandleFunc("/topics", handler.HandleTopics)
	mux.HandleFunc("/topics/", handler.HandleDeleteTopic)

	mux.HandleFunc("/publish/", handler.HandlePublish)

	mux.HandleFunc("/subscribe/", handler.HandleSubscribe)
	mux.HandleFunc("/subscribe", handler.HandleRegisterConsumer)

	mux.HandleFunc("/ack", handler.HandleAck)

	return mux
}
