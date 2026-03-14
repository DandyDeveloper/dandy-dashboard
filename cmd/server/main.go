package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dandydeveloper/dandy-dashboard/internal/config"
	"github.com/dandydeveloper/dandy-dashboard/internal/store"
	"github.com/dandydeveloper/dandy-dashboard/internal/widget"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/calendar"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/claude"
	"github.com/dandydeveloper/dandy-dashboard/internal/widgets/japanese"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	kv := mustOpenStore(cfg)
	defer kv.Close()

	registry := buildRegistry(cfg, kv)

	e := buildServer(cfg, registry)

	log.Printf("Dandy Dashboard running on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
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

func buildRegistry(cfg *config.Config, kv store.Store) *widget.Registry {
	registry := &widget.Registry{}

	registry.Register(claude.New(cfg.AnthropicAPIKey))

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

func buildServer(cfg *config.Config, registry *widget.Registry) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(cfg.AllowedOrigins, ","),
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, "X-Dashboard-Key"},
	}))

	if cfg.DashboardKey != "" {
		e.Use(apiKeyMiddleware(cfg.DashboardKey))
	}

	registry.Mount(e)

	e.GET("/api/widgets", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"widgets": registry.Slugs()})
	})

	return e
}

func apiKeyMiddleware(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api/") &&
				c.Request().Header.Get("X-Dashboard-Key") != key {
				return echo.ErrUnauthorized
			}
			return next(c)
		}
	}
}
