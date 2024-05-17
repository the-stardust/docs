package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/helper"
	"interview/models"
	"interview/router/request"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MockExam struct {
	ServicesBase
}

func NewMockExamService() *MockExam {
	return &MockExam{}
}

func (sf *MockExam) Create(req request.MockExamCreateRequest) (*models.MockExam, *models.MockExamSlotRoom, error) {
	var err error

	slotUserNum := 1 + 1
	if req.ExamMode == "结构化小组" {
		slotUserNum = req.MaxUserNum
	}
	teacherTeam := make([][]models.MockExamUser, 0)
	//userMap := new(models.InterviewGPT).GetUsersInfo([]string{req.Uid}, "402", 1)
	//teacherTeamItem := make([]models.MockExamUser, 0)
	//teacherTeamItem = append(teacherTeamItem, models.MockExamUser{
	//	UserId:   req.Uid,
	//	Avatar:   userMap[req.Uid].Avatar,
	//	UserName: userMap[req.Uid].Nickname,
	//})
	//teacherTeam = append(teacherTeam, teacherTeamItem)
	timeNow := time.Now()
	slotDate := timeNow.Format("2006-01-02")
	startT := timeNow.Format("2006-01-02 15:04:05")
	endT := timeNow.Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	exam := &models.MockExam{
		Title:             req.Title,
		ExamCategory:      req.ExamCategory,
		ExamChildCategory: req.ExamChildCategory,
		ExamMode:          req.ExamMode,
		ExamType:          1,
		Password:          req.Password,
		Status:            1,
		SignUpStartTime:   startT,
		SignUpEndTime:     endT,
		QuestionType:      2,
		ExamInfo: []models.ExamInfo{{
			StartTime:   startT,
			EndTime:     endT,
			ExamTime:    3600,
			TeacherTeam: teacherTeam,
			SlotUserNum: slotUserNum,
			SlotTimeList: []models.ExamSlotTime{
				{
					SlotDate:      slotDate,
					SlotStartTime: timeNow.Format("15:04:05"),
					SlotEndTime:   timeNow.Add(1 * time.Hour).Format("15:04:05"),
				},
			},
		}},
		Question:   make([]models.MockQuestion, 0),
		MaxUserNum: slotUserNum,
		CreatedBy:  req.Uid,
	}
	//

	_, err = sf.MockExamModel().Create(exam)

	// 创建房间
	var log = new(models.MockExamLog)
	log.UserId = req.Uid
	log.MockExamId = exam.Id.Hex()
	log.ExamCategory = exam.ExamCategory
	log.ExamChildCategory = exam.ExamChildCategory
	log.ExamMode = exam.ExamMode
	log.SlotDate = slotDate
	log.SlotStartTime = timeNow.Format("15:04:05")
	log.SlotEndTime = timeNow.Add(1 * time.Hour).Format("15:04:05")
	log.Status = 1
	log.Question = make([]models.MockQuestion, 0)

	room, err := sf.assignRoom(exam, log)
	if err != nil {
		return exam, room, err
	}

	log.TeacherTeam = room.TeacherTeam
	log.RoomId = room.Id.Hex()

	_, err = sf.MockExamLogModel().Create(log)
	sf.MockExamModel().Where(bson.M{"_id": exam.Id}).UpdateAtom(bson.M{"$inc": bson.M{"sign_up_user_num": 1}, "$set": bson.M{"updated_time": time.Now().Format("2006-01-02 15:04:05")}})

	return exam, room, err
}

func (sf *MockExam) GetExamRoom(roomId string) *models.MockExamSlotRoom {
	room := new(models.MockExamSlotRoom)
	sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Take(room)
	return room
}

// 房间成员变化
func (sf *MockExam) RoomMemberChange(param request.RoomMemberChangeRequest) error {
	var logs = new(models.MockExamLog)
	err := sf.MockExamLogModel().Where(bson.M{"room_id": param.RoomId}).Take(&logs)
	if err != nil {
		return err
	}

	var room = new(models.MockExamSlotRoom)
	sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(param.RoomId)}).Take(&room)

	var users = make([]models.MockExamUser, 0)
	for _, user := range room.Users {
		if user.UserId != param.Uid {
			users = append(users, user)
		}
	}
	room.Users = users
	_, err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": room.Id}).Update(room)
	_, err = sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(room.MockExamId)}).Update(bson.M{"sign_up_user_num": len(users)})
	return err
}

func (sf *MockExam) SetExamRoomVideoTaskId(roomId, taskId string) error {
	room := new(models.MockExamSlotRoom)
	sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Take(room)
	room.VideoTaskId = taskId
	if len(room.VideoArr) == 0 {
		room.VideoArr = make([]models.RoomVideo, 0)
	}
	room.VideoArr = append(room.VideoArr, models.RoomVideo{
		VideoTaskId: taskId,
		Users:       room.Users,
	})

	_, err := sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Update(room)
	return err
}

func (sf *MockExam) SetExamRoomVideoUrl(roomId, taskId, fileUrl string) error {
	room := new(models.MockExamSlotRoom)
	sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Take(room)
	room.VideoUrl = fileUrl
	users := make([]models.MockExamUser, 0)
	for i, item := range room.VideoArr {
		if item.VideoTaskId == taskId {
			room.VideoArr[i].VideoUrl = fileUrl
			users = room.VideoArr[i].Users
			break
		}
	}

	_, err := sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Update(room)

	// 更新到MockExamLog上
	for _, user := range users {
		var log = new(models.MockExamLog)
		err = sf.MockExamLogModel().Where(bson.M{"room_id": roomId, "user_id": user.UserId}).Take(log)
		if err != nil {
			continue
		}
		if !common.InArrCommon[string](fileUrl, log.VideoUrls) {
			log.VideoUrls = append(log.VideoUrls, fileUrl)
			sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(log)
		}
	}
	return err
}

func (sf *MockExam) SetRoomUseQuestion(roomId, questionId string) error {
	logs := make([]models.MockExamLog, 0)
	err := sf.MockExamLogModel().Where(bson.M{"room_id": roomId, "status": bson.M{"$ne": -1}}).Find(&logs)
	if err != nil {
		return err
	}
	newQ := models.MockQuestion{
		QuestionId: questionId,
	}
	for i := range logs {
		//_, err = sf.MockExamLogModel().Where(bson.M{"_id": logs[i].Id}).Update(&logs[i])
		_, err = sf.MockExamLogModel().Where(bson.M{"_id": logs[i].Id}).UpdateAtom(bson.M{"$addToSet": bson.M{"question": newQ}})
	}

	return err
}

// 结束模考
func (sf *MockExam) End(uid, roomId string) error {
	room := new(models.MockExamSlotRoom)
	err := sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Take(room)
	if err != nil {
		return err
	}
	_, err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": room.Id}).Update(bson.M{"status": 1, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
	if err != nil {
		return err
	}
	uids := make([]string, 0)
	for _, user := range room.Users {
		uids = append(uids, user.UserId)
	}

	_, err = sf.MockExamLogModel().Where(bson.M{"user_id": bson.M{"$in": uids}, "room_id": roomId, "status": bson.M{"$ne": -1}}).Update(bson.M{"status": 5, "updated_time": time.Now().Format("2006-01-02 15:04:05")})

	return err
}

func (sf *MockExam) List(filter bson.M, offset, limit int64) ([]*models.MockExam, int64, error) {
	list := make([]*models.MockExam, 0)
	total, err := sf.MockExamModel().Where(filter).Count()
	if err != nil {
		return list, total, err
	}
	err = sf.MockExamModel().Where(filter).Sort("-created_time").Skip(offset).Limit(limit).Find(&list)

	if err != nil {
		return list, total, err
	}

	list, _ = sf.getQuestionDetail(list, make([]models.MockExamLog, 0))

	return list, total, nil
}

func (sf *MockExam) getQuestionDetail(list []*models.MockExam, logList []models.MockExamLog) ([]*models.MockExam, []models.MockExamLog) {
	var questionIds = make(map[string]models.MockQuestion)
	var questions = make([]models.GQuestion, 0)
	var questionMap = make(map[string]models.GQuestion)
	if len(logList) > 0 {
		for _, exam := range logList {
			for _, question := range exam.Question {
				questionIds[question.QuestionId] = question
			}
		}
	} else {
		for _, exam := range list {
			for _, question := range exam.Question {
				questionIds[question.QuestionId] = question
			}
		}
	}
	if len(questionIds) == 0 {
		return list, logList
	}
	sf.GQuestionModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(common.GetMapKeys(questionIds))}}).Find(&questions)
	for i := range questions {
		questions[i].RequireAnswerDuration = questionIds[questions[i].Id.Hex()].RequireAnswerDuration
	}
	for _, i2 := range questions {
		questionMap[i2.Id.Hex()] = i2
	}
	if len(logList) > 0 {
		for i, exam := range logList {
			var questions2 = make([]models.GQuestion, 0)
			// 排序
			for _, question := range exam.Question {

				if _, ok := questionMap[question.QuestionId]; ok {
					questions2 = append(questions2, questionMap[question.QuestionId])
				}
			}
			logList[i].QuestionDetail = questions2
		}
	} else {
		for i, exam := range list {
			var questions2 = make([]models.GQuestion, 0)
			// 排序
			for _, question := range exam.Question {

				if _, ok := questionMap[question.QuestionId]; ok {
					questions2 = append(questions2, questionMap[question.QuestionId])
				}
			}
			list[i].QuestionDetail = questions2
		}
	}

	return list, logList
}

// 小程序端
func (sf *MockExam) List2(filter bson.M, uid string, offset, limit int64) ([]*models.MockExam, int64, error) {
	list := make([]*models.MockExam, 0)
	total, err := sf.MockExamModel().Where(filter).Count()
	if err != nil {
		return list, total, err
	}
	err = sf.MockExamModel().Where(filter).Sort("-sign_up_start_time").Skip(offset).Limit(limit).Find(&list)

	if err != nil {
		return list, total, err
	}
	ids := make([]string, 0)
	for _, exam := range list {
		ids = append(ids, exam.Id.Hex())
	}

	rooms := make([]*models.MockExamSlotRoom, 0)
	sf.MockExamSlotRoomModel().Where(bson.M{"mock_exam_id": bson.M{"$in": ids}}).Find(&rooms)
	roomMap := make(map[string]*models.MockExamSlotRoom, 0)
	for _, room := range rooms {
		roomMap[room.MockExamId] = room
	}

	for i := range list {
		list[i].UserIds = make([]string, 0)
		if room, ok := roomMap[list[i].Id.Hex()]; ok {
			list[i].SignUpUserNum = len(room.Users)
			for _, user := range room.Users {
				list[i].UserIds = append(list[i].UserIds, user.UserId)
			}
			if list[i].ExamType == 1 && len(room.TeacherTeam) > 0 {
				list[i].SignUpUserNum += 1
			}
		}
	}

	if uid != "" {
		logs := make([]*models.MockExamLog, 0)
		logMap := make(map[string]*models.MockExamLog, 0)
		sf.MockExamLogModel().Where(bson.M{"mock_exam_id": bson.M{"$in": ids}, "user_id": uid, "status": bson.M{"$ne": -1}}).Find(&logs)
		for _, log := range logs {
			logMap[log.MockExamId] = log
		}

		for i := range list {
			if _, ok := logMap[list[i].Id.Hex()]; ok {
				list[i].Log = logMap[list[i].Id.Hex()]

				cacheKey := string(rediskey.MockExamStatus) + list[i].Log.RoomId
				status, _ := helper.RedisHGetAll(cacheKey)
				if _, ok := status["start"]; ok {
					list[i].RoomStatus = "start"
				}
				if _, ok := status["close"]; ok {
					list[i].RoomStatus = "close"
				}
				if list[i].Log.Status == 5 {
					list[i].RoomStatus = "close"
				}
			}
		}
	}

	return list, total, nil
}

func (sf *MockExam) Edit(param request.MockexamEditRequest) (string, error) {
	var err error
	var mockexam = new(models.MockExam)
	if param.Id != "" {
		err = sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(mockexam)
		if err != nil {
			return "", err
		}

		// 如果已经报名 不允许修改时间段
		total, _ := sf.MockExamLogModel().Where(bson.M{"mock_exam_id": param.Id, "status": bson.M{"$ne": -1}}).Count()
		if total > 0 {
			var paramExamInfo string
			var examInfo string
			for _, i2 := range param.ExamInfo {
				paramExamInfo += i2.StartTime
				paramExamInfo += i2.EndTime
				paramExamInfo += strconv.Itoa(i2.ExamTime)
				paramExamInfo += strconv.Itoa(i2.RestTime)
			}
			for _, i2 := range mockexam.ExamInfo {
				examInfo += i2.StartTime
				examInfo += i2.EndTime
				examInfo += strconv.Itoa(i2.ExamTime)
				examInfo += strconv.Itoa(i2.RestTime)
			}
			if paramExamInfo != examInfo {
				return "", errors.New("已经有学员报名，无法修改")
			}

			var paramTeacher = make([]string, 0)
			var examTeacher = make([]string, 0)
			for _, i2 := range param.ExamInfo {
				for _, users := range i2.TeacherTeam {
					for _, user := range users {
						paramTeacher = append(paramTeacher, user.UserId)
					}
				}

			}
			for _, i2 := range mockexam.ExamInfo {
				for _, users := range i2.TeacherTeam {
					for _, user := range users {
						examTeacher = append(examTeacher, user.UserId)
					}
				}
			}
			sort.Strings(paramTeacher)
			sort.Strings(examTeacher)
			if strings.Join(paramTeacher, "") != strings.Join(examTeacher, "") {
				return "", errors.New("已经有学员报名，无法修改")
			}
		}
	}
	mockexam.Title = param.Title
	mockexam.SubTitle = param.SubTitle
	mockexam.Introduce = param.Introduce
	mockexam.ExamCategory = param.ExamCategory
	mockexam.ExamChildCategory = param.ExamChildCategory
	mockexam.ExamMode = param.ExamMode
	mockexam.ExamInfo = param.ExamInfo
	mockexam.SignUpStartTime = param.SignUpStartTime
	mockexam.SignUpEndTime = param.SignUpEndTime
	mockexam.Status = param.Status
	mockexam.QuestionType = param.QuestionType
	mockexam.Question = param.Question

	maxUserNum := 0
	for _, info := range mockexam.ExamInfo {
		maxUserNum += info.SlotUserNum * len(info.SlotTimeList) * len(info.TeacherTeam)
	}
	mockexam.MaxUserNum = maxUserNum

	if param.Id != "" {
		mockexam.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
		_, err = sf.MockExamModel().Where(bson.M{"_id": mockexam.Id}).Update(mockexam)
	} else {
		_, err = sf.MockExamModel().Create(mockexam)
	}
	return mockexam.Id.Hex(), err
}

// 报名
func (sf *MockExam) SignUp(uid string, param request.SignUpRequest, retry int) (*models.MockExamLog, error) {
	var err error
	var exam = new(models.MockExam)
	var log = new(models.MockExamLog)
	err = sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(param.ExamId)}).Take(exam)
	if err != nil {
		return log, err
	}

	nowTimestamp := time.Now().Unix()
	// 2024-05-09 12:25
	if len(exam.SignUpStartTime) == 16 {
		exam.SignUpStartTime = exam.SignUpStartTime + ":00"
		exam.SignUpEndTime = exam.SignUpEndTime + ":00"
	}
	startTime, _ := common.LocalTimeFromDateString(exam.SignUpStartTime)
	endTime, _ := common.LocalTimeFromDateString(exam.SignUpEndTime)
	if nowTimestamp < startTime.Unix() {
		return log, errors.New("报名还未开始")
	}
	if endTime.Unix() < nowTimestamp {
		return log, errors.New("报名已结束")
	}

	// 围观
	if exam.ExamType == 1 {
		if global.CONFIG.Env == "dev" {
			//if common.InArrCommon[string](uid, []string{"co5314au1osaiv97h7rg"}) {
			//	var roooom = new(models.MockExamSlotRoom)
			//	sf.MockExamSlotRoomModel().Where(bson.M{"mock_exam_id": exam.Id.Hex()}).Take(&roooom)
			//	log.RoomId = roooom.Id.Hex()
			//	return log, nil
			//}
		} else {
			if common.InArrCommon[string](uid, []string{"c68vn2uh04rug2i4hafg", "cmf24j8pumi421hc3img"}) {
				var roooom = new(models.MockExamSlotRoom)
				sf.MockExamSlotRoomModel().Where(bson.M{"mock_exam_id": exam.Id.Hex()}).Take(&roooom)
				log.RoomId = roooom.Id.Hex()
				return log, nil
			}
		}
	}

	err = sf.MockExamLogModel().Where(bson.M{"user_id": uid, "mock_exam_id": param.ExamId, "status": bson.M{"$ne": -1}}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return log, err
	}
	hasLog := true
	if err != nil && sf.MongoNoResult(err) {
		hasLog = false
	}
	// 自主练习的创建人 不需要报名
	if exam.ExamType != 1 && log.Status != 0 {
		return log, errors.New("无法修改")
	}
	slotUserNum := -1
	for _, info := range exam.ExamInfo {
		for _, _time := range info.SlotTimeList {
			if _time.SlotDate == param.SlotDate && _time.SlotStartTime == param.SlotStartTime && _time.SlotEndTime == param.SlotEndTime {
				slotUserNum = info.SlotUserNum
				if len(info.TeacherTeam) > 0 {
					slotUserNum = slotUserNum * len(info.TeacherTeam)
				}
				break
			}
		}
		if slotUserNum != -1 {
			break
		}
	}
	if slotUserNum == -1 {
		return log, errors.New("所选时间区间不存在")
	}

	// 加锁 防止多报
	lockKey := fmt.Sprintf("%s-%s-%s-%s", param.ExamId, param.SlotDate, param.SlotStartTime, param.SlotEndTime)
	if locked, _ := helper.RedisSetNx(lockKey, 5); !locked {
		if retry <= 0 {
			return log, errors.New("报名太火爆，请稍后重试")
		}
		time.Sleep(1 * time.Second)
		retry--
		return sf.SignUp(uid, param, retry)
	}
	defer helper.RedisDel([]string{lockKey})

	roooms := make([]models.MockExamSlotRoom, 0)
	sf.MockExamSlotRoomModel().Where(bson.M{"slot_date": param.SlotDate, "slot_start_time": param.SlotStartTime, "slot_end_time": param.SlotEndTime, "mock_exam_id": param.ExamId}).Find(&roooms)
	rooomUsers := 0
	for _, rooom := range roooms {
		rooomUsers += len(rooom.Users)
	}
	if rooomUsers >= slotUserNum {
		return log, errors.New("所选时间区间没有空余的报名名额")
	}
	log.UserId = uid
	log.MockExamId = param.ExamId
	log.ExamCategory = exam.ExamCategory
	log.ExamChildCategory = exam.ExamChildCategory
	log.ExamMode = exam.ExamMode
	log.SlotDate = param.SlotDate
	log.SlotStartTime = param.SlotStartTime
	log.SlotEndTime = param.SlotEndTime
	log.Question = make([]models.MockQuestion, 0)

	room, err := sf.assignRoom(exam, log)
	if err != nil {
		return log, err
	}

	log.TeacherTeam = room.TeacherTeam
	log.RoomId = room.Id.Hex()
	if hasLog {
		_, err = sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(log)
		sf.MockExamModel().Where(bson.M{"_id": exam.Id}).Update(bson.M{"updated_time": time.Now().Format("2006-01-02 15:04:05"), "sign_up_user_num": len(room.Users)})
	} else {
		_, err = sf.MockExamLogModel().Create(log)
		sf.MockExamModel().Where(bson.M{"_id": exam.Id}).UpdateAtom(bson.M{"$inc": bson.M{"sign_up_user_num": 1}, "$set": bson.M{"updated_time": time.Now().Format("2006-01-02 15:04:05")}})
	}

	return log, err
}

// 分配房间
func (sf *MockExam) assignRoom(exam *models.MockExam, log *models.MockExamLog) (*models.MockExamSlotRoom, error) {
	var err error
	var rooms = make([]models.MockExamSlotRoom, 0)
	sf.MockExamSlotRoomModel().Where(bson.M{
		"mock_exam_id": log.MockExamId, "slot_date": log.SlotDate, "slot_start_time": log.SlotStartTime,
		"slot_end_time": log.SlotEndTime,
	}).Find(&rooms)
	userMap := new(models.InterviewGPT).GetUsersInfo([]string{log.UserId}, "402", 1)
	room := new(models.MockExamSlotRoom)
	// 没有房间 创建房间
	if len(rooms) == 0 {
		room.Users = make([]models.MockExamUser, 0)
		room.Users = append(room.Users, models.MockExamUser{
			UserId:   log.UserId,
			UserName: userMap[log.UserId].Nickname,
			Avatar:   userMap[log.UserId].Avatar,
		})
		room.ExamType = exam.ExamType
		room.MockExamId = log.MockExamId
		room.SlotDate = log.SlotDate
		room.SlotStartTime = log.SlotStartTime
		room.SlotEndTime = log.SlotEndTime

		teacherTeam := make([]models.MockExamUser, 0)
		// 选择老师团队
		for _, info := range exam.ExamInfo {
			for _, slotTime := range info.SlotTimeList {
				if slotTime.SlotDate == room.SlotDate && slotTime.SlotStartTime == room.SlotStartTime && slotTime.SlotEndTime == room.SlotEndTime && len(info.TeacherTeam) > 0 {
					teacherTeam = info.TeacherTeam[0]
					break
				}
			}
		}
		if exam.ExamType == 0 && len(teacherTeam) == 0 {
			return room, errors.New("该时间段没有空闲老师")
		}
		room.TeacherTeam = teacherTeam
		_, err = sf.MockExamSlotRoomModel().Create(room)

	} else {
		// 判断看是否需要创建房间
		var examInfo models.ExamInfo
		for _, info := range exam.ExamInfo {
			for _, slotTime := range info.SlotTimeList {
				if slotTime.SlotDate == log.SlotDate && slotTime.SlotStartTime == log.SlotStartTime && slotTime.SlotEndTime == log.SlotEndTime {
					examInfo = info
					break
				}
			}
		}
		if examInfo.StartTime == "" {
			return room, errors.New("该时间段没有安排考试")
		}

		hasRoom := false
		for _, i2 := range rooms {
			if len(i2.Users) < examInfo.SlotUserNum {
				inRoom := false
				for _, user := range i2.Users {
					if user.UserId == log.UserId {
						inRoom = true
						break
					}
				}

				hasRoom = true
				if !inRoom {
					i2.Users = append(i2.Users, models.MockExamUser{
						UserId:   log.UserId,
						UserName: userMap[log.UserId].Nickname,
						Avatar:   userMap[log.UserId].Avatar,
					})
				}
				room = &i2
				break
			}
			if hasRoom {
				break
			}
		}

		if !hasRoom && len(rooms) >= len(examInfo.TeacherTeam) {
			return room, errors.New("该时间段考试已排满")
		}

		if !hasRoom {
			room.Users = make([]models.MockExamUser, 0)
			room.Users = append(room.Users, models.MockExamUser{
				UserId:   log.UserId,
				UserName: userMap[log.UserId].Nickname,
				Avatar:   userMap[log.UserId].Avatar,
			})
			room.MockExamId = log.MockExamId
			room.SlotDate = log.SlotDate
			room.SlotStartTime = log.SlotStartTime
			room.SlotEndTime = log.SlotEndTime
			room.ExamType = exam.ExamType

			// 选择老师团队
			hasUsedTeacherIds := make([]string, 0)
			for _, slotRoom := range rooms {
				for _, user := range slotRoom.TeacherTeam {
					hasUsedTeacherIds = append(hasUsedTeacherIds, user.UserId)
				}
			}

			teacherTeam := make([][]models.MockExamUser, 0)
			for _, info := range exam.ExamInfo {
				for _, slotTime := range info.SlotTimeList {
					if slotTime.SlotDate == room.SlotDate && slotTime.SlotStartTime == room.SlotStartTime && slotTime.SlotEndTime == room.SlotEndTime {
						teacherTeam = info.TeacherTeam
						break
					}
				}
			}
			teachers := make([]models.MockExamUser, 0)
			for _, users := range teacherTeam {
				if !common.InArrCommon[string](users[0].UserId, hasUsedTeacherIds) {
					teachers = users
					break
				}
			}
			if exam.ExamType == 0 && len(teachers) == 0 {
				return room, errors.New("该时间段没有空闲老师")
			}
			room.TeacherTeam = teachers
		}

		if hasRoom {
			_, err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": room.Id}).Update(room)
		} else {
			_, err = sf.MockExamSlotRoomModel().Create(room)
		}
	}
	// 如果之前关联的房间，从房间中去除
	if log.RoomId != "" && room.Id.Hex() != log.RoomId {
		oldRoom := new(models.MockExamSlotRoom)
		err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(log.RoomId)}).Take(oldRoom)
		if err == nil {
			users := make([]models.MockExamUser, 0)
			for _, user := range oldRoom.Users {
				if user.UserId != log.UserId {
					users = append(users, user)
				}
			}
			_, _ = sf.MockExamSlotRoomModel().Where(bson.M{"_id": oldRoom.Id}).Update(bson.M{"users": users})
		}
	}
	return room, err
}

// 签到
func (sf *MockExam) SignIn(uid, mockExamId string) error {
	var exam = new(models.MockExam)
	err := sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(mockExamId)}).Take(exam)
	if err != nil {
		return err
	}
	var log = new(models.MockExamLog)
	err = sf.MockExamLogModel().Where(bson.M{"user_id": uid, "mock_exam_id": exam.Id.Hex(), "status": bson.M{"$ne": -1}}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	if err != nil && sf.MongoNoResult(err) {
		return errors.New("未报名")
	}

	if log.Status > 1 {
		return nil
	}
	log.Status = 1
	_, err = sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(map[string]any{"status": 1, "updated_time": time.Now().Format("2006-01-02 15:04:05")})

	return err
}

// Cancel
func (sf *MockExam) Cancel(uid, mockExamId string) error {
	var exam = new(models.MockExam)
	err := sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(mockExamId)}).Take(exam)
	if err != nil {
		return err
	}
	var log = new(models.MockExamLog)
	err = sf.MockExamLogModel().Where(bson.M{"user_id": uid, "mock_exam_id": exam.Id.Hex(), "status": bson.M{"$ne": -1}}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	if err != nil && sf.MongoNoResult(err) {
		return errors.New("未报名")
	}

	// 老师已批改 不走取消
	if log.Status == 5 {
		return nil
	}
	_, err = sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(map[string]any{"status": -1, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
	sf.MockExamModel().Where(bson.M{"_id": exam.Id}).UpdateAtom(bson.M{"$inc": bson.M{"sign_up_user_num": -1}, "$set": bson.M{"updated_time": time.Now().Format("2006-01-02 15:04:05")}})

	var room = new(models.MockExamSlotRoom)
	err = sf.MockExamSlotRoomModel().Where(bson.M{"mock_exam_id": mockExamId, "users.user_id": uid}).Take(room)
	if err == nil {
		users := make([]models.MockExamUser, 0)
		for _, user := range room.Users {
			if user.UserId != uid {
				users = append(users, user)
			}
		}
		room.Users = users
		_, err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": room.Id}).Update(room)
	}

	return err
}

// 订阅
func (sf *MockExam) Subscribe(userId, openId, mockExamId string) error {
	var exam = new(models.MockExam)
	err := sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(mockExamId)}).Take(exam)
	if err != nil {
		return err
	}
	var log = new(models.MockExamLog)
	err = sf.MockExamLogModel().Where(bson.M{"user_id": userId, "mock_exam_id": exam.Id.Hex(), "status": bson.M{"$ne": -1}}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	if err != nil && sf.MongoNoResult(err) {
		return errors.New("未报名")
	}
	if log.Status > 1 {
		return errors.New("模考已在进行中，无需订阅")
	}
	if log.SubscribeStatus > 0 {
		return nil
	}

	key := fmt.Sprintf("%s%s", string(rediskey.MockExamSubscribe), mockExamId)
	err = helper.RedisHSet(key, userId, []byte(openId))
	if err != nil {
		return err
	}
	var examEndTime int64
	for _, info := range exam.ExamInfo {
		endTime, _ := common.LocalTimeFromDateString(info.EndTime + ":00")
		fmt.Println(info.EndTime, endTime.Unix())
		if examEndTime < endTime.Unix() {
			examEndTime = endTime.Unix()
		}
	}
	//考试结束时间+2小时
	expireTime := int(examEndTime + 7200 - time.Now().Unix())
	if expireTime > 0 {
		helper.RedisEXPIRE(key, expireTime)
		log.SubscribeStatus = 1
		log.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
		_, err = sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(log)
	} else {
		return errors.New("考试已结束，无需订阅")
	}

	return err
}

func (sf *MockExam) TeacherControl(uid string, param request.TeacherControlRequest) error {
	room := new(models.MockExamSlotRoom)
	err := sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(param.RoomId)}).Take(room)
	if err != nil {
		return err
	}
	exam := new(models.MockExam)
	err = sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(room.MockExamId)}).Take(exam)
	if err != nil {
		return err
	}

	//isTeacher := false
	//for _, info := range exam.ExamInfo {
	//	for _, teams := range info.TeacherTeam {
	//		for _, team := range teams {
	//			if uid == team.UserId {
	//				isTeacher = true
	//				break
	//			}
	//		}
	//		if isTeacher {
	//			break
	//		}
	//	}
	//	if isTeacher {
	//		break
	//	}
	//}
	//if !isTeacher {
	//	return errors.New("非老师无法修改")
	//}
	cacheKey := fmt.Sprintf("%s%s", string(rediskey.MockExamStatus), param.RoomId)
	controlStr, _ := helper.RedisHGet(cacheKey, "control")
	var control = make(map[string]map[string]int)
	if controlStr != "" {
		json.Unmarshal([]byte(controlStr), &control)
	}
	for userId, actArr := range param.Action {
		if _, ok := control[userId]; !ok {
			control[userId] = actArr
		} else {
			for act, val := range actArr {
				control[userId][act] = val
			}
		}
	}

	str, _ := json.Marshal(control)
	helper.RedisHSet(cacheKey, "control", str)
	helper.RedisEXPIRE(cacheKey, 14400)

	return err
}

func (sf *MockExam) GetTeacherControl(uid, roomId string) map[string]int {
	cacheKey := fmt.Sprintf("%s%s", string(rediskey.MockExamStatus), roomId)
	controlStr, _ := helper.RedisHGet(cacheKey, "control")
	var control = make(map[string]map[string]int)
	var userControl = make(map[string]int)
	if controlStr != "" {
		json.Unmarshal([]byte(controlStr), &control)
	}
	if _, ok := control[uid]; ok {
		userControl = control[uid]
	}

	return userControl
}

// info
func (sf *MockExam) Info(uid, mockExamId, roomId string, getDetail string) (*models.MockExam, error) {
	var exam = new(models.MockExam)
	err := sf.MockExamModel().Where(bson.M{"_id": sf.ObjectID(mockExamId)}).Take(exam)
	if err != nil {
		sf.SLogger().Error(err)
		return exam, err
	}
	exam.TemplateId = GetWeChatMockExamTemplateId()
	if getDetail != "1" {
		var logs = make([]models.MockExamLog, 0)
		err = sf.MockExamLogModel().Where(bson.M{"mock_exam_id": mockExamId, "status": bson.M{"$ne": -1}}).Find(&logs)
		logMap := make(map[string]int)
		for _, log := range logs {
			if _, ok := logMap[log.SlotDate+log.SlotStartTime+log.SlotEndTime]; !ok {
				logMap[log.SlotDate+log.SlotStartTime+log.SlotEndTime] = 0
			}
			logMap[log.SlotDate+log.SlotStartTime+log.SlotEndTime]++
		}
		for i, info := range exam.ExamInfo {
			for i2, slotTime := range info.SlotTimeList {
				if unum, ok := logMap[slotTime.SlotDate+slotTime.SlotStartTime+slotTime.SlotEndTime]; ok {
					exam.ExamInfo[i].SlotTimeList[i2].SignUpUserNum = unum
				}
			}

			exam.ExamInfo[i].RoomRole = "student"
			isTeacher := false
			for _, users := range info.TeacherTeam {
				for _, user := range users {
					if user.UserId == uid {
						isTeacher = true
						break
					}
				}
				if isTeacher {
					break
				}
			}
			if isTeacher {
				exam.ExamInfo[i].RoomRole = "teacher"
			}
		}
	}
	// 各个时间段人数

	// 处理报名信息， 几个人报名，各个时间段的报名人数
	var log = new(models.MockExamLog)
	err = sf.MockExamLogModel().Where(bson.M{"user_id": uid, "mock_exam_id": mockExamId, "status": bson.M{"$ne": -1}}).Take(log)
	if err == nil {
		exam.Log = log
		cacheKey := string(rediskey.MockExamStatus) + log.RoomId
		status, _ := helper.RedisHGetAll(cacheKey)
		if _, ok := status["start"]; ok {
			exam.RoomStatus = "start"
		}
		if _, ok := status["close"]; ok {
			exam.RoomStatus = "close"
		}
		if log.Status == 5 {
			exam.RoomStatus = "close"
		}
	}
	if getDetail == "1" {
		// 房间详情
		var room = new(models.MockExamSlotRoom)
		err = sf.MockExamSlotRoomModel().Where(bson.M{"mock_exam_id": mockExamId, "_id": sf.ObjectID(roomId)}).Take(room)
		if err != nil {
			sf.SLogger().Error(err)
			return exam, err
		}
		// 题目详情
		var questionIds = make(map[string]models.MockQuestion)
		var questions = make([]models.GQuestion, 0)
		var questions2 = make([]models.GQuestion, 0)

		for _, question := range exam.Question {
			questionIds[question.QuestionId] = question
		}
		sf.GQuestionModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(common.GetMapKeys(questionIds))}}).Find(&questions)

		for i := range questions {
			questions[i].RequireAnswerDuration = questionIds[questions[i].Id.Hex()].RequireAnswerDuration
		}
		// 排序
		for _, question := range exam.Question {
			for _, gQuestion := range questions {
				if gQuestion.Id.Hex() == question.QuestionId {
					questions2 = append(questions2, gQuestion)
					break
				}
			}
		}

		exam.Room = room
		exam.QuestionDetail = questions2
		// 每场考试的题总数
		exam.QuestionNum = 3
		if len(exam.QuestionDetail) < exam.QuestionNum {
			exam.QuestionNum = len(exam.QuestionDetail)
		}
	}
	return exam, nil
}

func (sf *MockExam) SlotTime(startTime, endTime time.Time, examTime, restTime int) []models.ExamSlotTime {
	eachtime := examTime + restTime
	diffTimeFloat := endTime.Sub(startTime).Seconds()
	slotNum := int(diffTimeFloat) / eachtime
	slotArr := make([]models.ExamSlotTime, 0)
	for i := 0; i < slotNum; i++ {
		slotStartTime := startTime.Add(time.Duration(i*eachtime) * time.Second)
		slotEndTime := slotStartTime.Add(time.Duration(examTime) * time.Second)
		slotArr = append(slotArr, models.ExamSlotTime{
			SlotDate:      startTime.Format("2006-01-02"),
			SlotEndTime:   slotEndTime.Format("15:04:05"),
			SlotStartTime: slotStartTime.Format("15:04:05"),
		})
	}
	return slotArr
}

func (sf *MockExam) IsTeacher(uid string, examId string) int {
	exam := new(models.MockExam)
	//err := sf.MockExamModel().Where(bson.M{"exam_info.teacher_team.user_id": uid, "status": 1}).Take(exam)
	var filter = bson.M{"exam_info.teacher_team": bson.M{"$elemMatch": bson.M{"$elemMatch": bson.M{"user_id": uid}}}, "status": 1}
	if examId != "" {
		filter["_id"] = sf.ObjectID(examId)
	}
	err := sf.MockExamModel().Where(filter).Take(exam)
	if err != nil {
		sf.SLogger().Error(err)
		return 0
	}
	return 1
}

// LogList
func (sf *MockExam) LogList(uid string, offset, limit int64) ([]models.MockExamSlotRoom, int64, error) {
	var filter = bson.M{}
	filter["$or"] = bson.A{bson.M{"teacher_team.user_id": uid, "users.0": bson.M{"$exists": true}},
		bson.M{"users.user_id": uid},
		//bson.M{"video_arr.users.user_id": uid, "video_arr.video_url": bson.M{"$ne": ""}},
		bson.M{"video_arr": bson.M{"$elemMatch": bson.M{"users.user_id": uid, "video_url": bson.M{"$ne": ""}}}},
	}
	var isTeacher = sf.IsTeacher(uid, "")
	//if isTeacher == 1 {
	//	filter = bson.M{"teacher_team.user_id": uid, "users.0": bson.M{"$exists": true}}
	//} else {
	//	filter = bson.M{"users.user_id": uid}
	//}
	rooms := make([]models.MockExamSlotRoom, 0)
	total, err := sf.MockExamSlotRoomModel().Where(filter).Count()
	if err != nil {
		return nil, total, err
	}
	err = sf.MockExamSlotRoomModel().Where(filter).Sort([]string{"-slot_date", "-slot_end_time"}...).Skip(offset).Limit(limit).Find(&rooms)
	if err != nil {
		sf.SLogger().Error(err)
		return nil, total, err
	}
	var mockExamIds = make([]string, 0)
	var roomIds = make([]string, 0)
	for _, exam := range rooms {
		mockExamIds = append(mockExamIds, exam.MockExamId)
		roomIds = append(roomIds, exam.Id.Hex())
	}
	exams := make([]models.MockExam, 0)
	err = sf.MockExamModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(mockExamIds)}}).Find(&exams)
	if err != nil {
		return rooms, total, err
	}
	examMap := make(map[string]models.MockExam)
	for _, exam := range exams {
		examMap[exam.Id.Hex()] = exam
	}

	logs := make([]models.MockExamLog, 0)
	var logFilter = bson.M{"room_id": bson.M{"$in": roomIds}, "status": bson.M{"$ne": -1}}
	if isTeacher != 1 {
		logFilter["user_id"] = uid
	}
	err = sf.MockExamLogModel().Where(logFilter).Find(&logs)
	if err != nil {
		return rooms, total, err
	}
	userIdMap := make(map[string]int)
	logMap := make(map[string][]models.MockExamLog)
	for _, log := range logs {
		if _, ok := logMap[log.RoomId]; !ok {
			logMap[log.RoomId] = make([]models.MockExamLog, 0)
		}
		logMap[log.RoomId] = append(logMap[log.RoomId], log)
		userIdMap[log.UserId] = 1
	}
	userService := new(User).GetGatewayUsersInfo(common.GetMapKeys(userIdMap), "402", 1)

	for i := range rooms {
		exam, ok := examMap[rooms[i].MockExamId]
		if !ok {
			continue
		}
		rooms[i].Title = exam.Title
		rooms[i].SubTitle = exam.SubTitle
		rooms[i].Introduce = exam.Introduce
		rooms[i].ExamCategory = exam.ExamCategory
		rooms[i].ExamChildCategory = exam.ExamChildCategory
		rooms[i].ExamMode = exam.ExamMode
		rooms[i].ExamStatus = exam.Status
		rooms[i].MaxUserNum = exam.ExamInfo[0].SlotUserNum
		if exam.ExamType == 0 {
			rooms[i].MaxUserNum++
		}

		userList := make([]models.MockExamUser, 0)
		if _, ok = logMap[rooms[i].Id.Hex()]; ok {
			for _, log := range logMap[rooms[i].Id.Hex()] {
				if uinfo, ok := userService[log.UserId]; ok {
					userList = append(userList, models.MockExamUser{
						UserId:    log.UserId,
						UserName:  uinfo.Nickname,
						Avatar:    uinfo.Avatar,
						LogStatus: log.Status,
					})
				}
			}
		}
		rooms[i].Logs = userList

		rooms[i].RoomRole = "student"
		isTeacher := false
		for _, user := range rooms[i].TeacherTeam {
			if user.UserId == uid {
				isTeacher = true
				break
			}
		}
		if isTeacher {
			rooms[i].RoomRole = "teacher"
		}

		// 处理视频
		if len(rooms[i].VideoArr) > 1 {
			tmpVideoUrls := make([]string, 0)
			for _, video := range rooms[i].VideoArr {
				for _, user := range video.Users {
					if user.UserId == uid && video.VideoUrl != "" {
						tmpVideoUrls = append(tmpVideoUrls, video.VideoUrl)
						break
					}
				}
			}
			if len(tmpVideoUrls) > 0 {
				rooms[i].VideoUrl = strings.Join(tmpVideoUrls, ",")
			}
		}
	}

	return rooms, total, err
}

func (sf *MockExam) LogDetail(uid, roomId string) ([]models.MockExamLog, error) {
	var logs = make([]models.MockExamLog, 0)
	var filter = bson.M{"room_id": roomId, "status": bson.M{"$ne": -1}}

	filter["$or"] = bson.A{bson.M{"teacher_team.user_id": uid}, bson.M{"user_id": uid}}
	//filter["teacher_team.user_id"] = uid
	//filter["user_id"] = uid
	err := sf.MockExamLogModel().Where(filter).Find(&logs)
	if err != nil {
		return nil, err
	}

	comments := sf.GetMarkComment()
	commentKeyArr := make([]string, 0)
	for _, comment := range comments {
		commentKeyArr = append(commentKeyArr, comment.Key)
	}
	roomIds := make([]string, 0)
	examIds := make([]string, 0)
	for i := range logs {
		logs[i].TeacherCorrect.MarkContentSort = commentKeyArr
		roomIds = append(roomIds, logs[i].RoomId)
		examIds = append(examIds, logs[i].MockExamId)
		cacheKey := string(rediskey.MockExamStatus) + logs[i].RoomId
		status, _ := helper.RedisHGetAll(cacheKey)
		if _, ok := status["start"]; ok {
			logs[i].RoomStatus = "start"
		}
		if _, ok := status["close"]; ok {
			logs[i].RoomStatus = "close"
		}
		if logs[i].Status == 5 {
			logs[i].RoomStatus = "close"
		}
	}

	exams := make([]*models.MockExam, 0)
	_, logs = sf.getQuestionDetail(exams, logs)

	rooms := make([]models.MockExamSlotRoom, 0)
	roomMap := make(map[string]models.MockExamSlotRoom)
	sf.MockExamSlotRoomModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(roomIds)}}).Find(&rooms)
	for _, room := range rooms {
		roomMap[room.Id.Hex()] = room
	}

	sf.MockExamSlotRoomModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(roomIds)}}).Find(&rooms)

	for i := range logs {
		if _, ok := roomMap[logs[i].RoomId]; ok {
			logs[i].Room = roomMap[logs[i].RoomId]
			if logs[i].Room.Status == 1 {
				logs[i].RoomStatus = "close"
			}
		}

		logs[i].RoomRole = "student"
		isTeacher := false
		for _, user := range logs[i].TeacherTeam {
			if user.UserId == uid {
				isTeacher = true
				break
			}
		}
		if isTeacher {
			logs[i].RoomRole = "teacher"
		}

		if len(logs[i].Question) > 0 {
			// 排序
			questions := make([]models.GQuestion, 0)
			questionIds := make([]string, 0)
			for _, info := range logs[i].Question {
				questionIds = append(questionIds, info.QuestionId)
			}
			questionIds = common.RemoveDuplicateElement(questionIds)
			for _, qid := range questionIds {
				for _, question := range logs[i].QuestionDetail {
					if question.Id.Hex() == qid {
						questions = append(questions, question)
						break
					}
				}
			}
			logs[i].QuestionDetail = questions
		}
	}

	return logs, err
}

func (sf *MockExam) GetMarkComment() []CorrectCommentItem {
	var comment = make([]CorrectCommentItem, 0)
	comment = append(comment, CorrectCommentItem{
		Key:    "思考时间",
		Title:  "思考时间",
		Values: []string{"较长", "一般", "较短"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "答题时间",
		Title:  "答题时间",
		Values: []string{"超时", "未超时"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "举止仪表",
		Title:  "举止仪表",
		Values: []string{"好", "中", "差"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "第一感觉",
		Title:  "第一感觉",
		Values: []string{"自信", "气势弱"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "语言表达",
		Title:  "语言表达",
		Values: []string{"非常流畅", "流畅", "卡顿"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "音量",
		Title:  "音量",
		Values: []string{"大", "适中", "小"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "语速",
		Title:  "语速",
		Values: []string{"快", "适中", "慢"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "语调",
		Title:  "语调",
		Values: []string{"抑扬顿挫", "较平"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "互动",
		Title:  "互动",
		Values: []string{"强", "一般", "弱"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "亲和力",
		Title:  "亲和力",
		Values: []string{"强", "一般", "弱"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "逻辑",
		Title:  "逻辑",
		Values: []string{"逻辑清晰", "一般", "逻辑混乱"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "内容",
		Title:  "内容",
		Values: []string{"跑题", "偏题", "准确", "优秀", "共情"},
	})

	return comment
}

// TeacherMark
func (sf *MockExam) TeacherMark(teacherId string, requestParam request.TeacherMarkRequest) error {
	log := new(models.MockExamLog)
	err := sf.MockExamLogModel().Where(bson.M{"room_id": requestParam.RoomId, "user_id": requestParam.UserId, "status": bson.M{"$ne": -1}}).Take(log)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	log.TeacherCorrect = requestParam.TeacherCorrect
	log.Status = 5

	_, err = sf.MockExamLogModel().Where(bson.M{"_id": log.Id}).Update(log)
	if err != nil {
		sf.SLogger().Error(err)
	}
	return err
}

// 录制过程中 中途加入新成员
func (sf *MockExam) VideoStartMidwayJoin(roomId, uid string) error {
	room := new(models.MockExamSlotRoom)
	err := sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Take(room)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}

	if room.VideoTaskId == "" {
		return nil
	}
	newVideoUser := models.MockExamUser{UserId: uid}
	for _, user := range room.Users {
		if user.UserId == uid {
			newVideoUser = user
		}
	}

	for i, video := range room.VideoArr {
		if video.VideoTaskId == room.VideoTaskId {
			isIn := false
			for _, user := range video.Users {
				if user.UserId == newVideoUser.UserId {
					isIn = true
					break
				}
			}
			if !isIn {
				video.Users = append(video.Users, newVideoUser)
				room.VideoArr[i] = video
			}
			break
		}
	}

	_, err = sf.MockExamSlotRoomModel().Where(bson.M{"_id": sf.ObjectID(roomId)}).Update(room)
	return err
}
