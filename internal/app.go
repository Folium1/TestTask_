package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type requestData struct {
	A uint64 `json:"a"`
	B uint64 `json:"b"`
}

type responseData struct {
	AFactorial uint64 `json:"a"`
	BFactorial uint64 `json:"b"`
}

// factorialHandler concurrently calculates the factorials of two numbers and returns them in JSON format.
func FactorialHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := r.Context().Value("data").(map[string]uint64)
	a := data["a"]
	b := data["b"]

	aCh := make(chan uint64)
	bCh := make(chan uint64)
	go сalculateFactorial(a, aCh)
	go сalculateFactorial(b, bCh)

	response := responseData{
		AFactorial: <-aCh,
		BFactorial: <-bCh,
	}

	json.NewEncoder(w).Encode(response)
}

// FactorialMiddleware decoding json request,validates
// then calls next with values in the context
func FactorialMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var requestData requestData

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			handleError(w, fmt.Errorf("Incorrect input"))
			return
		}

		ctx := createContext(r, requestData)
		next(w, r.WithContext(ctx), p)
	}
}

func createContext(r *http.Request, requestData requestData) context.Context {
	data := map[string]uint64{
		"a": requestData.A,
		"b": requestData.B,
	}

	return context.WithValue(r.Context(), "data", data)
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
