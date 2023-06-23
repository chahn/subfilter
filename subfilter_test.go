package subfilter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubFilter(t *testing.T) {
	// Define a simple handler that our middleware will wrap
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Foobar", "Hello, apple!")
		w.Write([]byte("Hello, world!"))
	})

	// Create a config for our middleware
	config := CreateConfig()
	config.Replacements = []ReplacementConfig{
		{Pattern: "hello", Replacement: "Hi", Flags: "i"},
		{Pattern: "apple(.)", Replacement: "banana?"},
	}

	// Create a handler that allows the next handler to be called only once
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})

	// Create an instance of our middleware
	middleware, err := New(context.TODO(), handler, config, "test")
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Use httptest to record the response
	rec := httptest.NewRecorder()

	// Call our middleware
	middleware.ServeHTTP(rec, req)

	// Check the response body
	expectedBody := "Hi, world!"
	if body := rec.Body.String(); body != expectedBody {
		t.Errorf("Unexpected response body: got %q want %q", body, expectedBody)
	}

	// Check the X-Foobar header
	expectedHeader := "Hi, banana?"
	if header := rec.Header().Get("X-Foobar"); header != expectedHeader {
		t.Errorf("Unexpected X-Foobar header: got %q want %q", header, expectedHeader)
	}
}
