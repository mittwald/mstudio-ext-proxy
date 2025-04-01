package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mittwald/mstudio-ext-proxy/pkg/bootstrap"
	"github.com/mittwald/mstudio-ext-proxy/pkg/controller"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/service"
	"github.com/mittwald/mstudio-ext-proxy/pkg/persistence"
	"github.com/mittwald/mstudio-ext-proxy/pkg/proxy"
)

func main() {
	config := bootstrap.ConfigFromEnv()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mongoClient := bootstrap.ConnectToMongodb(config.MongoDBURI)
	mongoDatabase := mongoClient.Database("mstudio_ext")
	mittwaldClient := bootstrap.BuildMittwaldAPIClientFromConfig(config, logger)
	authOptions := bootstrap.BuildAuthenticationOptions(config)

	instanceRepository := persistence.NewMongoExtensionInstanceRepository(mongoDatabase.Collection("instances"))
	sessionRepository := persistence.MustNewMongoSessionRepository(mongoDatabase.Collection("sessions"))

	sessionService := service.NewSessionService(mittwaldClient, sessionRepository, instanceRepository)

	webhookCtrl := controller.WebhookController{
		ExtensionInstanceRepository: instanceRepository,
		WebhookVerifier:             bootstrap.BuildWebhookVerifier(mittwaldClient),
		Logger:                      logger,
	}
	authCtrl := controller.UserAuthenticationController{
		Client:                mittwaldClient,
		SessionRepository:     sessionRepository,
		SessionService:        sessionService,
		InstanceRepository:    instanceRepository,
		Development:           config.Context == "dev",
		AuthenticationOptions: authOptions,
		Logger:                logger,
	}

	r := gin.New()
	r.LoadHTMLGlob("templates/*")

	rm := r.Group("/mstudio")
	rm.POST("/webhooks", webhookCtrl.HandleWebhookRequest)
	rm.GET("/auth/oneclick", authCtrl.HandleAuthenticationRequest)
	rm.GET("/auth/fake", authCtrl.HandleFakeAuthentication)
	rm.GET("/auth/current", authCtrl.HandleUserInfo)

	if authOptions.StaticPassword != "" {
		rm.Any("/auth/password", authCtrl.HandlePasswordAuthentication)
	}

	mux := http.NewServeMux()
	mux.Handle("/mstudio/", r)

	for prefix, proxyConfig := range config.Upstreams {
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		proxyHandler := proxy.Handler{
			HTTPClient:            http.DefaultClient,
			SessionRepository:     sessionRepository,
			SessionService:        sessionService,
			Configuration:         proxyConfig,
			Logger:                logger,
			AuthenticationOptions: authOptions,
		}

		mux.Handle(prefix, &proxyHandler)
	}

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
