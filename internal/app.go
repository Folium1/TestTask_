package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Run() {

	router := httprouter.New()

	router.POST("/calculate", factorialMiddleware(factorialHandler))

	slog.Info("Starting server on port 8989")
	err := http.ListenAndServe(":8989", router)
	if err != nil {
		slog.Error("Failed to start server:", err)
	}
}

type requestData struct {
	A uint64 `json:"a"`
	B uint64 `json:"b"`
}

type responseData struct {
	AFactorial uint64 `json:"a"`
	BFactorial uint64 `json:"b"`
}

// factorialHandler handles concurrently factorials of 2 nums and returns them in json
func factorialHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := r.Context().Value("data")

	convertedMap := data.(map[string]uint64)
	a := convertedMap["a"]
	b := convertedMap["b"]

	aCh := make(chan uint64)
	bCh := make(chan uint64)
	go сalculateFactorial(a, aCh)
	go сalculateFactorial(b, bCh)

	// wait for goroutines to finish calculation
	outputData := responseData{
		AFactorial: <-aCh,
		BFactorial: <-bCh,
	}

	json.NewEncoder(w).Encode(outputData)
}

// factorialMiddleware decoding json request,validates
// then calls next with values in the context
func factorialMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		slog.Info("/calculate POST")

		var reqData requestData

		// make validation and decoding json, only positive numbers will pass
		err := json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil {
			slog.Error(err.Error())
			handleError(w, fmt.Errorf("Incorrect input"))
			return
		}

		// put values in context
		ctx := context.WithValue(r.Context(), "data", map[string]uint64{
			"a": reqData.A,
			"b": reqData.B,
		})

		next(w, r.WithContext(ctx), p)
	}
}

func сalculateFactorial(num uint64, ch chan uint64) {
	var factorial uint64 = 1
	for i := uint64(2); i <= num; i++ {
		factorial *= i
	}
	ch <- factorial
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}
