package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"github.com/tinkershack/meteomunch/config"
	e "github.com/tinkershack/meteomunch/errors"
	"github.com/tinkershack/meteomunch/logger"
	"github.com/tinkershack/meteomunch/plumber"
	"github.com/tinkershack/meteomunch/providers"
)

func Serve(ctx context.Context, args []string) {
	logger := logger.NewTag("server")

	// The following are transient test routes for an early stage development convenience.
	// These will be cleanedup in the future.

	cfg, err := config.Get()
	logger.Debug("Config fetched", "config", cfg, "err", err)
	if err != nil {
		logger.Error(e.FATAL, "err", err, "description", "config not well formed")
		os.Exit(1)
	}
	logger.Debug("Config parsed successfully", "config", cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
		logger.Debug(r.URL.String())
	})

	mux.HandleFunc("GET /open-meteo", func(w http.ResponseWriter, r *http.Request) {
		// p, err := providers.NewOpenMeteoProvider(cfg)
		p, err := providers.New("open-meteo", cfg)
		if err != nil {
			logger.Error(e.FAIL, "err", err, "description", "Couldn't create OpenMeteoProvider")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bd, err := p.FetchData(plumber.NewCoordinates(10.9018379, 76.8998445))
		if err != nil {
			logger.Error(e.FAIL, "err", err, "description", "Couldn't fetch data from OpenMeteoProvider")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(bd); err != nil {
			logger.Error(e.FAIL, "err", err, "description", "Couldn't encode response data to JSON")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Debug("API Data fetched", "data", bd, "provider", "open-meteo")
	})

	mux.HandleFunc("GET /meteo", func(w http.ResponseWriter, r *http.Request) {
		// p, err := providers.NewMeteoBlueProvider(cfg)
		p, err := providers.New("meteoblue", cfg)
		if err != nil {
			logger.Error(e.FAIL, "err", err, "description", "Couldn't create MeteoBlueProvider")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// TO-DO : After defining the MB response, write it to the response
		bd, err := p.FetchData(plumber.NewCoordinates(11.0056, 76.9661))
		if err != nil {
			logger.Error(e.FAIL, "err", err, "description", "Couldn't fetch data from MeteoBlueProvider")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
		// fmt.Printf("Data: %+v\n", bd)
		logger.Debug("API Data fetched", "data", bd, "provider", "meteo-blue")
	})

	logger.Info("Ready, Plank? Serving Meteo Munch on " + cfg.Munch.Server.Hostname + ":" + cfg.Munch.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Munch.Server.Hostname, cfg.Munch.Server.Port), mux)
	logger.Error(e.FATAL, "err", err, "description", "Server killed!")
	os.Exit(-1)
}
