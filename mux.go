package main

import (
	"context"
	"net/http"

	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/config"
	"github.com/GenkiSugiyama/go_todo_app_2/handler"
	"github.com/GenkiSugiyama/go_todo_app_2/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// どんなハンドラー実装をどのパスで受け付けるかのルーティングを定義する関数
func NewMux(ctx context.Context, cfg *config.Config) (http.Handler, func(), error) {
	mux := chi.NewRouter()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	v := validator.New()
	db, cleanup, err := store.New(ctx, cfg)
	if err != nil {
		return nil, cleanup, err
	}
	r := &store.Repository{Clocker: clock.RealClocker{}}
	at := &handler.AddTask{DB: db, Repo: r, Validator: v}
	mux.Post("/tasks", at.ServeHTTP)
	lt := &handler.ListTask{DB: db, Repo: r}
	mux.Get("/tasks", lt.ServeHTTP)
	return mux, cleanup, nil
}
