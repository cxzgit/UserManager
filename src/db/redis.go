package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

// 定义 Redis 客户端和上下文
var (
	// RedisClient 对外暴露全局 Redis 客户端实例
	RedisClient *redis.Client
	// 定义上下文，后续 Redis 操作均使用该上下文
	Ctx = context.Background()
)

// InitRedis 初始化 Redis 连接
func InitRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,     // 如 "localhost:6379"
		Password: password, // 没有密码则传空字符串 ""
		DB:       db,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	log.Println("Redis 初始化成功")
}
