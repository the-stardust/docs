package jobs

import (
	"fmt"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sashabaranov/go-openai"
	"interview/common/rediskey"
	"interview/controllers/app"
	"interview/models"
	"interview/services"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/garyburd/redigo/redis"
)

func init() {
	gca := new(GPTCustomAnswer)
	RegisterOnceJob(gca)
}

type GPTCustomAnswer struct {
	JobBase
}

func (sf *GPTCustomAnswer) Do() {
	reConnectionCount := 0
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	sf.SLogger().Info("启动GPT生成回答任务消费队列")
	sw := sizedwaitgroup.New(10)
	var temperature float32 = 0
	var topP float32 = 0.95
	for {
		info, err := redis.Values(rdb.Do("BLPOP", rediskey.GPTCustomWaiting, 5))
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
						_, err = rdb.Do("RPUSH", rediskey.GPTCustomCompleted, answerLogId)
						if err != nil {
							sf.SLogger().Error(err)
						}
						var answerLog models.GCustomQuestion
						updateContent := ""
						err := sf.DB().Collection("g_custom_questions").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
						if err == nil {
							if answerLog.GPTAnswer.Content == "" {
								// systemContent := ""
								prompt := answerLog.Name
								var chatCompletionMessage []openai.ChatCompletionMessage

								if answerLog.AnswerType == 1 {
									contentChan := make(chan string, 2)
									innerSw := sizedwaitgroup.New(2)
									for _, tempAnswerType := range []int{1, 3} {
										innerSw.Add()
										go func(answerType int, tempChan chan string, innerSwg *sizedwaitgroup.SizedWaitGroup) {
											defer innerSwg.Done()
											// 原先只有提示，现在需要新增标准答案
											systemContentInCor := fmt.Sprintf("你是一名负责%s考试中面试环节的老师。", answerLog.ExamCategory)
											promptInCor := ""
											if answerType == 1 {
												promptInCor = fmt.Sprintf(`  面试题目会放在【】里。对于【%+v】，请提供一些提示，帮助学生组织好他们的回答。 `, answerLog.Name)
											} else {
												promptInCor = fmt.Sprintf(`  面试题目会放在【】里。对于【%+v】，生成合理的高分答案，答案以第一人称“我”叙述。 `, answerLog.Name)
											}
											var t models.CategoryGPTPrompt
											err = sf.DB().Collection("category_gpt_prompt").Where(bson.M{"exam_category": answerLog.ExamCategory, "answer_type": answerType, "question_category": answerLog.QuestionCategory}).Take(&t)
											if err == nil {
												systemContentInCor = t.SystemContent
												promptInCor = fmt.Sprintf(t.Prompt, answerLog.Name)
											}

											chatCompletionMessage = []openai.ChatCompletionMessage{
												{
													Role:    openai.ChatMessageRoleSystem,
													Content: systemContentInCor,
												},
												{
													Role:    openai.ChatMessageRoleUser,
													Content: promptInCor,
												},
											}
											answers, err := new(services.GPT).MakeAnswer1(temperature, topP, chatCompletionMessage, 1, false)
											if err == nil {
												if len(answers) > 0 {
													isSuccess = true
													if answerType == 1 {
														tempChan <- "【回答思路】:" + answers[0] + "\n"
													} else {
														tempChan <- "【标准答案】:" + answers[0] + "\n"
													}

												} else {
													isSuccess = false

													sf.SLogger().Errorf("gpt结果异常:%+v, err:%+v", answers, err)
												}

											} else {
												isSuccess = false

												sf.SLogger().Error(err)
											}

										}(tempAnswerType, contentChan, &innerSw)
									}
									innerSw.Wait()
									gptTips := ""
									gptGoodAnswer := ""
									close(contentChan)
									for gptString := range contentChan {
										if strings.Contains(gptString, "【回答思路】") {
											gptTips = gptString
										} else {
											gptGoodAnswer = gptString
										}
									}
									updateContent = gptTips + "\n" + gptGoodAnswer
								} else {
									chatCompletionMessage = []openai.ChatCompletionMessage{
										{
											Role:    openai.ChatMessageRoleUser,
											Content: prompt,
										},
									}
									answers, err := new(services.GPT).MakeAnswer1(temperature, topP, chatCompletionMessage, 1, false)
									if err == nil {
										if len(answers) > 0 {
											isSuccess = true
											updateContent = answers[0]
											//_, err := sf.DB().Collection("g_custom_questions").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Update(bson.M{"gpt_answer.content": answers[0]})
											//if err == nil {
											//	isSuccess = true
											//} else {
											//	sf.SLogger().Error(err)
											//}
										} else {
											sf.SLogger().Errorf("gpt结果异常:%+v, err:%+v", answers, err)
										}

									} else {
										sf.SLogger().Error(err)
									}
								}
							} else {
								sf.SLogger().Error("无效的数据")
							}
						} else {
							sf.SLogger().Error(err)
						}

						if !isSuccess {
							_, err = rdb.Do("RPUSH", rediskey.GPTCustomFailed, answerLogId)
							if err != nil {
								sf.SLogger().Error(err)
							}

							countInfo, err := new(app.Activity).GetGPTCanUseCount(answerLog.UserId)
							if err != nil {
								sf.SLogger().Error(err)
							}
							countInfo.AvailableCount += 1
							err = new(app.Activity).SetGPTCanUseCount(countInfo)
							if err != nil {
								sf.SLogger().Error(err)
							}

						} else {
							_, err = rdb.Do("SADD", rediskey.GPTCustomSuccess, answerLogId)
							if err != nil {
								sf.SLogger().Error(err)
							}

							// 统一更新mongo
							_, err := sf.DB().Collection("g_custom_questions").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Update(bson.M{"gpt_answer.content": updateContent})
							if err == nil {
								isSuccess = true
							} else {
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
				if reConnectionCount > 5 {
					sf.SLogger().Error("重试")
					// todo 消息通知
					break
				}
				time.Sleep(5 * time.Second)
				rdb = sf.RDBPool().Get()
				reConnectionCount += 1

			}
		}

	}
}
