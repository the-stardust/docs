package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/database"
	"sync"
	"time"
)

const (
	interviewMockTableName = "interview_mock"
)

var InterviewMockModel = &interviewMockModel{}

type interviewMockModel struct {
	DefaultStruct
}

const (
	MockStatusNotBegin = iota
	MockStatusApply
	MockStatusApplyOver
	MockStatusOver
)

type InterviewMock struct {
	DefaultField      `bson:",inline"`
	Title             string          `json:"title"`
	Desc              string          `json:"desc"`
	CurriculaID       string          `json:"curricula_id" bson:"curricula_id"`
	ExamStartTime     string          `json:"exam_start_time" bson:"exam_start_time"`
	ExamEndTime       string          `json:"exam_end_time" bson:"exam_end_time"`
	MockTime          string          `json:"mock_time" bson:"mock_time"`
	OneByOneTimeType  int             `json:"onebyone_time_type" bson:"onebyone_time_type"`
	SimulateTimeType  int             `json:"simulate_time_type" bson:"simulate_time_type"`
	Creator           string          `json:"creator" bson:"creator"`
	QuestionnaireList []Questionnaire `json:"questionnaire_list" bson:"questionnaire_list"`
	IsDeleted         int             `json:"is_deleted" bson:"is_deleted"`
}
type Questionnaire struct {
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	AnswerOptions []string `json:"answer_options" bson:"answer_options"`
}

func (i *InterviewMock) Status() int {
	location, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(location)
	if len(i.MockTime) < 11 {
		i.MockTime = i.MockTime + " 23:59:59"
	}
	start, _ := time.ParseInLocation("2006-01-02 15:04", i.ExamStartTime, location)
	end, _ := time.ParseInLocation("2006-01-02 15:04", i.ExamEndTime, location)
	mock, _ := time.ParseInLocation("2006-01-02 15:04:05", i.MockTime, location)
	if now.Before(start) {
		return MockStatusNotBegin
	} else if now.After(start) && now.Before(end) {
		return MockStatusApply
	} else if now.Before(mock) {
		return MockStatusApplyOver
	}
	return MockStatusOver
}

func (i *interviewMockModel) getCollection() *database.MongoWork {
	return i.DB().Collection(interviewMockTableName)
}

func (i *interviewMockModel) Create(param InterviewMock) error {
	if param.Id != primitive.NilObjectID {
		return fmt.Errorf("id is not nil")
	}
	param.Id = primitive.NewObjectID()
	_, err := i.getCollection().Create(&param)
	return err
}

func (i *interviewMockModel) Update(param InterviewMock) error {
	info, err := i.GetInfo(param.Id)
	if err != nil {
		return err
	}
	info = &param
	err = i.getCollection().Save(info)
	return err
}

func (i *interviewMockModel) Delete(id primitive.ObjectID) error {
	info, err := i.GetInfo(id)
	if err != nil {
		return err
	}
	info.IsDeleted = 1
	err = i.getCollection().Save(info)
	return err
}

func (i *interviewMockModel) GetInfo(id primitive.ObjectID) (*InterviewMock, error) {
	if id == primitive.NilObjectID {
		return nil, fmt.Errorf("id is nil")
	}
	var info InterviewMock
	err := i.getCollection().Where(bson.M{"_id": id, "is_deleted": 0}).Take(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
func (i *interviewMockModel) GetList(f bson.M) ([]InterviewMock, error) {
	var list []InterviewMock
	err := i.getCollection().Where(f).Sort("-created_time").Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (i *interviewMockModel) GetUserCurrIDAndApplyMockID(userID string) ([]string, []primitive.ObjectID) {
	var wg sync.WaitGroup
	var currIDs []string
	var applyMockID []primitive.ObjectID
	// 获取自己所属的考试id
	wg.Add(1)
	go func() {
		defer wg.Done()
		currIDs = InterviewClassModel.GetUserCurriculaIDs(userID)
	}()
	// 获取自己已经报名的模考,可能不属于自己的班级
	wg.Add(1)
	go func() {
		defer wg.Done()
		applyMockID, _ = ImaModel.GetUserApplyList(userID)
	}()
	wg.Wait()
	return currIDs, applyMockID
}

func (i *interviewMockModel) GetUserApplyList(currIDs []string, applyMockID []primitive.ObjectID) ([]InterviewMock, map[string]bool, error) {
	f := bson.M{"is_deleted": 0}
	f["$or"] = []bson.M{
		{"curricula_id": bson.M{"$in": currIDs}},
		{"_id": bson.M{"$in": applyMockID}},
	}
	list, err := InterviewMockModel.GetList(f)
	if err != nil {
		return nil, nil, err
	}
	// make一个 已报名的map，用于前台展示
	applyMap := make(map[string]bool)
	for _, v := range applyMockID {
		applyMap[v.Hex()] = true
	}
	return list, applyMap, nil
}
