package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"

	"go.uber.org/zap"
)

// CreatePost 发帖
func CreatePost(p *models.Post) (err error) {
	// 1. 生成post id
	p.ID = snowflake.GenID()
	// 2. 保存到数据库
	if err = mysql.CreatePost(p); err != nil {
		zap.L().Error("mysql.CreatePost failed",
			zap.Any("post", p),
			zap.Error(err))
		return
	}
	if err = redis.CreatePost(p.ID, p.CommunityID); err != nil {
		zap.L().Error("redis.CreatePost failed",
			zap.Any("post", p),
			zap.Error(err))
	}
	// 3. 返回
	return
}

// GetPostByID 根据帖子ID查询帖子数据
func GetPostByID(id int64) (data *models.ApiPostDetail, err error) {
	// 查询并组合我们需要的数据
	postData, err := mysql.GetPostByID(id)
	if err != nil {
		zap.L().Error("mysql.GetPostByID failed",
			zap.Int64("id", id),
			zap.Error(err))
		return
	}
	// 根据用户ID查询用户信息
	user, err := mysql.GetUserByID(postData.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserByID failed",
			zap.Int64("author_id", postData.AuthorID),
			zap.Error(err))
		return
	}
	// 根据社区ID查询社区信息
	community, err := mysql.GetCommunityDetailByID(postData.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID failed",
			zap.Int64("community_id", postData.CommunityID),
			zap.Error(err))
		return
	}
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            postData,
		CommunityDetail: community,
	}
	return
}

// GetPostList 获取帖子列表
func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	// 查询并组合我们需要的数据
	postData, err := mysql.GetPostList(page, size)
	if err != nil {
		zap.L().Error("mysql.GetPostList failed",
			zap.Error(err))
	}
	data = make([]*models.ApiPostDetail, 0, len(postData))
	for _, post := range postData {
		// 根据用户ID查询用户信息
		user, err := mysql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区ID查询社区信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}

// GetPostList2 获取帖子列表2
func GetPostList2(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 1. 去 Redis 查询 ID 列表
	ids, err := redis.GetPostIDInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetPostIDInOrder failed", zap.Error(err))
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDInorder success, return 0 data.")
		return
	}
	// 2. 根据 ID 去 mysql 查询帖子详细信息
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		zap.L().Error("mysql.GetPostListByIDs failed", zap.Error(err))
		return
	}
	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Error(err))
		return
	}
	// 3. 根据用户 ID 查询用户信息
	data = make([]*models.ApiPostDetail, 0, len(posts))
	for idx, post := range posts {
		// 根据用户ID查询用户信息
		user, err := mysql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区ID查询社区信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}

// GetCommunityPostList 获取社区帖子列表
func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 1. 去 Redis 查询 ID 列表
	ids, err := redis.GetCommunityPostIDsInOrder(p)
	if err != nil {
		zap.L().Error("redis.GetCommunityPostIDsInOrder failed", zap.Error(err))
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetCommunityPostIDsInOrder success, return 0 data.")
		return
	}
	// 2. 根据 ID 去 mysql 查询帖子详细信息
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		zap.L().Error("mysql.GetPostListByIDs failed", zap.Error(err))
		return
	}
	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		zap.L().Error("redis.GetPostVoteData failed", zap.Error(err))
		return
	}
	// 3. 根据用户 ID 查询用户信息
	data = make([]*models.ApiPostDetail, 0, len(posts))
	for idx, post := range posts {
		// 根据用户ID查询用户信息
		user, err := mysql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区ID查询社区信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}

// GetPostListNew 获取帖子列表 New
func GetPostListNew(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	if p.CommunityID == 0 {
		// 查询所有社区的帖子
		data, err = GetPostList2(p)
	} else {
		// 查询指定社区的帖子
		data, err = GetCommunityPostList(p)
	}
	if err != nil {
		zap.L().Error("logic.GetPostListNew failed", zap.Error(err))
		return nil, err
	}
	return data, err
}
