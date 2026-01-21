package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	common "github.com/peteraglen/slack-manager-common"
)

func TestNew(t *testing.T) {
	t.Parallel()

	client := New("http://example.com", WithRetryCount(5))

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.baseURL != "http://example.com" {
		t.Errorf("expected baseURL=http://example.com, got %s", client.baseURL)
	}

	if client.options.retryCount != 5 {
		t.Errorf("expected retryCount=5, got %d", client.options.retryCount)
	}
}

func TestConnect_EmptyURL(t *testing.T) {
	t.Parallel()

	client := New("")

	err := client.Connect(context.Background())

	if err == nil {
		t.Fatal("expected error for empty URL")
	}

	if err.Error() != "base URL must be set" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConnect_InvalidOptions(t *testing.T) {
	t.Parallel()

	client := New("http://example.com")
	// Force invalid options by setting nil logger
	client.options.requestLogger = nil

	err := client.Connect(context.Background())

	if err == nil {
		t.Fatal("expected error for invalid options")
	}

	if !strings.Contains(err.Error(), "invalid options") {
		t.Errorf("expected error to contain 'invalid options', got: %v", err)
	}
}

func TestConnect_PingFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := New(server.URL)

	err := client.Connect(context.Background())

	if err == nil {
		t.Fatal("expected error for ping failure")
	}

	if !strings.Contains(err.Error(), "failed to ping alerts API") {
		t.Errorf("expected error to contain 'failed to ping alerts API', got: %v", err)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain '500', got: %v", err)
	}
}

func TestConnect_Success(t *testing.T) {
	t.Parallel()

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)

	err := client.Connect(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if requestedPath != "/ping" {
		t.Errorf("expected path=/ping, got %s", requestedPath)
	}
}

func TestConnect_OnlyOnce(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)

	// First connect
	err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("first connect failed: %v", err)
	}

	// Second connect should be no-op
	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("second connect failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected ping to be called once, got %d", callCount)
	}
}

func TestConnect_SetsHeaders(t *testing.T) {
	t.Parallel()

	var contentType, accept, customHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		accept = r.Header.Get("Accept")
		customHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, WithRequestHeader("X-Custom", "custom-value"))

	err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	if contentType != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %s", contentType)
	}

	if accept != "application/json" {
		t.Errorf("expected Accept=application/json, got %s", accept)
	}

	if customHeader != "custom-value" {
		t.Errorf("expected X-Custom=custom-value, got %s", customHeader)
	}
}

func TestConnect_SetsBasicAuth(t *testing.T) {
	t.Parallel()

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, WithBasicAuth("user", "pass"))

	err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("expected Basic auth header, got %s", authHeader)
	}
}

func TestConnect_SetsTokenAuth(t *testing.T) {
	t.Parallel()

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, WithAuthScheme("Bearer"), WithAuthToken("my-token"))

	err := client.Connect(context.Background())
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	if authHeader != "Bearer my-token" {
		t.Errorf("expected 'Bearer my-token', got %s", authHeader)
	}
}

func TestSend_NilClient(t *testing.T) {
	t.Parallel()

	var client *Client

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for nil client")
	}

	if err.Error() != "alert client is nil" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_NotConnected(t *testing.T) {
	t.Parallel()

	client := New("http://example.com")

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for not connected client")
	}

	if err.Error() != "client not connected - call Connect() first" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_EmptyAlerts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)
	_ = client.Connect(context.Background())

	err := client.Send(context.Background())

	if err == nil {
		t.Fatal("expected error for empty alerts")
	}

	if err.Error() != "alerts list cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_NilAlert(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)
	_ = client.Connect(context.Background())

	err := client.Send(context.Background(), &common.Alert{}, nil, &common.Alert{})

	if err == nil {
		t.Fatal("expected error for nil alert")
	}

	if err.Error() != "alert at index 1 is nil" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_Success(t *testing.T) {
	t.Parallel()

	var capturedPath string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedBody = make([]byte, r.ContentLength)
		_, _ = r.Body.Read(capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)
	_ = client.Connect(context.Background())

	alert := &common.Alert{
		Header: "Test Alert",
	}
	err := client.Send(context.Background(), alert)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if capturedPath != "/alerts" {
		t.Errorf("expected path=/alerts, got %s", capturedPath)
	}

	if !strings.Contains(string(capturedBody), "Test Alert") {
		t.Errorf("expected body to contain 'Test Alert', got: %s", capturedBody)
	}
}

func TestSend_HTTPError_JSONErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "validation failed: header is required"}`))
	}))
	defer server.Close()

	client := New(server.URL, WithRetryCount(0))
	_ = client.Connect(context.Background())

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for HTTP error")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain '400', got: %v", err)
	}

	// Should extract the error message from JSON
	if !strings.Contains(err.Error(), "validation failed: header is required") {
		t.Errorf("expected error to contain 'validation failed: header is required', got: %v", err)
	}
}

func TestSend_HTTPError_PlainTextResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := New(server.URL, WithRetryCount(0))
	_ = client.Connect(context.Background())

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for HTTP error")
	}

	// Should fall back to raw body for non-JSON response
	if !strings.Contains(err.Error(), "Bad Request") {
		t.Errorf("expected error to contain 'Bad Request', got: %v", err)
	}
}

func TestSend_HTTPError_JSONWithoutErrorField(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "something went wrong"}`))
	}))
	defer server.Close()

	client := New(server.URL, WithRetryCount(0))
	_ = client.Connect(context.Background())

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for HTTP error")
	}

	// Should fall back to raw body when JSON doesn't have "error" field
	if !strings.Contains(err.Error(), `{"message": "something went wrong"}`) {
		t.Errorf("expected error to contain raw JSON body, got: %v", err)
	}
}

func TestSend_HTTPError_EmptyResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(server.URL, WithRetryCount(0))
	_ = client.Connect(context.Background())

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for HTTP error")
	}

	if !strings.Contains(err.Error(), "(empty error body)") {
		t.Errorf("expected error to contain '(empty error body)', got: %v", err)
	}
}

func TestSend_RequestError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	client := New(server.URL, WithRetryCount(0))
	_ = client.Connect(context.Background())

	// Close server to cause connection error on Send
	server.Close()

	err := client.Send(context.Background(), &common.Alert{})

	if err == nil {
		t.Fatal("expected error for request failure")
	}

	if !strings.Contains(err.Error(), "POST") {
		t.Errorf("expected error to mention POST, got: %v", err)
	}
}

func TestConnect_RequestError(t *testing.T) {
	t.Parallel()

	// Use a URL that will fail to connect
	client := New("http://localhost:1", WithRetryCount(0))

	err := client.Connect(context.Background())

	if err == nil {
		t.Fatal("expected error for connection failure")
	}

	if !strings.Contains(err.Error(), "failed to ping alerts API") {
		t.Errorf("expected error to contain 'failed to ping alerts API', got: %v", err)
	}
}

func TestSend_MultipleAlerts(t *testing.T) {
	t.Parallel()

	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alerts" {
			capturedBody = make([]byte, r.ContentLength)
			_, _ = r.Body.Read(capturedBody)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)
	_ = client.Connect(context.Background())

	alerts := []*common.Alert{
		{Header: "Alert 1"},
		{Header: "Alert 2"},
		{Header: "Alert 3"},
	}
	err := client.Send(context.Background(), alerts...)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	bodyStr := string(capturedBody)
	if !strings.Contains(bodyStr, "Alert 1") ||
		!strings.Contains(bodyStr, "Alert 2") ||
		!strings.Contains(bodyStr, "Alert 3") {
		t.Errorf("expected body to contain all alerts, got: %s", bodyStr)
	}
}

func TestSend_JSONFormat(t *testing.T) {
	t.Parallel()

	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alerts" {
			capturedBody = make([]byte, r.ContentLength)
			_, _ = r.Body.Read(capturedBody)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL)
	_ = client.Connect(context.Background())

	alert := &common.Alert{
		Header: "Test Header",
		Text:   "Test Text",
	}
	err := client.Send(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the JSON structure
	var result struct {
		Alerts []struct {
			Header string `json:"header"`
			Text   string `json:"text"`
		} `json:"alerts"`
	}
	if err := json.Unmarshal(capturedBody, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(result.Alerts))
	}

	if result.Alerts[0].Header != "Test Header" {
		t.Errorf("expected header='Test Header', got %s", result.Alerts[0].Header)
	}

	if result.Alerts[0].Text != "Test Text" {
		t.Errorf("expected text='Test Text', got %s", result.Alerts[0].Text)
	}
}
