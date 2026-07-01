package queue

import (
	"context"
	"fmt"
	"local/complex-web/worker/database"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

const (
	redisStream = "stream"
	redisGroup  = "worker"
)

var defaultConfig = redisConfig{
	Host: "localhost",
	Port: 6379,
}

type redisConfig struct {
	Host string
	Port int
}

func (rc *redisConfig) toAddr() string {
	return fmt.Sprintf("%s:%d", rc.Host, rc.Port)
}

type RedisService struct {
	redisClient *redis.Client
}

func (rs *RedisService) Close() error {
	if rs != nil && rs.redisClient != nil {
		err := rs.redisClient.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (rs *RedisService) CreateGroup(ctx context.Context) error {
	cmd := rs.redisClient.XGroupCreateMkStream(ctx, redisStream, redisGroup, "0")
	if cmd.Err() != nil && strings.Contains(cmd.Err().Error(), "BUSYGROUP Consumer Group name already exists") {
		return nil
	}

	return cmd.Err()
}

func (rs *RedisService) Write(ctx context.Context, value any) error {
	cmd := rs.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: redisStream,
		Values: map[string]any{
			"value": value,
		},
	})

	return cmd.Err()
}

func (rs *RedisService) RunWorker(ctx context.Context, consumerName string, dbService *database.DBPostgres) error {
	for {
		streams, err := rs.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    redisGroup,
			Consumer: consumerName,
			Streams:  []string{redisStream, ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			slog.Error("failed to read from redis stream", "error", err)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				slog.Info("received redis message",
					"id", msg.ID,
					"values", msg.Values,
				)

				num, err := strconv.Atoi(msg.Values["value"].(string))
				if err != nil {
					slog.Error("failed to parse redis value", "error", err)
					continue
				}
				calc := fib(int(num))

				errDb := dbService.Insert(ctx, num, calc)
				if errDb != nil {
					slog.Error("failed to insert data into database", "error", errDb)
				}

				if err := rs.redisClient.XAck(ctx, redisStream, redisGroup, msg.ID).Err(); err != nil {
					slog.Error("failed to ack redis message", "id", msg.ID, "error", err)
				}
			}
		}
	}
}

func InitRedisService(ctx context.Context) (*RedisService, error) {
	cfg, err := readRedisConfig()
	if err != nil {
		slog.Warn("Failed to read redis config", "error", err)
		cfg = &defaultConfig
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.toAddr(),
	})

	status := rdb.Ping(ctx)
	if status.Err() != nil {
		return nil, status.Err()
	}

	srv := new(RedisService{
		redisClient: rdb,
	})

	return srv, srv.CreateGroup(ctx)
}

func readRedisConfig() (*redisConfig, error) {
	host := os.Getenv("REDIS_HOST")

	port, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, err
	}

	return new(redisConfig{
		Host: host,
		Port: port,
	}), nil
}

func fib(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}

	return b
}
