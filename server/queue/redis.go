package queue

import (
	"context"
	"fmt"
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
