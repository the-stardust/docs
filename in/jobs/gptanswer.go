package jobs

import (
	"fmt"
	"interview/common/rediskey"
	"interview/models"
	"interview/services"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/remeh/sizedwaitgroup"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/garyburd/redigo/redis"
)

func init() {
	// ga := new(GPTQuestionAnswer)
	// RegisterOnceJob(ga)
}

type GPTQuestionAnswer struct {
	JobBase
}

func (sf *GPTQuestionAnswer) Do() {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	sf.SLogger().Info("启动GPT生成试题答案任务消费队列")
	sw := sizedwaitgroup.New(10)
	for {
		info, err := redis.Values(rdb.Do("BLPOP", rediskey.GPTQuestionAnswerWaiting, 5))
		if err == nil {
			sw.Add()
			go func(infoArr []interface{}, swg *sizedwaitgroup.SizedWaitGroup) {
				defer swg.Done()
				rdb := sf.RDBPool().Get()
				defer rdb.Close()
				if len(infoArr) > 1 {
					if by, ok := infoArr[1].([]byte); ok {
						isSuccess := false
						questionId := string(by)
						_, err = rdb.Do("RPUSH", rediskey.GPTQuestionAnswerCompleted, questionId)
						if err != nil {
							sf.SLogger().Error(err)
						}
						var question models.GQuestion
						err := sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(questionId)}).Take(&question)
						if err == nil {
							systemContent := "你是一名负责公务员考试中面试环节的老师."
							prompt := fmt.Sprintf(`  面试题目会放在【】里。
						对于【%+v】，请提供一些提示，帮助学生组织好他们的回答。 `, question.GetWantedQuestionContent())
							if question.CategoryId != "" {
								var category models.GQuestionCategory
								err := sf.DB().Collection("g_interview_question_category").Where(bson.M{"_id": sf.ObjectID(question.CategoryId)}).Take(&category)
								if err == nil {
									prompt = fmt.Sprintf(`  面试题目会放在【】里。
						对于【%+v】，请根据%+v提供一些提示，帮助学生组织好他们的回答。 `, question.GetWantedQuestionContent(), category.Prompt)
								} else {
									sf.SLogger().Error(err)
								}
							}
							chatCompletionMessage := []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleSystem,
									Content: systemContent,
								},
								{
									Role:    openai.ChatMessageRoleUser,
									Content: prompt,
								},
							}
							var temperature float32 = 0
							var topP float32 = 0.95
							answers, err := new(services.GPT).MakeAnswer1(temperature, topP, chatCompletionMessage, 1, false)
							if err == nil {
								if len(answers) > 0 {
									_, err := sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(questionId)}).Update(bson.M{"gpt_answer.content": answers[0]})
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
						if !isSuccess {
							_, err = rdb.Do("RPUSH", rediskey.GPTQuestionAnswerFailed, question)
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
