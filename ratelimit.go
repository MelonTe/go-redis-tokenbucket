package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// 预编译LUA脚本,获取令牌的核心
var bucketScript = redis.NewScript(`
--[[
	参数:
	1.key:令牌桶的key
	2.tokenGenerateRate:令牌生成速率,ns/perToken
	3.tokenLimit:令牌桶最大令牌容量
	4.tokenInitNums:令牌桶初始令牌数量
	5.curTime:当前时间戳,单位为ns
	6.tokenRequest:请求的令牌数量
	7.expires:令牌桶的过期时间,单位为ms
]]
local key = KEYS[1]
local tokenGenerateRate = tonumber(ARGV[1])
local tokenLimit = tonumber(ARGV[2])
local tokenInitNums = tonumber(ARGV[3])
local curTime = tonumber(ARGV[4])
local tokenRequest = tonumber(ARGV[5])
local expires = tonumber(ARGV[6])
--获取当前桶的状态
local bucket = redis.call("HMGET",key,"tokens","timestamp")
local tokens = tonumber(bucket[1])
local timestamp = tonumber(bucket[2])

--校验桶是否存在，不存在则初始化桶
if not tokens or not timestamp then
	tokens = tokenInitNums
	timestamp = curTime
end

--计算需要填充的令牌数量
local delta = curTime - timestamp
if delta < 0 then delta = 0 end
local fillTokens = math.floor(delta / tokenGenerateRate)
--计算当前桶的令牌数量
tokens = math.min(tokens + fillTokens, tokenLimit)

--决定是否允许请求
local allow = 0
if tokens >= tokenRequest then
	--允许请求，更新桶的状态
	tokens = tokens - tokenRequest
	--更新桶的时间戳
	timestamp = timestamp + fillTokens * tokenGenerateRate
	allow = 1
end

--更新桶的状态
redis.call("HSET",key,"tokens",tokens,"timestamp",timestamp)
redis.call("PEXPIRE",key,expires)
--返回结果
return allow
`)

// TokenBucketLimiter 基于 Redis 的分布式令牌桶限流器
type TokenBucketLimiter struct {
	client         *redis.Client // Redis 客户端
	rate           time.Duration // 令牌生成速率，单位是时间间隔（e.g. 600*time.Nanosecond）
	capacity       int64         // 桶容量（最大令牌数）
	initTokens     int64         // 初始令牌数
	expireDuration time.Duration // Redis Key 的过期时长（e.g. 10*time.Millisecond）
}

type Option func(*TokenBucketLimiter)

// NewTokenBucketLimiter 创建一个新的令牌桶限流器
func NewTokenBucketLimiter(client *redis.Client, opts ...Option) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		client: client,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	optionFix(limiter)

	return limiter
}

// optionFix 修正配置
func optionFix(l *TokenBucketLimiter) {
	if l.rate <= 0 {
		l.rate = time.Millisecond // 默认 1 ms/令牌 → 1000 QPS
	}
	if l.capacity <= 0 {
		l.capacity = 1000
	}
	if l.initTokens <= 0 {
		l.initTokens = l.capacity // 启动时桶内满令牌
	}
	if l.expireDuration <= 0 {
		// 过期时长 = refill 完全两次所需时间
		l.expireDuration = time.Duration(l.capacity) * l.rate * 2
		if l.expireDuration < time.Millisecond {
			l.expireDuration = time.Millisecond
		}
	}
	if l.capacity < l.initTokens {
		l.initTokens = l.capacity // 初始令牌数不能超过桶容量
	}
}

// SetRate 设置生成速率（纳秒级），例如 10*time.Microsecond → 100k QPS
func SetRate(d time.Duration) Option {
	return func(l *TokenBucketLimiter) { l.rate = d }
}

// SetCapacity 设置桶容量
func SetCapacity(c int64) Option {
	return func(l *TokenBucketLimiter) { l.capacity = c }
}

// SetInitTokens 设置初始令牌数
func SetInitTokens(n int64) Option {
	return func(l *TokenBucketLimiter) { l.initTokens = n }
}

// SetExpireDuration 设置 Key 过期时长
func SetExpireDuration(d time.Duration) Option {
	return func(l *TokenBucketLimiter) { l.expireDuration = d }
}

// Allow 尝试获取 tokenRequest 个令牌，返回 true 表示允许
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string, tokenRequest int64) (bool, error) {
	// 转换参数
	rateNs := l.rate.Nanoseconds()
	capv := l.capacity
	initv := l.initTokens
	nowNs := time.Now().UnixNano()
	expMs := l.expireDuration.Milliseconds()

	res, err := bucketScript.Run(ctx, l.client,
		[]string{key},
		rateNs, capv, initv, nowNs, tokenRequest, expMs,
	).Result()
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}
