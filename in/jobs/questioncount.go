package jobs

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	questionCountJob := new(QuestionCountJob)
	RegisterWorkConnectJob(questionCountJob, 60*3, 2)
}

type QuestionCountJob struct {
	JobBase
}

func (sf *QuestionCountJob) GetJobs() ([]interface{}, error) {
	var err error
	var qc []models.QuestionCategory
	var filter = bson.M{}
	err = sf.DB().Collection("question_category").Where(filter).Find(&qc)
	if err != nil {
		return []interface{}{}, err
	}
	var ids []interface{}
	for _, v := range qc {
		ids = append(ids, v.Id)
	}
	return ids, nil
}

func (sf *QuestionCountJob) Do(objectID interface{}) error {
	sf.SLogger().Info("启动各分类试题计数任务")
	var err error
	var qc = models.QuestionCategory{}
	err = sf.DB().Collection("question_category").Where(bson.M{"_id": objectID.(primitive.ObjectID)}).Take(&qc)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		return err
	}

	qc.Categorys = sf.ExamDeal(qc.ExamCategory, qc.ExamChildCategory, qc.Categorys, []string{})
	qc.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
	err = sf.DB().Collection("question_category").Save(&qc)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}

	return nil
}
func (sf *QuestionCountJob) ExamDeal(examCategory string, examChildCategory string, qc []models.QuestionCategoryItem, qks []string) []models.QuestionCategoryItem {
	type tempData struct {
		ID struct {
			QuestionReal int `bson:"question_real"`
		} `bson:"_id"`
		Count float64 `bson:"count"`
	}
	for i, vo := range qc {
		var tempQks []string
		tempQks = append(tempQks, qks...)
		tempQks = append(tempQks, vo.Title)
		filter := bson.M{"status": 5}

		for j, questionCategory := range tempQks {
			filter[fmt.Sprintf("question_category.%d", j)] = questionCategory
		}
		filter["exam_category"] = examCategory
		if examChildCategory != "" {
			filter["exam_child_category"] = examChildCategory
		}

		var tempResp []tempData
		aggregateF := bson.A{bson.M{"$match": filter},
			bson.M{"$group": bson.M{"_id": bson.M{"question_real": "$question_real"}, "count": bson.M{"$sum": 1}}}}
		err := sf.DB().Collection("g_interview_questions").Aggregate(aggregateF, &tempResp)
		if err == nil {
			for _, ii := range tempResp {
				tempRespLength := len(tempResp)
				if ii.ID.QuestionReal == 0 {
					qc[i].NotRealQuestionCount = int64(ii.Count)
					if tempRespLength == 1 {
						qc[i].RealQuestionCount = 0
					}
				} else {
					qc[i].RealQuestionCount = int64(ii.Count)
					if tempRespLength == 1 {
						qc[i].NotRealQuestionCount = 0
					}
				}
			}
			qc[i].QuestionCount = qc[i].NotRealQuestionCount + qc[i].RealQuestionCount
		} else {
			qc[i].NotRealQuestionCount = 0
			qc[i].RealQuestionCount = 0
			qc[i].QuestionCount = 0
		}
		qc[i].ShortName = ""
		for _, r := range []rune(vo.Title) {
			if string(r) == "" {
				continue
			}
			qc[i].ShortName += common.FirstLetterOfPinYin(string(r))
		}

		if len(vo.ChildCategory) > 0 {
			qc[i].ChildCategory = sf.ExamDeal(examCategory, examChildCategory, vo.ChildCategory, tempQks)
		}
	}
	return qc
}
