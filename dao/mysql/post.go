package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

// CreatePost 创建一个新帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := "insert into Post(post_id, author_id,  community_id, title, content) values(?,?,?,?,?)" // status, create_time
	_, err = db.Exec(sqlStr, p.ID, p.AuthorID, p.CommunityID, p.Title, p.Content)
	if err != nil {
		return err
	}
	return
}

// GetPostByID 根据帖子ID查询指定帖子的详细信息
func GetPostByID(id int64) (data *models.Post, err error) {
	data = new(models.Post)
	sqlStr := `select post_id, author_id, community_id, title, content, create_time from post where post_id = ?`
	if err = db.Get(data, sqlStr, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Warn("there is no data in post")
			err = nil
		}
	}
	return
}

// GetPostList 获取帖子列表 帖子由新到旧排序
func GetPostList(page, size int64) (data []*models.Post, err error) {
	sqlStr := `select post_id, author_id, community_id, title, content, create_time 
	from post 
	ORDER BY create_time
	DESC   # 默认ASC
    limit ?,?`
	data = make([]*models.Post, 0, 2)
	err = db.Select(&data, sqlStr, (page-1)*size, size)
	return
}

// GetPostListByIDs 根据给定的ID列表查询帖子数据
func GetPostListByIDs(ids []string) (data []*models.Post, err error) {
	zap.L().Debug("GetPostListByIDs", zap.Strings("ids", ids))
	sqlStr := `select post_id, author_id, community_id, title, content, create_time
			   from post
			   where post_id in (?)
			   order by FIND_IN_SET(post_id, ?)`
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	// sqlx.In()会帮我们转义查询语句中的 ?
	query = db.Rebind(query)
	err = db.Select(&data, query, args...)
	return
}
