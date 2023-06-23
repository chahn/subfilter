package subfilter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Replacements []ReplacementConfig `json:"replacements,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type ReplacementConfig struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Flags       string `json:"flags,omitempty"`
}

type Replacement struct {
	Pattern     *regexp.Regexp
	Replacement string
}

type SubFilter struct {
	next         http.Handler
	name         string
	replacements []Replacement
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	var replacements []Replacement
	for _, replacementConfig := range config.Replacements {
		pattern := replacementConfig.Pattern

		if replacementConfig.Flags == "i" {
			// If "i" flag is set, add case insensitivity to the pattern
			pattern = "(?i)" + pattern
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		replacements = append(replacements, Replacement{Pattern: re, Replacement: replacementConfig.Replacement})
	}
	return &SubFilter{
		next:         next,
		name:         name,
		replacements: replacements,
	}, nil
}

type responseWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (r *SubFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Modify the request to prevent it from accepting compressed responses
	req.Header.Del("Accept-Encoding")

	// Wrap the ResponseWriter
	w := &responseWriter{
		ResponseWriter: rw,
		body:           &bytes.Buffer{},
		status:         http.StatusOK, // Default to 200 OK
	}

	r.next.ServeHTTP(w, req)

	// Declare body here
	body := w.body.String()

	// Check Content-Type of the response
	contentType := w.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "application/xml") {
		for _, rep := range r.replacements {
			newBody := rep.Pattern.ReplaceAllString(body, rep.Replacement)
			if newBody != body {
				fmt.Printf("Body replacement occurred: %s => %s\n", rep.Pattern.String(), rep.Replacement)
			}
			body = newBody // continue with the updated body for the next iteration
		}
	}

	// Update the "Content-Length" header to reflect the length of the modified body
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))

	// Replace headers with modified values
	for key, values := range rw.Header() {
		newValues := make([]string, len(values))
		for i, value := range values {
			temp := value
			for _, rep := range r.replacements {
				newValue := rep.Pattern.ReplaceAllString(temp, rep.Replacement)
				if newValue != temp {
					fmt.Printf("Header replacement occurred for header '%s': %s => %s\n", key, rep.Pattern.String(), rep.Replacement)
				}
				temp = newValue // continue with the updated value for the next iteration
			}
			newValues[i] = temp
		}

		rw.Header().Del(key) // Remove the old values
		for _, newValue := range newValues {
			rw.Header().Add(key, newValue) // Add the new values
		}
	}

	rw.WriteHeader(w.status)
	rw.Write([]byte(body))

}
