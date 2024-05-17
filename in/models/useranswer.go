package models

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"interview/common"
	"interview/database"
	"strings"
	"time"
)

type UserAnswer struct {
	DefaultField        `bson:",inline"`
	UserID              string `json:"user_id" bson:"user_id"`
	ExamCategory        string `json:"exam_category" bson:"exam_category"`
	ExamChildCategory   string `json:"exam_child_category" bson:"exam_child_category"`
	QuestionCategoryStr string `json:"question_category_str" bson:"question_category_str"`
	QuestionID          string `json:"question_id" bson:"question_id"`
	Provence            string `json:"provence" bson:"provence"`
	JobTag              string `json:"job_tag" bson:"job_tag"`
	QuestionRealType    int    `json:"question_real_type" bson:"question_real_type"`
}

type QuestionSimpleItem struct {
	QuestionID string `json:"question_id" bson:"question_id"`
	CreateTime string `json:"create_time" bson:"create_time"`
	Year       int    `json:"year" bson:"year"`   // 年份
	Month      int    `json:"month" bson:"month"` // 月份
	Day        int    `json:"day" bson:"day"`     // 日
}

type UserRecordValue struct {
	LastViewQuestionID             string `json:"last_view_question_id"`
	LastViewQuestionIDIndex        int    `json:"last_view_question_id_index"`
	CurrCategoryMaxCreateTime      int64  `json:"curr_category_max_create_time"`
	CurrCategoryMaxCreateTimeQID   string `json:"curr_category_max_create_time_qid"`
	CurrCategoryMaxCreateTimeIndex int    `json:"curr_category_max_create_time_index"`
}

// 前缀树 插入
func (k *KeypointStatisticsResp) FindChild(word string) int {
	if k.Child == nil || len(k.Child) == 0 {
		return -1
	}
	res := -1
	for index, child := range k.Child {
		if child.Title == word {
			res = index
			break
		}
	}
	return res
}

const (
	UserAnswerTable = "g_interview_user_answer"
)

func (u *UserAnswer) getCollection() *database.MongoWork {
	return u.DB().Collection(UserAnswerTable)
}

func (u *UserAnswer) GetUserMap(uid, examCategory, examChildCategory, jobTag, provence string, questionRealInt int) (map[string][]string, error) {
	res := make(map[string][]string)
	filter := bson.M{"exam_category": examCategory}
	if examChildCategory != "" {
		filter["exam_child_category"] = examChildCategory
	}
	if jobTag != "" {
		filter["job_tag"] = jobTag
	}
	if provence != "" {
		filter["provence"] = provence
	}
	filter["question_real_type"] = questionRealInt
	filter["user_id"] = uid
	var list []UserAnswer
	err := u.getCollection().Where(filter).Find(&list)
	if err != nil {
		return res, err
	}
	for _, v := range list {
		if _, ok := res[v.QuestionCategoryStr]; ok {
			res[v.QuestionCategoryStr] = common.AppendSet(res[v.QuestionCategoryStr], []string{v.QuestionID})
		} else {
			res[v.QuestionCategoryStr] = []string{v.QuestionID}
		}
	}
	return res, nil
}

func (u *UserAnswer) CreateLog(answerLog GAnswerLog, question GQuestion) error {
	if len(question.QuestionCategory) == 0 {
		return nil
	}
	filter := bson.M{
		"user_id":     answerLog.UserId,
		"question_id": question.Id.Hex(),
	}
	var info UserAnswer
	err := u.getCollection().Where(filter).Take(&info)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if !info.Id.IsZero() {
		return nil
	}
	ins := UserAnswer{
		UserID:              answerLog.UserId,
		ExamCategory:        question.ExamCategory,
		ExamChildCategory:   question.ExamChildCategory,
		QuestionCategoryStr: strings.Join(question.QuestionCategory, "_"),
		QuestionID:          question.Id.Hex(),
		Provence:            question.Province,
		JobTag:              question.JobTag,
		QuestionRealType:    int(question.QuestionReal),
	}

	_, err = u.getCollection().Create(&ins)
	return err
}

func (u *UserAnswer) GetUserAnswerQID(uid, examCate, examChildCate, provence, jobTag string, questionCate []string, realType int) []string {
	var res []string
	if uid == "" {
		return res
	}
	filter := bson.M{"user_id": uid}
	if examCate != "" {
		filter["exam_category"] = examCate
	}
	if examChildCate != "" {
		filter["exam_child_category"] = examCate
	}
	if provence != "" {
		filter["provence"] = provence
	}
	if jobTag != "" {
		filter["job_tag"] = jobTag
	}
	if len(questionCate) > 0 {
		filter["question_category_str"] = strings.Join(questionCate, "_")
	}
	filter["question_real_type"] = realType
	var list []UserAnswer
	s := time.Now()
	err := u.getCollection().Where(filter).Find(&list)
	fmt.Println("cost:", time.Since(s).Milliseconds())
	if err != nil {
		return res
	}
	for _, v := range list {
		res = append(res, v.QuestionID)
	}
	return res
}
