package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"
	"strconv"

	"go.uber.org/zap"
)

// 投票功能
// 基于用户投票的相关算法： 阮一峰投票算法

// 本项目使用简化版的投票分数
// 投一票加 432 分，取消投票减 432 分  86400(s) / 200 = 432 --> 200张赞成票可以给你的帖子续一天

/* 投票的几种情况
1. direction = 1, 有两种情况：
	- 之前没有投过票，现在投赞成票    --> 更新分数和投票记录
	- 之前投过反对票，现在改投赞成票   --> 更新分数和投票记录
2. direction = 0， 有两种情况：
	- 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录
	- 之前投过反对票，现在要取消投票  --> 更新分数和投票记录
3. direction = -1， 有两种情况：
	- 之前没有投过票，现在投反对票    --> 更新分数和投票记录
	- 之前投过赞成票，现在改投反对票  --> 更新分数和投票记录

投票的限制：
	- 每个帖子只能在自发表之日起一个星期之内投票，超过一个星期就不能投票了，禁止挖坟
	- 帖子到期后，将Redis中保存的赞成票票数以及反对票票数保存到mysql中
	- 到期之后，删除那个帖子的KeyPostVotedZSetPF
	- 同一个帖子，同一个用户只能投一次票（无论是赞成票还是反对票）
	- 如果用户之前投过赞成票，现在又要投反对票，应该取消之前的赞成票，只留下反对票
*/

// VoteForPost 为帖子投票
func VoteForPost(userID int64, p *models.ParamVoteData) (err error) {
	zap.L().Debug("logic.VoteForPost: ",
		zap.Int64("userID", userID),
		zap.String("postID", p.PostID),
		zap.Int8("direction", p.Direction),
		zap.Error(err))
	return redis.VoteForPost(strconv.Itoa(int(userID)), p.PostID, float64(p.Direction))
}
