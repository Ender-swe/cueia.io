package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
)

func run(ctx context.Context) error {
	conn, err := database.Open(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	dir := config.MigrationsDir("migrations")
	return database.Migrate(ctx, conn, os.DirFS(dir))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}
