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

	"confvision-infrastructure-monitor/internal/api"
	"confvision-infrastructure-monitor/internal/collect"
	"confvision-infrastructure-monitor/internal/config"
)

func main() {
	log.SetFlags(0)
	cfg, err := config.Load()
	if err != nil {
		log.Printf("nivel=erro evento=config detalhe=%q", err.Error())
		os.Exit(1)
	}
	if cfg.Secret == "" {
		log.Printf("nivel=erro evento=segredo_ausente")
		os.Exit(1)
	}
	eng := collect.NewEngine(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go eng.Run(ctx)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           api.New(cfg, eng).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shut); err != nil {
			log.Printf("nivel=erro evento=shutdown")
		}
	}()

	log.Printf("nivel=info evento=escuta tls=%t", cfg.TLSEnabled())
	if cfg.TLSEnabled() {
		err = srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("nivel=erro evento=http")
		os.Exit(1)
	}
}
