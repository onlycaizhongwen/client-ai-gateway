package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"client-ai-gateway/internal/access"
	"client-ai-gateway/internal/audit"
	"client-ai-gateway/internal/config"
	gatewayruntime "client-ai-gateway/internal/runtime"
	"client-ai-gateway/internal/trace"
)

func main() {
	configPath := flag.String("config", "configs/dev.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store, err := trace.NewJSONLStoreWithRetention(cfg.TraceStorePath, cfg.TraceRetentionMax)
	if err != nil {
		log.Fatalf("open trace store: %v", err)
	}
	manager, err := gatewayruntime.NewManager(*configPath, store)
	if err != nil {
		log.Fatalf("build runtime: %v", err)
	}
	defer manager.Close()

	auditStore, err := audit.NewJSONLStoreWithRetention(cfg.AuditStorePath, cfg.AuditRetentionMax)
	if err != nil {
		log.Fatalf("open audit store: %v", err)
	}
	handler := access.NewRuntimeHandler(manager, store).WithAudit(auditStore)
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("client AI gateway listening on http://%s", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
