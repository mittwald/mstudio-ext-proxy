package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mittwald/mstudio-ext-proxy/pkg/bootstrap"
	"github.com/mittwald/mstudio-ext-proxy/pkg/controller"
	"github.com/mittwald/mstudio-ext-proxy/pkg/persistence"
	"github.com/mittwald/mstudio-ext-proxy/pkg/proxy"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	webhookCtrl := controller.WebhookController{
		ExtensionInstanceRepository: instanceRepository,
		WebhookVerifier:             bootstrap.BuildWebhookVerifier(mittwaldClient),
		Logger:                      logger,
	}
	authCtrl := controller.UserAuthenticationController{
		Client:                mittwaldClient,
		SessionRepository:     sessionRepository,
		InstanceRepository:    instanceRepository,
		Development:           config.Context == "dev",
		AuthenticationOptions: authOptions,
	}

	r := gin.New()
	r.LoadHTMLGlob("templates/*")

	rm := r.Group("/mstudio")
	rm.POST("/webhooks", webhookCtrl.HandleWebhookRequest)
	rm.GET("/auth/oneclick", authCtrl.HandleAuthenticationRequest)
	rm.GET("/auth/fake", authCtrl.HandleFakeAuthentication)

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
			Configuration:         proxyConfig,
			Logger:                logger,
			AuthenticationOptions: authOptions,
			ProxyBufferSize:       config.ProxyBufferSize,
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
