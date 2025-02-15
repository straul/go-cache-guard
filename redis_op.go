package guard

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

// ******************** 基础设置 ********************

const (
	ExpireTypeRandom = "random"
	LockKeySuffix    = "_lock"
)

type RedisClientInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// other redis methods...
}

// RedisOptions Custom Redis connection options
type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}

// RedisHandlerConfig Config for RedisHandler
type RedisHandlerConfig struct {
	Client       RedisClientInterface
	Options      *RedisOptions
	BackfillFunc func(ctx context.Context, key string) (string, error) // 回源函数
}

type RedisHandler struct {
	client          RedisClientInterface
	slidingExpire   bool
	slidingDuration time.Duration
	autoBackfill    bool
	expireType      string
	randomDuration  time.Duration
	locker          *redislock.Client
	backfillFunc    func(ctx context.Context, key string) (string, error)
}

// NewRedisHandler Create a new RedisHandler
func NewRedisHandler(config RedisHandlerConfig) (*RedisHandler, error) {
	var client RedisClientInterface
	if config.Client != nil {
		client = config.Client
	} else if config.Options != nil {
		// Convert RedisOptions to *redis.Options
		redisOptions := &redis.Options{
			Addr:     config.Options.Addr,
			Password: config.Options.Password,
			DB:       config.Options.DB,
		}
		client = NewGoRedisClient(redisOptions)
	} else {
		return nil, errors.New("either Client or Options must be provided")
	}

	locker := redislock.New(client.(*GoRedisClient).client)

	return &RedisHandler{
		client:       client,
		locker:       locker,
		backfillFunc: config.BackfillFunc,
	}, nil
}

// SetAutoBackfill 设置是否自动回源
func (h *RedisHandler) SetAutoBackfill(autoBackfill bool) {
	h.autoBackfill = autoBackfill
}

// SetSlidingExpire 设置滑动延期
func (h *RedisHandler) SetSlidingExpire(slidingExpire bool) {
	h.slidingExpire = slidingExpire
}

// SetSlidingDuration 设置滑动延期时长
func (h *RedisHandler) SetSlidingDuration(duration time.Duration) {
	h.slidingDuration = duration
}

// SetExpireType 设置有效期类型
func (h *RedisHandler) SetExpireType(expireType string) {
	h.expireType = expireType
}

// SetRandomDuration 设置随机延长时长
func (h *RedisHandler) SetRandomDuration(duration time.Duration) {
	h.randomDuration = duration
}

// CheckExpireTypeRandom 检查有效期类型是否为随机
func (h *RedisHandler) CheckExpireTypeRandom() bool {
	return h.expireType == ExpireTypeRandom
}

// ******************** 缓存操作 ********************

// ReadKey Read from cache
func (h *RedisHandler) ReadKey(ctx context.Context, key string) (string, error) {
	value, err := h.client.Get(ctx, key)
	if err == nil {
		if h.slidingExpire {
			h.WriteKey(ctx, key, value, h.slidingDuration)
		}
		return value, nil
	}

	if h.autoBackfill && h.backfillFunc != nil {
		// lock
		lock, err := h.locker.Obtain(ctx, key+LockKeySuffix, 10*time.Second, nil)
		if err != nil {
			return "", err
		}
		defer lock.Release(ctx)

		// backfill
		value, err = h.backfillFunc(ctx, key)
		if err != nil {
			return "", err
		}

		// write
		h.WriteKey(ctx, key, value, h.slidingDuration)
	}

	return value, err
}

// WriteKey Write to cache
func (h *RedisHandler) WriteKey(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if h.CheckExpireTypeRandom() {
		expiration += time.Duration(rand.Int63n(int64(h.randomDuration)))
	}
	return h.client.Set(ctx, key, value, expiration)
}
