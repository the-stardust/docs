package jobs

import (
	"fmt"
	"interview/common/rediskey"
	"interview/models"
	"interview/services"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sashabaranov/go-openai"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/garyburd/redigo/redis"
)

func init() {
	gct := new(GPTComment)
	RegisterOnceJob(gct)
}

type GPTComment struct {
	JobBase
}

func (sf *GPTComment) Do() {
	reConnectionCount := 0
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	sf.SLogger().Info("启动GPT生成点评任务消费队列")
	sw := sizedwaitgroup.New(10)
	for {
		info, err := redis.Values(rdb.Do("BLPOP", rediskey.GPTCommentWaiting, 5))
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
						_, err = rdb.Do("RPUSH", rediskey.GPTCommentCompleted, answerLogId)
						if err != nil {
							sf.SLogger().Error(err)
						}
						var answerLog models.GAnswerLog
						err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
						if err == nil {
							if len(answerLog.Answer) > 0 && answerLog.Answer[0].VoiceText != "" {
								tempAnswer := []rune(answerLog.Answer[0].VoiceText)
								if len(tempAnswer) < 70 {
									isSuccess = true
									time.Sleep(5 * time.Second)
									_, err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Update(bson.M{"gpt_comment.content": "抱歉,你提供的回答不足以让我进行有效的点评和建议。请提供更多详细回答，让我能够理解情境和内容。谢谢!"})
									if err == nil {
										isSuccess = true
									} else {
										sf.SLogger().Error(err)
									}
								} else {
									var question models.GQuestion
									err := sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(answerLog.QuestionId)}).Take(&question)
									if err == nil {
										systemContent := ""
										prompt := ""
										var t models.CategoryGPTPrompt
										err = sf.DB().Collection("category_gpt_prompt").Where(bson.M{"exam_category": question.ExamCategory, "answer_type": 0}).Take(&t)
										if err != nil {
											systemContent = fmt.Sprintf("你是一名负责%s考试中面试环节的面试官。", question.ExamCategory)
											prompt = fmt.Sprintf(`你的任务是根据面试中的题目内容，对学生的回答进行点评。  
										面试中的题目内容和学生答案是被【】括起来的文本。  
										按以下步骤完成这个任务：1.对学生的回答进行点评，点评内容需要包括学生回答是否流畅、考虑问题是否全面，对策是否具有针对性。  
									  2.针对点评内容提出针对性建议。\n【题目】：%s\n【学生答案】：%s `, question.GetWantedQuestionContent(), answerLog.Answer[0].VoiceText)
										} else {
											systemContent = t.SystemContent
											prompt = fmt.Sprintf(t.Prompt, question.GetWantedQuestionContent(), answerLog.Answer[0].VoiceText)
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
												_, err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Update(bson.M{"gpt_comment.content": answers[0]})
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
								}

							} else {
								sf.SLogger().Error("无效的回答")
							}
						} else {
							sf.SLogger().Error(err)
						}
						if !isSuccess {
							_, err = rdb.Do("RPUSH", rediskey.GPTCommentFailed, answerLogId)
							if err != nil {
								sf.SLogger().Error(err)
							}
						} else {
							_, err = rdb.Do("SADD", rediskey.GPTCommentSuccess, answerLogId)
							if err != nil {
								sf.SLogger().Error(err)
							}
							services.NewQuestionSet().UpdateReviewLogCorrect(answerLog)
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
				if reConnectionCount > 5 {
					sf.SLogger().Error("重试")
					// todo 消息通知

				}
				time.Sleep(5 * time.Second)
				rdb = sf.RDBPool().Get()
				reConnectionCount += 1

			}
		}

	}
}
