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
	A int `json:"a"`
	B int `json:"b"`
}

func (r requestData) isValid() bool {
	if r.A < 0 || r.B < 0 {
		return false
	}
	return true
}

type responseData struct {
	AFactorial int `json:"a!"`
	BFactorial int `json:"b!"`
}

func calculateFactorial(num int, ch chan int) {
	ch <- (num * calculate(num-1))
}

func calculate(num int) int {
	factorial := 1
	for i := 2; i <= num; i++ {
		factorial *= i
	}
	return factorial
}

func calculateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ctx := r.Context()
	data := ctx.Value("data")

	convertedMap := data.(map[string]int)
	a := convertedMap["a"]
	b := convertedMap["b"]

	aCh := make(chan int)
	bCh := make(chan int)
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

		if !reqData.isValid() {
			handleError(w, fmt.Errorf("Incorrect input"))
			return
		}
		ctx := context.WithValue(r.Context(), "data", map[string]int{
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
