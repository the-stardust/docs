package dao

import (
	"fmt"
	"interview/common/rediskey"
)

type RedisDao struct {
	baseDao
}

func InitRedisDao() *RedisDao {
	return &RedisDao{}
}

// set 必须设定存在时间
func (r RedisDao) SetExCurriculaTitleRedis(id, curriculaTitle string, expireTime int64) error {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaTitleString), id)
	err := r.setExRedisDao(key, curriculaTitle, expireTime)
	return err
}

// 邀请码
func (r RedisDao) SetInviteCodeRedis(code, id string, expireTime int64) error {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaCodeString), code)
	err := r.setExRedisDao(key, id, expireTime)
	return err
}

// get
func (r RedisDao) GetCurriculaTitleRedis(id string) (string, error) {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaTitleString), id)
	value, err := r.getRedisDao(key)
	return value, err
}

// get邀请码
func (r RedisDao) GetInviteCodeRedis(code string) (string, error) {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaCodeString), code)
	value, err := r.getRedisDao(key)
	return value, err
}

// Exists 超级用户权限
func (r RedisDao) ExistsCurriculaAdminRedis(id string) (bool, error) {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), id)
	flag, err := r.existsRedisDao(key)
	return flag, err
}

// HKeys 超级用户权限
func (r RedisDao) HKeysCurriculaAdminRedis(id string) ([]string, error) {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), id)
	values, err := r.hKeysRedisDao(key)
	return values, err
}

// 删除邀请码
func (r RedisDao) DeleteInviteCodeRedis(code string) error {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaCodeString), code)
	err := r.deleteRedisDao(key)
	return err
}

// map 是否存在权限
func (r RedisDao) HExistsCurriculaAdminRedis(userId, curriculaId string) (bool, error) {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), userId)
	return r.hExistsRedisDao(key, curriculaId)
}

// 设置map key value
func (r RedisDao) HSetCurriculaAdminRedis(userId, curriculaId string, value interface{}) error {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), userId)
	return r.hSetRedisDao(key, curriculaId, value)
}

// 批量设置map key value
func (r RedisDao) MHSetQuestionIdNameRedis(questionIdMap map[string]interface{}) error {
	key := rediskey.InterviewQuestionId2Name
	return r.mHSetRedisDao(string(key), questionIdMap)
}

// map 删除 key value
func (r RedisDao) HDelCurriculaAdminRedis(userId string, curriculaIdList []string) error {
	key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), userId)
	return r.HDelRedisDao(key, curriculaIdList)
}
