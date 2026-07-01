package main

import (
	"context"
	"local/complex-web/worker/database"
	"local/complex-web/worker/queue"
	"log/slog"
	"os"
	"sync"
)

func main() {
	ctx := context.Background()

	queueService, err := queue.InitRedisService(ctx)
	if err != nil {
		slog.Error("Failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer func() {
		err := queueService.Close()
		if err != nil {
			slog.Error("Failed to close redis", "err", err)
			os.Exit(1)
		}
	}()

	dbService, err := database.GetConnectionPostgres()
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}

	wg.Go(
		func() {
			if err := queueService.RunWorker(ctx, "worker-1", dbService); err != nil {
				slog.Error("redis worker stopped", "error", err)
			}
		},
	)

	wg.Wait()

	defer func() {
		err := recover()
		if err != nil {
			slog.Error("redis worker stopped", "error", err)
		}
	}()
}
