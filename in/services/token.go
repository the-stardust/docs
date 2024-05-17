package services

import (
	"encoding/json"
	"interview/common"
	"interview/common/global"
)

type UserToken struct {
	ServicesBase
}

// type UploadRes struct {
// 	Status string `json:"status"`
// 	Data   struct {
// 		Url string `json:"url"`
// 	} `json:"data"`
// }

func (sf *UserToken) Token2User(token string, retryTimes int) string {
	//获取 用户头像昵称
	res, err := common.HttpGet(global.CONFIG.ServiceUrls.TokenUrl + "/get_guid?token=" + token)
	if err == nil {
		type UserRes struct {
			UserId string `json:"guid"`
		}
		r := UserRes{}
		err = json.Unmarshal(res, &r)
		if err == nil {
			return r.UserId
		} else {
			return ""
		}
	} else {
		sf.SLogger().Error(err)
		if retryTimes > 0 {
			return sf.Token2User(token, retryTimes-1)
		}
	}
	return ""
}
