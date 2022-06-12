package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/api"
	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/provider"
	"github.com/hikjik/go-musthave-diploma-tpl/internal/app/storage/pg"
	"github.com/hikjik/go-musthave-diploma-tpl/pkg/wpool"
)

func main() {
	cfg := ReadConfig()

	storage, err := pg.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp := wpool.New(cfg.WorkersCount)
	go wp.Run(ctx)
	defer wp.Wait()

	rs := &api.Resources{
		AuthKey:    []byte(cfg.AuthKey),
		Storage:    storage,
		Provider:   provider.New(cfg.Accrual),
		WorkerPool: wp,
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: rs.Routes(),
	}

	idle := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		if err = srv.Shutdown(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("Failed to shutdown HTTP server: %v")
		}
		close(idle)
		cancel()
	}()
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Warn().Err(err).Msg("ListenAndServe failed")
	}
	<-idle
}
