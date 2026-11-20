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

	"github.com/Ender-swe/cueia.io/backend/internal/config"
	"github.com/Ender-swe/cueia.io/backend/internal/document"
	"github.com/Ender-swe/cueia.io/backend/internal/middleware"
	"github.com/Ender-swe/cueia.io/backend/internal/session"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/provider/azure"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run(ctx context.Context) error {
	settings, err := config.LoadAPI()
	if err != nil {
		return err
	}
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return errors.New("invalid database configuration")
	}
	defer db.Close()
	machines, err := azure.FromEnvironment()
	if err != nil {
		return errors.New("invalid Azure workspace provider configuration")
	}
	store := &workspace.Service{DB: db, Provider: machines, Deployment: os.Getenv("CUEIA_DEPLOYMENT_ID"), ProviderName: "azure", Region: os.Getenv("AZURE_LOCATION"), Mode: settings.Mode}
	sessions := session.Service{DB: db, Verify: session.Clerk(ctx, settings.ClerkIssuer, settings.AuthorizedParty), Provision: store.Provision}
	mux := routes(sessions.Authenticate, nil)
	workspaceRoutes(mux, sessions, store)
	signer, err := document.Load(ctx, db)
	if err != nil {
		return errors.New("cannot load deployment signing key")
	}
	previewRoutes(mux, sessions, store, signer)
	server := &http.Server{Addr: settings.Addr, Handler: middleware.Protect(mux, settings.AuthorizedParty, middleware.NewLimiter()), ReadHeaderTimeout: 5 * time.Second}
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
