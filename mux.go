package main

import "net/http"

// どんなハンドラー実装をどのパスで受け付けるかのルーティングを定義する関数
func NewMux() http.Handler {
	mux := http.NewServeMux()
	// /healthにアクセスが来たら、JSON形式で{"status":"ok"}を返すハンドラー関数を登録する
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
