package redis

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-redis/redis"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

//go:generate mockery --name RedisClientInterface
type RedisClientInterface interface {
	InitRedisClient(address, password, caPath string, database int, tlsEnabled bool) RedisClientInterface
	Ping() (string, error)
	Addr() string
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Close() error
}

func (r RedisClient) InitRedisClient(address, password, caPath string, database int, tlsEnabled bool) RedisClientInterface {
	var tlsConfig *tls.Config
	if tlsEnabled {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caPath))

		// Setup TLS client
		tlsConfig = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
		}
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:      address,
		Password:  password,
		DB:        database,
		TLSConfig: tlsConfig,
	})

	return r
}

func (r RedisClient) Addr() string {
	return r.client.Options().Addr
}

func (r RedisClient) Ping() (string, error) {
	return r.client.Ping().Result()
}

func (r RedisClient) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r RedisClient) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(key, value, expiration).Err()
}

func (r RedisClient) Close() error {
	return r.client.Close()
}
