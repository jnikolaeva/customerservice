package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sirupsen/logrus"

	postgresadapter "github.com/jnikolaeva/eshop-common/postgres"

	"github.com/jnikolaeva/customerservice/internal/probes"
	"github.com/jnikolaeva/customerservice/internal/customer/application"
	usertransport "github.com/jnikolaeva/customerservice/internal/customer/infrastructure/transport"

	"github.com/jnikolaeva/customerservice/internal/customer/infrastructure/postgres"
)

const (
	appName     = "customerservice"
	defaultPort = "8080"
)

func main() {
	serverAddr := ":" + envString("APP_PORT", defaultPort)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	})

	errorLogger := gokitlog.NewJSONLogger(gokitlog.NewSyncWriter(os.Stderr))
	errorLogger = level.NewFilter(errorLogger, level.AllowDebug())
	errorLogger = gokitlog.With(errorLogger,
		"appName", appName,
		"@timestamp", gokitlog.DefaultTimestampUTC,
	)

	identityProviderUrl := envString("IDP_URL", "")
	if identityProviderUrl == "" {
		logger.Fatal("environment variable IDP_URL is not set")
	}

	connConfig, err := postgresadapter.ParseEnvConfig(appName)
	if err != nil {
		logger.Fatal(err.Error())
	}
	connectionPool, err := postgresadapter.NewConnectionPool(connConfig)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer connectionPool.Close()

	repository := postgres.New(connectionPool)
	service := application.NewAuthService(application.NewService(repository))
	endpoints := usertransport.MakeEndpoints(service, identityProviderUrl)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/", usertransport.MakeHandler("/api/v1/customers", endpoints, errorLogger))
	mux.Handle("/ready", probes.MakeReadyHandler())
	mux.Handle("/live", probes.MakeLiveHandler())

	srv := startServer(serverAddr, mux, logger)

	waitForShutdown(srv)
	logger.Info("shutting down")
}

func startServer(serverAddr string, handler http.Handler, logger *logrus.Logger) *http.Server {
	srv := &http.Server{Addr: serverAddr, Handler: handler}

	go func() {
		logger.WithFields(logrus.Fields{"url": serverAddr}).Info("starting the server")
		logger.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func waitForShutdown(srv *http.Server) {
	killSignalChan := make(chan os.Signal, 1)
	signal.Notify(killSignalChan, os.Kill, os.Interrupt, syscall.SIGTERM)

	<-killSignalChan
	_ = srv.Shutdown(context.Background())
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
