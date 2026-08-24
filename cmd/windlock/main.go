package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry-110/internal/api"
)

func main() {
	address := flag.String("listen", "127.0.0.1:21210", "HTTP listen address")
	dataDirectory := flag.String("data", "./data", "journal and snapshot directory")
	flag.Parse()

	system, err := api.NewSystem(*dataDirectory, time.Now().UTC())
	if err != nil {
		slog.Error("initialize WindLock", "error", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr: *address, Handler: api.NewServer(system).Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}

	interrupt, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-interrupt.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			slog.Error("shutdown WindLock", "error", err)
		}
	}()

	slog.Info("WindLock listening", "address", *address)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("serve WindLock", "error", err)
		os.Exit(1)
	}
	if interrupt.Err() != nil && !errors.Is(interrupt.Err(), context.Canceled) {
		fmt.Fprintln(os.Stderr, interrupt.Err())
	}
}
