package mysql

import (
	"bluebell/models"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"time"
)

// PersistPost 持久化帖子数据，包括帖子分数和投票数据，在一个事务中执行
func PersistPost(postScores []*models.PostScore, postVotes []*models.PostVoteData, retryCount, timeout int) (err error) {
	for attempt := 1; attempt <= retryCount; attempt++ {
		// 设置函数超时时间
		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		err = PersistPostOnce(postScores, postVotes)
		cancel() // 调用 cancel 释放资源
		if err == nil {
			return // 如果成功，直接返回结果
		}

		// 检查是否是超时错误
		if errors.Is(timeoutctx.Err(), context.DeadlineExceeded) {
			zap.L().Warn("PersistPost timeout, retrying...",
				zap.Int("attempt", attempt),
				zap.Int("retryCount", retryCount),
				zap.Error(timeoutctx.Err()),
			)
		} else {
			zap.L().Warn("PersistPost failed, retrying...",
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
	return
}

// PersistPostOnce 持久化帖子数据，包括帖子分数和投票数据，在一个事务中执行
func PersistPostOnce(postScores []*models.PostScore, postVotes []*models.PostVoteData) error {
	// 开始一个新的事务
	tx, err := db.Beginx()
	if err != nil {
		zap.L().Error("failed to begin transaction", zap.Error(err))
		return err
	}

	// 1. 将帖子分数数据持久化到 MySQL 中
	fmt.Printf("%+v\n", postScores)
	if len(postScores) > 0 {
		_, err = tx.NamedExec(`
        INSERT INTO post_scores (post_id, score)
        VALUES (:post_id, :score)
        ON DUPLICATE KEY UPDATE score = VALUES(score)`, postScores)
		if err != nil {
			zap.L().Error("failed to insert post scores", zap.Error(err))
			_ = tx.Rollback() // 回滚事务
			return err
		}
	}
	// 2. 将帖子投票数据持久化到 MySQL 中
	if len(postVotes) > 0 {
		_, err = tx.NamedExec(`
        INSERT INTO post_votes (post_id, user_id, direction)
        VALUES (:post_id, :user_id, :direction)
        ON DUPLICATE KEY UPDATE direction = VALUES(direction)`, postVotes)
		if err != nil {
			zap.L().Error("failed to insert post votes", zap.Error(err))
			_ = tx.Rollback() // 回滚事务
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		zap.L().Error("failed to commit transaction", zap.Error(err))
		return err
	}

	zap.L().Info("successfully persisted post scores and post votes")
	return nil
}
