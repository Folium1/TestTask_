package internal

import (
	// "bytes"
	// "context"
	// "encoding/json"
	// "fmt"
	// "net/http"
	// "net/http/httptest"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	// "github.com/julienschmidt/httprouter"
)

func TestCalculateFactorial(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    uint64
		expected uint64
	}{
		{1, 1},
		{2, 2},
		{3, 6},
		{4, 24},
		{5, 120},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("calculateFactorial(%d) = %d", c.input, c.expected), func(t *testing.T) {
			ch := make(chan uint64)
			go сalculateFactorial(c.input, ch)
			result := <-ch

			if result != c.expected {
				t.Errorf("expected factorial of %d to be %d, got %d", c.input, c.expected, result)
			}
		})
	}
}

func TestCalculateMiddleware(t *testing.T) {
	t.Parallel()

	next := httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Simulate the next handler in the chain
		data := r.Context().Value("data")
		convertedMap := data.(map[string]uint64)
		a := convertedMap["a"]
		b := convertedMap["b"]

		outputData := responseData{
			AFactorial: a,
			BFactorial: b,
		}

		json.NewEncoder(w).Encode(outputData)
	})

	cases := []struct {
		input    string
		expected map[string]uint64
	}{
		{`{"a":1,"b":2}`, map[string]uint64{"a": 1, "b": 2}},
		{`{"a":3,"b":4}`, map[string]uint64{"a": 3, "b": 4}},
		{`{"a":-1},"b":-5`, map[string]uint64{"error": 0}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("calculateMiddleware(%s)", c.input), func(t *testing.T) {
			buff := bytes.NewBufferString(c.input)

			req, err := http.NewRequest("GET", "/", buff)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), "data", c.expected)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			p := httprouter.Params{}
			factorialMiddleware(next)(w, req, p)

			var result responseData
			err = json.NewDecoder(w.Body).Decode(&result)
			if err != nil {
				t.Error(err)
				return
			}

			if result.AFactorial != c.expected["a"] || result.BFactorial != c.expected["b"] {
				t.Errorf("expected data to be %+v, got %+v", c.expected, result)
			}
		})
	}
}
