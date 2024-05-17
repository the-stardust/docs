package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/global"
	"interview/models"
	"strconv"
	"strings"
	"time"
)

type questionSet struct {
	ServicesBase
}

func NewQuestionSet() *questionSet {
	return new(questionSet)
}

func (sf *questionSet) List(filter bson.M, offset, limit int64, id string) ([]models.Review, int64, error) {
	var err error
	resp := make([]models.Review, 0)

	total, err := sf.ReviewModel().Where(filter).Sort("-updated_time").Count()
	if err != nil {
		return resp, total, err
	}
	err = sf.ReviewModel().Where(filter).Sort("-updated_time").Skip(offset).Limit(limit).Find(&resp)
	if err != nil {
		return resp, total, err
	}

	if id != "" {
		for i, review := range resp {
			resp[i].Questions = sf.getReviewQuestions(review.PaperId)
		}
	}
	if len(resp) > 0 {
		reviewIdList := make([]string, 0)
		for _, review := range resp {
			reviewIdList = append(reviewIdList, review.Id.Hex())
		}
		reviewIdCountMap, _ := sf.AggregateReviewIdCount(reviewIdList)
		reviewIdUserCountMap, _ := sf.AggregateReviewIdUserCount(reviewIdList)
		for i, review := range resp {
			resp[i].AnswerCount = reviewIdCountMap[review.Id.Hex()].Total
			resp[i].AnswerPersonCount = reviewIdUserCountMap[review.Id.Hex()].Total
		}
	}
	return resp, total, nil
}

func (sf *questionSet) AggregateReviewIdCount(reviewIdList []string) (map[string]models.ReviewExerciseTotalCount,
	error) {
	var err error
	resp := make(map[string]models.ReviewExerciseTotalCount, 0)
	if len(reviewIdList) == 0 {
		return resp, nil
	}
	resultList := make([]models.ReviewExerciseTotalCount, 0)
	err = sf.GAnswerLogModel().Aggregate(bson.A{
		bson.M{"$match": bson.M{
			"review_id": bson.M{"$in": reviewIdList}},
		},
		bson.M{"$group": bson.M{
			"_id":   "$review_id",
			"total": bson.M{"$sum": 1},
		}},
	}, &resultList)
	if err != nil {
		return nil, err
	}
	if len(resultList) > 0 {
		for _, result := range resultList {
			resp[result.Id] = result
		}
	}
	return resp, err
}

func (sf *questionSet) AggregateReviewIdUserCount(reviewIdList []string) (map[string]models.ReviewExerciseTotalCount,
	error) {
	var err error
	resp := make(map[string]models.ReviewExerciseTotalCount, 0)
	if len(reviewIdList) == 0 {
		return resp, nil
	}
	resultList := make([]models.ReviewExerciseTotalCount, 0)
	err = sf.GAnswerLogModel().Aggregate(bson.A{
		bson.M{"$match": bson.M{
			"review_id": bson.M{"$in": reviewIdList}},
		},
		bson.M{"$group": bson.M{
			"_id":         "$review_id",
			"uniqueCount": bson.M{"$addToSet": "$user_id"},
		}},
		bson.M{"$project": bson.M{
			"_id":   1,
			"total": bson.M{"$size": "$uniqueCount"},
		}},
	}, &resultList)
	if err != nil {
		return nil, err
	}
	if len(resultList) > 0 {
		for _, result := range resultList {
			resp[result.Id] = result
		}
	}
	return resp, err
}

func (sf *questionSet) getReviewQuestions(paperId string) []models.GQuestion {
	paper := new(models.Paper)
	sf.PaperModel().Where(bson.M{"_id": sf.ObjectID(paperId)}).Take(paper)
	sortgquestions := make([]models.GQuestion, 0)
	if paper.Id.Hex() != "" {
		gquestions := make([]models.GQuestion, 0)
		sf.GQuestionModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(paper.QuestionIds)}}).Find(&gquestions)
		// 排序
		for _, questionId := range paper.QuestionIds {
			for _, gquestion := range gquestions {
				if gquestion.Id.Hex() == questionId {
					sortgquestions = append(sortgquestions, gquestion)
					break
				}
			}
		}
	}
	return sortgquestions
}

func (sf *questionSet) Edit(id string, title, managerId, managerName string, status int8, scoreType int, questionIds []string, examCate, examChildCate string) (*models.Review, error) {
	var err error
	var total int64
	review := new(models.Review)
	if id != "" {
		err = sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Take(review)
		if err != nil && !sf.MongoNoResult(err) {
			return review, err
		}

		total, _ = sf.ReviewLogModel().Where(bson.M{"review_id": id}).Count()
	}
	shortId := common.DecimalConversionBelow32(0, 32)

	paper := new(models.Paper)
	if review.PaperId != "" {
		sf.PaperModel().Where(bson.M{"_id": sf.ObjectID(review.PaperId)}).Take(paper)
		if strings.Join(paper.QuestionIds, ",") != strings.Join(questionIds, ",") && total > 0 {
			return review, errors.New("已经有学员练习，无法修改")
		}
		if review.ScoreType != scoreType && total > 0 {
			return review, errors.New("已经有学员练习，无法修改")
		}
		paper.QuestionIds = questionIds
		paper.Title = title
		paper.Status = status
		paper.ExamCategory = examCate
		paper.ExamChildCategory = examChildCate
		sf.PaperModel().Where(bson.M{"_id": sf.ObjectID(review.PaperId)}).Update(paper)
	} else {
		paper.PaperType = 3
		paper.ShortId = shortId
		paper.QuestionIds = questionIds
		paper.Title = title
		paper.ManagerId = managerId
		paper.ManagerName = managerName
		paper.Status = status
		paper.ExamCategory = examCate
		paper.ExamChildCategory = examChildCate
		sf.PaperModel().Create(paper)
	}

	review.Title = title
	review.ManagerId = managerId
	review.ManagerName = managerName
	review.Status = status
	review.PaperId = paper.Id.Hex()
	review.ExamCategory = examCate
	review.ExamChildCategory = examChildCate

	review.ScoreType = scoreType

	if id == "" {
		review.ShortId = shortId
		_, err = sf.ReviewModel().Create(review)
	} else {
		_, err = sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Update(review)
	}

	return review, err
}

func (sf *questionSet) EditClass(id, classId, className, classReviewDateStr string) error {

	qset := new(models.Review)
	err := sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Take(qset)
	if err != nil {
		return err
	}

	class := models.QSetClassInfo{
		ClassReviewDate: classReviewDateStr,
		ClassName:       className,
		ClassID:         classId,
	}

	qset.Class = class
	_, err = sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Update(qset)
	return err
}

func (sf *questionSet) EditCourse(id, courseId, courseName, chapterId, chapterName, courseWorkBtnName string) error {
	review := new(models.Review)
	err := sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Take(review)
	if err != nil {
		return err
	}

	course := models.QSetCourse{
		CourseId:          courseId,
		CourseName:        courseName,
		ChapterId:         chapterId,
		ChapterName:       chapterName,
		CourseWorkBtnName: courseWorkBtnName,
	}

	review.Course = course
	_, err = sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(id)}).Update(review)
	if err == nil {
		sf.relevanceCourse(review.Course.CourseId, review.Course.ChapterId, review.Id.Hex(), review.Course.CourseWorkBtnName+"(面试AI)")
	}

	return err
}

func (sf *questionSet) ClassMate(classId string) map[string]string {
	// {guid: name}
	classMateUidNameList := make(map[string]string)
	//获取 用户头像昵称
	res, err := common.HttpGet(global.CONFIG.ServiceUrls.QuestionBankUrl + "/question-bank/app/n/v1/class/students/v2?app_code=402&class_id=" + classId)
	if err == nil {
		type student struct {
			Guid     string `json:"g_uuid"`
			RealName string `json:"real_name"`
		}
		type ClassMateData struct {
			StudentList []student `json:"student_list"`
		}
		type ClassMateRes struct {
			Code int           `json:"code"`
			Data ClassMateData `json:"data"`
			Msg  string        `json:"msg"`
		}
		r := ClassMateRes{}
		err = json.Unmarshal(res, &r)
		if err == nil {
			for _, s := range r.Data.StudentList {
				classMateUidNameList[s.Guid] = s.RealName
			}
		} else {
			sf.SLogger().Error("class/students err:", err)
		}
	} else {
		sf.SLogger().Error("class/students err", err)
	}
	return classMateUidNameList

}

func (sf *questionSet) WorkInfo(reviewId string) (*models.Review, error) {
	review := new(models.Review)
	err := sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(reviewId)}).Take(review)
	if err != nil {
		return review, err
	}

	logs := make([]models.ReviewLog, 0)
	if review.Class.ClassID != "" {
		classMate := sf.ClassMate(review.Class.ClassID)
		if len(classMate) == 0 {
			return review, errors.New("班级未关联学员")
		}
		err = sf.ReviewLogModel().Where(bson.M{"review_id": reviewId, "user_id": bson.M{"$in": common.GetMapKeys(classMate)}}).Find(&logs)
	} else {
		err = sf.ReviewLogModel().Where(bson.M{"review_id": reviewId}).Find(&logs)
	}
	if err != nil {
		return review, err
	}
	review.Questions = sf.getReviewQuestions(review.PaperId)

	var (
		correctNum,
		answerNum int
	)
	var questionCorrectNum = make(map[string]int64)
	var questionAnswerNum = make(map[string]int64)

	for _, log := range logs {
		for _, question := range log.Questions {
			for _, answerLog := range question.AnswerLogs {
				if answerLog.CorrectStatus == 1 {
					correctNum++
					questionCorrectNum[question.QuestionId]++
				}
			}
			answerNum += len(question.AnswerLogs)
			questionAnswerNum[question.QuestionId] += int64(len(question.AnswerLogs))
		}
	}
	for i := range review.Questions {
		review.Questions[i].CorrectCount = questionCorrectNum[review.Questions[i].Id.Hex()]
		review.Questions[i].AnswerCount = questionAnswerNum[review.Questions[i].Id.Hex()]
	}
	review.AnswerCount = answerNum
	review.CorrectCount = correctNum

	return review, err
}

func (sf *questionSet) LogInfo(uid, reviewId string) (*models.ReviewLog, error) {
	log := new(models.ReviewLog)
	review := new(models.Review)
	err := sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(reviewId)}).Take(review)
	if err != nil {
		return log, err
	}

	err = sf.ReviewLogModel().Where(bson.M{"user_id": uid, "review_id": reviewId}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return log, err
	}
	if err == nil {
		log.ScoreType = review.ScoreType
		log.ReviewName = review.Title
		log.ReviewStatus = review.Status
		sf.MakeQuestionInfo(log, review.PaperId)
		//// 处理批改信息
		sf.MakeCorrectInfo(log)
	}

	return log, err
}

func (sf *questionSet) MakeLog(uid, reviewId string, reMake int8, preview string) (*models.ReviewLog, error) {
	log, err := sf.LogInfo(uid, reviewId)
	if (err != nil && !sf.MongoNoResult(err)) || (err == nil && reMake == 0) {
		return log, err
	}

	review := new(models.Review)
	err = sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(reviewId)}).Take(review)
	if err != nil {
		return log, err
	}
	log = new(models.ReviewLog)
	log.ExamChildCategory = review.ExamChildCategory
	log.ExamCategory = review.ExamCategory
	log.UserId = uid
	log.ReviewId = review.Id.Hex()
	log.ReviewName = review.Title
	log.ReviewStatus = review.Status
	log.Class = review.Class
	log.Course = review.Course
	log.Questions = make([]models.LogQuestion, 0)

	sf.MakeQuestionInfo(log, review.PaperId)

	// 预览不生成log
	if preview == "" && review.Status > -1 {
		_, err = sf.ReviewLogModel().Create(log)
	}

	return log, err
}

func (sf *questionSet) MakeQuestionInfo(log *models.ReviewLog, paperId string) {
	paper := new(models.Paper)
	sf.PaperModel().Where(bson.M{"_id": sf.ObjectID(paperId)}).Take(paper)

	questions := make([]models.GQuestion, 0)
	sf.GQuestionModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(paper.QuestionIds)}}).Find(&questions)
	questionsMap := make(map[string]models.GQuestion)
	for _, i2 := range questions {
		questionsMap[i2.Id.Hex()] = i2
	}
	if len(log.Questions) > 0 {
		for i, question := range log.Questions {
			if _, ok := questionsMap[question.QuestionId]; !ok {
				continue
			}
			log.Questions[i].Name = questionsMap[question.QuestionId].Name
			log.Questions[i].Desc = questionsMap[question.QuestionId].Desc
			log.Questions[i].Tags = questionsMap[question.QuestionId].Tags
			log.Questions[i].Answer = questionsMap[question.QuestionId].Answer
			log.Questions[i].Thinking = questionsMap[question.QuestionId].Thinking
			log.Questions[i].CategoryId = questionsMap[question.QuestionId].CategoryId
			log.Questions[i].JobTag = questionsMap[question.QuestionId].JobTag
			log.Questions[i].QuestionSource = questionsMap[question.QuestionId].QuestionSource
			log.Questions[i].TTSUrl = questionsMap[question.QuestionId].TTSUrl
			log.Questions[i].NameStruct = questionsMap[question.QuestionId].NameStruct
			log.Questions[i].NameDesc = questionsMap[question.QuestionId].NameDesc
			log.Questions[i].QuestionContentType = questionsMap[question.QuestionId].QuestionContentType
		}
	} else {
		for _, qid := range paper.QuestionIds {
			i2, ok := questionsMap[qid]
			if !ok {
				continue
			}
			log.Questions = append(log.Questions, models.LogQuestion{
				QuestionId:        i2.Id.Hex(),
				ExamCategory:      i2.ExamCategory,
				ExamChildCategory: i2.ExamChildCategory,
				QuestionCategory:  i2.QuestionCategory,
				AnswerLogs:        make([]models.GAnswerLogBase, 0),

				Name:                i2.Name,
				Desc:                i2.Desc,
				Tags:                i2.Tags,
				Answer:              i2.Answer,
				Thinking:            i2.Thinking,
				CategoryId:          i2.CategoryId,
				JobTag:              i2.JobTag,
				QuestionSource:      i2.QuestionSource,
				TTSUrl:              i2.TTSUrl,
				NameStruct:          i2.NameStruct,
				NameDesc:            i2.NameDesc,
				QuestionContentType: i2.QuestionContentType,
			})
		}
	}
}

func (sf *questionSet) MakeCorrectInfo(log *models.ReviewLog) {
	for i, question := range log.Questions {
		for _, answerLog := range question.AnswerLogs {
			if answerLog.CorrectStatus > 0 {
				log.Questions[i].CorrectNum++
			}
		}
	}
}

// 答题后更新log中的数据
func (sf *questionSet) UpdateReviewLog(reviewLogId string, answerLog models.GAnswerLog) error {
	log := new(models.ReviewLog)
	err := sf.ReviewLogModel().Where(bson.M{"user_id": answerLog.UserId, "_id": sf.ObjectID(reviewLogId)}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	log.AnswerCount = 0
	for i, question := range log.Questions {
		if question.QuestionId == answerLog.QuestionId {
			log.Questions[i].AnswerStatus = 1
			log.Questions[i].LastAnswerTime = time.Now().Format("2006-01-02 15:04:05")
			log.Questions[i].AnswerLogs = append(log.Questions[i].AnswerLogs, models.GAnswerLogBase{
				GAnswerLogId: answerLog.Id.Hex(),
				VoiceLength:  answerLog.Answer[0].VoiceLength, // answerLog.Answer长度一直是1条 所以取0
				VoiceText:    answerLog.Answer[0].VoiceText,
				VoiceUrl:     answerLog.Answer[0].VoiceUrl,
			})
		}

		// 计算下已做的数量
		if log.Questions[i].AnswerStatus == 1 {
			log.AnswerCount++
		}
	}
	// 已回答 == 题数 说明都做完了
	if log.AnswerCount == len(log.Questions) {
		log.Status = 5
	}

	_, err = sf.ReviewLogModel().Where(bson.M{"_id": log.Id}).Update(log)
	return err
}

// GPT点评后更新log中的数据
func (sf *questionSet) UpdateReviewLogCorrect(answerLog models.GAnswerLog) error {
	if answerLog.ReviewId == "" {
		return nil
	}
	review := new(models.Review)
	err := sf.ReviewModel().Where(bson.M{"_id": sf.ObjectID(answerLog.ReviewId)}).Take(review)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	if review.ScoreType != 1 {
		return nil
	}

	log := new(models.ReviewLog)
	err = sf.ReviewLogModel().Where(bson.M{"_id": sf.ObjectID(answerLog.ReviewLogId)}).Take(log)
	if err != nil && !sf.MongoNoResult(err) {
		return err
	}
	correctNum := 0
	log.CorrectStatus = 1 // log的点评状态置为部分点评
	for i, question := range log.Questions {
		if question.QuestionId == answerLog.QuestionId {
			qcorrectNum := 0
			log.Questions[i].CorrectStatus = 1 // 题的点评状态
			for j, alog := range question.AnswerLogs {
				if alog.GAnswerLogId == answerLog.Id.Hex() {
					log.Questions[i].AnswerLogs[j].CorrectStatus = 1 // 回答记录的点评状态
				}
				if log.Questions[i].AnswerLogs[j].CorrectStatus > 0 {
					qcorrectNum++
				}
			}
			if qcorrectNum == len(question.AnswerLogs) {
				log.Questions[i].CorrectStatus = 2
			}
		}
		if log.Questions[i].CorrectStatus == 2 {
			correctNum++
		}
	}
	// 已点评 == 题数 说明都做完了
	if correctNum == len(log.Questions) {
		log.CorrectStatus = 2
	}
	_, _ = sf.GAnswerLogModel().Where(bson.M{"_id": answerLog.Id}).Update(bson.M{"correct_status": 1})
	_, err = sf.ReviewLogModel().Where(bson.M{"_id": log.Id}).Update(log)
	return err
}

// 学员的作业列表
func (sf *questionSet) ClassReviewList(uid, classId, teacher string, offset, limit int64) ([]models.Review, int64, error) {
	var err error

	resp := make([]models.Review, 0)

	total, err := sf.ReviewModel().Where(bson.M{"class.class_id": classId}).Count()
	if err != nil {
		return resp, total, err
	}
	err = sf.ReviewModel().Where(bson.M{"class.class_id": classId}).Sort("-created_time").Skip(offset).Limit(limit).Find(&resp)
	if err != nil {
		return resp, total, err
	}

	if len(resp) > 0 && uid != "" {
		var logs = make([]models.ReviewLog, 0)
		var papers = make([]models.Paper, 0)
		var logsMap = make(map[string]models.ReviewLog)
		var papersMap = make(map[string]models.Paper)
		var paperIds []string
		var reviewIds []string
		var filter = bson.M{}
		filter["user_id"] = uid
		for _, review := range resp {
			reviewIds = append(reviewIds, review.Id.Hex())
			paperIds = append(paperIds, review.PaperId)
		}
		filter["review_id"] = bson.M{"$in": reviewIds}
		//filter["class.class_id"] = classId
		sf.ReviewLogModel().Where(filter).Find(&logs)
		for _, log := range logs {
			logsMap[log.ReviewId] = log
		}
		sf.PaperModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(paperIds)}}).Find(&papers)
		for _, log := range papers {
			papersMap[log.Id.Hex()] = log
		}
		for i := range resp {
			if log, ok := logsMap[resp[i].Id.Hex()]; ok {
				resp[i].AnswerCount = log.AnswerCount
			}
			if paper, ok := papersMap[resp[i].PaperId]; ok {
				resp[i].QuestionCount = len(paper.QuestionIds)
			}
		}
	}

	return resp, total, nil
}

// 学员的作业记录列表
func (sf *questionSet) ReviewLogList(uid string, offset, limit int64) ([]models.ReviewLog, int64, error) {
	var err error
	var logs = make([]models.ReviewLog, 0)

	total, err := sf.ReviewLogModel().Where(bson.M{"user_id": uid}).Count()
	if err != nil {
		return logs, total, err
	}
	err = sf.ReviewLogModel().Where(bson.M{"user_id": uid}).Sort("-created_time").Skip(offset).Limit(limit).Find(&logs)
	if err != nil {
		return logs, total, err
	}

	reviews := make([]models.Review, 0)
	reviewIds := make(map[string]models.Review)
	for _, log := range logs {
		reviewIds[log.ReviewId] = models.Review{}
	}
	sf.ReviewModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(common.GetMapKeys(reviewIds))}}).Find(&reviews)
	for _, review := range reviews {
		reviewIds[review.Id.Hex()] = review
	}

	for i, log := range logs {
		if _, ok := reviewIds[log.ReviewId]; !ok {
			continue
		}
		logs[i].ReviewName = reviewIds[log.ReviewId].Title
		logs[i].ReviewStatus = reviewIds[log.ReviewId].Status
		logs[i].QuestionCount = len(log.Questions)
	}
	return logs, total, nil
}

type CorrectCommentItem struct {
	Key    string   `json:"key"`
	Title  string   `json:"title"`
	Values []string `json:"values"`
}

// 跟 models.TeacherCorrectComment 是对应关系
func (sf *questionSet) GetCorrectComment() []CorrectCommentItem {
	var comment = make([]CorrectCommentItem, 0)
	comment = append(comment, CorrectCommentItem{
		Key:    "content",
		Title:  "内容",
		Values: []string{"跑题", "偏题", "准确", "优秀", "共情"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "speed",
		Title:  "语速",
		Values: []string{"偏快", "适中", "偏慢"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "interaction",
		Title:  "互动",
		Values: []string{"无互动", "有互动", "互动恰当"},
	})
	comment = append(comment, CorrectCommentItem{
		Key:    "confident",
		Title:  "自信程度",
		Values: []string{"自信", "不自信", "过度自信"},
	})

	return comment
}

// 更新answerlog 和 reviewlog
func (sf *questionSet) Correct(uid, reviewLogId, answerLogId string, comment models.TeacherCorrectComment) error {
	reviewLog := new(models.ReviewLog)
	err := sf.ReviewLogModel().Where(bson.M{"_id": sf.ObjectID(reviewLogId)}).Take(reviewLog)
	if err != nil {
		return err
	}

	answerLog := new(models.GAnswerLog)
	err = sf.GAnswerLogModel().Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Take(answerLog)
	if err != nil {
		return err
	}

	answerLog.TeacherCorrectContent = comment
	answerLog.CorrectStatus = 1

	hasCorrect := false
	for i, i2 := range reviewLog.Questions {
		if answerLog.QuestionId != i2.QuestionId {
			continue
		}
		for i3, i4 := range i2.AnswerLogs {
			if i4.GAnswerLogId == answerLogId {
				reviewLog.Questions[i].AnswerLogs[i3].CorrectStatus = 1
				hasCorrect = true
				break
			}
		}
		if hasCorrect {
			break
		}
	}
	var allCorrect = true
	reviewLog.CorrectStatus = 1
	for i, i2 := range reviewLog.Questions {
		reviewLog.Questions[i].CorrectNum = 0
		for _, i4 := range i2.AnswerLogs {
			if i4.CorrectStatus != 1 {
				allCorrect = false
			}
			reviewLog.Questions[i].CorrectNum++
			reviewLog.Questions[i].CorrectStatus = 1
			if reviewLog.Questions[i].CorrectNum == len(i2.AnswerLogs) {
				reviewLog.Questions[i].CorrectStatus = 2
			}
		}
	}
	if allCorrect {
		reviewLog.CorrectStatus = 2
	}

	_, err = sf.GAnswerLogModel().Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Update(answerLog)
	if err != nil {
		return err
	}
	_, err = sf.ReviewLogModel().Where(bson.M{"_id": sf.ObjectID(reviewLogId)}).Update(reviewLog)
	return err
}

func (sf *questionSet) ClassReview(classIds []string) []models.Review {
	reviews := make([]models.Review, 0)
	sf.ReviewModel().Where(bson.M{"class.class_id": bson.M{"$in": classIds}}).Find(reviews)

	return reviews
}

func (sf *questionSet) relevanceCourse(courseId string, lessionId string, homeworkId string, homework_name string) {
	cId := 0
	if courseId != "" {
		tcid, err := strconv.Atoi(courseId)
		if err == nil {
			cId = tcid
		}
	}
	lId := 0
	if lessionId != "" {
		tlid, err := strconv.Atoi(lessionId)
		if err == nil {
			lId = tlid
		}
	}
	p := map[string]interface{}{"course_id": cId, "lesson_id": lId, "homework_id": homeworkId, "homework_name": homework_name}
	httpRes, err := common.HttpPostJson(fmt.Sprintf("%s/course_manager/api/course/homework/save", global.CONFIG.ServiceUrls.RelevanceCourseUrl), p)
	if err != nil {
		sf.Logger().Error(err.Error())
	}
	var res = struct {
		Code int `json:"code"`
	}{}
	err = json.Unmarshal(httpRes, &res)
	if err != nil && res.Code != 1 {
		sf.SLogger().Error(err, string(httpRes))
	}
	sf.SLogger().Info("course/homework/saves res:" + string(httpRes))
}
