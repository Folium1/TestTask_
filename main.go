package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.POST("/calculate", calculateMiddleware(calculateHandler))

	fmt.Println("Starting server on port 8989")
	err := http.ListenAndServe(":8989", router)
	if err != nil {
		fmt.Println("Failed to start server:", err)
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

func calculateFactorial(num uint64, ch chan uint64) {
	ch <- (num * calculate(num-1))
}

func calculate(num uint64) uint64 {
	var factorial uint64 = 1
	for i := uint64(2); i <= num; i++ {
		factorial *= i
	}
	return factorial
}

func calculateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := r.Context().Value("data")

	convertedMap := data.(map[string]uint64)
	a := convertedMap["a"]
	b := convertedMap["b"]

	aCh := make(chan uint64)
	bCh := make(chan uint64)
	go calculateFactorial(a, aCh)
	go calculateFactorial(b, bCh)

	outputData := responseData{
		AFactorial: <-aCh,
		BFactorial: <-bCh,
	}

	json.NewEncoder(w).Encode(outputData)
}

func calculateMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var reqData requestData
		err := json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil {
			handleError(w, fmt.Errorf("Incorrect input"))
			return
		}

		ctx := context.WithValue(r.Context(), "data", map[string]uint64{
			"a": reqData.A,
			"b": reqData.B,
		})

		next(w, r.WithContext(ctx), p)
	}
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}
