package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/config"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/handler"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/server"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/service"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/cache"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/multitemplate"
)

func Run() {
	cfg := config.Get()

	multitemplate.AddLayout("email", "layouts/email.html")
	multitemplate.AddTemplate("adminNewOrganization", "email", "adminNewOrganization.html")
	multitemplate.AddTemplate("userNewOrganization", "email", "userNewOrganization.html")
	multitemplate.AddTemplate("userVerification", "email", "userVerification.html")
	multitemplate.AddTemplate("userVerificationWithInvite", "email", "userVerificationWithInvite.html")
	multitemplate.AddTemplate("userPasswordReset", "email", "userPasswordReset.html")
	multitemplate.AddTemplate("userInvite", "email", "userInvite.html")

	db, err := repository.NewMariaDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	} else {
		log.Printf("DB connection established")
	}

	cache := cache.NewMemoryCache()
	repos := repository.NewRepository(db)
	service := service.NewService(repos, cache)
	handlers := handler.NewHandler(service)

	r := handlers.InitRoutes()

	srv := server.NewServer(cfg, r)

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatalln("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Stop(shutdownCtx)
		if err != nil {
			log.Fatalln(err)
		}
		serverStopCtx()
	}()

	// Run the server
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error occurred while running http server: %s\n", err.Error())
		}
	}()
	log.Printf("Server is running on http://%s:%s\n", cfg.Server.Host, cfg.Server.Port)

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
