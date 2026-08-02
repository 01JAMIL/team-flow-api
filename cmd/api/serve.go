package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *application) serve(h http.Handler) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      h,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on port %d\n", app.config.port)

	return server.ListenAndServe()
}
