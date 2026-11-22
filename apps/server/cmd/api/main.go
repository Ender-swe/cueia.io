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

	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/config"
	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/routes"
	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/util"
	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/azure"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
)

func run(ctx context.Context) error {
	// Load deployment settings and keep the database open for the server's lifetime.
	settings, configErr := config.Load()
	sessionConfig, sessionErr := session.LoadConfig()
	if err := errors.Join(configErr, sessionErr); err != nil {
		return err
	}

	db, err := util.OpenDatabase(ctx, settings.DatabaseURL)
	if err != nil {
		return errors.New("invalid database configuration")
	}
	defer db.Close()

	// Give the session service a callback to provision workspaces.
	store := &workspace.Service{DB: db, Deployment: settings.Deployment, Mode: settings.Mode}
	sessions := sessionConfig.NewService(ctx, db, store.Provision)

	// Sample mode can use a local fixture directory. Other configurations use Azure.
	if dir := settings.TestVolumeRoot; dir != "" && settings.Mode == "sample" {
		store.Provider = &providertest.Fixture{Directory: dir}
		store.ProviderName, store.Region = "internal", "local"
	} else {
		machines, err := azure.FromEnvironment()
		if err != nil {
			return errors.New("invalid Azure workspace provider configuration")
		}
		store.Provider, store.ProviderName, store.Region = machines, "azure", settings.AzureLocation
	}

	// Load the deployment's signing key for document previews.
	signer, err := document.Load(ctx, db)
	if err != nil {
		return errors.New("cannot load deployment signing key")
	}
	// Local fixture tests can shift the signer's clock to exercise preview expiration.
	if offset := settings.TestPreviewClockOffset; store.ProviderName == "internal" && offset != "" {
		advance, err := time.ParseDuration(offset)
		if err != nil {
			return errors.New("invalid test preview clock offset")
		}
		signer.Now = func() time.Time { return time.Now().Add(advance) }
	}

	server := &http.Server{Addr: settings.Addr, Handler: routes.New(sessions, store, signer), ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)

	// Serve in a goroutine so run can wait for either a server error or cancellation.
	go func() { done <- server.ListenAndServe() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Use a fresh context to give active requests up to five seconds to finish.
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
		// ListenAndServe returns ErrServerClosed during a normal shutdown.
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
