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

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/azure"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run(ctx context.Context) error {
	settings, configErr := config.LoadAPI()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return errors.New("invalid database configuration")
	}
	defer db.Close()
	store := &workspace.Service{DB: db, Deployment: os.Getenv("CUEIA_DEPLOYMENT_ID"), Mode: settings.Mode}
	sessions, sessionErr := session.NewService(ctx, db, store.Provision)
	if err := errors.Join(configErr, sessionErr); err != nil {
		return err
	}
	if dir := os.Getenv("CUEIA_TEST_VOLUME_ROOT"); dir != "" && settings.Mode == "sample" {
		store.Provider = &providertest.Fixture{Directory: dir}
		store.ProviderName, store.Region = "internal", "local"
	} else {
		machines, err := azure.FromEnvironment()
		if err != nil {
			return errors.New("invalid Azure workspace provider configuration")
		}
		store.Provider, store.ProviderName, store.Region = machines, "azure", os.Getenv("AZURE_LOCATION")
	}
	mux := routes(sessions.Authenticate, nil)
	workspaceRoutes(mux, sessions, store)
	signer, err := document.Load(ctx, db)
	if err != nil {
		return errors.New("cannot load deployment signing key")
	}
	if offset := os.Getenv("CUEIA_TEST_PREVIEW_CLOCK_OFFSET"); store.ProviderName == "internal" && offset != "" {
		advance, err := time.ParseDuration(offset)
		if err != nil {
			return errors.New("invalid test preview clock offset")
		}
		signer.Now = func() time.Time { return time.Now().Add(advance) }
	}
	previewRoutes(mux, sessions, store, signer)
	server := &http.Server{Addr: settings.Addr, Handler: sessions.Protect(mux, middleware.NewLimiter()), ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}
