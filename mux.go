package main

import (
	"context"
	"net/http"

	"github.com/GenkiSugiyama/go_todo_app_2/auth"
	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/config"
	"github.com/GenkiSugiyama/go_todo_app_2/handler"
	"github.com/GenkiSugiyama/go_todo_app_2/service"
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
	// ハンドラの初期化に必要なvalidatorやdb, repositoryの初期化
	v := validator.New()
	db, cleanup, err := store.New(ctx, cfg)
	if err != nil {
		return nil, cleanup, err
	}
	clocker := clock.RealClocker{}
	r := &store.Repository{Clocker: clock.RealClocker{}}

	rcli, err := store.NewKVS(ctx, cfg)
	if err != nil {
		return nil, cleanup, err
	}
	jwter, err := auth.NewJWTer(rcli, clocker)
	if err != nil {
		return nil, cleanup, err
	}
	l := &handler.Login{
		Service: &service.Login{
			DB:             db,
			Repo:           r,
			TokenGenerator: jwter,
		},
		Validator: v,
	}
	mux.Post("/login", l.ServeHTTP)

	// ハンドラを初期化し、パスとハンドラの処理を紐付けしている
	at := &handler.AddTask{
		Service:   &service.AddTask{DB: db, Repo: r},
		Validator: v,
	}
	lt := &handler.ListTask{
		Service: &service.ListTask{DB: db, Repo: r},
	}
	mux.Route("/tasks", func(r chi.Router) {
		r.Use(handler.AuthMiddleware(jwter))
		r.Post("/", at.ServeHTTP)
		r.Get("/", lt.ServeHTTP)
	})
	ru := &handler.RegisterUser{
		Service:   &service.RegisterUser{DB: db, Repo: r},
		Validator: v,
	}
	mux.Post("/register", ru.ServeHTTP)
	return mux, cleanup, nil
}
