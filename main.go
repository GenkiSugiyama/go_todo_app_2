package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	if len(os.Args) != 2 {
		log.Printf("need port number\n")
		os.Exit(1)
	}
	p := os.Args[1]
	l, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Fatalf("failed to listen port %s: %v\n", p, err)
	}
	if err := run(context.Background(), l); err != nil {
		log.Printf("failed to terminate server: %v\n", err)
	}
}

// 引数にListenerを受け取るようにする
func run(ctx context.Context, l net.Listener) error {
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
