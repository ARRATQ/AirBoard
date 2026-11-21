package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
)

// Délais du serveur HTTP. Sans eux, un client lent (ou malveillant) garde une
// connexion et un goroutine ouverts indéfiniment.
const (
	// En-têtes : défense contre Slowloris.
	serverReadHeaderTimeout = 10 * time.Second
	// Requête complète, corps compris : un envoi de 12 Mo (limite des uploads) doit
	// passer sur une connexion lente. En prod, nginx met le corps en tampon avant de
	// le transmettre, ce délai protège surtout un accès direct au backend.
	serverReadTimeout = 2 * time.Minute
	// Traitement + écriture de la réponse (exports, rapports). Le WebSocket du chat
	// n'est pas concerné : net/http efface les délais d'une connexion détournée
	// (Hijack), puis le chat pose ses propres délais (ping/pong, écriture).
	serverWriteTimeout = 2 * time.Minute
	// Connexions keep-alive inactives.
	serverIdleTimeout = 2 * time.Minute
	// Arrêt propre : laisser finir les requêtes en cours, sous les 10 s que Docker
	// accorde entre SIGTERM et SIGKILL.
	serverShutdownTimeout = 8 * time.Second
)

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
}

// runServer sert les requêtes jusqu'à SIGINT/SIGTERM, puis arrête le serveur
// proprement (requêtes en cours terminées) et ferme la base.
func runServer(srv *http.Server, db *gorm.DB) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	stop() // un second signal tue le processus sans attendre
	log.Println("Arrêt demandé, fin des requêtes en cours…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(errors.New("arrêt du serveur"), shutdownErr)
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Avertissement: fermeture de la base: %v", err)
		}
	}
	log.Println("Serveur arrêté")
	return shutdownErr
}
