package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestRun(t *testing.T) {
	t.Skip("リファクタリング中のため、テストをスキップします")

	// localhost:0でリクエストをリッスンするListenerを作成してrun()関数に渡す
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen port %v\n", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return run(ctx)
	})
	in := "message"
	url := fmt.Sprintf("http://%s/%s", l.Addr().String(), in)
	t.Logf("try to request to %q\n", url)
	rsp, err := http.Get(url)
	if err != nil {
		t.Errorf("failed to get: %+v\n", err)
	}
	defer rsp.Body.Close()
	got, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	want := fmt.Sprintf("Hello, %s!", in)
	if string(got) != want {
		t.Errorf("want %q, got %q\n", want, got)
	}

	cancel()

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
