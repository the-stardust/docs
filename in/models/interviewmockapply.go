package models

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"interview/database"
)

const (
	interviewMockApplyTableName = "interview_mock_apply"
	statusApply                 = 1
	statusUnApply               = 2
)

var ImaModel = &imaModel{}

type imaModel struct {
	DefaultStruct
}

type InterviewMockApply struct {
	DefaultField      `bson:",inline"`
	UserID            string          `json:"user_id" bson:"user_id"`
	UserName          string          `json:"user_name" bson:"user_name"`
	ClassID           string          `json:"class_id" bson:"class_id"`
	ClassName         string          `json:"class_name" bson:"class_name"`
	Mobile            string          `json:"mobile" bson:"mobile"`
	InterviewMockID   string          `json:"interview_mock_id" bson:"interview_mock_id"`
	OneByOneTimeType  int             `json:"onebyone_time_type" bson:"onebyone_time_type"`
	SimulateTimeType  int             `json:"simulate_time_type" bson:"simulate_time_type"`
	Status            int             `json:"status"`
	QuestionnaireInfo []Questionnaire `json:"questionnaire_info" bson:"questionnaire_info"`
}

func (i *imaModel) getCollection() *database.MongoWork {
	return i.DB().Collection(interviewMockApplyTableName)
}

func (i *imaModel) Create(param InterviewMockApply) error {
	if param.Id != primitive.NilObjectID {
		return fmt.Errorf("id is not nil")
	}
	param.Id = primitive.NewObjectID()
	param.Status = statusApply
	_, err := i.getCollection().Create(&param)
	return err
}

func (i *imaModel) CheckRepeatApply(interviewMockID, userID, userName, mobile string) bool {
	var info InterviewMockApply
	f := bson.M{
		"status":            statusApply,
		"interview_mock_id": interviewMockID,
		"$or": []bson.M{
			{"mobile": mobile, "user_name": userName},
			{"user_id": userID},
		},
	}
	err := i.getCollection().Where(f).Take(&info)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return false
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false
	}
	return true
}

func (i *imaModel) GetApplyInfoByUser(userID, interviewMockID string) (*InterviewMockApply, error) {
	if userID == "" {
		return nil, fmt.Errorf("un login")
	}
	var info InterviewMockApply
	err := i.getCollection().Where(bson.M{"user_id": userID, "status": statusApply, "interview_mock_id": interviewMockID}).Take(&info)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &info, nil
	}
	return &info, nil
}

func (i *imaModel) ApplyCancel(userID, interviewMockID string) error {
	info, err := i.GetApplyInfoByUser(userID, interviewMockID)
	if err != nil {
		return err
	}
	if info == nil {
		return nil
	}
	info.Status = statusUnApply
	err = i.getCollection().Save(info)
	return err
}

func (i *imaModel) GetApplyListByFilter(f bson.M) ([]InterviewMockApply, error) {
	var list []InterviewMockApply
	f["status"] = statusApply
	err := i.getCollection().Where(f).Sort("+class_id").Find(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetApplyListMap 获取以interview_mock_id 为key的map
func (i *imaModel) GetApplyListMap(mockIDs []string) (map[string]InterviewMockApply, error) {
	var list []InterviewMockApply
	if len(mockIDs) == 0 {
		return nil, nil
	}
	filter := bson.M{"interview_mock_id": bson.M{"$in": mockIDs}, "status": statusApply}
	err := i.getCollection().Where(filter).Find(&list)
	if err != nil {
		return nil, err
	}
	res := make(map[string]InterviewMockApply)
	for _, item := range list {
		res[item.InterviewMockID] = item
	}
	return res, nil
}

func (i *imaModel) GetUserApplyList(userID string) ([]primitive.ObjectID, error) {
	var list []InterviewMockApply
	f := bson.M{"status": statusApply, "user_id": userID}
	err := i.getCollection().Where(f).Find(&list)
	if err != nil {
		return nil, err
	}
	mockIDs := make([]primitive.ObjectID, 0, len(list))
	for _, item := range list {
		mockIDs = append(mockIDs, i.ObjectID(item.InterviewMockID))
	}
	return mockIDs, nil
}
