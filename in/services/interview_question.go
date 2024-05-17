package services

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/rediskey"
	"interview/helper"
	"interview/models"
	"strings"
	"sync"
	"time"
)

type InterviewQuestionService struct {
	ServicesBase
}

func InitInterviewQuestionService() InterviewQuestionService {
	return InterviewQuestionService{ServicesBase{}}
}
func (s InterviewQuestionService) GetQuestionValues(question models.GQuestion, index int, cate string) string {
	f := bson.M{"exam_category": question.ExamCategory}
	if question.ExamChildCategory != "" {
		f["exam_child_category"] = question.ExamChildCategory
	}
	f["status"] = 5
	f["question_real"] = question.QuestionReal
	if cate != "全部" {
		k := fmt.Sprintf("question_category.%d", index)
		f[k] = cate
	}
	questions := make([]models.GQuestion, 0)
	err := s.DB().Collection("g_interview_questions").Where(f).Sort([]string{"-year", "-month", "-day", "-created_time"}...).Fields(bson.M{"_id": 1, "created_time": 1}).Find(&questions)
	if err != nil {
		return ""
	}
	if len(questions) == 0 {
		return ""
	}
	val := models.UserRecordValue{
		LastViewQuestionID:        question.Id.Hex(),
		LastViewQuestionIDIndex:   0,
		CurrCategoryMaxCreateTime: 0,
	}
	// 找到这个 question 在整个 list 的 index
	for ii, q := range questions {
		if q.Id.Hex() == question.Id.Hex() {
			val.LastViewQuestionIDIndex = ii
			break
		}
	}
	// 记录当前 list 最新的 created_time,用于判断是否有新题
	_, _, val.CurrCategoryMaxCreateTime = FindMaxCreateTimeQuestion(questions)
	valStr, _ := json.Marshal(val)
	return string(valStr)
}

// FindMaxCreateTimeQuestion  返回 qid, qid 在 list 的 index 以及 created_time 的时间戳
func FindMaxCreateTimeQuestion(questions []models.GQuestion) (string, int, int64) {
	maxCreated, err := common.LocalTimeFromDateString(questions[0].CreatedTime)
	if err != nil {
		return "", 0, 0
	}
	qid := questions[0].Id.Hex()
	index := 0
	for j := 1; j < len(questions); j++ {
		currCrTime, err := common.LocalTimeFromDateString(questions[j].CreatedTime)
		if err != nil {
			continue
		}
		if maxCreated.Before(currCrTime) {
			maxCreated = currCrTime
			qid = questions[j].Id.Hex()
			index = j
		}
	}
	return qid, index, maxCreated.Unix()
}

func (s InterviewQuestionService) RecordUserQuestion(uid, jobTag, province string, question models.GQuestion) {
	// 记录 redis, 用户访问最新的 question_id 与 当前的 index
	if len(question.QuestionCategory) == 0 {
		return
	}
	start := time.Now()
	examCate := question.ExamCategory
	if question.ExamChildCategory != "" {
		examCate += "_" + question.ExamChildCategory
	}
	key := fmt.Sprintf(string(rediskey.UserAnswerKeyPointRecord), uid, examCate, jobTag, province, question.QuestionReal)
	type item struct {
		cate string
		val  string
	}
	// 先 get
	resChan := make(chan item, len(question.QuestionCategory))
	var wg sync.WaitGroup
	for index := range question.QuestionCategory {
		wg.Add(1)
		go func(question models.GQuestion, index int) {
			defer wg.Done()
			val := s.GetQuestionValues(question, index, question.QuestionCategory[index])
			allCate := strings.Join(question.QuestionCategory[:index+1], "_")
			resChan <- item{cate: allCate, val: val}
		}(question, index)
	}

	wg.Wait()
	close(resChan)
	values := make(map[string]interface{})
	for res := range resChan {
		values[res.cate] = res.val
	}
	err := helper.RedisHMSet(key, values)
	if err != nil {
		s.SLogger().Error(err)
	}
	fmt.Println("recordUserQuestion cost:", time.Since(start).Milliseconds())
}

// RecordUserKeypoint 记录用户足迹 上次浏览的 qid 与 index
func (s InterviewQuestionService) RecordUserKeypoint(uid string, jobTag, province string, question models.GQuestion) {
	// 记录 redis, 用户访问最新的 question_id 与 当前的 index
	if len(question.QuestionCategory) == 0 {
		return
	}
	start := time.Now()
	examCate := question.ExamCategory
	if question.ExamChildCategory != "" {
		examCate += "_" + question.ExamChildCategory
	}
	key := fmt.Sprintf(string(rediskey.UserAnswerKeyPointRecord), uid, examCate, jobTag, province, question.QuestionReal)
	type item struct {
		cate string
		val  models.UserRecordValue
	}
	// 先 get
	oldRecord, err := helper.RedisHGetAll(key)
	resChan := make(chan item, len(question.QuestionCategory)+1)
	var wg sync.WaitGroup
	for index := range question.QuestionCategory {
		wg.Add(1)
		go func(question models.GQuestion, index int, oldRecord map[string]string) {
			defer wg.Done()
			lastIndex := s.GetCateQuestionIndex(question, question.QuestionCategory[index])
			allCate := strings.Join(question.QuestionCategory[:index+1], "_")
			var oldVal models.UserRecordValue
			if oldCateRecord, ok := oldRecord[allCate]; ok {
				err = json.Unmarshal([]byte(oldCateRecord), &oldVal)
			}
			oldVal.LastViewQuestionID = question.Id.Hex()
			oldVal.LastViewQuestionIDIndex = lastIndex
			resChan <- item{
				cate: allCate,
				val:  oldVal,
			}
		}(question, index, oldRecord)
	}
	// `全部` tab 也要处理
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastIndex := s.GetCateQuestionIndex(question, "全部")
		var oldVal models.UserRecordValue
		if oldCateRecord, ok := oldRecord["全部"]; ok {
			err = json.Unmarshal([]byte(oldCateRecord), &oldVal)
		}
		oldVal.LastViewQuestionID = question.Id.Hex()
		oldVal.LastViewQuestionIDIndex = lastIndex
		resChan <- item{
			cate: "全部",
			val:  oldVal,
		}
	}()

	wg.Wait()
	close(resChan)
	values := make(map[string]interface{})
	for res := range resChan {
		valByte, _ := json.Marshal(res.val)
		values[res.cate] = string(valByte)
	}
	err = helper.RedisHMSet(key, values)
	if err != nil {
		s.SLogger().Error(err)
	}
	fmt.Println("recordUserQuestion cost:", time.Since(start).Milliseconds())
}

// GetCateQuestionIndex 查询一个 qid 在一个 list 中的 index
func (s InterviewQuestionService) GetCateQuestionIndex(question models.GQuestion, cate string) int {
	f := bson.M{"exam_category": question.ExamCategory}
	if question.ExamChildCategory != "" {
		f["exam_child_category"] = question.ExamChildCategory
	}
	f["status"] = 5
	f["question_real"] = question.QuestionReal
	if cate != "全部" {
		f["question_category"] = bson.M{"$in": []string{cate}}
	}
	questions := make([]models.GQuestion, 0)
	err := s.DB().Collection("g_interview_questions").Where(f).Sort([]string{"-year", "-month", "-day", "-created_time"}...).Fields(bson.M{"_id": 1, "created_time": 1}).Find(&questions)
	if err != nil {
		return 0
	}
	lastIndex := 0
	// 找到这个 question 在整个 list 的 index
	for ii, q := range questions {
		if q.Id.Hex() == question.Id.Hex() {
			lastIndex = ii
			break
		}
	}
	return lastIndex
}

func DelHasNew(record map[string]string, questions []models.GQuestion, questionCategory []string, currQID string) (bool, string, int) {
	if len(questions) == 0 {
		return false, "", 0
	}
	key := strings.Join(questionCategory, "_")
	if _, ok := record[key]; !ok {
		return false, "", 0
	}
	var info models.UserRecordValue
	err := json.Unmarshal([]byte(record[key]), &info)
	if err != nil {
		return false, "", 0
	}
	currQuestionIDIndex := 0
	for i, question := range questions {
		if question.Id.Hex() == currQID {
			currQuestionIDIndex = i
			break
		}
	}
	// 找到当前最新的题
	maxQId, maxQIdIndex, maxIDCreateTime := FindMaxCreateTimeQuestion(questions)
	// 有新题 并且新题排在当前试题的前面
	if info.CurrCategoryMaxCreateTime != maxIDCreateTime && maxQIdIndex < currQuestionIDIndex {
		return true, maxQId, maxQIdIndex
	}

	return false, "", 0
}

func FindQIDList(allCate string, root *models.KeypointStatisticsResp, res *[]string) {
	if root == nil {
		return
	}
	if root.AllCate == allCate {
		*res = root.AllQuestionID
		return
	}
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i := range root.Child {
		FindQIDList(allCate, &root.Child[i], res)
	}
	return
}
