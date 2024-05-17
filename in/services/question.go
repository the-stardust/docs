package services

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/models"
)

type Question struct {
	ServicesBase
}

func (sf *Question) GetSimilarStrQuestion(material, contentAnswer string) []string {
	duplicateQuestion := make([]string, 0)
	if (material + contentAnswer) == "" {
		return duplicateQuestion
	}
	hash := StrSimHash(material + contentAnswer)
	if hash == "" {
		return duplicateQuestion
	}
	qh := new(models.QuestionHash)
	err := sf.DB().Collection("question_hash").Where(bson.M{"hash": hash}).Take(qh)
	if err != nil {
		return duplicateQuestion
	}

	for _, qid := range qh.QuestionIds {
		question := new(models.GQuestion)
		err = sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(qid)}).Fields(bson.M{"name": 1}).Take(question)
		if err != nil {
			sf.SLogger().Error(err)
			continue
		}

		percent := SimilarText(sf.GetQuestionStr(*question), material+contentAnswer)
		if percent > 0.9 {
			if !common.InArr(qid, duplicateQuestion) {
				duplicateQuestion = append(duplicateQuestion, qid)
			}
			// 去掉材料部分重新比较一下
			//percent = SimilarText(sf.GetNoMaterialQuestionStr(*question), contentAnswer)
			//if percent > 0.94 {
			//	if !common.InArr(qid, duplicateQuestion) {
			//		duplicateQuestion = append(duplicateQuestion, qid)
			//	}
			//}
		}
	}

	return duplicateQuestion
}

func (sf *Question) GetQuestionStr(question models.GQuestion) string {
	// 材料 + 题干
	simshaStr := ""
	simshaStr += question.Name
	//for _, content := range question.QuestionMaterial {
	//	simshaStr += (content.ContentToStr() + " ")
	//}
	//simshaStr += question.QuestionContent.ContentToStr()
	//for _, option := range question.QuestionOptions {
	//	simshaStr += (option.ContentToStr() + " ")
	//}
	//simshaStr += question.QuestionAnswer
	return simshaStr
}

//
//func (sf *Question) GetNoMaterialQuestionStr(question models.GQuestion) string {
//	// 题干 + 选项 + 答案
//	simshaStr := ""
//	simshaStr += question.Name
//	for _, option := range question.QuestionOptions {
//		simshaStr += (option.ContentToStr() + " ")
//	}
//	simshaStr += question.QuestionAnswer
//	return simshaStr
//}

func (sf *Question) GetQuestionSimilarText(QuestionIds []primitive.ObjectID) ([]models.GQuestion, []string, error) {
	questions := make([]models.GQuestion, 0)
	var questionStr = make([]string, 0)
	// hash相同的题
	err := sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": bson.M{"$in": QuestionIds}}).Find(&questions)
	if err != nil {
		return questions, questionStr, err
	}
	for _, question := range questions {
		questionStr = append(questionStr, sf.GetQuestionStr(question))
	}
	return questions, questionStr, nil
}

func (sf *Question) GetQuestions(filter bson.M, questionIds []string) ([]models.GQuestion, error) {
	questions := make([]models.GQuestion, 0)
	sortedQuestions := make([]models.GQuestion, 0)
	// 排序
	err := sf.DB().Collection("g_interview_questions").Where(filter).Find(&questions)
	for _, questionId := range questionIds {
		for _, question := range questions {
			if question.Id.Hex() == questionId {
				sortedQuestions = append(sortedQuestions, question)
				break
			}
		}
	}
	return sortedQuestions, err
}
