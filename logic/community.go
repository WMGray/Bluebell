package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
)

// GetCommunityList 查询所有的社区（community_id, community_name）列表
func GetCommunityList() (data []*models.Community, err error) {
	// 查询所有的社区（community_id, community_name）列表
	return mysql.GetCommunityList()
}

func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	return mysql.GetCommunityDetailByID(id)
}
