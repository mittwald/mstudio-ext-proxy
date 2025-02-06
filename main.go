package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mittwald/mstudio-ext-proxy/pkg/bootstrap"
	"github.com/mittwald/mstudio-ext-proxy/pkg/controller"
	"github.com/mittwald/mstudio-ext-proxy/pkg/persistence"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func main() {
	mongoClient := bootstrap.ConnectToMongodb(os.Getenv("MONGODB_URI"))
	mongoDatabase := mongoClient.Database("mstudio_ext")
	mittwaldClient := bootstrap.BuildMittwaldAPIClientFromEnv()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	instanceRepository := persistence.NewMongoExtensionInstanceRepository(mongoDatabase.Collection("instances"))
	sessionRepository := persistence.NewMongoSessionRepository(mongoDatabase.Collection("sessions"))

	webhookCtrl := controller.WebhookController{
		ExtensionInstanceRepository: instanceRepository,
		WebhookVerifier:             bootstrap.BuildWebhookVerifier(mittwaldClient),
		Logger:                      logger,
	}
	authCtrl := controller.UserAuthenticationController{
		Client:            mittwaldClient,
		SessionRepository: sessionRepository,
		Development:       true,
	}

	r := gin.New()
	rm := r.Group("/mstudio")
	rm.POST("/webhooks", webhookCtrl.HandleWebhookRequest)
	rm.GET("/auth", authCtrl.HandleAuthenticationRequest)

	mux := http.NewServeMux()
	mux.Handle("/mstudio/", r)

	s := http.Server{
		Handler: mux,
		Addr:    getListenAddr(),
	}

	logger.Info("listening", "server.addr", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func getListenPort() int64 {
	if p := os.Getenv("PORT"); p != "" {
		pi, err := strconv.ParseInt(p, 10, 32)
		if err != nil {
			panic(err)
		}

		return pi
	}

	return 8000
}

func getListenAddr() string {
	return fmt.Sprintf(":%d", getListenPort())
}
