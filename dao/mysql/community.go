package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

func GetCommunityList() (data []*models.Community, err error) {
	// 查询所有的社区（community_id, community_name)
	sqlStr := "select community_id, community_name from community"
	if err = db.Select(&data, sqlStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Warn("there is no data in community")
			err = nil
		}
	}
	return
}

// GetCommunityDetailByID 根据ID查询指定的社区详情
func GetCommunityDetailByID(id int64) (cd *models.CommunityDetail, err error) {
	sqlStr := "select community_id, community_name, introduction, create_time from community where community_id = ?"
	cd = new(models.CommunityDetail)
	if err := db.Get(cd, sqlStr, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Warn("there is no data in community")
			err = ErrorInvalidID
		}
	}
	return cd, err
}
