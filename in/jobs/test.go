package jobs

type VoiceToTextJob struct {
	JobBase
}

func init() {
	voiceToText := new(VoiceToTextJob)
	RegisterWorkConnectJob(voiceToText, 60)
}

func (sf *VoiceToTextJob) GetJobs() ([]interface{}, error) {
	// answerFilter := bson.M{"answer.voice_text": "", "answer.status": 0}
	// commentFilter := bson.M{"comment.voice_text": "", "comment.status": 0}
	// var answers []models.AnswerLog
	// var comments []models.AnswerComment
	// err := sf.DB().Collection("interview_answer_logs").Where(answerFilter).Find(&answers)
	// if err != nil {
	// 	sf.SLogger().Error(err)
	// }
	// err = sf.DB().Collection("interview_comment_logs").Where(commentFilter).Find(&comments)
	// if err != nil {
	// 	sf.SLogger().Error(err)
	// }

	// rdb := sf.RDBPool().Get()
	// defer rdb.Close()

	// answerVoiceUrls := make([]string, len(answers))
	// for _, answer := range answers {
	// 	_id := answer.Id.Hex()
	// 	voiceUrl := answer.Answer[0].VoiceUrl
	// 	// 如果不在redis里,继续添加
	// 	answerKey := fmt.Sprintf("%s:%s:%s", rediskey.ProName, "answer", _id)
	// 	_, err := redis.Values(rdb.Do("GET", answerKey))
	// 	if err != nil { // todo
	// 		sf.SLogger().Error(err)
	// 		answerVoiceUrls = append(answerVoiceUrls, voiceUrl)
	// 	}
	// }

	// commentVoiceUrls := make([]string, len(comments))
	// for _, comment := range comments {
	// 	_id := comment.Id.Hex()
	// 	voiceUrl := comment.Comment.VoiceUrl
	// 	// 如果不在redis里,继续添加
	// 	commentKey := fmt.Sprintf("%s:%s:%s", rediskey.ProName, "comment", _id)
	// 	_, err := redis.Values(rdb.Do("GET", commentKey))
	// 	if err != nil {
	// 		sf.SLogger().Error(err)
	// 		commentVoiceUrls = append(commentVoiceUrls, voiceUrl)
	// 	}
	// }

	// filter := bson.M{"reviews.end_time": bson.M{"$lt": endTime, "$gt": startTime}}
	// exams := []models.MockExam{}
	// err := sf.DB().Collection("mock_exams").Where(filter).Find(&exams)
	// if err != nil {
	// 	sf.SLogger().Error(err)
	// 	return []interface{}{}, err
	// }
	// if len(exams) > 0 {
	// 	for _, v := range exams {
	// 		sf.SLogger().Debugf("需要处理模考：%+v", v.MockExamName)
	// 		for _, review := range v.Reviews {
	// 			end, err := time.ParseInLocation("2006-01-02 15:04:05", review.EndTime, time.Local)
	// 			if err == nil {
	// 				if startTimestamp < end.Unix() && endTimestamp.Unix() > end.Unix() {
	// 					sf.SLogger().Debugf("需要处理得测评：%+v", review.ReviewName)
	// 					//查询 测评下尚未完成得log
	// 					f := bson.M{"status": 0, "review_id": review.ReviewId}
	// 					testLogs := []struct {
	// 						Id primitive.ObjectID `bson:"_id"`
	// 					}{}
	// 					err := sf.DB().Collection("user_test_log").Where(f).Find(&testLogs)
	// 					if err == nil {
	// 						for _, lg := range testLogs {
	// 							resp = append(resp, lg.Id)
	// 						}
	// 					}
	// 				}
	// 			} else {
	// 				sf.SLogger().Error(err)
	// 			}
	// 		}

	// 	}
	// }
	return make([]interface{}, 0), nil
}

func (sf *VoiceToTextJob) Do(objectID interface{}) error {
	// var err error
	// logRes := models.TestLog{}
	// err = sf.DB().Collection("user_test_log").Where(bson.M{"_id": objectID}).Take(&logRes)
	// if err == nil {
	// 	if logRes.Status == 0 {
	// 		answers := []app.UserAnswer{}
	// 		for _, v := range logRes.Questions {
	// 			t := app.UserAnswer{
	// 				QuestionId:   v.QuestionId,
	// 				Timed:        v.Timed,
	// 				Answer:       v.UserAnswer,
	// 				CalculusInfo: v.CalculusInfo,
	// 			}
	// 			answers = append(answers, t)
	// 		}
	// 		gc := gin.Context{}
	// 		gc.Set("APP-SOURCE-TYPE", logRes.SourceType)
	// 		autoLog := models.AutoSubmitLog{LogId: logRes.Id.Hex()}
	// 		info := fmt.Sprintf("UserId:%+v\nReviewId:%+v\nLogId:%+v\nanswers:%+v\n", logRes.UserId, logRes.ReviewId, logRes.Id.Hex(), answers)
	// 		autoLog.Info = info
	// 		sf.DB().Collection("auto_submit_log").Create(&autoLog)
	// 		new(app.Answer).SubmitTestLog(logRes.UserId, logRes.ReviewId, logRes.Id.Hex(), answers, logRes.Timed, logRes.AnsweredIndex, 5, logRes.SubmitType, logRes.ScanImage, "", 0, &gc)
	// 	}
	// } else {
	// 	sf.SLogger().Error(err)
	// 	return err
	// }
	return nil
}

/*
   def download_and_create_task(self, data, category_name):
       _id = data.get("_id")
       str_id = str(_id)
       # 如果已经在redis中，代表已经上传音频，跳过本条数据
       if self.redis_db.get(f"{category_name}:{str_id}"):
           return

       if category_name == "answer":
           voice_url = data.get(category_name)[0].get("voice_url")
       else:
           voice_url = data.get(category_name).get("voice_url")
       voice_data = requests.get(voice_url, headers=self.headers).content
       voice_suffix_name = voice_url[-3:]
       voice_path = rf'./audio/{self.today_date}-{category_name}-{str_id}.{voice_suffix_name}'
       with open(voice_path, "wb") as f:
           f.write(voice_data)
       voice_file_url = self.get_file_url(voice_path)
       # os.remove(voice_path)  # 删除文件
       task_id_info = self.task_create(voice_file_url)
       task_id = task_id_info.get("data").get("task_id")
       self.redis_pipeline.set(f"{category_name}:{str_id}", json.dumps([str_id, task_id, voice_file_url, voice_suffix_name]))
*/

func downloadVoiceUrlContentAndSave(voiceUrl string) {
	// voiceData, err := common.HttpGet(voiceUrl)
	// if err != nil{
	// 	fmt.Println(err)
	// }

	// with os.Open()
}
