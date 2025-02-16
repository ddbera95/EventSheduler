package server

import (
	"EventTrigger/data"
	"EventTrigger/event"
	"EventTrigger/web"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	data.Init()
	event.Init()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	web.Init(router)

	port := os.Getenv("PORT")
	certificate := os.Getenv("CERTIFICATE")
	privateKey := os.Getenv("PRIVATE_KEY")
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port), // Port to listen on
		Handler: router,
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started on port %s", port)
		if err := srv.ListenAndServeTLS(certificate, privateKey); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v", err)
	}

	log.Println("Shutting down server...")

}
