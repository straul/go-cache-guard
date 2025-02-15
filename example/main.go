package main

import (
	"context"
	"fmt"
	"time"

	guard "go-cache-guard"

	"github.com/go-redis/redis/v8" // 示例中使用与 redis_client.go 中不同的 redis 客户端
)

// MyCustomRedisClient 是你自定义的 Redis 客户端，使用 go-redis v8
type MyCustomRedisClient struct {
	client *redis.Client
}

// NewMyCustomRedisClient 初始化一个新的 go-redis v8 客户端
func NewMyCustomRedisClient(options *redis.Options) *MyCustomRedisClient {
	return &MyCustomRedisClient{
		client: redis.NewClient(options),
	}
}

// Get 实现 RedisClientInterface 的 Get 方法
func (c *MyCustomRedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set 实现 RedisClientInterface 的 Set 方法
func (c *MyCustomRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func main() {
	// 使用自定义客户端
	myClient := NewMyCustomRedisClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 使用配置结构体创建 RedisHandler
	handler, err := guard.NewRedisHandler(guard.RedisHandlerConfig{
		Client: myClient,
	})
	if err != nil {
		fmt.Println("Error creating RedisHandler:", err)
		return
	}

	// 使用 handler 进行 Redis 操作
	err = handler.WriteKey(context.Background(), "key", "value", 10*time.Second)
	if err != nil {
		fmt.Println("Error writing key:", err)
	}

	value, err := handler.ReadKey(context.Background(), "key")
	if err != nil {
		fmt.Println("Error reading key:", err)
	} else {
		fmt.Println("Value:", value)
	}
}
