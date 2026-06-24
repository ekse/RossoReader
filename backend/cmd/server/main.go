package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ekse/rssreader/internal/bootstrap"
	"github.com/ekse/rssreader/internal/db"
	"github.com/ekse/rssreader/internal/fetcher"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/middleware"
	"github.com/ekse/rssreader/internal/scheduler"
	"github.com/ekse/rssreader/internal/store/pgstore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/rssreader?sslmode=disable"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("connecting to database...")
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	log.Printf("running migrations...")
	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	store := pgstore.New(pool)

	log.Printf("running bootstrap...")
	if err := bootstrap.Run(ctx, pool, store); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	f := fetcher.NewHTTPFetcher()
	sched := scheduler.New(store, f)

	h := handlers.New(store, sched, f)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS)
	r.Use(chimw.Recoverer)

	r.Mount("/", h.MountRouter())

	log.Printf("initial feed fetch...")
	if err := sched.FetchAll(ctx); err != nil {
		log.Printf("initial feed fetch error: %v", err)
	}

	sched.Start()

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on :%s", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down...")
	sched.Stop()
	server.Shutdown(context.Background())
	log.Printf("shutdown complete")
}
