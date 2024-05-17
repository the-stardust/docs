package jobs

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/models"
	"interview/services"
	"sync"
	"time"
)

func init() {
	gct := new(QuestionHash)
	RegisterOnceJob(gct)
}

type QuestionHash struct {
	JobBase
}

func (sf *QuestionHash) Do() {
	for {
		sf.SLogger().Info("启动题目文字题hash入库")

		sw := sync.WaitGroup{}
		allTotal, _ := sf.DB().Collection("g_interview_questions").Count()
		total := allTotal/16 + 1
		for i := 0; i < 16; i++ {
			sw.Add(1)
			go func(sk int64) {
				defer sw.Done()
				var err error
				var skip int64 = 0
				var limit int64 = 1000
				if limit > total {
					limit = total
				}
				for {
					questions := make([]models.GQuestion, 0)
					startTime := time.Now().Unix()
					hashmap := make(map[string][]string)
					_ = sf.DB().Collection("g_interview_questions").Skip((sk * total) + skip*limit).Limit(limit).Find(&questions)
					for _, question := range questions {
						simHashStr := question.Name
						hashStr := services.StrSimHash(simHashStr)
						if _, ok := hashmap[hashStr]; !ok {
							hashmap[hashStr] = make([]string, 0)
						}
						hashmap[hashStr] = append(hashmap[hashStr], question.Id.Hex())
					}

					for hs, item := range hashmap {
						qh := new(models.QuestionHash)
						err = sf.DB().Collection("question_hash").Where(bson.M{"hash": hs}).Take(qh)
						if err != nil && !sf.MongoNoResult(err) {
							fmt.Println(err)
							continue
						}
						if sf.MongoNoResult(err) {
							qh.Hash = hs
							qh.QuestionIds = make([]string, 0)
							qh.QuestionIds = append(qh.QuestionIds, item...)
							qh.QuestionIdsTotal = len(qh.QuestionIds)
							_, err = sf.DB().Collection("question_hash").Where(bson.M{"hash": hs}).Create(qh)
							if err != nil {
								sf.SLogger().Error(err)
							}
						} else {
							qh.QuestionIds = append(qh.QuestionIds, item...)
							qh.QuestionIds = common.RemoveDuplicateElement(qh.QuestionIds)
							qh.QuestionIdsTotal = len(qh.QuestionIds)
							_, err = sf.DB().Collection("question_hash").Where(bson.M{"hash": hs}).Update(qh)
							if err != nil {
								sf.SLogger().Error(err)
							}
						}
					}
					skip++
					fmt.Println(sk, "批次  ", skip*limit, "/", total, " cost:", (time.Now().Unix() - startTime), "秒")
					if skip*limit >= total {
						break
					}
				}
			}(int64(i))
		}
		sw.Wait()
		time.Sleep(60 * time.Second)
		//helper.RedisRPUSH("questions_hash_done", "1")
	}
}
