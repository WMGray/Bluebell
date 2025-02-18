package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/setting"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Redis 数据库持久化

type Persistence struct {
	cfg          *setting.RedisPersistenceConfig // 配置项
	lastSyncTime time.Time                       // 上一次同步成功时间
	mu           sync.Mutex                      // 用于保护 lastSyncTime
	cron         *cron.Cron                      // cron 实例
}

// NewPersistence 初始化持久化实例
func NewPersistence(cfg *setting.RedisPersistenceConfig) (error, *Persistence) {
	// 配置校验
	if err := validateConfig(cfg); err != nil {
		zap.L().Fatal("invalid persistence config", zap.Error(err))
		return err, nil
	}
	// 初始化持久化实例
	return nil, &Persistence{
		cfg:          cfg,
		lastSyncTime: time.Time{},
		mu:           sync.Mutex{},
		cron:         cron.New(cron.WithSeconds()),
	}
}

// validateConfig 校验 Redis 持久化配置项
func validateConfig(cfg *setting.RedisPersistenceConfig) error {
	if cfg.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0, got %d", cfg.Interval)
	}
	if cfg.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be greater than 0, got %d", cfg.BatchSize)
	}
	if cfg.RetryCount < 0 {
		return fmt.Errorf("retry_count must be non-negative, got %d", cfg.RetryCount)
	}
	if cfg.ScoreFixedDays < 0 {
		return fmt.Errorf("score_fixed_days must be non-negative, got %d", cfg.ScoreFixedDays)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0, got %d", cfg.Timeout)
	}
	if cfg.LogLevel == "" {
		return fmt.Errorf("log_level cannot be empty")
	}
	return nil
}

// Start 启动持久化任务
func (p *Persistence) Start() (err error) {
	// 获取配置项的 interval 字段，设置定时任务
	spec := fmt.Sprintf("@every %ds", p.cfg.Interval) // 每隔多久执行一次
	_, err = p.cron.AddFunc(spec, func() {
		if err := p.persistData(); err != nil {
			zap.L().Error("persist data failed", zap.Error(err))
			return
		}
		p.updateLastSyncTime() // 更新上次同步成功时间
	}) // 添加定时任务
	if err != nil {
		return
	}
	p.cron.Start() // 启动定时任务
	zap.L().Info("start persistence cron job success")
	return
}

// Stop 停止持久化任务
func (p *Persistence) Stop() {
	p.cron.Stop()
	zap.L().Info("stop persistence cron job success")
}

// persistData 执行数据持久化的具体逻辑
func (p *Persistence) persistData() error {
	// 从 Redis 中获取需要持久化的数据，并持久化到MySQL中
	/*
		p.cfg.BatchSize // 一次性处理的数据量（包括读取和写入）
		p.cfg.Timeout // 一次处理的超时时间
		p.cfg.RetryCount // 持久化失败时的重试次数
		p.cfg.ScoreFixedDays //保留多少天的数据，超过这个天数，Score这个表中该数据就不再进行维护
	*/
	// 1. 从 Redis 中获取帖子分数数据
	postScores, err := redis.FetchPostScores(p.cfg.ScoreFixedDays, p.cfg.RetryCount, p.cfg.Timeout)
	if err != nil {
		zap.L().Error("failed to fetch post scores from redis", zap.Error(err))
		return err
	}
	if len(postScores) == 0 {
		zap.L().Warn("no post scores fetched from redis")
	}

	// 2 从 Redis 中获取帖子投票数据
	postVotes, err := redis.FetchPostVotes(p.cfg.RetryCount, p.cfg.Timeout)
	if err != nil {
		zap.L().Error("failed to fetch post votes from redis", zap.Error(err))
		return err
	}

	// 3. 将数据持久化到 MySQL 中
	if err := mysql.PersistPost(postScores, postVotes, p.cfg.RetryCount, p.cfg.Timeout); err != nil {
		zap.L().Error("failed to persist data", zap.Error(err))
		return err
	}

	zap.L().Info("data persisted successfully")
	return nil
}

// updateLastSyncTime 更新上次同步成功时间
func (p *Persistence) updateLastSyncTime() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastSyncTime = time.Now()
}

// GetLastSyncTime 获取上一次同步成功时间
func (p *Persistence) GetLastSyncTime() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastSyncTime
}

// retryFunc 定义重试逻辑
func retryFunc(retries int, delay time.Duration, operation func() error) error {
	var lastErr error
	for i := 0; i <= retries; i++ {
		if i > 0 {
			time.Sleep(delay)
		} // 重试时增加一定延迟

		// 执行操作
		if err := operation(); err != nil {
			// 失败时记录日志
			zap.L().Warn("operation failed", zap.Error(err))
			lastErr = err
			continue
		}
		// 成功
		return nil
	}

	// 超过最大重试次数仍失败
	return fmt.Errorf("operation failed after %d retries: %w", retries, lastErr)
}
