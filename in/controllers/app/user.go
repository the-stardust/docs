package app

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/helper"
	"interview/models"
	"interview/services"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type User struct {
	controllers.Controller
}

func (sf *User) UserFeedback(c *gin.Context) {
	var err error
	var param struct {
		DataInfo   models.DataInfo `json:"data_info"`
		FastRemark []string        `json:"fast_remark"`
		Remark     string          `json:"remark"`
		ImageList  []string        `json:"image_list"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.Remark == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if len(param.ImageList) > 4 {
		sf.Error(common.CodeInvalidParam, c, "最多允许上传4张图片!")
		return
	} else if len(param.ImageList) == 0 {
		param.ImageList = []string{}
	}
	uid := c.GetHeader("X-XTJ-UID")
	feedback := models.UserFeedback{
		UserId:     uid,
		FastRemark: param.FastRemark,
		Remark:     param.Remark,
		DataInfo:   param.DataInfo,
		SourceType: strconv.Itoa(sf.GetAppSourceType(c)),
		ImageList:  param.ImageList,
	}
	_, err = sf.DB().Collection("user_feedback").Create(&feedback)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

// SaveUserChoiceStatus 保存用户选择状态（是否勾选用户须知等）
func (sf *User) SaveUserChoiceStatus(c *gin.Context) {
	paramMap := make(map[string]interface{})
	//var param struct {
	//	UseNotice    bool `json:"use_notice" `
	//	UseTip       bool `json:"use_tip" `
	//	PracticeMode int8 `json:"practice_mode"`
	//}
	err := c.ShouldBindJSON(&paramMap)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	uid := c.GetHeader("X-XTJ-UID")
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	keyName := fmt.Sprintf("%s%s", rediskey.UserChoice, uid)
	// values := []interface{}{keyName, "use_notice", param.UseNotice, "use_tip", param.UseTip, "practice_mode", param.PracticeMode}
	values := []interface{}{keyName}
	for k, v := range paramMap {
		if boolV, ok := v.(bool); ok {
			values = append(values, []interface{}{k, boolV}...)
		} else if stringV, ok := v.(string); ok {
			values = append(values, []interface{}{k, stringV}...)
		} else {
			values = append(values, []interface{}{k, int8(v.(float64))}...)
		}
	}
	_, err = rdb.Do("HMSET", values...)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

// GetUserChoiceStatus 获取用户选择状态（是否勾选用户须知等）
func (sf *User) GetUserChoiceStatus(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()

	keyName := fmt.Sprintf("%s%s", rediskey.UserChoice, uid)
	res, err := redis.Values(rdb.Do("HGETALL", keyName))
	if err == nil {
		userChoiceStatus := models.UserChoiceStatus{}
		err = redis.ScanStruct(res, &userChoiceStatus)
		if err == nil {
			if userChoiceStatus.ExamCategory == "" {
				userChoiceStatus.ExamCategory = "事业单位"
			}
			respM := common.StructToMap(&userChoiceStatus)
			//respM := make(map[string]any)
			//respM["exam_child_category"] = userChoiceStatus.ExamChildCategory
			//respM["province"] = userChoiceStatus.Province
			//respM["city"] = userChoiceStatus.City
			//respM["exam_category"] = userChoiceStatus.ExamCategory
			//respM["district"] = userChoiceStatus.District
			//respM["use_notice"] = userChoiceStatus.UseNotice
			//respM["use_tip"] = userChoiceStatus.UseTip
			//respM["practice_mode"] = userChoiceStatus.PracticeMode
			//respM["job_tag"] = userChoiceStatus.JobTag
			//respM["question_real_info"] = userChoiceStatus.QuestionRealInfo
			//respM["question_answer_after_tip"] = userChoiceStatus.QuestionAnswerAfterTip
			//respM["guid_tips"] = userChoiceStatus.GuidTips
			//respM["neimeng_prediction"] = userChoiceStatus.NeimengPrediction

			m, _ := helper.RedisHGetAll(keyName)
			for s, s2 := range m {
				if strings.Contains(s, "activity_") {
					respM[s] = s2
				}
			}
			sf.Success(respM, c)
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
		}
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
	}
}

func (sf *User) JSCode2session(c *gin.Context) {
	var param struct {
		Code string `json:"code" `
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	res, err := services.GetOpenIdByCode(param.Code)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	sf.Success(res, c)
}

// IsShowClass 当前用户是否显示班级
func (sf *User) IsShowClass(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	if uid == "" {
		sf.Error(common.CodeServerBusy, c, "当前用户id为空！")
		return
	}
	IsCanSeeAllClass := false
	isShow := helper.RedisSIsMember(string(rediskey.InterviewAdminAuthUserIds), uid)
	if !isShow {
		filter := bson.M{"is_deleted": 0, "members": bson.M{"$elemMatch": bson.M{"user_id": uid, "status": 0}}}
		count, err := sf.DB().Collection("interview_class").Where(filter).Count()
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		if count > 0 {
			isShow = true
		}
	} else {
		IsCanSeeAllClass = true
	}
	sf.Success(map[string]bool{"is_show": isShow, "is_can_see_all_class": IsCanSeeAllClass}, c)
}

// WechatGroup
func (sf *User) WechatGroup(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.WechatStudyGroup))
	if str == "" {
		sf.Success([][]string{
			{"2024国考面试提前学", "https://res.xtjzx.cn/2311704964867.pic_17a9409f41649865.jpg"},
		}, c)
		return
	}
	nowStr := time.Now().Format("20060102")
	resp := make([][]string, 0)
	resp2 := make([][]string, 0)
	json.Unmarshal([]byte(str), &resp)
	for _, stringss := range resp {
		if stringss[2] == "-1" || stringss[2] > nowStr {
			resp2 = append(resp2, stringss)
		}
	}
	sf.Success(resp2, c)
}
