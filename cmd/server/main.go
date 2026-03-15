package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/dandydeveloper/dandy-dashboard/internal/config"
	"github.com/dandydeveloper/dandy-dashboard/internal/httputil"
	"github.com/dandydeveloper/dandy-dashboard/internal/middleware"
	"github.com/dandydeveloper/dandy-dashboard/internal/store"
	"github.com/dandydeveloper/dandy-dashboard/internal/widget"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/calendar"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/claude"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/japanese"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	kv := mustOpenStore(cfg)
	defer kv.Close()

	registry := buildRegistry(cfg, kv, logger)
	srv := buildServer(cfg, registry, logger)

	logger.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func mustOpenStore(cfg *config.Config) store.Store {
	if cfg.StoreURL == "" {
		if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
			log.Fatalf("creating data dir: %v", err)
		}
	}
	kv, err := store.New(cfg.StoreURL, cfg.DataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	return kv
}

func buildRegistry(cfg *config.Config, kv store.Store, logger *slog.Logger) *widget.Registry {
	registry := &widget.Registry{}

	registry.Register(claude.New(cfg.AnthropicAPIKey, logger))

	japaneseWidget, err := japanese.New(kv, cfg.WaniKaniToken)
	if err != nil {
		log.Fatalf("japanese widget: %v", err)
	}
	registry.Register(japaneseWidget)

	calendarWidget, err := calendar.New(cfg.GoogleCredentialsJSON, cfg.GoogleCalendarID)
	if err != nil {
		log.Fatalf("calendar widget: %v", err)
	}
	registry.Register(calendarWidget)

	return registry
}

func buildServer(cfg *config.Config, registry *widget.Registry, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()

	registry.Mount(mux)

	mux.HandleFunc("GET /api/widgets", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"widgets": registry.Slugs()})
	})

	handler := middleware.Chain(mux,
		middleware.Recover(logger),
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.CORS(strings.Split(cfg.AllowedOrigins, ",")),
		middleware.APIKey(cfg.DashboardKey),
	)

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}
}
