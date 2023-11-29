package main

import (
	"log/slog"
	"net/http"

	"factorial/internal"

	"github.com/julienschmidt/httprouter"
)

func main() {
	Run()
}

func Run() {
	router := httprouter.New()

	router.POST("/calculate", internal.FactorialMiddleware(internal.FactorialHandler))

	slog.Info("Starting server on port 8989")
	err := http.ListenAndServe(":8989", router)
	if err != nil {
		slog.Error("Failed to start server:", err)
	}
}
