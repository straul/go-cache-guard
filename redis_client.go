package guard

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// GoRedisClient 使用 go-redis 作为默认客户端
type GoRedisClient struct {
	client *redis.Client
}

// NewGoRedisClient 初始化一个新的 go-redis 客户端
func NewGoRedisClient(options *redis.Options) *GoRedisClient {
	return &GoRedisClient{
		client: redis.NewClient(options),
	}
}

// Get 实现 RedisClientInterface 的 Get 方法
func (c *GoRedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set 实现 RedisClientInterface 的 Set 方法
func (c *GoRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
