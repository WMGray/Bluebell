package redis

import (
	"bluebell/models"
	"context"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetPostVoteData 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	//data = make([]int64, 0, len(ids))
	//for _, id := range ids {
	//	key := getRedisKey(KeyPostVotedZSetPF + id)
	//	v := client.ZCount(ctx, key, "1", "1").Val()
	//	data = append(data, v)
	//}
	// 使用pipeline一次发送多条命令，减少RTT
	pipeline := client.Pipeline()
	data = make([]int64, 0, len(ids))
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		pipeline.ZCount(ctx, key, "1", "1")
	}
	cmders, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

// getIDsFormKey 按照分数从大到小的顺序查询指定数量的元素
func getIDsFormKey(key string, page, size int64) ([]string, error) {
	// 确定要查询的起始点和终止点
	start := (page - 1) * size
	end := start + size - 1
	// ZRevRange 按照分数从大到小的顺序查询指定数量的元素
	return client.ZRevRange(ctx, key, start, end).Result()
}

// GetPostIDInOrder 根据给定的orderType获取帖子ID
func GetPostIDInOrder(p *models.ParamPostList) ([]string, error) {
	// 1. 根据用户请求中携带的order参数确定要查询的redis key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}

	return getIDsFormKey(key, p.Page, p.Size)
}

// GetCommunityPostIDsInOrder 根据社区ID和给定的orderType获取帖子ID
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 1.根据用户请求中携带的order参数确定要查询的redis key
	orderkey := KeyPostTimeZSet       // 默认是时间
	if p.Order == models.OrderScore { // 按照分数请求
		orderkey = KeyPostScoreZSet
	}

	// 使用zinterstore 把分区的帖子set与帖子分数的zset生成一个新的zset
	// 针对新的zset 按之前的逻辑取数据

	// 社区的key
	cKey := KeyCommunitySetPF + strconv.Itoa(int(p.CommunityID))

	// 利用缓存key减少zinterstore执行的次数 缓存key
	key := orderkey + strconv.Itoa(int(p.CommunityID))
	if client.Exists(ctx, key).Val() < 1 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(ctx, key, &redis.ZStore{
			Keys:      []string{cKey, orderkey}, // 两个key的交集
			Aggregate: "MAX",                    // 将两个zset函数聚合的时候 求最大值
		}) // zinterstore 计算
		pipeline.Expire(ctx, key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec(ctx)
		if err != nil {
			return nil, err
		}
	}
	// 存在的就直接根据key查询ids
	return getIDsFormKey(key, p.Page, p.Size)
}

// CreatePost 创建帖子
func CreatePost(postID, communityID int64) error {
	pipeline := client.TxPipeline() // 获取一个事务
	// 帖子时间
	pipeline.ZAdd(ctx, getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 帖子分数
	pipeline.ZAdd(ctx, getRedisKey(KeyPostScoreZSet), redis.Z{
		Score:  0,
		Member: postID,
	})
	// 把帖子id加到社区的set
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(ctx, cKey, postID)
	// 提交事务
	_, err := pipeline.Exec(ctx)
	return err
}

// GetPostIDsByTimeRange  获取指定时间范围内的帖子id
func GetPostIDsByTimeRange(ctx context.Context, expiredDays int) ([]string, error) {
	// 计算过期时间的阈值
	now := time.Now().Unix()
	expiredThreshold := now - int64(expiredDays*24*60*60) // expiredDays 表示保留天数

	// 从 KeyPostTimeZSet 中获取帖子ID
	postTimes, err := client.ZRangeByScoreWithScores(ctx, getRedisKey(KeyPostTimeZSet),
		&redis.ZRangeBy{
			Min: fmt.Sprintf("%d", expiredThreshold), // 只获取未过期的帖子
			Max: "+inf",                              // 获取未来时间的帖子
		}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post times: %w", err)
	}

	// 提取帖子ID列表
	postIDs := make([]string, 0, len(postTimes))
	for _, postTime := range postTimes {
		postIDs = append(postIDs, postTime.Member.(string)) // 提取帖子ID
	}
	return postIDs, nil
}

// GetPostScoreByIDs 获取帖子的分数
//func GetPostScoreByIDs(ids []string) (data []int64, err error) {
//	// 从 KeyPostScoreZSet 中获取帖子分数
//	scores, err := client.ZMScore(ctx, getRedisKey(KeyPostScoreZSet), ids...).Result()
//	if err != nil {
//		return nil, err
//	}
//
//	// 组装数据
//	data = make([]int64, 0, len(ids))
//	for _, score := range scores {
//		data = append(data, int64(score))
//	}
//	return
//}

// GetPostScoreByIDs 获取帖子的分数
func GetPostScoreByIDs(ctx context.Context, postIDs []string) (map[string]float64, error) {
	pipeline := client.Pipeline()
	cmds := make(map[string]*redis.FloatCmd, len(postIDs))

	for _, postID := range postIDs {
		key := getRedisKey(KeyPostScoreZSet)
		cmds[postID] = pipeline.ZScore(ctx, key, postID) // 按帖子 ID 获取分数
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}

	scores := make(map[string]float64, len(postIDs))
	for postID, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			zap.L().Warn("failed to fetch post score", zap.String("postID", postID), zap.Error(err))
			continue
		}
		scores[postID] = cmd.Val()
	}
	return scores, nil
}
