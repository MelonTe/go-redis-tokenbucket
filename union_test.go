package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

func TestTokenBucketLimiterConcurrent(t *testing.T) {
	// 连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "xxx",
		DB:       0,
	})

	// 测试Redis连接
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("连接Redis失败: %v", err)
	}
	defer redisClient.Close()

	// 创建令牌桶限流器，初始容量为90
	limiter := NewTokenBucketLimiter(redisClient,
		SetCapacity(90),              // 桶容量为90
		SetInitTokens(90),            // 初始令牌数为90
		SetRate(20*time.Millisecond), // 每20ms生成一个令牌
	)

	// 测试key
	testKey := "test:ratelimit:concurrent"
	// 清除可能存在的旧数据
	redisClient.Del(ctx, testKey)

	// 定义总请求数
	concurrentRequests := 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	// 用于统计成功请求
	successCount := 0

	// 启动并发请求
	for i := 0; i < concurrentRequests; i++ {
		go func(requestID int) {
			defer wg.Done()
			// 每个请求尝试获取1个令牌
			allowed, err := limiter.Allow(ctx, testKey, 1)
			if err != nil {
				t.Logf("请求[%d]发生错误: %v", requestID, err)
				return
			}

			status := "失败"
			if allowed {
				status = "成功"
				successCount++
			}
			t.Logf("请求[%d]获取令牌%s", requestID, status)
		}(i + 1)
	}

	// 等待所有请求完成
	wg.Wait()

	// 输出最终结果
	t.Logf("总请求数: %d, 成功请求数: %d", concurrentRequests, successCount)
}
