package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"library/internal/db"
	"library/internal/handlers"
	"library/internal/store"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://library:library@localhost:5440/library?sslmode=disable")
	addr := getenv("ADDR", ":8080")

	database, err := db.Open(dsn)
	if err != nil {
		log.Fatalf("connect to postgres: %v\n(is it running? try: docker compose up -d)", err)
	}
	defer database.Close()

	authors := store.NewAuthorStore(database)
	api := &handlers.API{
		DB:      database,
		Authors: authors,
		Books:   store.NewBookStore(database, authors),
		Members: store.NewMemberStore(database),
		Loans:   store.NewLoanStore(database),
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		log.Printf("listening on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
