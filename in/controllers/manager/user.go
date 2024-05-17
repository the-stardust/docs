package manager

import (
	"interview/common"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/helper"
	"interview/models"
	"interview/models/managerresp"
	"interview/services"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	controllers.Controller
}

// UserFeedbackList 用户问题反馈列表
func (sf *User) UserFeedbackList(c *gin.Context) {
	var err error
	var param struct {
		SourceType   string `json:"source_type"`
		KeyWords     string `json:"keywords"`
		PageIndex    int64  `json:"page_index"`
		PageSize     int64  `json:"page_size"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		FeedbackType string `json:"feedback_type"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	filter := bson.M{}
	if param.SourceType != "" {
		filter["source_type"] = param.SourceType
	}
	if param.StartTime != "" && param.EndTime != "" {
		filter["updated_time"] = bson.M{"$gte": param.StartTime, "$lte": param.EndTime}
	}

	if param.FeedbackType != "" {
		filter["fast_remark"] = bson.M{"$in": bson.A{param.FeedbackType}}
	}
	if param.KeyWords != "" {
		filter["$or"] = bson.A{bson.M{"remark": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.KeyWords)}}},
			bson.M{"data_info.url": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.KeyWords)}}},
			bson.M{"_id": sf.ObjectID(param.KeyWords)},
			bson.M{"user_id": param.KeyWords},
			bson.M{"fast_remark": param.KeyWords},
		}
	}

	var feedbacks []models.UserFeedback
	err = sf.DB().Collection("user_feedback").Where(filter).Skip(offset).Sort("-created_time").Limit(limit).Find(&feedbacks)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	count, err := sf.DB().Collection("user_feedback").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var feedbackResp []managerresp.FeedBackResp
	for _, i := range feedbacks {
		feedbackResp = append(feedbackResp, managerresp.FeedBackResp{Id: i.Id.Hex(),
			CreatedTime: i.CreatedTime,
			UserId:      i.UserId,
			FastRemark:  i.FastRemark,
			Remark:      i.Remark,
			SourceType:  i.SourceType,
		})
	}

	sf.Success(managerresp.FeedBackList{List: feedbackResp, TotalCount: count}, c)
}

// UserFeedbackDetail 问题反馈详情信息
func (sf *User) UserFeedbackDetail(c *gin.Context) {
	var err error
	id := c.Query("feedback_id")
	if id == "" {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var feedback models.UserFeedback
	err = sf.DB().Collection("user_feedback").Where(bson.M{"_id": sf.ObjectID(id)}).Take(&feedback)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	userMobileMap := new(services.User).GetMobileInfoFromMysql([]string{feedback.UserId})
	feedback.NickName = userMobileMap[feedback.UserId].NickName
	feedback.Avatar = userMobileMap[feedback.UserId].Avatar
	feedback.MobileMask = userMobileMap[feedback.UserId].MobileMask
	feedback.MobileID = userMobileMap[feedback.UserId].MobileID

	sf.Success(feedback, c)
}

func (sf *User) SaveIsShowClassMembers(c *gin.Context) {
	var param struct {
		Uid    string `json:"uid"`
		Action string `json:"action"`
	}
	err := c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.Action != "add" && param.Action != "delete" {
		sf.Error(common.CodeInvalidParam, c, "action传值有误，只能是新增或者删除！")
		return
	}

	if param.Action == "add" {
		var addSlice = []interface{}{rediskey.InterviewAdminAuthUserIds, param.Uid}
		err = helper.RedisSADD(addSlice...)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		sf.Success(nil, c, "添加成功！")
	} else {
		err = helper.RedisSREM(string(rediskey.InterviewAdminAuthUserIds), param.Uid)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		sf.Success(nil, c, "删除成功！")
	}
}

func (sf *User) GetIsShowClassMembers(c *gin.Context) {
	uids, err := helper.RedisSMembers(string(rediskey.InterviewAdminAuthUserIds))
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	// 增加头像和昵称
	var list = []map[string]interface{}{}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo(uids, appCode, 1)
	for _, i := range userMap {
		tempMap := make(map[string]interface{})
		tempMap["id"] = i.UserId
		tempMap["user_name"] = i.Nickname
		tempMap["avatar"] = i.Avatar
		list = append(list, tempMap)
	}

	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = len(uids)
	sf.Success(resultInfo, c)
}

func (sf *User) GPTCountInfo(c *gin.Context) {
	var param struct {
		Uid   string `json:"uid"`  // 多个uid用逗号隔开
		Type  int    `json:"type"` // 1 增加表 2不限次数
		Count int    `json:"count"`
	}
	err := c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uids := strings.Split(param.Uid, ",")
	if param.Type == 1 {
		var gptcountinfos = make([]models.GPTCountInfo, 0)
		err = sf.DB().Collection("gpt_count_info").Where(bson.M{"user_id": bson.M{"$in": uids}}).Find(&gptcountinfos)

		var existsUid = make([]string, 0)
		for _, item := range gptcountinfos {
			existsUid = append(existsUid, item.UserID)
			// 增加
			item.AvailableCount += param.Count
			item.TotalCount += param.Count
			sf.DB().Collection("gpt_count_info").Where(bson.M{"_id": item.Id}).Update(&item)
		}
		// 增加不存在uid的
		for _, i2 := range uids {
			if common.InArrCommon[string](i2, existsUid) {
				continue
			}

			var gptcountinfo = new(models.GPTCountInfo)
			gptcountinfo.UserID = i2
			gptcountinfo.AvailableCount = 10 + param.Count
			gptcountinfo.TotalCount = 10 + param.Count
			sf.DB().Collection("gpt_count_info").Create(gptcountinfo)
		}
	} else if param.Type == 2 {
		for _, uid := range uids {
			if !helper.RedisSIsMember("interview:temporary_uids", uid) {
				helper.RedisSADD("interview:temporary_uids", uid)
			}
		}
	}

	sf.Success("success", c)
}
