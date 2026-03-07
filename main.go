package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GenkiSugiyama/go_todo_app_2/config"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
	l   net.Listener
}

func NewServer(l net.Listener, mux http.Handler) *Server {
	return &Server{
		srv: &http.Server{Handler: mux},
		l:   l,
	}
}

func (s *Server) Run(ctx context.Context) error {
	// 1. graceful shutdownのためにシグナルを受け取るContextを作成する
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// *errgroup.Group型を使ってエラーを含むゴルーチン間の並行処理を実装する
	eg, ctx := errgroup.WithContext(ctx)
	// 2. eg.Go()を使ってfunc() errorの処理を別ゴルーチンで実行する（サーバー起動）
	eg.Go(func() error {
		// 5. s.srv.Shutdown()が呼び出されると、s.srv.Serve()が終了する
		if err := s.srv.Serve(s.l); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("failed to close: %+v\n", err)
			return err
		}
		return nil
	})

	// 3. キャンセル通知が発生するまで処理を待機する
	<-ctx.Done()
	// 4. ctx経由でキャンセルが通知されたら、サーバーをシャットダウンする
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v\n", err)
	}

	// 6. eg.Go()メソッドが終了するのを待機して、eg.Go()内のメソッドでエラーがあれば返す
	return eg.Wait()
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to run: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := config.New()
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v\n", cfg.Port, err)
	}
	url := fmt.Sprintf("http://%s", l.Addr().String())
	log.Printf("start with: %v\n", url)
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}
	eg, ctx := errgroup.WithContext(ctx)

	// s.ListenAndServe()が終了すると、eg.Go()の処理も終了する
	eg.Go(func() error {
		// s.Shutdown()が呼び出されると、s.ListenAndServe()が終了する
		if err := s.Serve(l); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("failed to close server: %+v\n", err)
			return err
		}
		return nil
	})

	// Contextによるキャンセル通知が発生するまで処理を待機する
	<-ctx.Done()
	// Contextによるキャンセル通知が発生したら、サーバーをシャットダウンする
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown server: %+v\n", err)
	}

	// eg.Go()メソッドが終了するのを待機して、エラーがあれば返す
	return eg.Wait()
}
