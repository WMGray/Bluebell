package redis

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

/*
	投票的几种情况

1. direction = 1, 有两种情况：
  - 之前没有投过票，现在投赞成票    --> 更新分数和投票记录   + 432 * 1
  - 之前投过反对票，现在改投赞成票   --> 更新分数和投票记录  + 432 * 2

2. direction = 0， 有两种情况：
  - 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录   - 432 * 1
  - 之前投过反对票，现在要取消投票  --> 更新分数和投票记录   + 432 * 1

3. direction = -1， 有两种情况：
  - 之前没有投过票，现在投反对票    --> 更新分数和投票记录   - 432 * 1
  - 之前投过赞成票，现在改投反对票  --> 更新分数和投票记录   - 432 * 2

投票的限制：
  - 每个帖子只能在自发表之日起一个星期之内投票，超过一个星期就不能投票了，禁止挖坟
  - 帖子到期后，将Redis中保存的赞成票票数以及反对票票数保存到mysql中
  - 到期之后，删除那个帖子的KeyPostVotedZSetPF
  - 同一个帖子，同一个用户只能投一次票（无论是赞成票还是反对票）
  - 如果用户之前投过赞成票，现在又要投反对票，应该取消之前的赞成票，只留下反对票
*/
const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 // 每一票的分数
)

var (
	ErrorVoteTimeExpire = errors.New("投票时间已过")
	ErrorVoteRepeat     = errors.New("不允许重复投票")
)

func VoteForPost(userID, postID string, direction float64) error {
	// 1. 判断投票限制
	// 获取帖子的发布时间
	postTime := client.ZScore(ctx, getRedisKey(KeyPostTimeZSet), postID).Val()
	fmt.Println("postTime: ", postTime)
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		// 超过一个星期 --> 不允许投票了
		return ErrorVoteTimeExpire
	}
	// 2 和 3 需要放到同一个事务中进行操作
	// 2. 更新分数
	// 查询当前用户给该帖子的投票记录
	ov := client.ZScore(ctx, getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()
	if ov == direction {
		// 这一次投票的值和上一次投票的一致，就提示不允许重复投票
		return ErrorVoteRepeat
	}
	var op float64
	if direction > ov {
		// 现在的值 > 之前的值
		op = 1
	} else {
		op = -1
	}
	diff := op * math.Abs(ov-direction) // 计算两次投票的差值

	pipline := client.TxPipeline() // 开启事务
	pipline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), diff*scorePerVote, postID)
	// 3. 记录用户为该帖子投票的记录
	if direction == 0 { // 移除投票记录
		client.ZRem(ctx, getRedisKey(KeyPostVotedZSetPF+postID), userID)
	} else {
		pipline.ZAdd(ctx, getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{
			Score:  direction, // 分数 -- 赞成/反对
			Member: userID,
		})
	}
	_, err := pipline.Exec(ctx)
	return err
}
