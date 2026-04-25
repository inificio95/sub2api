package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort    = 8080
	defaultHost    = "0.0.0.0"
	appVersion     = "1.0.0"
)

// Config holds the application configuration
type Config struct {
	Host    string
	Port    int
	Debug   bool
	Token   string
}

func main() {
	cfg := parseConfig()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	setupRoutes(r, cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("sub2api v%s starting on %s", appVersion, addr)

	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}

// parseConfig reads configuration from flags and environment variables.
// Environment variables take precedence over defaults; flags override env vars.
func parseConfig() *Config {
	cfg := &Config{}

	// Determine defaults from environment
	envPort := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			envPort = parsed
		}
	}
	envHost := defaultHost
	if h := os.Getenv("HOST"); h != "" {
		envHost = h
	}
	envToken := os.Getenv("API_TOKEN")

	flag.StringVar(&cfg.Host, "host", envHost, "host address to listen on")
	flag.IntVar(&cfg.Port, "port", envPort, "port to listen on")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable debug mode")
	flag.StringVar(&cfg.Token, "token", envToken, "optional API token for authentication")
	flag.Parse()

	return cfg
}

// setupRoutes registers all HTTP routes on the provided router.
func setupRoutes(r *gin.Engine, cfg *Config) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": appVersion,
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	if cfg.Token != "" {
		v1.Use(tokenAuthMiddleware(cfg.Token))
	}

	// Subscription conversion endpoint
	v1.GET("/sub", handleSubscription)
}

// tokenAuthMiddleware validates the Bearer token in the Authorization header.
func tokenAuthMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}
		// Strip "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token != expectedToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// handleSubscription is a placeholder for the subscription conversion handler.
// The actual conversion logic will be implemented in the handler package.
func handleSubscription(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}
	// TODO: delegate to subscription service
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "url": url})
}
