package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GenkiSugiyama/go_todo_app_2/config"
)

func TestNewMux(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)

	ctx := context.Background()
	cf, err := config.New()
	if err != nil {
		t.Error("failed to parse config: %w", err)
	}
	sut, cleanup, err := NewMux(ctx, cf)
	if err != nil {
		t.Error("failed to initialize mux: %w", err)
	}
	defer cleanup()
	sut.ServeHTTP(w, r)
	resp := w.Result()
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Error("want status code 200, but got", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v\n", err)
	}

	want := `{"status":"ok"}`
	fmt.Printf("got: %s\n", got)
	if string(got) != want {
		t.Errorf("want %q, got %q\n", want, got)
	}
}
