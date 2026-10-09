package application

import (
	"context"
	"log"
	"path/filepath"

	"github.com/brienze1/xenon/internal/infra/dependency/consumer"
	"github.com/brienze1/xenon/internal/infra/dependency/router"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/pkg/logger"
)

func Start() {
	ctx := context.Background()
	sink, err := logger.SetFileSink(filepath.Join(properties.Properties().Application.DataDir, "logs", "xenon.log"))
	if err != nil {
		logger.Warn(ctx, err, "Log file unavailable, logging to stdout")
	} else {
		defer func() { _ = sink.Close() }()
	}
	logger.Info(ctx, "Starting application", map[string]any{"version": properties.Properties().Application.Version})

	consumer.GetFileChangeConsumer().Register(ctx)
	if err := router.GetRouter().Route().Run(); err != nil {
		log.Fatal(err)
	}
}
