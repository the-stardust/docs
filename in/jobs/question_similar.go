package jobs

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/models"
	"interview/services"
	"time"
)

func init() {
	gct := new(QuestionSimilar)
	RegisterOnceJob(gct)
}

type QuestionSimilar struct {
	JobBase
}

func (sf *QuestionSimilar) Do() {
	for {
		sf.SLogger().Info("启动题目文字相似度计算")

		total, _ := sf.DB().Collection("question_hash").Where(bson.M{"questionidstotal": bson.M{"$gt": 1}}).Count()
		sf.SLogger().Info("total", total)
		perTotal := total / 16
		perTotal += 1
		for i := 0; i < 16; i++ {
			go func(sk int64) {
				var skip int64 = 0
				var limit int64 = 1000
				if limit > perTotal {
					limit = perTotal
				}

				for {
					questionhash := make([]models.QuestionHash, 0)
					startTime := time.Now().Unix()
					_ = sf.DB().Collection("question_hash").Where(bson.M{"questionidstotal": bson.M{"$gt": 1}}).Skip((sk * perTotal) + skip*limit).Limit(limit).Find(&questionhash)

					if len(questionhash) == 0 {
						break
					}
					for _, questionh := range questionhash {
						// 把字符串id转为ObjectId
						tempIds := make([]primitive.ObjectID, 0)
						for _, tempId := range questionh.QuestionIds {
							tempIds = append(tempIds, sf.ObjectID(tempId))
						}
						questions, questionStr, err := new(services.Question).GetQuestionSimilarText(tempIds)
						if err != nil {
							continue
						}

						var duplicateQuestion = make([]string, 0)
						// 2、比较多个 查出其中相同的 插入到另一个表中
						for k := 0; k < len(questionStr)-1; k++ {
							for j := k + 1; j < len(questionStr); j++ {
								percent := services.SimilarText(questionStr[k], questionStr[j])
								if percent > 0.9 {
									if !common.InArr(questions[k].Id.Hex(), duplicateQuestion) {
										duplicateQuestion = append(duplicateQuestion, questions[k].Id.Hex())
									}
									//// 去掉材料部分重新比较一下
									//percent = services.SimilarText(services.NewQuestion().GetNoMaterialQuestionStr(questions[k]), services.NewQuestion().GetNoMaterialQuestionStr(questions[j]))
									//if percent > 0.9 {
									//	if !common.InArr(questions[k].Id, duplicateQuestion) {
									//		duplicateQuestion = append(duplicateQuestion, questions[k].Id)
									//	}
									//	if !common.InArr(questions[j].Id, duplicateQuestion) {
									//		duplicateQuestion = append(duplicateQuestion, questions[j].Id)
									//	}
									//}
								}
							}
						}

						if len(duplicateQuestion) > 0 {
							sf.DB().Collection("question_similar").Create(&models.QuestionSimilar{
								Hash:        questionh.Hash,
								QuestionIds: duplicateQuestion,
								SimilarType: 1,
							})
						}
					}

					skip++
					sf.SLogger().Info(sk, "批次 ", skip*limit, "/", perTotal, " cost:", (time.Now().Unix() - startTime), "秒")
					if skip*limit >= perTotal {
						break
					}
				}

				sf.SLogger().Info("done")
			}(int64(i))
		}
		time.Sleep(60 * time.Second)

	}
}
