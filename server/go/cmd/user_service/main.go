package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xlunch/user-service/internal/db"
	"github.com/0xlunch/user-service/internal/routes"
	"github.com/go-chi/chi/v5"
)

// Postgres connection string
const connectionString string = "user=pdao password=parallel host=localhost port=5432 dbname=mydb sslmode=verify-ca pool_max_conns=10"

// main run
func run(ctx context.Context, w io.Writer, args []string) error {
	iCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// create gin router
	r := chi.NewRouter()

	// new database pool
	db, err := db.NewDB(connectionString)
	if err != nil {
		return err
	}
	defer db.Pool.Close()

	// setup routes
	routes.SetupRoutes(r, db)

	// http wrap for graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func(iCtx context.Context) {
		<-iCtx.Done()

		fmt.Fprintln(w, "Shutting down service...")

		ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdown()

		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(w, "%s\n", err)
			os.Exit(1)
		}
	}(iCtx)

	return srv.ListenAndServe()
}

// main
func main() {
	// handle run errors
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
