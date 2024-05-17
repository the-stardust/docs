package app

import (
	"context"
	"errors"
	"fmt"
	"interview/common"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/es"
	"interview/helper"
	"interview/models"
	"interview/services"
	"interview/util"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
)

type Interview struct {
	controllers.Controller
}

// 面试班列表
func (sf *Interview) ClassList(c *gin.Context) {
	var err error
	var param struct {
		PageIndex int64 `json:"page_index"`
		PageSize  int64 `json:"page_size"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}}
	filter["$or"] = bson.A{bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid, "status": 0, "identity_type": bson.M{"$lt": 2}}}}, bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid, "status": 0}}, "status": 5}}
	totalCount, err := sf.DB().Collection("interview_class").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	var classes = []models.InterviewClass{}
	err = sf.DB().Collection("interview_class").Where(filter).Fields(bson.M{"members": 0}).Skip(offset).Sort("-created_time").Limit(limit).Find(&classes)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resp := map[string]interface{}{"total_count": totalCount, "list": classes}
	sf.Success(resp, c)
}

func (sf *Interview) MGetCurriculaTitleRedis(idList []string) ([]string, error) {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	keyList := make([]interface{}, 0)
	for _, id := range idList {
		key := fmt.Sprintf(string(rediskey.InterviewCurriculaTitleString), id)
		keyList = append(keyList, key)
	}
	valueList, err := redis.Strings(rdb.Do("MGET", keyList...))
	return valueList, err
}

func (sf *Interview) MSetCurriculaTitleRedis(idNameMap map[string]string) error {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	for id, name := range idNameMap {
		key := fmt.Sprintf(string(rediskey.InterviewCurriculaTitleString), id)
		_, err := rdb.Do("SETEX", key, 60*15, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sf *Interview) MGetCurriculaTitle(idList []string) (map[string]string, error) {
	curriculaTitleMap := make(map[string]string, 0)
	if len(idList) != 0 {
		newIdList := util.SetSliceList(idList)
		values, err := sf.MGetCurriculaTitleRedis(newIdList)
		if err != nil {
			return curriculaTitleMap, err
		}
		noCacheIdList := make([]string, 0)
		for index, value := range values {
			if value != "" {
				curriculaTitleMap[newIdList[index]] = value
			} else {
				noCacheIdList = append(noCacheIdList, newIdList[index])
			}
		}
		if len(noCacheIdList) != 0 {
			var modelList []models.Curricula
			err = sf.DB().Collection(models.CurriculaTableName).Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(
				noCacheIdList)},
				"is_delete": bson.M{"$ne": 1}}).Find(&modelList)
			if err != nil {
				return curriculaTitleMap, err
			}
			for _, model := range modelList {
				curriculaTitleMap[model.Id.Hex()] = model.CurriculaTitle
			}
			go func() {
				_ = sf.MSetCurriculaTitleRedis(curriculaTitleMap)
			}()
		}
	}
	return curriculaTitleMap, nil
}

// ClassListV2 面试班列表
func (sf *Interview) ClassListV2(c *gin.Context) {
	var err error
	var param struct {
		Page     int64 `json:"page" form:"page"`
		PageSize int64 `json:"page_size" form:"page_size"`
	}
	err = c.ShouldBindQuery(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	if uid == "" {
		sf.SLogger().Error("用户未登录")
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 未删除的
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}}
	// 自己创建的和自己是管理员(无论是否关闭)  自己加入的
	filter["$or"] = bson.A{bson.M{"creator_user_id": uid}, bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid,
		"status": 0}}, "status": 5}, bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid,
		"identity_type": 0}}}}
	//key := fmt.Sprintf(string(rediskey.InterviewCurriculaAdminUserIdHash), uid)
	//curriculaIdQueryList, _ := sf.hKeysRedisDao(key)
	//if len(curriculaIdQueryList) != 0 {
	//	filter["$or"] = bson.A{bson.M{"creator_user_id": uid}, bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid,
	//		"status": 0}}, "status": 5}, bson.M{"curricula_id": bson.M{"$in": curriculaIdQueryList}}}
	//} else {
	//	// 自己创建的(无论是否关闭)  自己加入的
	//	filter["$or"] = bson.A{bson.M{"creator_user_id": uid}, bson.M{"members": bson.M{"$elemMatch": bson.M{"user_id": uid,
	//		"status": 0}}, "status": 5}}
	//}
	classes := make([]models.InterviewClass, 0)
	totalCount, err := sf.DB().Collection("interview_class").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	// 没有数据不再二次查询
	if totalCount == 0 {
		resp := map[string]interface{}{"total_count": totalCount, "list": classes}
		sf.Success(resp, c)
		return
	}
	offset, limit := sf.PageLimit(param.Page, param.PageSize)
	err = sf.DB().Collection("interview_class").Where(filter).Fields(bson.M{"members": 0}).Skip(offset).Sort(
		"status", "-created_time").Limit(limit).Find(&classes)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	list := make([]models.InterviewClassAndClassCount, 0)
	curriculaIdList := make([]string, 0)
	for _, class := range classes {
		if class.CurriculaId != "" {
			curriculaIdList = append(curriculaIdList, class.CurriculaId)
		}
	}
	curriculaIdTitleMap, err := sf.MGetCurriculaTitle(curriculaIdList)
	if err != nil {
		return
	}
	for _, class := range classes {
		title := ""
		if class.CurriculaId != "" {
			title = curriculaIdTitleMap[class.CurriculaId]
		}
		list = append(list, models.InterviewClassAndClassCount{
			InterviewClass: class,
			QuestionCount:  0,
			CurriculaTitle: title,
		})
	}
	resp := map[string]interface{}{"total_count": totalCount, "list": list}
	sf.Success(resp, c)
}

func (r Interview) hKeysRedisDao(key string) ([]string, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	values, err := redis.Strings(rdb.Do("HKEYS", key))
	if err != nil {
		msg := fmt.Sprintf("HKEYS key=%s err=%s", key, err.Error())
		r.SLogger().Error(msg)
		return values, err
	}
	return values, nil
}

func closeRedisConnect(conn redis.Conn) error {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// 面试班详情
func (sf *Interview) ClassInfo(c *gin.Context) {
	var err error
	var param struct {
		ClassId   string `json:"class_id"`
		ClassCode int64  `json:"class_code"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"is_deleted": bson.M{"$ne": 1}}
	if param.ClassId != "" {
		filter["_id"] = sf.ObjectID(param.ClassId)
	} else if param.ClassCode != 0 {
		filter["class_code"], _ = sf.dealClassCode(param.ClassCode)
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var class models.InterviewClass
	err = sf.DB().Collection("interview_class").Where(filter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "没有该班级哦")
		return
	}
	curriculaTitle := ""
	if class.CurriculaId != "" {
		curriculaIdTitleMap, _ := sf.MGetCurriculaTitle([]string{
			class.CurriculaId,
		})
		if curriculaIdTitleMap != nil && len(curriculaIdTitleMap) != 0 {
			curriculaTitle = curriculaIdTitleMap[class.CurriculaId]
		}
	}
	rand.Seed(time.Now().UnixNano())
	code := int64(rand.Intn(9))
	cString := fmt.Sprintf("%d%d", class.ClassCode, code)
	ci, err := strconv.Atoi(cString)
	if err == nil {
		class.ClassTeacherCode = int64(ci)
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var userName = ""
	identityType := -1
	if class.CreatorUserId == uid {
		identityType = 0
	}
	//if class.CurriculaId != "" {
	//	isShowFlag := services.NewCurriculaSrv().CheckCurriculaAdminRedis(uid, class.CurriculaId)
	//	if isShowFlag {
	//		identityType = 0
	//	}
	//}
	var memberStatus int8 = 2
	teacherCount := 0
	studentCount := 0
	teacherArr := []models.Member{}
	studentArr := []models.Member{}
	if len(class.Members) > 0 {
		uids := []string{}
		for _, v := range class.Members {
			if v.UserId == uid {
				userName = v.Name
				identityType = int(v.IdentityType)
				memberStatus = v.Status
			}
			uids = append(uids, v.UserId)
		}
		appCode := c.GetString("APP-CODE")
		memberMap := new(models.InterviewClass).GetMembersInfo(class.Id.Hex(), uids, appCode, 1)

		for i, v := range class.Members {
			class.Members[i].Avatar = memberMap[v.UserId].Avatar
			if v.IdentityType == 2 && v.Status == 0 {
				class.Members[i].IdentityName = "学生"
				studentArr = append(studentArr, class.Members[i])
				studentCount++
			} else if v.IdentityType == 1 && v.Status == 0 {
				class.Members[i].IdentityName = "老师"
				teacherArr = append(teacherArr, class.Members[i])
				teacherCount++
			} else if v.IdentityType == 0 {
				if v.UserId == class.CreatorUserId {
					class.Members[i].IdentityName = "创建者"
				} else {
					class.Members[i].IdentityName = "管理者"
				}
				teacherArr = append(teacherArr, class.Members[i])
				teacherCount++
			}

		}
	}
	sort.Slice(studentArr, func(i, j int) bool {
		return studentArr[i].Idx > studentArr[j].Idx
	})
	sort.Slice(teacherArr, func(i, j int) bool {
		return teacherArr[i].Idx > teacherArr[j].Idx
	})
	class.Members = append(teacherArr, studentArr...)

	class.UserName = userName
	class.IdentityType = int8(identityType)
	class.MemberStatus = memberStatus
	class.StudentCount = studentCount
	class.TeacherCount = teacherCount

	sf.Success(models.InterviewClassAndClassCount{
		class, 0, curriculaTitle,
	}, c)
}

// 面试班保存
func (sf *Interview) ClassSave(c *gin.Context) {
	var err error
	var param struct {
		ClassId              string `json:"class_id"`
		Name                 string `json:"name"`                     // 班级名称
		IsTeacherCQ          int8   `json:"is_teacher_cq"`            //老师创建试题 0不可以 1可以
		IsStudentCQ          int8   `json:"is_student_cq"`            //学生创建试题 0不可以 1可以
		IsStudentSeeAnswer   int8   `json:"is_student_see_answer"`    //学生看别人回答 0不可以 1可以
		IsStudentSeeComment  int8   `json:"is_student_see_comment"`   //学生看别人评论 0不可以 1可以
		IsAnswerTimeOutAlert int8   `json:"is_answer_time_out_alert"` //答题超时提醒
		StructuredGroup      int8   `json:"structured_group"`         //结构化小组模式
		AnswerTimeOut        int64  `json:"answer_time_out"`          //答题超时时间(仅作为提醒)
		UserName             string `json:"user_name"`
		Status               int8   `json:"status"`     //5正常 9关闭(学员不可见)
		IsDeleted            int8   `json:"is_deleted"` // 是否删除班级，0否1是
		CurriculaId          string `json:"curricula_id"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var class models.InterviewClass
	if param.ClassId == "" {
		//创建
		rand.Seed(time.Now().UnixNano())
		code := int64(rand.Intn(899999) + 100000)
		members := []models.Member{{UserId: uid, IdentityType: 0, Name: param.UserName}}
		if param.CurriculaId != "" {
			var model models.Curricula
			err = sf.DB().Collection(models.CurriculaTableName).Where(bson.M{"_id": sf.ObjectID(param.CurriculaId),
				"is_delete": bson.M{"$ne": 1}}).Take(&model)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeInvalidParam, c)
				return
			}
			for _, listModel := range model.AdminList {
				if listModel.AdminId != uid {
					members = append(members, models.Member{
						UserId:       listModel.AdminId,
						Name:         listModel.Name,
						IdentityType: 0,
						Status:       0,
					})
				}
			}
		}
		class = models.InterviewClass{
			Name:                 param.Name,
			CreatorUserId:        uid,
			CreatorUserName:      param.UserName,
			Members:              members,
			MemberCount:          1,
			IsTeacherCQ:          param.IsTeacherCQ,
			IsStudentCQ:          param.IsStudentCQ,
			IsStudentSeeAnswer:   param.IsStudentSeeAnswer,
			IsStudentSeeComment:  param.IsStudentSeeComment,
			IsAnswerTimeOutAlert: param.IsAnswerTimeOutAlert,
			StructuredGroup:      param.StructuredGroup,
			AnswerTimeOut:        param.AnswerTimeOut,
			ClassCode:            code,
			Status:               5,
			IsDeleted:            0,
			CurriculaId:          param.CurriculaId,
		}
		_, err = sf.DB().Collection("interview_class").Create(&class)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		err = sf.DB().Collection("interview_class").Where(bson.M{"_id": sf.ObjectID(param.ClassId), "is_delete": bson.M{"$ne": 1}}).Take(&class)
		if err != nil {
			sf.Error(common.CodeServerBusy, c)
			sf.SLogger().Error(err)
			return
		}
		if param.UserName != "" {
			if class.CreatorUserId == uid {
				class.CreatorUserName = param.UserName
			}
			for i, member := range class.Members {
				if member.UserId == uid {
					class.Members[i].Name = param.UserName
				}
			}
		}
		isAdminFlag := false
		isTeacherAndStudentFlag := false
		for _, member := range class.Members {
			if member.UserId == uid && member.IdentityType == 0 && member.Status == 0 {
				isAdminFlag = true
			}
			if member.UserId == uid && member.IdentityType != 0 && member.Status == 0 {
				isTeacherAndStudentFlag = true
			}
		}
		if isAdminFlag {
			class.Name = param.Name
			class.IsTeacherCQ = param.IsTeacherCQ
			class.IsStudentCQ = param.IsStudentCQ
			class.IsStudentSeeAnswer = param.IsStudentSeeAnswer
			class.IsStudentSeeComment = param.IsStudentSeeComment
			class.IsAnswerTimeOutAlert = param.IsAnswerTimeOutAlert
			class.StructuredGroup = param.StructuredGroup
			class.AnswerTimeOut = param.AnswerTimeOut
			class.Status = param.Status
			class.IsDeleted = param.IsDeleted
		} else {
			if !isTeacherAndStudentFlag {
				sf.Error(common.CodeServerBusy, c, "您不是该班级的成员")
				sf.SLogger().Error(err)
				return
			}
		}
		// 替换考试id
		if param.CurriculaId != class.CurriculaId {
			adminList := make([]models.Member, 0)
			membersMap := make(map[string]struct{}, 0)
			// 创建者写入
			adminList = append(adminList, models.Member{
				UserId:       class.CreatorUserId,
				Name:         class.CreatorUserName,
				IdentityType: 0,
				Status:       0,
			})
			membersMap[class.CreatorUserId] = struct{}{}
			// 考试管理员写入
			if param.CurriculaId != "" {
				var model models.Curricula
				err = sf.DB().Collection(models.CurriculaTableName).Where(bson.M{"_id": sf.ObjectID(param.CurriculaId),
					"is_delete": bson.M{"$ne": 1}}).Take(&model)
				if err != nil {
					sf.SLogger().Error(err)
					sf.Error(common.CodeInvalidParam, c)
					return
				}
				for _, listModel := range model.AdminList {
					if _, ok := membersMap[listModel.AdminId]; !ok {
						adminList = append(adminList, models.Member{
							UserId:       listModel.AdminId,
							Name:         listModel.Name,
							IdentityType: 0,
							Status:       0,
						})
						membersMap[listModel.AdminId] = struct{}{}
					}
				}
			}
			studentList := make([]models.Member, 0)
			for _, v := range class.Members {
				if v.IdentityType != 0 {
					_, flag := membersMap[v.UserId]
					if !flag {
						studentList = append(studentList, v)
					}
				}
			}
			class.Members = append(adminList, studentList...)
		}
		class.CurriculaId = param.CurriculaId
		err = sf.DB().Collection("interview_class").Save(&class)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		err = new(models.InterviewClass).SetRedisMemberInfo(class.Id.Hex(), uid, []interface{}{"name", param.UserName, "user_id", uid}...)
		if err != nil {
			sf.SLogger().Error(err)
		}
	}
	sf.Success(map[string]string{"class_id": class.Id.Hex()}, c)
}

// 班级操作成员
func (sf *Interview) ClassChangeMember(c *gin.Context) {
	var err error
	var param struct {
		ClassId      string `json:"class_id"`
		ClassCode    int64  `json:"class_code"`
		UserId       string `json:"user_id"`
		Name         string `json:"name"`          // 成员姓名
		Status       int8   `json:"status"`        //0添加 1删除
		IdentityType int8   `json:"identity_type"` //身份类型 1老师2学生

	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"is_deleted": 0}
	if param.ClassId != "" {
		filter["_id"] = sf.ObjectID(param.ClassId)

	} else if param.ClassCode != 0 {
		filter["class_code"], param.IdentityType = sf.dealClassCode(param.ClassCode)
	}
	var class models.InterviewClass
	err = sf.DB().Collection("interview_class").Where(filter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "没有该班级哦")
		return
	}
	// 班级已关闭
	if class.Status != 5 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "该班级已经关闭，请联系客服")
		return
	}
	members := class.Members
	var userMember models.Member
	tempMembers := make([]models.Member, 0)
	// 判断判断是否已经加入过
	isHave := false
	if param.Status == 1 {
		// 移除成员
		for _, v := range members {
			if v.UserId == param.UserId {
				if v.IdentityType == 0 {
					sf.SLogger().Error(errors.New("无法移除管理员的信息"))
					sf.Error(common.CodeServerBusy, c, "无法移除管理员的信息")
					return
				}
				v.Status = 2
				isHave = true
				userMember = v
			}
			tempMembers = append(tempMembers, v)
		}
	} else {
		//添加成员
		uid := c.GetHeader("X-XTJ-UID")
		if uid == "" {
			sf.Error(common.CodeServerBusy, c, "用户未登录，请登录")
			sf.SLogger().Error(errors.New("用户未登录，请登录"))
			return
		}
		param.UserId = uid
		for _, v := range members {
			if v.UserId == param.UserId {
				// 管理权限无法操作 新增或者删除
				if v.IdentityType == 0 {
					sf.SLogger().Error(errors.New("无法操作管理员的信息"))
					sf.Error(common.CodeServerBusy, c, "无法操作管理员的信息")
					return
				}
				isHave = true
				// 以前删除 变更状态
				if v.Status != param.Status {
					v.Status = param.Status
				}
				if param.IdentityType != 0 {
					v.IdentityType = param.IdentityType
				}
				if param.Name != "" {
					v.Name = param.Name
				}
				userMember = v
			}
			tempMembers = append(tempMembers, v)
		}
		if !isHave {
			addUserMember := models.Member{
				UserId:       param.UserId,
				Name:         param.Name,
				IdentityType: param.IdentityType,
				Status:       0,
			}
			tempMembers = append(tempMembers, addUserMember)
			userMember = addUserMember
		}
	}
	// 更新成员list
	members = tempMembers
	memberCount := 0
	for _, member := range members {
		if member.Status == 0 {
			memberCount += 1
		}
	}
	_, err = sf.DB().Collection("interview_class").Where(bson.M{"_id": class.Id}).Update(
		map[string]interface{}{"members": members, "updated_time": time.Now().Format("2006-01-02 15:04:05"),
			"member_count": memberCount})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	go func(userMember models.Member) {
		var gError error
		defer func() {
			if recoverError := recover(); recoverError != nil {
				sf.SLogger().Error("panic:", recoverError)
			}
		}()
		defer func() {
			if gError != nil {
				sf.SLogger().Error(gError)
			}
		}()
		uid := userMember.UserId
		// 需要删除的 成员信息
		if userMember.Status == 2 {
			gError = new(models.InterviewClass).DeleteRedisMemberInfo(class.Id.Hex(), uid)
		} else {
			// 新增或者修改
			tempUser := new(services.User).GetGatewayUsersInfo([]string{uid}, "402", 1)
			gError = new(models.InterviewClass).SetRedisMemberInfo(class.Id.Hex(), uid, []interface{}{"name",
				param.Name, "avatar", tempUser[uid].Avatar, "nickname", tempUser[uid].Nickname, "user_id", uid, "identity_type", param.IdentityType, "status", 0}...)
		}
		return

	}(userMember)

	sf.Success(map[string]string{"class_id": class.Id.Hex()}, c)
}

// SaveInterviewQuestion 保存试题
func (sf *Interview) SaveInterviewQuestion(c *gin.Context) {
	var param struct {
		QuestionId       string `json:"question_id"` // 试题ID
		ClassId          string `json:"class_id" binding:"required"`
		Name             string `json:"name" binding:"required"` // 试题名称
		Desc             string `json:"desc"`                    // 试题描述
		IsTeacherComment int8   `json:"is_teacher_comment"`      //老师点评 0不可以 1可以
		IsStudentComment int8   `json:"is_student_comment"`      //学生点评 0不可以 1可以
		Status           int32  `json:"status"`                  // 试题状态
		AnswerStyle      int8   `json:"answer_style"`            //0默认都可以看见试题 1.排队模式只有老师放行学生 学生才能看到题
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")

	// 获取老师id和学生id，校验修改试题的人是否属于班级成员
	teacherIDs := []string{}
	studentIDS := []string{}
	roleIDS := []string{}
	var class models.InterviewClass
	classFilter := bson.M{"_id": sf.ObjectID(param.ClassId), "status": 5, "is_deleted": 0}
	err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "班级不存在")
		return
	}
	for _, c := range class.Members {
		if c.IdentityType == 1 {
			teacherIDs = append(teacherIDs, c.UserId)
		} else if c.IdentityType == 2 {
			studentIDS = append(studentIDS, c.UserId)
		} else if c.IdentityType == 0 {
			roleIDS = append(roleIDS, c.UserId)
		}
	}
	//isShowFlag := false
	//if class.CurriculaId != "" {
	//	isShowFlag = services.NewCurriculaSrv().CheckCurriculaAdminRedis(uid, class.CurriculaId)
	//}

	// 如果不是班级内成员，不允许操作
	if !common.InArr(uid, teacherIDs) && !common.InArr(uid, studentIDS) && !common.InArr(uid, roleIDS) {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "不是班级内成员，不允许创建或修改试题")
		return
	}

	var question models.InterviewQuestion
	// 如果存在试题ID，代表是修改操作
	if param.QuestionId != "" {
		questionFilter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
		err = sf.DB().Collection("interview_questions").Where(questionFilter).Take(&question)
		if err != nil {
			if sf.MongoNoResult(err) {
				sf.Error(common.CodeServerBusy, c, "试题不存在，不允许修改！")
				return
			} else {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
		}
		question.Name = param.Name
		question.Desc = param.Desc
		question.IsTeacherComment = param.IsTeacherComment
		question.IsStudentComment = param.IsStudentComment
		question.Status = param.Status
		question.AnswerStyle = param.AnswerStyle
		if question.AllowAnswerUserIds == nil {
			question.AllowAnswerUserIds = make([]string, 0)
		}
		//if isShowFlag {
		//	err = sf.DB().Collection("interview_questions").Save(&question)
		//	if err != nil {
		//		sf.SLogger().Error(err)
		//		sf.Error(common.CodeServerBusy, c)
		//		return
		//	}
		//}
		// 如果不允许老师修改，即只能创建者修改
		if question.IsTeacherCanChangeStatus == 0 {
			if question.CreatorUserId == uid || common.InArr(uid, roleIDS) {
				err = sf.DB().Collection("interview_questions").Save(&question)
				if err != nil {
					sf.SLogger().Error(err)
					sf.Error(common.CodeServerBusy, c)
					return
				}
			} else {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "你不是创建者，不能修改试题状态！")
				return
			}
		} else {
			// 允许老师修改
			if question.CreatorUserId == uid || common.InArr(uid, teacherIDs) || common.InArr(uid, roleIDS) {
				// 符合条件，允许修改
				err = sf.DB().Collection("interview_questions").Save(&question)
				if err != nil {
					sf.SLogger().Error(err)
					sf.Error(common.CodeServerBusy, c)
					return
				}
			} else {
				// 不符合条件，不允许修改
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "你不是创建者或老师，不能修改试题状态！")
				return
			}
		}

	} else {
		// 新增试题
		question.ClassId = param.ClassId
		question.Name = param.Name
		question.Desc = param.Desc
		question.CreatorUserId = uid
		question.IsTeacherComment = param.IsTeacherComment
		question.IsStudentComment = param.IsStudentComment
		question.AnswerStyle = param.AnswerStyle
		question.Status = param.Status
		question.AllowAnswerUserIds = make([]string, 0)
		_, err = sf.DB().Collection("interview_questions").Create(&question)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	rdb.Do("HSET", rediskey.InterviewQuestionId2Name, question.Id.Hex(), question.Name)
	sf.Success(nil, c)
}

// DelInterviewQuestion 改变试题状态
func (sf *Interview) ChangeInterviewQuestionStatus(c *gin.Context) {
	var err error
	var param struct {
		QuestionId string `json:"question_id" binding:"required" msg:"invalid question_id"` // 试题ID
		Status     int32  `json:"status"`                                                   // 试题状态
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var question models.InterviewQuestion

	filter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
	err = sf.DB().Collection("interview_questions").Where(filter).Take(&question)
	if err == nil {
		classFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "_id": sf.ObjectID(question.ClassId)}
		var class models.InterviewClass
		err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
		isDeleteFlag := false
		isOfflineFlag := false
		// 创建者才能删除 上下架
		if uid == question.CreatorUserId {
			isDeleteFlag = true
			isOfflineFlag = true
		}
		for _, member := range class.Members {
			if member.UserId == uid && member.Status == 0 {
				if member.IdentityType == 0 {
					isDeleteFlag = true
					isOfflineFlag = true
				} else if member.IdentityType == 1 {
					isOfflineFlag = true
				}
			}
		}
		if param.Status == 9 {
			// 创建者或者 管理员才能删除
			if !isDeleteFlag {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "删除失败!身份权限不足!")
				return
			}
		} else if param.Status == 0 || param.Status == 5 {
			//上下架 只能创建者或者老师
			if !isOfflineFlag {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "操作失败!身份权限不足")
				return
			}

		}
		_, err := sf.DB().Collection("interview_questions").Where(filter).Update(map[string]interface{}{"status": param.Status, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "删除失败!")
			return
		}
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "删除失败!!试题不存在")
		return

	}
	sf.Success(nil, c)
}

// GetInterviewQuestions 查看班级下所有试题
func (sf *Interview) GetInterviewQuestions(c *gin.Context) {
	classID := c.Query("class_id")
	uid := c.GetHeader("X-XTJ-UID")
	pageIndex := c.DefaultQuery("page_index", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	if classID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	pageIndexNumber, err := strconv.Atoi(pageIndex)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	pageSizeNumber, err := strconv.Atoi(pageSize)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	appCode := c.GetString("APP-CODE")
	identityType := int8(2)
	isSuperAdmin := sf.CheckAdmin(uid)
	// 如果是大管理员，身份直接变为班级里的管理员
	if isSuperAdmin {
		identityType = 0
	} else {
		memberMap := new(models.InterviewClass).GetMembersInfo(classID, []string{uid}, appCode, 1)
		if tempValue, isExists := memberMap[uid]; isExists {
			identityType = tempValue.IdentityType
		} else {
			sf.Error(common.CodeServerBusy, c, "非大管理员或班级成员，无权查看班级下试题！")
			return
		}
	}
	// identityType := memberMap[uid].IdentityType
	filter := bson.M{"class_id": classID, "status": 5}
	if identityType != 2 {
		//管理员和老师可以看到未上架试题
		filter["status"] = bson.M{"$in": []int8{0, 5}}
	} else {
		//学生只能看到 上架的 或者自己创建的
		filter = bson.M{"$or": bson.A{bson.M{"creator_user_id": uid, "class_id": classID, "status": bson.M{"$in": []int8{0, 5}}}, filter}}
	}
	var questions = []models.InterviewQuestion{}
	offset, limit := sf.PageLimit(int64(pageIndexNumber), int64(pageSizeNumber))
	err = sf.DB().Collection("interview_questions").Where(filter).Sort("-updated_time").Skip(offset).Limit(limit).Find(&questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	for i, v := range questions {
		//学员看下是否有能看题
		if identityType == 2 {
			if v.AnswerStyle == 1 {
				v.AllowAnswerUserIds = append(v.AllowAnswerUserIds, v.CreatorUserId)
				if common.InArr(uid, v.AllowAnswerUserIds) {
					questions[i].IsAllowAnswer = 1
				}
			} else {
				questions[i].IsAllowAnswer = 1
			}
		} else {
			//老师和管理员都可以看到
			questions[i].IsAllowAnswer = 1
		}
	}
	totalCount, _ := sf.DB().Collection("interview_questions").Where(filter).Count()
	resultInfo := make(map[string]interface{})
	resultInfo["question_list"] = questions
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)

}

// GetInterviewQuestion 查看试题详情
func (sf *Interview) GetInterviewQuestion(c *gin.Context) {
	QuestionID := c.Query("question_id")
	if QuestionID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var question models.InterviewQuestion
	filter := bson.M{"_id": sf.ObjectID(QuestionID)}
	err := sf.DB().Collection("interview_questions").Where(filter).Take(&question)
	appCode := c.GetString("APP-CODE")
	memberMap := new(models.InterviewClass).GetMembersInfo(question.ClassId, []string{question.CreatorUserId}, appCode, 1)
	question.UserName = memberMap[question.CreatorUserId].Name
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "试题不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	sf.Success(question, c)
}

// 查看排队模式下试题放行学员情况
func (sf *Interview) GetInterviewQuestionStudents(c *gin.Context) {
	var err error
	questionId := c.Query("question_id")
	list := []models.Member{}
	var question models.InterviewQuestion
	err = sf.DB().Collection("interview_questions").Where(bson.M{"_id": sf.ObjectID(questionId)}).Take(&question)
	if err == nil {
		var class models.InterviewClass
		err = sf.DB().Collection("interview_class").Where(bson.M{"_id": sf.ObjectID(question.ClassId)}).Take(&class)
		if err == nil {
			uids := []string{}
			for _, v := range class.Members {
				uids = append(uids, v.UserId)
			}
			appCode := c.GetString("APP-CODE")
			memberMap := new(models.InterviewClass).GetMembersInfo(class.Id.Hex(), uids, appCode, 1)
			for _, v := range class.Members {
				if v.IdentityType == 2 && v.Status == 0 {
					if question.AnswerStyle == 1 {
						question.AllowAnswerUserIds = append(question.AllowAnswerUserIds, question.CreatorUserId)
						if common.InArr(v.UserId, question.AllowAnswerUserIds) {
							v.IsAllowAnswer = 1
						}
					} else {
						v.IsAllowAnswer = 1
					}
					v.Avatar = memberMap[v.UserId].Avatar

					list = append(list, v)
				}
			}

		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "没有该班级哦")
			return
		}

	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	sf.Success(resultInfo, c)
}

// 试题放行学员情况
func (sf *Interview) InterviewQuestionAllowStudent(c *gin.Context) {
	var err error
	questionId := c.Query("question_id")
	allowUserId := c.Query("allow_user_id")
	var question models.InterviewQuestion
	filter := bson.M{"_id": sf.ObjectID(questionId)}
	err = sf.DB().Collection("interview_questions").Where(filter).Take(&question)
	if err == nil {
		_, err = sf.DB().Collection("interview_questions").Where(filter).UpdateAtom(bson.M{"$addToSet": bson.M{"allow_answer_user_ids": allowUserId}, "$set": bson.M{"updated_time": time.Now().Format("2006-01-02 15:04:05")}})
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

// SaveAnswerLog 保存答案记录
func (sf *Interview) SaveAnswerLog(c *gin.Context) {
	var param struct {
		ClassId     string  `json:"class_id" binding:"required" `
		QuestionId  string  `json:"question_id" binding:"required" `
		VoiceUrl    string  `bson:"voice_url" json:"voice_url"`       //语音url
		VoiceLength float64 `bson:"voice_length" json:"voice_length"` //语音时长
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	// 查询是否已经提交答案
	var question models.InterviewQuestion
	err = sf.DB().Collection("interview_questions").Where(bson.M{"_id": sf.ObjectID(param.QuestionId)}).Take(&question)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "试题查询失败")
		return
	}
	var answerLog models.GAnswerLog
	answerLog.LogType = 2
	answerLog.UserId = uid
	answerLog.ClassId = param.ClassId
	answerLog.QuestionId = param.QuestionId
	answerLog.QuestionName = question.Name
	answerLog.PracticeType = int8(11)
	answerLog.Answer = []models.GAnswer{{VoiceUrl: param.VoiceUrl, VoiceContent: []models.VoiceContent{}, VoiceLength: sf.TransitionFloat64(param.VoiceLength, -1)}}
	_, err = sf.DB().Collection("g_interview_answer_logs").Create(&answerLog)
	// es中插入
	tempEsClient, err := es.CreateEsClient()
	if err == nil {
		err = es.AnswerLogAddToEs(tempEsClient, context.Background(), []models.GAnswerLog{answerLog})
		if err != nil {
			sf.SLogger().Error(err)
		}
	} else {
		sf.SLogger().Error(err)
	}
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	// logFilter := bson.M{"user_id": uid, "class_id": param.ClassId, "question_id": param.QuestionId}
	// err = sf.DB().Collection("interview_answer_logs").Where(logFilter).Take(&answerLog)
	// if err != nil {
	// 	if sf.MongoNoResult(err) {
	// 		answerLog.UserId = uid
	// 		answerLog.ClassId = param.ClassId
	// 		answerLog.QuestionId = param.QuestionId
	// 		answerLog.Answer = []models.Answer{{VoiceUrl: param.VoiceUrl, VoiceContent: []models.VoiceContent{}, VoiceLength: sf.TransitionFloat64(param.VoiceLength, -1)}}
	// 		_, err = sf.DB().Collection("interview_answer_logs").Create(&answerLog)
	// 		if err != nil {
	// 			sf.SLogger().Error(err)
	// 			sf.Error(common.CodeServerBusy, c)
	// 			return
	// 		}
	// 	} else {

	// 		sf.SLogger().Error(err)
	// 		sf.Error(common.CodeServerBusy, c)
	// 		return
	// 	}
	// } else {
	// 	if len(answerLog.Answer) > 0 {
	// 		if answerLog.Answer[0].VoiceUrl != param.VoiceUrl {
	// 			answerLog.Answer = []models.Answer{{VoiceUrl: param.VoiceUrl, VoiceLength: sf.TransitionFloat64(param.VoiceLength, -1)}}
	// 		}
	// 	} else {
	// 		answerLog.Answer = []models.Answer{{VoiceUrl: param.VoiceUrl, VoiceLength: sf.TransitionFloat64(param.VoiceLength, -1)}}
	// 	}
	// 	err = sf.DB().Collection("interview_answer_logs").Save(&answerLog)
	// 	if err != nil {
	// 		sf.SLogger().Error(err)
	// 		sf.Error(common.CodeServerBusy, c)
	// 		return
	// 	}
	// }

	sf.Success(map[string]interface{}{"log_id": answerLog.Id.Hex()}, c)
}

// GetAnswerLogs 查看班级下所有答案记录
func (sf *Interview) GetAnswerLogs(c *gin.Context) {
	questionID := c.Query("question_id")
	answerUserId := c.Query("answer_user_id")
	classId := c.Query("class_id")
	pageIndex := c.DefaultQuery("page_index", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	uid := c.GetHeader("X-XTJ-UID")
	pageIndexNumber, err := strconv.Atoi(pageIndex)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	pageSizeNumber, err := strconv.Atoi(pageSize)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var logs []models.GAnswerLog
	filter := bson.M{"log_type": 2} // log_type为2的是面试云的答题log
	if questionID != "" {
		filter["question_id"] = questionID
	} else {

		filter["user_id"] = uid
	}
	if answerUserId != "" {
		filter["user_id"] = answerUserId
	}
	// 如果出现班级id
	if classId != "" {
		classFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "_id": sf.ObjectID(classId)}
		var class models.InterviewClass
		err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "没有该班级哦")
			return
		}
		if class.CurriculaId != "" {
			curriculaFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "curricula_id": class.CurriculaId}
			var classes []models.InterviewClass
			err = sf.DB().Collection("interview_class").Where(curriculaFilter).Find(&classes)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
			classIdList := make([]string, 0)
			for _, interviewClass := range classes {
				classIdList = append(classIdList, interviewClass.Id.Hex())
			}
			if len(classIdList) != 0 {
				filter["class_id"] = bson.M{"$in": classIdList}
			}
		} else {
			filter["class_id"] = classId
		}
	}
	totalCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(filter).Count()
	offset, limit := sf.PageLimit(int64(pageIndexNumber), int64(pageSizeNumber))
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Sort("-updated_time").Skip(offset).Limit(limit).Find(&logs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	if len(logs) > 0 {
		uidList := make([]string, 0)
		logIdList := make([]string, 0)
		for _, v := range logs {
			logIdList = append(logIdList, v.Id.Hex())
			uidList = append(uidList, v.UserId)
		}
		appCode := c.GetString("APP-CODE")
		memberMap := new(models.InterviewClass).GetMembersInfo(logs[0].ClassId, uidList, appCode, 1)
		// 获取点评
		var commentLogList []models.AnswerComment
		filterQuery := bson.M{"answer_log_id": bson.M{"$in": logIdList}}
		_ = sf.DB().Collection("interview_comment_logs").Where(filterQuery).Find(&commentLogList)
		teacherCommentMap := make(map[string]struct{}, 0)
		for _, comment := range commentLogList {
			_, ok := teacherCommentMap[comment.AnswerLogId]
			if !ok {
				teacherCommentMap[comment.AnswerLogId] = struct{}{}
			}
		}
		for i, v := range logs {
			logs[i].UserName = memberMap[v.UserId].Name
			logs[i].Avatar = memberMap[v.UserId].Avatar
			logs[i].QuestionName = new(models.InterviewQuestion).GetQuestionName(v.QuestionId)
			_, ok1 := teacherCommentMap[v.Id.Hex()]
			if ok1 {
				logs[i].IsTeacherComment = true
			}
		}
	}
	resultInfo := make(map[string]interface{})
	resultInfo["logs_list"] = logs
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)

}

// GetAnswerLogsV2 查看班级下所有答案记录
func (sf *Interview) GetAnswerLogsV2(c *gin.Context) {
	var param struct {
		QuestionId string `json:"question_id" form:"question_id"`
		ClassId    string `json:"class_id" form:"class_id"`
		Page       int64  `json:"page" form:"page"`
		PageSize   int64  `json:"page_size" form:"page_size"`
		UserId     string
	}
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	if uid == "" {
		sf.SLogger().Error(errors.New("用户未登录"))
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.UserId = uid
	filter := bson.M{"log_type": 2} // log_type为2的是面试云的答题log
	if param.QuestionId == "" {
		sf.SLogger().Error("question_id不能为空")
		sf.Error(common.CodeInvalidParam, c)
		return
	} else {
		filter["question_id"] = param.QuestionId
	}
	if param.ClassId != "" {
		filter["class_id"] = param.ClassId
	} else {
		sf.SLogger().Error("class_id不能为空")
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 验证用户权限是否可以查看该班级下的回答
	var class models.InterviewClass
	err := sf.DB().Collection("interview_class").Where(bson.M{"_id": sf.ObjectID(param.ClassId)}).Take(&class)
	if err != nil {
		sf.SLogger().Error(fmt.Sprintf("未查找到相应class id=%s", param.ClassId))
		sf.Error(common.CodeServerBusy, c)
		return
	}
	isShowFlag := false
	teacherIdList := make([]string, 0)
	for _, member := range class.Members {
		if member.IdentityType <= 1 && member.Status == 0 {
			teacherIdList = append(teacherIdList, member.UserId)
			if member.UserId == param.UserId {
				isShowFlag = true
			}
		}
	}
	// 检查是否是考试课程管理员
	//if !isShowFlag {
	//	if class.CurriculaId != "" {
	//		isShowFlag = services.NewCurriculaSrv().CheckCurriculaAdminRedis(param.UserId, class.CurriculaId)
	//	}
	//}
	//超级管理员
	if !isShowFlag {
		isShowFlag = sf.CheckAdmin(param.UserId)
	}
	// 不是管理员也不是老师
	if !isShowFlag {
		sf.SLogger().Error("用户权限不足，无法查看所有回答")
		sf.Error(common.CodeServerBusy, c, "用户权限不足，无法查看所有回答")
		return
	}
	aggregateFilter := bson.A{
		bson.M{"$match": bson.M{"class_id": param.ClassId, "question_id": param.QuestionId}},
		bson.M{"$group": bson.M{"_id": "$user_id", "count": bson.M{"$sum": 1}, "all": bson.M{"$last": "$$ROOT"}}},
		bson.M{"$sort": bson.M{"all.created_time": -1}},
	}
	countAggregateFilter := append(aggregateFilter, bson.M{"$count": "total"})
	pageAggregateFilter := append(aggregateFilter, bson.M{"$skip": (param.Page - 1) * param.PageSize},
		bson.M{"$limit": param.PageSize})
	var total []struct {
		Total int64 `json:"total"`
	}
	logList := make([]models.AggregateGAnswerLogRes, 0)
	err = sf.DB().Collection("g_interview_answer_logs").Aggregate(countAggregateFilter, &total)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var res models.AggregateGAnswerLogListRes
	if total == nil || total[0].Total == 0 {
		res.Total = 0
		res.List = logList
		sf.Success(res, c)
		return
	}
	var logAggregateList []models.AggregateGAnswerLog
	err = sf.DB().Collection("g_interview_answer_logs").Aggregate(pageAggregateFilter, &logAggregateList)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	for _, logAggregate := range logAggregateList {
		logList = append(logList, models.AggregateGAnswerLogRes{
			GAnswerLog:           logAggregate.All,
			Count:                logAggregate.Count,
			IsTeacherComment:     false,
			IsTeacherCommentUser: false,
		})
	}
	if len(logList) > 0 {
		uidList := make([]string, 0)
		for _, v := range logList {
			uidList = append(uidList, v.UserId)
		}
		// 获取用户信息
		appCode := c.GetString("APP-CODE")
		memberMap := new(models.InterviewClass).GetMembersInfo(logList[0].ClassId, uidList, appCode, 1)
		// 获取点评
		var commentLogList []models.AnswerComment
		filterQuery := bson.M{"user_id": bson.M{"$in": teacherIdList}, "commented_user_id": bson.M{"$in": uidList},
			"class_id": param.ClassId, "question_id": param.QuestionId}
		_ = sf.DB().Collection("interview_comment_logs").Where(filterQuery).Find(&commentLogList)
		teacherCommentMap := make(map[string]struct{}, 0)
		teacherCommentAnswerLogIdMap := make(map[string]struct{}, 0)
		for _, comment := range commentLogList {
			_, ok := teacherCommentMap[comment.CommentedUserId]
			if !ok {
				teacherCommentMap[comment.CommentedUserId] = struct{}{}
			}
			_, ok1 := teacherCommentAnswerLogIdMap[comment.AnswerLogId]
			if !ok1 {
				teacherCommentAnswerLogIdMap[comment.AnswerLogId] = struct{}{}
			}
		}
		for i, v := range logList {
			logList[i].UserName = memberMap[v.UserId].Name
			logList[i].Avatar = memberMap[v.UserId].Avatar
			logList[i].QuestionName = new(models.InterviewQuestion).GetQuestionName(v.QuestionId)
			_, ok := teacherCommentMap[v.UserId]
			if ok {
				logList[i].IsTeacherCommentUser = true
			}
			_, ok1 := teacherCommentAnswerLogIdMap[v.Id.Hex()]
			if ok1 {
				logList[i].IsTeacherComment = true
			}
		}
	}
	res.Total = total[0].Total
	res.List = logList
	sf.Success(res, c)

}

func (sf *Interview) GetExcludeAnswerLogsV2(c *gin.Context) {
	var param struct {
		QuestionId         string `json:"question_id" form:"question_id"`
		ClassId            string `json:"class_id" form:"class_id"`
		Page               int64  `json:"page" form:"page"`
		PageSize           int64  `json:"page_size" form:"page_size"`
		AnswerUserId       string `json:"answer_user_id" form:"answer_user_id"`
		ExcludeAnswerLogId string `json:"exclude_answer_log_id" form:"exclude_answer_log_id"`
	}
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"log_type": 2} // log_type为2的是面试云的答题log
	if param.QuestionId != "" {
		filter["question_id"] = param.QuestionId
	}
	if param.AnswerUserId != "" {
		filter["user_id"] = param.AnswerUserId
	}
	if param.ClassId != "" {
		filter["class_id"] = param.ClassId
	}
	if param.ExcludeAnswerLogId != "" {
		filter["_id"] = bson.M{"$ne": sf.ObjectID(param.ExcludeAnswerLogId)}
	}
	uid := c.GetHeader("X-XTJ-UID")
	if uid == "" {
		sf.SLogger().Error(errors.New("用户未登录"))
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 验证用户权限是否可以查看该班级下的回答
	var class models.InterviewClass
	err := sf.DB().Collection("interview_class").Where(bson.M{"_id": sf.ObjectID(param.ClassId)}).Take(&class)
	if err != nil {
		sf.SLogger().Error(fmt.Sprintf("未查找到相应class id=%s", param.ClassId))
		sf.Error(common.CodeServerBusy, c)
		return
	}
	isShowFlag := false
	teacherIdList := make([]string, 0)
	for _, member := range class.Members {
		if member.IdentityType <= 1 && member.Status == 0 {
			teacherIdList = append(teacherIdList, member.UserId)
			if member.UserId == uid {
				isShowFlag = true
			}
		}
	}
	// 检查是否是考试课程管理员
	//if !isShowFlag {
	//	if class.CurriculaId != "" {
	//		isShowFlag = services.NewCurriculaSrv().CheckCurriculaAdminRedis(uid, class.CurriculaId)
	//	}
	//}
	// 不是管理员也不是老师
	if !isShowFlag {
		sf.SLogger().Error("用户权限不足，无法查看所有回答")
		sf.Error(common.CodeServerBusy, c)
		return
	}
	totalCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(filter).Count()
	offset, limit := sf.PageLimit(param.Page, param.PageSize)
	var logs []models.GAnswerLog
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Sort("-updated_time").Skip(offset).Limit(
		limit).Find(&logs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	list := make([]models.AggregateGAnswerLogRes, 0)
	if len(logs) > 0 {
		uidList := make([]string, 0)
		logIdList := make([]string, 0)
		for _, v := range logs {
			uidList = append(uidList, v.UserId)
			logIdList = append(logIdList, v.Id.Hex())
		}
		appCode := c.GetString("APP-CODE")
		memberMap := new(models.InterviewClass).GetMembersInfo(logs[0].ClassId, uidList, appCode, 1)
		// 获取点评
		var commentLogList []models.AnswerComment
		// 必须是老师才算是点评
		filterQuery := bson.M{"answer_log_id": bson.M{"$in": logIdList}, "user_id": bson.M{"$in": teacherIdList}}
		_ = sf.DB().Collection("interview_comment_logs").Where(filterQuery).Find(&commentLogList)
		teacherCommentMap := make(map[string]struct{}, 0)
		for _, comment := range commentLogList {
			_, ok := teacherCommentMap[comment.AnswerLogId]
			if !ok {
				teacherCommentMap[comment.AnswerLogId] = struct{}{}
			}
		}
		for _, v := range logs {
			v.UserName = memberMap[v.UserId].Name
			v.Avatar = memberMap[v.UserId].Avatar
			v.QuestionName = new(models.InterviewQuestion).GetQuestionName(v.QuestionId)
			isTeacherComment := false
			_, ok := teacherCommentMap[v.Id.Hex()]
			if ok {
				isTeacherComment = true
			}
			list = append(list, models.AggregateGAnswerLogRes{
				GAnswerLog:       v,
				Count:            0,
				IsTeacherComment: isTeacherComment,
			})
		}
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["total"] = totalCount
	sf.Success(resultInfo, c)

}

// GetAnswerLog 查看答案记录详情
func (sf *Interview) GetAnswerLog(c *gin.Context) {
	logID := c.Query("answer_log_id")
	if logID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var log models.GAnswerLog
	filter := bson.M{"log_type": 2, "_id": sf.ObjectID(logID)}
	err := sf.DB().Collection("g_interview_answer_logs").Where(filter).Take(&log)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "答案记录不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	appCode := c.GetString("APP-CODE")
	memberMap := new(models.InterviewClass).GetMembersInfo(log.ClassId, []string{log.UserId}, appCode, 1)
	log.UserName = memberMap[log.UserId].Name
	log.Avatar = memberMap[log.UserId].Avatar
	sf.Success(log, c)
}

// GetAnswerLog 删除答题记录
func (sf *Interview) DelAnswerLog(c *gin.Context) {
	logID := c.Query("answer_log_id")
	if logID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	filter := bson.M{"_id": sf.ObjectID(logID), "user_id": uid, "log_type": 2}
	_, err := sf.DB().Collection("g_interview_answer_logs").Where(filter).Update(map[string]interface{}{"is_deleted": 1, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "删除失败!")
		return
	}
	sf.Success(nil, c)
}

// SaveAnswerComment2 保存点评
func (sf *Interview) SaveAnswerComment2(c *gin.Context) {
	var param struct {
		AnswerLogId string         `json:"answer_log_id"  binding:"required" ` //回答记录id
		Comment     models.Comment `json:"comment"  binding:"required" `       //评论内容
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	uid := c.GetHeader("X-XTJ-UID")
	// 查询答案记录，获取试题ID和班级ID
	var answerLog models.GAnswerLog
	err = sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 2, "_id": sf.ObjectID(param.AnswerLogId)}).Take(&answerLog)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "查询答案记录失败")
		return
	}

	// 校验试题是否正常
	var question models.InterviewQuestion
	filter := bson.M{"_id": sf.ObjectID(answerLog.QuestionId)}
	err = sf.DB().Collection("interview_questions").Where(filter).Take(&question)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "试题不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		if question.Status == 9 {
			sf.Error(common.CodeServerBusy, c, "试题已删除,不支持继续点评!")
			return
		}
	}

	// 获取老师id和学生id
	teacherIDs := []string{}
	studentIDS := []string{}
	roleIDS := []string{}
	var class models.InterviewClass
	classFilter := bson.M{"_id": sf.ObjectID(question.ClassId), "status": 5, "is_deleted": 0}
	err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "班级不存在")
		return
	}
	for _, c := range class.Members {
		if c.IdentityType == 1 {
			teacherIDs = append(teacherIDs, c.UserId)
		} else if c.IdentityType == 2 {
			studentIDS = append(studentIDS, c.UserId)
		} else if c.IdentityType == 0 {
			roleIDS = append(roleIDS, c.UserId)
		}
	}

	// 如果不是班级内成员，不允许点评
	if !common.InArr(uid, teacherIDs) && !common.InArr(uid, studentIDS) && !common.InArr(uid, roleIDS) {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "不是班级内成员，不允许点评")
		return
	}

	IsTeacherComment := question.IsTeacherComment
	IsStudentComment := question.IsStudentComment
	// 如果不允许老师点评
	if IsTeacherComment == 0 {
		if common.InArr(uid, teacherIDs) {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "老师不允许点评")
			return
		}
	}
	// 如果不允许学生点评
	if IsStudentComment == 0 {
		if common.InArr(uid, studentIDS) {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "学生不允许点评")
			return
		}
	}

	var comment models.AnswerComment
	comment.UserId = uid
	comment.ClassId = answerLog.ClassId
	comment.CommentedUserId = answerLog.UserId
	comment.AnswerLogId = param.AnswerLogId
	comment.QuestionId = answerLog.QuestionId
	comment.StructuredGroup = 1
	comment.Comment = param.Comment
	comment.Comment.GAnswer.VoiceLength = sf.TransitionFloat64(comment.Comment.GAnswer.VoiceLength, -1)

	var commentData models.AnswerComment
	commentFilter := bson.M{"answer_log_id": param.AnswerLogId, "user_id": uid}
	err = sf.DB().Collection("interview_comment_logs").Where(commentFilter).Take(&commentData)
	if err != nil {
		if sf.MongoNoResult(err) {
			_, err = sf.DB().Collection("interview_comment_logs").Create(&comment)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		// 存在更改
		commentData.UserId = uid
		commentData.ClassId = answerLog.ClassId
		commentData.CommentedUserId = answerLog.UserId
		commentData.AnswerLogId = param.AnswerLogId
		commentData.QuestionId = answerLog.QuestionId
		commentData.StructuredGroup = 1
		commentData.Comment = param.Comment
		comment.Comment.GAnswer.VoiceLength = sf.TransitionFloat64(comment.Comment.GAnswer.VoiceLength, -1)
		err = sf.DB().Collection("interview_comment_logs").Save(&commentData)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}

	sf.Success(nil, c)
}

// SaveAnswerComment 保存点评
func (sf *Interview) SaveAnswerComment(c *gin.Context) {
	var param struct {
		AnswerLogId string         `json:"answer_log_id"  binding:"required" ` //回答记录id
		Comment     models.Comment `json:"comment"  binding:"required" `       //评论内容
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if !common.InArr(param.Comment.Content, []string{"跑题", "偏题", "准确", "优秀", "共情"}) {
		sf.Error(common.InvalidId, c, "点评中的内容不符合要求")
		return
	}

	if !common.InArr(param.Comment.Speed, []string{"偏快", "适中", "偏慢"}) {
		sf.Error(common.InvalidId, c, "点评中的语速不符合要求")
		return
	}
	if !common.InArr(param.Comment.Interaction, []string{"无互动", "有互动", "互动恰当"}) {
		sf.Error(common.InvalidId, c, "点评中的互动不符合要求")
		return
	}
	if !common.InArr(param.Comment.Confident, []string{"自信", "不自信", "过度自信"}) {
		sf.Error(common.InvalidId, c, "点评中的自信不符合要求")
		return
	}
	commentKey := param.Comment.Content + "-" + param.Comment.Speed + "-" + param.Comment.Interaction + "-" + param.Comment.Confident
	grade := common.GetGrade(commentKey)
	param.Comment.Grade = grade

	uid := c.GetHeader("X-XTJ-UID")
	// 查询答案记录，获取试题ID和班级ID
	var answerLog models.GAnswerLog
	err = sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 2, "_id": sf.ObjectID(param.AnswerLogId)}).Take(&answerLog)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "查询答案记录失败")
		return
	}

	// 校验试题是否正常
	var question models.InterviewQuestion
	filter := bson.M{"_id": sf.ObjectID(answerLog.QuestionId)}
	err = sf.DB().Collection("interview_questions").Where(filter).Take(&question)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "试题不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		if question.Status == 9 {
			sf.Error(common.CodeServerBusy, c, "试题已删除,不支持继续点评!")
			return
		}
	}

	// 获取老师id和学生id
	teacherIDs := []string{}
	studentIDS := []string{}
	roleIDS := []string{}
	var class models.InterviewClass
	classFilter := bson.M{"_id": sf.ObjectID(question.ClassId), "status": 5, "is_deleted": 0}
	err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "班级不存在")
		return
	}
	for _, c := range class.Members {
		if c.IdentityType == 1 {
			teacherIDs = append(teacherIDs, c.UserId)
		} else if c.IdentityType == 2 {
			studentIDS = append(studentIDS, c.UserId)
		} else if c.IdentityType == 0 {
			roleIDS = append(roleIDS, c.UserId)
		}
	}

	// 如果不是班级内成员，不允许点评
	if !common.InArr(uid, teacherIDs) && !common.InArr(uid, studentIDS) && !common.InArr(uid, roleIDS) {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "不是班级内成员，不允许点评")
		return
	}

	IsTeacherComment := question.IsTeacherComment
	IsStudentComment := question.IsStudentComment
	// 如果不允许老师点评
	if IsTeacherComment == 0 {
		if common.InArr(uid, teacherIDs) {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "老师不允许点评")
			return
		}
	}
	// 如果不允许学生点评
	if IsStudentComment == 0 {
		if common.InArr(uid, studentIDS) {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "学生不允许点评")
			return
		}
	}

	var comment models.AnswerComment
	comment.UserId = uid
	comment.ClassId = answerLog.ClassId
	comment.CommentedUserId = answerLog.UserId
	comment.AnswerLogId = param.AnswerLogId
	comment.QuestionId = answerLog.QuestionId
	comment.Comment = param.Comment
	comment.StructuredGroup = 0
	comment.Comment.GAnswer.VoiceLength = sf.TransitionFloat64(comment.Comment.GAnswer.VoiceLength, -1)

	var commentData models.AnswerComment
	commentFilter := bson.M{"answer_log_id": param.AnswerLogId, "user_id": uid}
	err = sf.DB().Collection("interview_comment_logs").Where(commentFilter).Take(&commentData)
	if err != nil {
		if sf.MongoNoResult(err) {
			_, err = sf.DB().Collection("interview_comment_logs").Create(&comment)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		// 存在更改
		commentData.UserId = uid
		commentData.ClassId = answerLog.ClassId
		commentData.CommentedUserId = answerLog.UserId
		commentData.AnswerLogId = param.AnswerLogId
		commentData.QuestionId = answerLog.QuestionId
		commentData.Comment = param.Comment
		commentData.StructuredGroup = 0
		comment.Comment.GAnswer.VoiceLength = sf.TransitionFloat64(comment.Comment.GAnswer.VoiceLength, -1)
		err = sf.DB().Collection("interview_comment_logs").Save(&commentData)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}

	sf.Success(nil, c)
}

// GetAnswerComments 个人或试题所有点评列表
func (sf *Interview) GetAnswerComments(c *gin.Context) {
	var param struct {
		CommentedUserId string `json:"commented_user_id"` //被点评人
		CommentUserId   string `json:"comment_user_id"`   //点评人
		AnswerLogId     string `json:"answer_log_id"`     //答题id
		PageIndex       int64  `json:"page_index"`
		PageSize        int64  `json:"page_size"`
		IsAdmin         int8   `json:"is_admin"` //管理员查看所有
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	appCode := c.GetString("APP-CODE")
	var comments = []models.AnswerComment{}
	var filter = bson.M{}
	uid := c.GetHeader("X-XTJ-UID")
	if param.IsAdmin == 1 {
		if !sf.CheckAdmin(uid) {
			sf.Error(common.PermissionDenied, c)
			return
		}
	} else {
		if param.AnswerLogId != "" {
			filter = bson.M{"answer_log_id": param.AnswerLogId}
		} else {
			filter = bson.M{"user_id": uid}
		}
		if param.CommentedUserId != "" {
			filter = bson.M{"commented_user_id": param.CommentedUserId}
		}
		if param.CommentUserId != "" {
			filter["user_id"] = param.CommentUserId
		}
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("interview_comment_logs").Where(filter).Sort("-updated_time").Skip(offset).Limit(limit).Find(&comments)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	if len(comments) > 0 {
		uids := []string{}
		for _, v := range comments {
			uids = append(uids, v.UserId)
		}
		memberMap := new(models.InterviewClass).GetMembersInfo(comments[0].ClassId, uids, appCode, 1)
		for i, v := range comments {
			comments[i].UserName = memberMap[v.UserId].Name
			comments[i].Avatar = memberMap[v.UserId].Avatar
			comments[i].QuestionName = new(models.InterviewQuestion).GetQuestionName(v.QuestionId)
		}
	}
	totalCount, _ := sf.DB().Collection("interview_comment_logs").Where(filter).Count()
	resultInfo := make(map[string]interface{})
	resultInfo["comment_list"] = comments
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

// GetAnswerComment 点评详情
func (sf *Interview) GetAnswerComment(c *gin.Context) {
	answerCommentID := c.Query("answer_comment_id")
	if answerCommentID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var data models.AnswerComment
	filter := bson.M{"_id": sf.ObjectID(answerCommentID)}
	err := sf.DB().Collection("interview_comment_logs").Where(filter).Take(&data)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "点评记录不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	appCode := c.GetString("APP-CODE")
	uids := []string{}
	uids = append(uids, data.UserId)
	uids = append(uids, data.CommentedUserId)
	memberMap := new(models.InterviewClass).GetMembersInfo(data.ClassId, uids, appCode, 1)
	data.UserName = memberMap[data.UserId].Name
	data.Avatar = memberMap[data.UserId].Avatar
	data.CommentedUserAvatar = memberMap[data.CommentedUserId].Avatar
	data.CommentedUserName = memberMap[data.CommentedUserId].Name
	data.QuestionName = new(models.InterviewQuestion).GetQuestionName(data.QuestionId)
	sf.Success(data, c)
}

// CommentFeedBack2 点评反馈
func (sf *Interview) CommentFeedBack2(c *gin.Context) {
	var param struct {
		CommentId string         `json:"comment_log_id"  binding:"required" ` //回答记录id
		Comment   models.Comment `json:"comment"  binding:"required" `        //评论内容
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var commentData models.AnswerComment
	commentFilter := bson.M{"_id": sf.ObjectID(param.CommentId)}
	err = sf.DB().Collection("interview_comment_logs").Where(commentFilter).Take(&commentData)
	if err != nil {
		if sf.MongoNoResult(err) {
			if err != nil {
				sf.Error(common.CodeServerBusy, c, "点评不存在，无法对评价进行反馈")
				return
			}
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		// 判断是否属于自己的点评
		uid := c.GetHeader("X-XTJ-UID")
		if uid != commentData.CommentedUserId {
			if err != nil {
				sf.Error(common.CodeServerBusy, c, "这个点评不属于你，无法对评价进行反馈")
				return
			}
		}
		commentData.ReplyComment = param.Comment
		err = sf.DB().Collection("interview_comment_logs").Save(&commentData)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}

	sf.Success(nil, c)
}

// CommentFeedBack 点评反馈
func (sf *Interview) CommentFeedBack(c *gin.Context) {
	var param struct {
		CommentId string       `json:"comment_log_id"  binding:"required" ` //回答记录id
		Reply     models.Reply `json:"reply" binding:"required" `           //反馈
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var commentData models.AnswerComment
	commentFilter := bson.M{"_id": sf.ObjectID(param.CommentId)}
	err = sf.DB().Collection("interview_comment_logs").Where(commentFilter).Take(&commentData)
	if err != nil {
		if sf.MongoNoResult(err) {
			if err != nil {
				sf.Error(common.CodeServerBusy, c, "点评不存在，无法对评价进行反馈")
				return
			}
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		// 判断是否属于自己的点评
		uid := c.GetHeader("X-XTJ-UID")
		if uid != commentData.CommentedUserId {
			if err != nil {
				sf.Error(common.CodeServerBusy, c, "这个点评不属于你，无法对评价进行反馈")
				return
			}
		}
		commentData.Reply = param.Reply
		err = sf.DB().Collection("interview_comment_logs").Save(&commentData)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}

	sf.Success(nil, c)
}
func (sf *Interview) dealClassCode(code int64) (classCode int64, identityType int8) {
	if code > 999999 {
		return int64(code / 10), 1
	}
	return code, 2
}

func (sf *Interview) IsAdmin(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	sf.Success(map[string]interface{}{"is_admin": sf.CheckAdmin(uid)}, c)
}
func (sf *Interview) CheckAdmin(uid string) bool {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	key := rediskey.InterviewAdminAuthUserIds
	exists, _ := redis.Bool(rdb.Do("EXISTS", key))
	if exists {
		isHave, _ := redis.Bool(rdb.Do("SISMEMBER", key, uid))
		if isHave {
			return true
		} else {
			return false
		}
	} else {
		redis.Bool(rdb.Do("SADD", key, "bnmdu1mh04rug2h87ds0"))
	}
	return false
}

// 面试班列表
func (sf *Interview) AllClassList(c *gin.Context) {
	var err error
	var param struct {
		PageIndex int64 `json:"page_index"`
		PageSize  int64 `json:"page_size"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	isShow := helper.RedisSIsMember(string(rediskey.InterviewAdminAuthUserIds), uid)
	if !isShow {
		sf.Error(common.CodeServerBusy, c, "暂无权限查看全部班级！")
		return
	}

	filter := bson.M{"status": 5, "is_deleted": 0}
	totalCount, err := sf.DB().Collection("interview_class").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	var classes []models.InterviewClass
	err = sf.DB().Collection("interview_class").Where(filter).Fields(bson.M{"members": 0}).Skip(offset).Sort("-created_time").Limit(limit).Find(&classes)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resp := map[string]interface{}{"total_count": totalCount, "list": classes}
	sf.Success(resp, c)
}

// 面试班列表
func (sf *Interview) RemoveStudent(c *gin.Context) {
	var err error
	var param struct {
		ClassId    string   `json:"class_id"`
		ToClassId  string   `json:"to_class_id"`
		UserIdList []string `json:"user_id_list"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if len(param.UserIdList) == 0 {
		sf.Error(common.CodeInvalidParam, c, "请选择要移除的学生")
		return
	}
	// 验证操作人员是否有权限
	uid := c.GetHeader("X-XTJ-UID")
	classFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "_id": sf.ObjectID(param.ClassId)}
	var class models.InterviewClass
	err = sf.DB().Collection("interview_class").Where(classFilter).Take(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "移动的班级不存在，请检查")
		return
	}
	uidIdentityFlag := false
	for _, member := range class.Members {
		if len(class.Members) == 1 {
			sf.Error(common.CodeServerBusy, c, "班级中只有一个人，无法移动")
			return
		}
		if uid == member.UserId && member.IdentityType <= 1 && member.Status == 0 {
			uidIdentityFlag = true
			break
		}
	}
	if !uidIdentityFlag {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "用户移动权限不足")
		return
	}
	studentList := make([]models.Member, 0)
	memberList := make([]models.Member, 0)
	for _, member := range class.Members {
		// 学生才能被移除
		isFind := false
		if member.IdentityType == 2 && member.Status == 0 {
			for _, userId := range param.UserIdList {
				if member.UserId == userId {
					isFind = true
					studentList = append(studentList, member)
				}
			}
		}
		if isFind {
			member.Status = 2
		}
		memberList = append(memberList, member)
	}
	class.Members = memberList
	// 验证目标班级是否存在
	toClassFilter := bson.M{"is_deleted": bson.M{"$ne": 1}, "_id": sf.ObjectID(param.ToClassId)}
	var toClass models.InterviewClass
	err = sf.DB().Collection("interview_class").Where(toClassFilter).Take(&toClass)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "目标的班级不存在，请检查")
		return
	}
	toClassMembers := make([]models.Member, 0)
	for _, member := range studentList {
		isFind := false
		for _, toMember := range toClass.Members {
			if toMember.UserId == member.UserId && toMember.IdentityType == 2 && toMember.Status == 0 {
				isFind = true
			}
		}
		if !isFind {
			toClassMembers = append(toClassMembers, member)
		}
	}
	toClass.Members = append(toClass.Members, toClassMembers...)
	// 更新班级
	err = sf.DB().Collection("interview_class").Where(classFilter).Save(&class)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	err = sf.DB().Collection("interview_class").Where(toClassFilter).Save(&toClass)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}
