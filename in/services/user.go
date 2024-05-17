package services

import (
	"encoding/json"
	"errors"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/models"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
)

type User struct {
	ServicesBase
}

func NewUser() *User {
	return &User{}
}

type CacheUser struct {
	UserId     string `redis:"user_id"`
	Mobile     string `redis:"mobile"`
	Nickname   string `redis:"nickname"`
	Avatar     string `redis:"avatar"`
	MobileMask string `redis:"mobile_mask"`
	MobileID   string `redis:"mobile_id"`
	Province   string `redis:"province"`
	City       string `redis:"city"`
}

func (sf *User) getManyRedisUserInfo(uids []string) map[string]CacheUser {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	resp := map[string]CacheUser{}
	for _, uid := range uids {
		rdb.Send("HGETALL", rediskey.UserInfo+rediskey.RedisKey(uid))
	}
	r, err := redis.Values(rdb.Do(""))
	if err == nil {
		for _, user := range r {
			r, err := redis.Values(user, nil)
			if err == nil {
				user := CacheUser{}
				err = redis.ScanStruct(r, &user)
				if err == nil {
					resp[user.UserId] = user
				} else {
					sf.SLogger().Error(err)
					return resp
				}
			} else {
				sf.SLogger().Error(err)
				return resp
			}

		}
	} else {
		sf.SLogger().Error(err)
		return resp
	}
	return resp
}

type CacheUserInfo struct {
	UserId       string
	FieldsValues []interface{}
}

func (sf *User) setManyRedisUserInfo(users []CacheUserInfo) error {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	for _, v := range users {
		key := rediskey.UserInfo + rediskey.RedisKey(v.UserId)
		fieldsValues := []interface{}{key}
		fieldsValues = append(fieldsValues, v.FieldsValues...)
		rdb.Send("HMSET", fieldsValues...)
		rand.Seed(time.Now().UnixNano())
		random := rand.Intn(100000)
		rdb.Send("expire", key, 86400*30+random)
	}
	_, err = rdb.Do("")
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	return nil
}

// GetMobileInfoFromMysql 获取脱敏手机号信息，先尝试从redis中获取，如果没有，再从mysql中查询并缓存
func (sf *User) GetMobileInfoFromMysql(userIDs []string) map[string]models.MobileInfo {
	var userInfoMap = map[string]models.MobileInfo{}
	loopTimes := len(userIDs) / 1000
	remain := len(userIDs) % 1000
	if remain > 0 {
		loopTimes++
	}
	for i := 1; i <= loopTimes; i++ {
		var tempUids []string
		if i == loopTimes {
			tempUids = userIDs[((i - 1) * 1000) : ((i-1)*1000)+(len(userIDs)-((i-1)*1000))]
		} else {
			tempUids = userIDs[((i - 1) * 1000) : ((i-1)*1000)+1000]
		}
		ids := []string{}
		for _, v := range tempUids {
			if _, ok := userInfoMap[v]; !ok {
				ids = append(ids, v)
			}
		}
		r := sf.getMobileInfoFromMysql(ids)
		for k, v := range r {
			userInfoMap[k] = v
		}
	}
	return userInfoMap
}

func (sf *User) getMobileInfoFromMysql(userIDs []string) map[string]models.MobileInfo {
	var userInfoMap = map[string]models.MobileInfo{}
	var uids = []string{}
	userMap := sf.getManyRedisUserInfo(userIDs)
	for _, v := range userIDs {
		if user, ok := userMap[v]; ok && user.MobileMask != "" {
			userInfoMap[v] = models.MobileInfo{
				MobileID:   user.MobileID,
				MobileMask: user.MobileMask,
				Province:   user.Province,
				City:       user.City,
				NickName:   user.Nickname,
				Avatar:     user.Avatar,
			}
			continue
		}
		uids = append(uids, v)
	}

	if len(uids) > 0 {
		var mobileInfos []models.MobileInfo
		sf.Mysql().Table("xtj_mobile").Select("xtj_mobile.mobile_id,xtj_mobile.mobile_mask, xtj_user_info.province, xtj_user_info.city,  xtj_user_info.nick_name, xtj_user_info.avatar,xtj_user_info.guid").
			Joins("INNER JOIN xtj_user_info ON xtj_mobile.guid = xtj_user_info.guid").
			Where("xtj_mobile.guid IN (?)", userIDs).
			Scan(&mobileInfos)
		users := []CacheUserInfo{}
		for _, v := range mobileInfos {
			userInfoMap[v.GUID] = models.MobileInfo{
				MobileID:   v.MobileID,
				MobileMask: v.MobileMask,
				Province:   v.Province,
				City:       v.City,
				NickName:   v.NickName,
				Avatar:     v.Avatar,
			}
			users = append(users, CacheUserInfo{UserId: v.GUID, FieldsValues: []interface{}{"mobile_mask", v.MobileMask, "mobile_id", v.MobileID, "user_id", v.GUID, "province", v.Province, "city", v.City, "nickname", v.NickName, "avatar", v.Avatar}})
		}
		sf.setManyRedisUserInfo(users)
	}
	return userInfoMap
}

func (sf *User) GetGatewayUsersInfo(userIds []string, appCode string, retryTimes int) map[string]models.GatewayUserInfo {
	//获取 用户头像昵称
	var userInfoMap = map[string]models.GatewayUserInfo{}
	uids := userIds

	if len(uids) > 0 {
		type GatewayParam struct {
			Ids     []string `json:"ids"`
			AppCode string   `json:"appCode"`
		}
		pm := GatewayParam{
			Ids:     uids,
			AppCode: appCode}
		res, err := common.HttpPostJson(global.CONFIG.ServiceUrls.UserLoginUrl+"/new-userinfo/users-info", pm)
		type TempRes struct {
			Msg  string                            `json:"msg"`
			Code string                            `json:"code"`
			Data map[string]models.GatewayUserInfo `json:"data"`
		}
		if err == nil {
			r := TempRes{}
			err = json.Unmarshal(res, &r)
			if err != nil {
				sf.SLogger().Error(err)
			} else {
				for _, v := range r.Data {
					userInfoMap[v.Guid] = v
				}
			}
		} else {
			sf.SLogger().Error(err)
			if retryTimes > 0 {
				return sf.GetGatewayUsersInfo(uids, appCode, retryTimes-1)
			}
		}

	}
	return userInfoMap
}

type GetUserMapAbs struct {
	GUuid    string `json:"g_uuid" gorm:"g_uuid"`
	RealName string `json:"real_name" gorm:"real_name"`
}

func (this *User) GetUserMap(userIdList []string) (map[string]GetUserMapAbs, error) {
	req := map[string]interface{}{
		"user_id": userIdList,
	}
	reqUrl := global.CONFIG.ServiceUrls.CrmDataServiceUrl + "/question-bank-crm-data/inner/fcUser/getUserMap"
	respData, err := common.HttpPostJson(reqUrl, req)
	if err != nil {
		return nil, err
	}

	type DataRespV2 struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	result := make(map[string]GetUserMapAbs)
	resp := &DataRespV2{Data: &result}
	_ = json.Unmarshal(respData, resp)
	if resp.Code > 0 {
		return nil, errors.New(resp.Message)
	}

	return result, nil
}
