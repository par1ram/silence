package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client представляет Redis клиент с дополнительными методами
type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
	prefix string
}

// Config конфигурация Redis клиента
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Prefix   string
}

// NewClient создает новый Redis клиент
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "silence"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis client connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("prefix", cfg.Prefix))

	return &Client{
		rdb:    rdb,
		logger: logger,
		prefix: cfg.Prefix,
	}, nil
}

// Close закрывает соединение с Redis
func (c *Client) Close() error {
	return c.rdb.Close()
}

// key создает ключ с префиксом
func (c *Client) key(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Set устанавливает значение с TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.Set(ctx, c.key(key), data, ttl).Err(); err != nil {
		c.logger.Error("failed to set value in Redis",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}

// Get получает значение из Redis
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, c.key(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete удаляет ключ
func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, c.key(key)).Err(); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// Exists проверяет существование ключа
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, c.key(key)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// SetNX устанавливает значение только если ключ не существует
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	ok, err := c.rdb.SetNX(ctx, c.key(key), data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set value NX: %w", err)
	}

	return ok, nil
}

// Increment увеличивает числовое значение
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.rdb.Incr(ctx, c.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	return val, nil
}

// IncrementBy увеличивает числовое значение на указанное число
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := c.rdb.IncrBy(ctx, c.key(key), value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment by: %w", err)
	}
	return val, nil
}

// Expire устанавливает TTL для ключа
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.rdb.Expire(ctx, c.key(key), ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expire: %w", err)
	}
	return nil
}

// TTL получает оставшееся время жизни ключа
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.rdb.TTL(ctx, c.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// Hash операции с хеш-таблицами

// HSet устанавливает поле в хеш-таблице
func (c *Client) HSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.HSet(ctx, c.key(key), field, data).Err(); err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}

	return nil
}

// HGet получает поле из хеш-таблицы
func (c *Client) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	data, err := c.rdb.HGet(ctx, c.key(key), field).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// HDel удаляет поле из хеш-таблицы
func (c *Client) HDel(ctx context.Context, key string, field string) error {
	if err := c.rdb.HDel(ctx, c.key(key), field).Err(); err != nil {
		return fmt.Errorf("failed to delete hash field: %w", err)
	}
	return nil
}

// HExists проверяет существование поля в хеш-таблице
func (c *Client) HExists(ctx context.Context, key string, field string) (bool, error) {
	exists, err := c.rdb.HExists(ctx, c.key(key), field).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check hash field existence: %w", err)
	}
	return exists, nil
}

// HIncrBy увеличивает числовое значение поля в хеш-таблице
func (c *Client) HIncrBy(ctx context.Context, key string, field string, incr int64) (int64, error) {
	val, err := c.rdb.HIncrBy(ctx, c.key(key), field, incr).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment hash field: %w", err)
	}
	return val, nil
}

// HGetAll получает все поля из хеш-таблицы
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	data, err := c.rdb.HGetAll(ctx, c.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields: %w", err)
	}
	return data, nil
}

// HKeys получает все ключи из хеш-таблицы
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	keys, err := c.rdb.HKeys(ctx, c.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash keys: %w", err)
	}
	return keys, nil
}

// List операции со списками

// LPush добавляет элемент в начало списка
func (c *Client) LPush(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.LPush(ctx, c.key(key), data).Err(); err != nil {
		return fmt.Errorf("failed to push to list: %w", err)
	}

	return nil
}

// RPush добавляет элемент в конец списка
func (c *Client) RPush(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.RPush(ctx, c.key(key), data).Err(); err != nil {
		return fmt.Errorf("failed to push to list: %w", err)
	}

	return nil
}

// LPop удаляет и возвращает первый элемент списка
func (c *Client) LPop(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.LPop(ctx, c.key(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to pop from list: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// LRange получает элементы списка в диапазоне
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	data, err := c.rdb.LRange(ctx, c.key(key), start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range: %w", err)
	}
	return data, nil
}

// LLen получает длину списка
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	length, err := c.rdb.LLen(ctx, c.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length: %w", err)
	}
	return length, nil
}

// Set операции с множествами

// SAdd добавляет элемент в множество
func (c *Client) SAdd(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.SAdd(ctx, c.key(key), data).Err(); err != nil {
		return fmt.Errorf("failed to add to set: %w", err)
	}

	return nil
}

// SRem удаляет элемент из множества
func (c *Client) SRem(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.SRem(ctx, c.key(key), data).Err(); err != nil {
		return fmt.Errorf("failed to remove from set: %w", err)
	}

	return nil
}

// SIsMember проверяет принадлежность элемента множеству
func (c *Client) SIsMember(ctx context.Context, key string, value interface{}) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	isMember, err := c.rdb.SIsMember(ctx, c.key(key), data).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership: %w", err)
	}

	return isMember, nil
}

// SMembers получает все элементы множества
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := c.rdb.SMembers(ctx, c.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return members, nil
}

// SCard получает количество элементов в множестве
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	count, err := c.rdb.SCard(ctx, c.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get set cardinality: %w", err)
	}
	return count, nil
}

// Sorted Set операции с отсортированными множествами

// ZAdd добавляет элемент в отсортированное множество
func (c *Client) ZAdd(ctx context.Context, key string, score float64, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.ZAdd(ctx, c.key(key), redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to sorted set: %w", err)
	}

	return nil
}

// ZRem удаляет элемент из отсортированного множества
func (c *Client) ZRem(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.rdb.ZRem(ctx, c.key(key), data).Err(); err != nil {
		return fmt.Errorf("failed to remove from sorted set: %w", err)
	}

	return nil
}

// ZRange получает элементы отсортированного множества в диапазоне
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	data, err := c.rdb.ZRange(ctx, c.key(key), start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range: %w", err)
	}
	return data, nil
}

// ZIncrBy увеличивает счетчик элемента в отсортированном множестве
func (c *Client) ZIncrBy(ctx context.Context, key string, increment float64, member interface{}) (float64, error) {
	data, err := json.Marshal(member)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal member: %w", err)
	}

	score, err := c.rdb.ZIncrBy(ctx, c.key(key), increment, string(data)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment sorted set member: %w", err)
	}

	return score, nil
}

// ZCard получает количество элементов в отсортированном множестве
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	count, err := c.rdb.ZCard(ctx, c.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get sorted set cardinality: %w", err)
	}
	return count, nil
}

// Pipeline операции с пайплайнами

// Pipeline создает новый пайплайн
func (c *Client) Pipeline() redis.Pipeliner {
	return c.rdb.Pipeline()
}

// Transaction выполняет транзакцию
func (c *Client) Transaction(ctx context.Context, fn func(redis.Pipeliner) error) error {
	_, err := c.rdb.TxPipelined(ctx, fn)
	return err
}

// Health проверяет состояние соединения
func (c *Client) Health(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Keys получает все ключи по шаблону
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := c.rdb.Keys(ctx, c.key(pattern)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	// Убираем префикс из ключей
	result := make([]string, len(keys))
	prefixLen := len(c.prefix) + 1
	for i, key := range keys {
		if len(key) > prefixLen {
			result[i] = key[prefixLen:]
		} else {
			result[i] = key
		}
	}

	return result, nil
}

// FlushDB очищает текущую базу данных
func (c *Client) FlushDB(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}

// Info получает информацию о Redis сервере
func (c *Client) Info(ctx context.Context) (string, error) {
	return c.rdb.Info(ctx).Result()
}
