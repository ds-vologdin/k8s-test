package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	maxShutdownTime = 10 * time.Second
)

var (
	port = flag.Uint("port", 8000, "server port")
)

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"OK\"}"))
}

func root(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ROOT HANDLER] request: %v %v", r.Method, r.URL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("root"))
}

func newServer(address string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz/", health)
	mux.HandleFunc("/", root)
	mux.Handle("/metrics", promhttp.Handler())
	return &http.Server{Addr: address, Handler: mux}
}

func main() {
	flag.Parse()
	address := fmt.Sprintf(":%d", *port)
	log.Printf("Start HTTP server %s", address)

	srv := newServer(address)

	shutdown := make(chan struct{})
	defer close(shutdown)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		sig := <-stop
		log.Printf("Got signal '%v', the graceful shutdown will start", sig)

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, maxShutdownTime)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		} else {
			log.Print("HTTP server has been shutdown")
		}
		shutdown <- struct{}{}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Print("Wait for the shutdown server")
	<-shutdown
}
