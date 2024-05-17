package jobs

import (
	"fmt"
	"interview/common/rediskey"
	"interview/models"
	"interview/services"
	"time"

	"github.com/remeh/sizedwaitgroup"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/garyburd/redigo/redis"
)

func init() {
	// ga := new(GPTAnswer)
	// RegisterOnceJob(ga)
}

type GPTAnswer struct {
	JobBase
}
type VInfo struct {
	Url        string `redis:"url"`
	Quality    string `redis:"quality"`
	QuestionId string `redis:"question_id"`
}

func (sf *GPTAnswer) Do() {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	sf.SLogger().Info("启动GPT生成答案任务消费队列")
	sw := sizedwaitgroup.New(10)
	for {
		info, err := redis.Values(rdb.Do("BLPOP", rediskey.GPTAnswerWaiting, 5))
		if err == nil {
			sw.Add()
			go func(infoArr []interface{}, swg *sizedwaitgroup.SizedWaitGroup) {
				defer swg.Done()
				rdb := sf.RDBPool().Get()
				defer rdb.Close()
				if len(infoArr) > 1 {
					if by, ok := infoArr[1].([]byte); ok {
						isSuccess := false
						answerLogId := string(by)
						_, err = rdb.Do("RPUSH", rediskey.GPTAnswerCompleted, answerLogId)
						if err != nil {
							sf.SLogger().Error(err)
						}
						var log models.GAnswerLog
						err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 1, "_id": sf.ObjectID(answerLogId)}).Take(&log)
						if err == nil {
							var question models.GQuestion
							err := sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(log.QuestionId)}).Take(&question)
							if err == nil {
								systemContent := "你是一名负责公务员考试中面试环节的老师."
								promptPrefix := `  面试题目会放在【】里。  
							对于【%+v】，请提供一些提示，帮助学生组织好他们的回答。 `
								var temperature float32 = 0
								var topP float32 = 0.95
								key := rediskey.InterviewGPTAnswerSet
								exists, _ := redis.Bool(rdb.Do("EXISTS", key))
								if !exists {
									err := new(models.InterviewGPT).SetRedisGPTSet(string(key), "system_content", systemContent, "prompt_prefix", promptPrefix, "temperature", temperature, "top_p", topP)
									if err != nil {
										sf.SLogger().Error(err)
									}
								}
								set := new(models.InterviewGPT).GetRedisGPTSet(string(key))
								if set != nil {
									systemContent = set.SystemContent
									promptPrefix = set.PromptPrefix
									temperature = set.Temperature
									topP = set.TopP
								}

								prompt := fmt.Sprintf(promptPrefix, question.GetWantedQuestionContent())
								answers, err := new(services.GPT).MakeAnswer(temperature, topP, systemContent, prompt, 1)
								if err == nil {
									if len(answers) > 0 {
										_, err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 1, "_id": sf.ObjectID(answerLogId)}).Update(bson.M{"gpt_answer.comment": answers[0]})
										if err == nil {
											isSuccess = true
										} else {
											sf.SLogger().Error(err)
										}
									} else {
										sf.SLogger().Errorf("gpt结果异常:%+v", answers)
									}

								} else {
									sf.SLogger().Error(err)
								}
							} else {
								sf.SLogger().Error(err)
							}
						} else {
							sf.SLogger().Error(err)
						}
						if !isSuccess {
							_, err = rdb.Do("RPUSH", rediskey.GPTAnswerFailed, answerLogId)
							if err != nil {
								sf.SLogger().Error(err)
							}
						}
					} else {
						sf.SLogger().Error("值不是字符串", infoArr)
					}
				} else {
					sf.SLogger().Error("队列值异常", infoArr)
				}
			}(info, &sw)

		} else {
			if err.Error() != "redigo: nil returned" {
				sf.SLogger().Error(err)
				time.Sleep(5 * time.Second)
			}

		}
	}
}
