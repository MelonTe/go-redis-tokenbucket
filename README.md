# go-redis-tokenbucket

A distributed token bucket rate limiter implementation based on Redis and Lua script.(Need to use with go-redis/v9)

## Installation

```bash
go get github.com/MelonTe/go-redis-tokenbucket
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/MelonTe/go-redis-tokenbucket/ratelimit"
	"time"
)

func main() {
	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a new token bucket limiter
	// Default: 1000 QPS, bucket capacity of 1000
	limiter := ratelimit.NewTokenBucketLimiter(redisClient)

	// Try to acquire a token
	ctx := context.Background()
	key := "rate:limit:example"
	allowed, err := limiter.Allow(ctx, key, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if allowed {
		fmt.Println("Request allowed")
		// Process the request...
	} else {
		fmt.Println("Request rejected (rate limited)")
	}
}
```

### Custom Configuration

You can customize the token bucket using these options:

```go
limiter := ratelimit.NewTokenBucketLimiter(redisClient,
	// Set token generation rate to 100000 tokens per second (10ns per token)
	ratelimit.SetRate(10 * time.Nanosecond),

	// Set bucket capacity to 200 tokens
	ratelimit.SetCapacity(200),

	// Set initial token count to 50
	ratelimit.SetInitTokens(50),

	// Set Redis key expiration to 1 minute
	ratelimit.SetExpireDuration(time.Minute),
)
```

## Configuration Options

| Option            | Description                                         | Default               |
| ----------------- | --------------------------------------------------- | --------------------- |
| SetRate           | Token generation rate                               | 1ms (1000 QPS)        |
| SetCapacity       | Maximum number of tokens in the bucket              | 1000                  |
| SetInitTokens     | Initial number of tokens when creating a new bucket | Same as capacity      |
| SetExpireDuration | Redis key expiration time                           | 2 * (capacity * rate) |

---

# go-redis-tokenbucket

基于 Redis 和 Lua 脚本实现的分布式令牌桶限流器（需要与 go-redis/v9 一起使用）

## 安装

```bash
go get github.com/MelonTe/go-redis-tokenbucket
```

## 使用方法

### 基本示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/MelonTe/go-redis-tokenbucket/ratelimit"
	"time"
)

func main() {
	// 连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       0,  // 使用默认数据库
	})

	// 创建一个新的令牌桶限流器
	// 默认: 1000 QPS, 桶容量为1000
	limiter := ratelimit.NewTokenBucketLimiter(redisClient)

	// 尝试获取一个令牌
	ctx := context.Background()
	key := "rate:limit:example"
	allowed, err := limiter.Allow(ctx, key, 1)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if allowed {
		fmt.Println("请求允许")
		// 处理请求...
	} else {
		fmt.Println("请求被拒绝（速率限制）")
	}
}
```

### 自定义配置

你可以使用这些选项自定义令牌桶：

```go
limiter := ratelimit.NewTokenBucketLimiter(redisClient,
	// 设置令牌生成速率为每秒100000个令牌（每令牌10纳秒）
	ratelimit.SetRate(10 * time.Nanosecond),

	// 设置桶容量为200个令牌
	ratelimit.SetCapacity(200),

	// 设置初始令牌数为50
	ratelimit.SetInitTokens(50),

	// 设置Redis键过期时间为1分钟
	ratelimit.SetExpireDuration(time.Minute),
)
```

## 配置选项

| 选项              | 描述                     | 默认值                      |
| ----------------- | ------------------------ | --------------------------- |
| SetRate           | 令牌生成速率             | 1 毫秒 (1000 QPS)           |
| SetCapacity       | 桶中最大令牌数量         | 1000                        |
| SetInitTokens     | 创建新桶时的初始令牌数量 | 与容量相同                  |
| SetExpireDuration | Redis 键过期时间         | 2 * (桶容量 * 令牌生成速率) |

---
