package redis

import (
	"bluebell/models"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// FetchPostScores 从 Redis 中获取帖子分数数据
func FetchPostScores(expiredDays int, retryCount, timeout int) (data []*models.PostScore, err error) {
	// 从 Redis 中获取一次帖子分数数据
	for attempt := 1; attempt <= retryCount; attempt++ {
		// 设置函数超时时间
		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		data, err = fetchPostScoresOnce(timeoutctx, expiredDays)
		cancel() // 调用 cancel 释放资源
		if err == nil {
			return data, nil // 如果成功，直接返回结果
		}

		// 检查是否是超时错误
		if errors.Is(timeoutctx.Err(), context.DeadlineExceeded) {
			zap.L().Warn("FetchPostScores timeout, retrying...",
				zap.Int("attempt", attempt),
				zap.Int("retryCount", retryCount),
				zap.Error(timeoutctx.Err()),
			)
		} else {
			zap.L().Warn("FetchPostScores failed, retrying...",
				zap.Int("attempt", attempt),
				zap.Int("retryCount", retryCount),
				zap.Error(err),
			)
		}

		// 如果是最后一次重试，不再继续
		if attempt < retryCount {
			time.Sleep(time.Duration(attempt) * time.Second) // 退避时间
		}
	}
	zap.L().Error("FetchPostScores failed after all retries", zap.Error(err))
	// 重试结束后仍失败，返回错误
	return nil, err
}

// FetchPostVotes 从 Redis 中获取帖子投票数据
func FetchPostVotes(retryCount, timeout int) (data []*models.PostVoteData, err error) {
	for attempt := 1; attempt <= retryCount; attempt++ {
		// 设置超时时间
		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		data, err = fetchPostVotesOnce(timeoutctx)
		cancel()
		if err == nil {
			return data, nil // 如果成功，直接返回结果
		}

		// 检查是否是超时错误
		if errors.Is(timeoutctx.Err(), context.DeadlineExceeded) {
			zap.L().Warn("FetchPostVotes timeout, retrying...",
				zap.Int("attempt", attempt),
				zap.Int("retryCount", retryCount),
				zap.Error(timeoutctx.Err()),
			)
		} else {
			zap.L().Warn("FetchPostVotes failed, retrying...",
				zap.Int("attempt", attempt),
				zap.Int("retryCount", retryCount),
				zap.Error(err),
			)
		}

		// 如果是最后一次重试，不再继续
		if attempt < retryCount {
			time.Sleep(time.Duration(attempt) * time.Second) // 退避时间
		}
	}

	// 重试结束后仍失败，记录错误日志
	zap.L().Error("FetchPostVotes failed after all retries", zap.Error(err))
	return nil, err
}

// fetchPostScoresOnce 从 Redis 中获取一次帖子分数数据
func fetchPostScoresOnce(ctx context.Context, expiredDays int) (data []*models.PostScore, err error) {
	/*
		从 Redis 中获取需要持久化的数据
		expiredDays: 保留多少天的数据，超过这个天数，Score这个表中该数据就不再进行维护
		batchSize: 一次性处理的数据量（包括读取和写入）
		timeout: 一次处理的超时时间
		retryCount: 持久化失败时的重试次数
	*/
	data = make([]*models.PostScore, 0)
	// 1. 提取未过期的帖子
	postIDs, err := GetPostIDsByTimeRange(ctx, expiredDays)
	if err != nil {
		zap.L().Error("failed to fetch unexpired post IDs", zap.Error(err))
		return nil, err
	}

	// 2. 从 KeyPostScoreZSet 中获取帖子分数
	scoresMap, err := GetPostScoreByIDs(ctx, postIDs)
	if err != nil {
		zap.L().Error("failed to fetch post scores", zap.Error(err))
		return nil, err
	}

	// 4. 组装数据
	for _, postID := range postIDs {
		if score, ok := scoresMap[postID]; ok {
			data = append(data, &models.PostScore{
				ID:    mustParseInt64(postID),
				Score: int64(score),
			})
		}
	}
	return data, nil
}

// fetchPostVotesOnce 从 Redis 中获取一次帖子投票数据
func fetchPostVotesOnce(ctx context.Context) (data []*models.PostVoteData, err error) {
	// 获取所有的帖子ID
	// 1. 从 Redis 中获取所有帖子 ID
	postIDs, err := client.ZRange(ctx, getRedisKey(KeyPostTimeZSet), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post IDs from Redis: %w", err)
	}

	// 2. 使用 Pipeline 批量获取帖子投票数据
	pipeline := client.Pipeline()
	cmders := make(map[string]*redis.ZSliceCmd, len(postIDs)) // 存储每个命令的返回结果

	// 构造 Pipeline 命令
	for _, postID := range postIDs {
		key := getRedisKey(KeyPostVotedZSetPF + postID)
		cmders[postID] = pipeline.ZRangeWithScores(ctx, key, 0, -1) // 获取每个帖子所有的投票记录
	}
	// 执行 Pipeline
	_, err = pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("FetchPostVotes pipeline exec failed", zap.Error(err))
		return nil, err
	}

	// 3. 解析 Pipeline 返回结果
	data = make([]*models.PostVoteData, 0)
	for _, postID := range postIDs {
		// 检查每条命令的执行结果
		cmd := cmders[postID]
		if err := cmd.Err(); err != nil {
			zap.L().Warn("failed to fetch votes for post", zap.String("postID", postID), zap.Error(err))
			continue
		}

		// 获取投票记录
		votes, _ := cmd.Result()
		if len(votes) == 0 {
			// 如果没有投票记录，生成默认值
			//data = append(data, &models.PostVoteData{
			//	PostID: mustParseInt64(postID),
			//})
			continue
		}

		// 打印vote的值、数据类型
		for direction, userID := range votes {
			userID := userID.Member.(string)

			data = append(data, &models.PostVoteData{
				PostID:    mustParseInt64(postID),
				UserID:    mustParseInt64(userID),
				Direction: int32(direction),
			})
		}
	}
	return data, nil
}

// mustParseInt64 将字符串解析为 int64
func mustParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64) // 将字符串 s 解析为 int64，基数为 10
	return v
}
