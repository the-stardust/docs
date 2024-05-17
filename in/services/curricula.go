package services

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/garyburd/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"interview/dao"
	"interview/models"
	"interview/params"
	"interview/util"
	"strconv"
	"time"
)

const SuperAdminId = "c68vn2uh04rug2i4hafg"

type CurriculaSrv struct {
	ServicesBase
	RedisDao *dao.RedisDao
}

func NewCurriculaSrv() *CurriculaSrv {
	return &CurriculaSrv{
		ServicesBase: ServicesBase{},
		RedisDao:     dao.InitRedisDao(),
	}
}

func (c CurriculaSrv) CheckCurriculaAdminRedis(uid string, curriculaId string) bool {
	exists, _ := c.RedisDao.ExistsCurriculaAdminRedis(uid)
	// 管理员hash 存在   验证权限
	if exists {
		if curriculaId != "" {
			isHave, _ := c.RedisDao.HExistsCurriculaAdminRedis(uid, curriculaId)
			if isHave {
				return true
			} else {
				return false
			}
		} else {
			return true
		}

	} else {
		// 初始 超级账户
		if uid == SuperAdminId {
			return true
		}
	}
	return false
}

func (c CurriculaSrv) QueryCurriculaListService(param params.CurriculaListRequestParam) (params.
	CurriculaListResponseParam, error) {
	var res params.CurriculaListResponseParam
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}, "admin_list": bson.M{"$elemMatch": bson.M{"admin_id": param.
		UserId}}}
	if param.Id != "" {
		filter["_id"] = c.ObjectID(param.Id)
	}
	if param.Status != 0 {
		filter["status"] = param.Status
	}
	totalCount, err := c.DB().Collection(models.CurriculaTableName).Where(filter).Count()
	if err != nil {
		return res, err
	}
	if totalCount == 0 {
		res.Data = make([]params.CurriculaParam, 0)
		return res, nil
	}
	offset, limit := c.PageLimit(param.Page, param.PageSize)
	curriculaList := make([]params.CurriculaParam, 0)
	err = c.DB().Collection(models.CurriculaTableName).Where(filter).Skip(offset).
		Sort("-status", "-created_time").Limit(limit).Find(
		&curriculaList)
	if err != nil {
		return res, err
	}
	for i, curriculaParam := range curriculaList {
		for _, admin := range curriculaParam.AdminList {
			if admin.AdminId == param.UserId {
				curriculaList[i].Name = admin.Name
			}
		}
	}
	res.Total = totalCount
	res.Data = curriculaList
	return res, nil
}

func (c CurriculaSrv) QueryCurriculaSelectListService(param params.CurriculaListRequestParam) (params.
	CurriculaListResponseParam, error) {
	var res params.CurriculaListResponseParam
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}}
	if param.Id != "" {
		filter["_id"] = c.ObjectID(param.Id)
	}
	if param.Status != 0 {
		filter["status"] = param.Status
	} else {
		filter["status"] = 1
	}
	totalCount, err := c.DB().Collection(models.CurriculaTableName).Where(filter).Count()
	if err != nil {
		return res, err
	}
	if totalCount == 0 {
		res.Data = make([]params.CurriculaParam, 0)
		return res, nil
	}
	offset, limit := c.PageLimit(param.Page, param.PageSize)
	curriculaList := make([]params.CurriculaParam, 0)
	err = c.DB().Collection(models.CurriculaTableName).Fields(bson.M{"admin_list": 0}).Where(filter).Skip(offset).
		Sort("-created_time").Limit(limit).Find(
		&curriculaList)
	if err != nil {
		return res, err
	}
	res.Total = totalCount
	res.Data = curriculaList
	return res, nil
}

func (c CurriculaSrv) CreateCurriculaService(param params.CurriculaRequestParam) error {
	// 更新操作
	var model models.Curricula
	if param.Id != "" {
		err := c.DB().Collection(models.CurriculaTableName).Where(bson.M{"_id": c.ObjectID(param.Id),
			"is_delete": bson.M{"$ne": 1}}).Take(&model)
		if err != nil {
			return err
		}
		if param.CurriculaTitle != "" {
			model.CurriculaTitle = param.CurriculaTitle
		}
		if param.Sort != 0 {
			model.Sort = param.Sort
		}
		if param.IsDelete != 0 {
			model.IsDelete = param.IsDelete
		}
		if param.Status != 0 {
			model.Status = param.Status
		}
		if param.Name != "" {
			// 更新创建者名称
			if model.CreateUserId == param.CreateUserId {
				model.Name = param.Name
			}
			for index, listModel := range model.AdminList {
				if listModel.AdminId == param.CreateUserId {
					listModel.Name = param.Name
					model.AdminList[index] = listModel
				}
			}
		}
		model.UpdatedTime = util.GetNowFmtString()
		err = c.DB().Collection(models.CurriculaTableName).Save(&model)
		if err != nil {
			return err
		}
	} else {
		model.Sort = param.Sort
		model.CreateUserId = param.CreateUserId
		model.CurriculaTitle = param.CurriculaTitle
		model.Name = param.Name
		model.AdminList = append(model.AdminList, models.AdminListModel{
			AdminId:     param.CreateUserId,
			Name:        param.Name,
			Type:        0,
			CreatedTime: util.GetNowFmtString(),
		})
		model.Status = param.Status
		_, err := c.DB().Collection(models.CurriculaTableName).Create(&model)
		if err != nil {
			return err
		}
	}
	go func() {
		err := c.RedisDao.HSetCurriculaAdminRedis(model.CreateUserId, model.Id.Hex(), "1")
		if err != nil {
			return
		}
		// 设置 考试课程标题
		err = c.RedisDao.SetExCurriculaTitleRedis(model.Id.Hex(), model.CurriculaTitle, 60*15)
		if err != nil {
			return
		}
	}()
	return nil
}

func (c CurriculaSrv) CurriculaInviteCodeService(param params.CurriculaInviteCodeRequestParam) (params.
	CurriculaInviteCodeResponseParam, error) {
	var res params.CurriculaInviteCodeResponseParam
	b := []byte(param.Id)
	h := md5.New()
	h.Write(b)
	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	s := []byte(salt)
	h.Write(s)
	inviteCode := hex.EncodeToString(h.Sum(nil))
	res.InviteCode = inviteCode
	go func() {
		err := c.RedisDao.SetInviteCodeRedis(inviteCode, param.Id, 60*15)
		if err != nil {
			return
		}
	}()
	return res, nil

}

func (c CurriculaSrv) CurriculaInviteUseService(param params.CurriculaInviteCodeUseRequestParam) (params.CurriculaInviteCodeUseResponseParam, error) {
	var res params.CurriculaInviteCodeUseResponseParam
	id, err := c.RedisDao.GetInviteCodeRedis(param.InviteCode)
	if err != nil {
		if err == redis.ErrNil {
			res.Tips = "邀请码已使用或者已过期"
			return res, nil
		} else {
			return res, err
		}
	}
	var model models.Curricula
	err = c.DB().Collection(models.CurriculaTableName).Where(bson.M{"_id": c.ObjectID(id),
		"is_delete": bson.M{"$ne": 1}}).Take(&model)
	if err != nil {
		if c.MongoNoResult(err) {
			return res, errors.New("课程不存在")
		}
		return res, err
	}
	for _, admin := range model.AdminList {
		// 已经加入了  不再处理了
		if admin.AdminId == param.UserId {
			res.Tips = "已经加入该考试的管理员"
			return res, nil
		}
	}
	nowString := util.GetNowFmtString()
	model.AdminList = append(model.AdminList, models.AdminListModel{
		AdminId:     param.UserId,
		Type:        1,
		CreatedTime: nowString,
		Name:        param.Name,
	})
	model.UpdatedTime = nowString
	err = c.DB().Collection(models.CurriculaTableName).Save(&model)
	if err != nil {
		return res, err
	}
	go func() {
		_ = c.RedisDao.HSetCurriculaAdminRedis(param.UserId, model.Id.Hex(), "1")

		_ = c.RedisDao.DeleteInviteCodeRedis(param.InviteCode)

		var classes []models.InterviewClass
		filter := bson.M{"is_deleted": bson.M{"$ne": 1}, "curricula_id": model.Id.Hex()}
		err = c.DB().Collection("interview_class").Where(filter).Find(
			&classes)
		for _, class := range classes {
			isFind := false
			for index, member := range class.Members {
				if member.UserId == param.UserId {
					isFind = true
					if member.Status != 0 {
						class.Members[index].Status = 0
					}
					// 0415调整为班级管理员
					if member.IdentityType != 0 {
						class.Members[index].IdentityType = 0
					}
					class.Members[index].Name = param.Name
				}
			}
			if !isFind {
				// 0415调整为班级管理员
				class.Members = append(class.Members, models.Member{
					UserId:       param.UserId,
					Name:         param.Name,
					IdentityType: 0,
				})
			}
			classFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "_id": class.Id}
			_ = c.DB().Collection("interview_class").Where(classFilter).Save(&class)
		}
	}()
	return res, nil

}

func (c CurriculaSrv) CurriculaAdminUserDeleteService(param params.CurriculaAdminUserIdRequestParam) error {
	var modelList []models.Curricula
	filterQuery := bson.M{"is_delete": bson.M{"$ne": 1}, "admin_list": bson.M{"$elemMatch": bson.M{"admin_id": param.
		UserId}}}
	if param.Id != "" {
		filterQuery["_id"] = c.ObjectID(param.Id)
	}
	err := c.DB().Collection(models.CurriculaTableName).Where(filterQuery).Find(&modelList)
	if err != nil {
		return err
	}
	curriculaIdList := make([]string, 0)
	for _, model := range modelList {
		newAdminList := make([]models.AdminListModel, 0)
		for _, admin := range model.AdminList {
			if admin.AdminId != param.UserId {
				newAdminList = append(newAdminList, admin)
			}
		}
		curriculaIdList = append(curriculaIdList, model.Id.Hex())
		model.AdminList = newAdminList
		model.UpdatedTime = util.GetNowFmtString()
		err = c.DB().Collection(models.CurriculaTableName).Save(&model)
		if err != nil {
			return err
		}
	}
	if len(curriculaIdList) != 0 {
		go func() {
			err = c.RedisDao.HDelCurriculaAdminRedis(param.UserId, curriculaIdList)
			if err != nil {
				return
			}
		}()
	}
	return nil
}

func (c CurriculaSrv) CurriculaAdminUserListService(param params.CurriculaAdminListRequestParam) (params.CurriculaParam, error) {
	var res params.CurriculaParam
	filterQuery := bson.M{"is_delete": bson.M{"$ne": 1}, "_id": c.ObjectID(param.Id)}
	err := c.DB().Collection(models.CurriculaTableName).Where(filterQuery).Take(&res)
	if err != nil {
		return res, err
	}
	userIdList := make([]string, 0)
	for _, admin := range res.AdminList {
		userIdList = append(userIdList, admin.AdminId)
	}
	tempUser := new(User).GetGatewayUsersInfo(userIdList, param.AppCode, 1)
	for index, admin := range res.AdminList {
		res.AdminList[index].Avatar = tempUser[admin.AdminId].Avatar
	}
	return res, nil
}

func (c CurriculaSrv) AdminSaveInterviewQuestionService(param params.AdminSaveInterviewQuestionRequestParam) error {
	var class []models.InterviewClass
	classFilter := bson.M{"_id": bson.M{"$in": c.ObjectIDs(param.ClassIdList)}, "is_deleted": bson.M{"$ne": 1},
		"curricula_id": param.CurriculaId}
	err := c.DB().Collection("interview_class").Where(classFilter).Find(&class)
	if err != nil {
		c.SLogger().Error(err)
		return err
	}
	if len(class) != len(param.ClassIdList) {
		c.SLogger().Error(err)
		return err
	}
	questionIdNameMap := make(map[string]interface{}, 0)
	for _, classId := range param.ClassIdList {
		question := models.InterviewQuestion{}
		question.ClassId = classId
		question.Name = param.Name
		question.Desc = param.Desc
		question.CreatorUserId = param.UserId
		question.IsTeacherComment = param.IsTeacherComment
		question.IsStudentComment = param.IsStudentComment
		question.AnswerStyle = param.AnswerStyle
		question.Status = param.Status
		question.AllowAnswerUserIds = make([]string, 0)
		question.SourceType = 1
		_, err = c.DB().Collection("interview_questions").Create(&question)
		if err != nil {
			c.SLogger().Error(err)

			return err
		}
		questionIdNameMap[question.Id.Hex()] = question.Name
	}
	go func() {
		redisError := c.RedisDao.MHSetQuestionIdNameRedis(questionIdNameMap)
		if redisError != nil {
			return
		}
	}()
	return err
}

func (c CurriculaSrv) AdminClassListService(param params.AdminClassListRequestParam) (
	map[string]interface{}, error) {
	res := make(map[string]interface{}, 0)
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}, "curricula_id": param.CurriculaId}
	totalCount, err := c.DB().Collection("interview_class").Where(filter).Count()
	if err != nil {
		c.SLogger().Error(err)
		return res, err
	}
	if totalCount == 0 {
		res["total"] = totalCount
		res["list"] = make([]models.InterviewClass, 0)
	}
	offset, limit := c.PageLimit(param.Page, param.PageSize)
	var classes []models.InterviewClass
	err = c.DB().Collection("interview_class").Where(filter).Skip(offset).Sort("-created_time").Limit(limit).Find(
		&classes)
	if err != nil {
		c.SLogger().Error(err)
		return res, err
	}
	// 老师学生上传试题统计
	classCountMap := make(map[string]int64, 0)
	if param.StartTime != "" && param.EndTime != "" {
		classIdList := make([]string, 0)
		for _, class := range classes {
			classIdList = append(classIdList, class.Id.Hex())
		}
		if len(classIdList) != 0 {
			groupClassList := make([]params.GroupClassParam, 0)
			aggregateFilter := bson.A{
				bson.M{"$match": bson.M{"class_id": bson.M{"$in": classIdList}, "source_type": bson.M{"$ne": 1},
					"created_time": bson.M{"$gte": param.StartTime, "$lte": param.EndTime}}},
				bson.M{"$group": bson.M{"_id": "$class_id", "count": bson.M{"$sum": 1}}},
			}
			err = c.DB().Collection("interview_questions").Aggregate(aggregateFilter, &groupClassList)
			if err != nil {
				c.SLogger().Error(err)
				return nil, err
			}
			for _, classParam := range groupClassList {
				classCountMap[classParam.ClassId] = classParam.Count
			}
		}
	}
	list := make([]models.InterviewClassAndClassCount, 0)
	for _, class := range classes {
		count := classCountMap[class.Id.Hex()]
		list = append(list, models.InterviewClassAndClassCount{
			InterviewClass: class,
			QuestionCount:  count,
		})
	}
	res["list"] = list
	res["total"] = totalCount
	return res, nil
}
