package main

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	api := router.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/get-config", handleSerializeUserConfig)

	router.PathPrefix("/")

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("x-streamdeck-version", version)
			w.Header().Set("cache-control", "no-cache")

			h.ServeHTTP(w, r)
		})
	})
}

func handleSerializeUserConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(userConfig); err != nil {
		log.WithError(err).Error("Unable to marshal user config")
		http.Error(w, errors.Wrap(err, "marshal user config").Error(), http.StatusInternalServerError)
		return
	}
}
