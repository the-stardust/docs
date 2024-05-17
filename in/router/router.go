package router

import (
	"interview/controllers"
	"interview/controllers/app"
	"interview/controllers/manager"
	"interview/middlewares"
	"net/http"

	nice "github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
)

/**
 * @name: 注册路由
 * @msg: 用于注册全局路由
 * @param {*gin.Engine}
 * @return {*}
 */
func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(new(middlewares.Cors).HandlerFunc())
	r.Use(nice.Recovery(new(middlewares.Recovery).HandlerFunc))
	// k8s心跳检测
	r.GET("/interview/healthy", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"K8s": "healthy",
		})
	})
	midSign := new(middlewares.Sign)
	category := new(controllers.Category)
	interview := new(app.Interview)
	interviewMock := new(app.InterviewMockController)
	mockApply := new(app.ImaController)
	amockexam := new(app.MockExam)
	questionset := new(app.QuestionSet)
	paper := new(app.Paper)
	upload := new(app.Upload)
	aMorningRead := new(app.MorningRead)
	mMorningReadTag := new(manager.MorningReadTag)

	curriculaControl := app.GetInitControl()
	home := new(app.HomeController)
	rqu := r.Group("/interview/app/u", new(middlewares.App).HandlerFunc())
	{
		v1 := rqu.Group("/v1")
		{
			v1.POST("/class/list", interview.ClassList)
			v1.POST("/all/class/list", interview.AllClassList)
			v1.POST("/class/save", interview.ClassSave)
			v1.POST("/class/info", interview.ClassInfo)
			v1.POST("/class/member/save", interview.ClassChangeMember)
			v1.POST("/question/save", interview.SaveInterviewQuestion)
			v1.GET("/question/list", interview.GetInterviewQuestions)
			v1.GET("/question/info", interview.GetInterviewQuestion)
			v1.GET("/question/students", interview.GetInterviewQuestionStudents)
			v1.GET("/question/allow/student", interview.InterviewQuestionAllowStudent)
			v1.POST("/question/status/save", interview.ChangeInterviewQuestionStatus)

			v1.POST("/answer-log/save", interview.SaveAnswerLog)
			v1.GET("/answer-log/list", interview.GetAnswerLogs)
			v1.GET("/answer-log/info", interview.GetAnswerLog)
			v1.GET("/answer-log/del", interview.DelAnswerLog)

			v1.POST("/comment/save", interview.SaveAnswerComment)
			v1.POST("/comment/save2", interview.SaveAnswerComment2)
			v1.POST("/comment/list", interview.GetAnswerComments)
			v1.GET("/comment/info", interview.GetAnswerComment)
			v1.POST("/comment/feedback", interview.CommentFeedBack)
			v1.POST("/comment/feedback2", interview.CommentFeedBack2)
			v1.GET("/isadmin", interview.IsAdmin)
			v1.POST("/remove/student", interview.RemoveStudent)
			curriculaGroup := v1.Group("/curricula")
			{
				// 是否否展示考试tab页
				curriculaGroup.GET("/is_show", curriculaControl.CurriculaIsShow)
				// 考试课程列表
				curriculaGroup.GET("/list", curriculaControl.CurriculaList)
				// 考试课程列表
				curriculaGroup.GET("select/list", curriculaControl.CurriculaSelectList)
				// 课程保存
				curriculaGroup.POST("/save", curriculaControl.CurriculaSave)
				//用户超级管理员列表
				curriculaGroup.GET("/admin_user/list", curriculaControl.CurriculaAdminUserList)
				//删除用户超级管理员
				curriculaGroup.POST("/admin_user/delete", curriculaControl.CurriculaAdminUserDelete)
				// 生成邀请码
				curriculaGroup.POST("/invite_code/create", curriculaControl.CurriculaInviteCode)
				// 检查是否已经是管理员
				curriculaGroup.GET("/check/admin", curriculaControl.CheckCurriculaAdmin)
				// 消费邀请码
				curriculaGroup.GET("/invite_code/use", curriculaControl.CurriculaInviteUse)
				// 管理员试题上传
				curriculaGroup.POST("/question/save", curriculaControl.AdminSaveInterviewQuestion)
				// 管理员班级列表
				curriculaGroup.GET("/class/list", curriculaControl.AdminClassList)
			}
			// 面试云模考
			// 是否可以创建模考
			v1.GET("/interview/mock/permission", interviewMock.CreatePermission)
			v1.GET("/interview/mock/list", interviewMock.List)
			v1.POST("/interview/mock/create", interviewMock.Create)
			v1.POST("/interview/mock/update", interviewMock.Update)
			v1.POST("/interview/mock/delete", interviewMock.Delete)
			v1.GET("/interview/mock/info", interviewMock.Info)
			v1.GET("/interview/mock/unapply/list", interviewMock.UnApplyList)
			// 学生报名参加模考
			v1.POST("/interview/mock/apply", mockApply.ApplyMock)
			v1.GET("/interview/mock/apply/users", mockApply.ApplyUserList)
			// 学生取消报名
			v1.POST("/interview/mock/apply/cancel", mockApply.ApplyCancel)
			// 导出模考详情
			v1.POST("/interview/mock/export", mockApply.Export)
			// 判断某用户是否报名过某个模考
			v1.POST("/interview/mock/user/is_apply", mockApply.IsApply)

		}
		v2 := rqu.Group("/v2")
		{
			v2.GET("/class/list", interview.ClassListV2)
			v2.GET("/answer_log/list", interview.GetAnswerLogsV2)
			v2.GET("/exclude_answer/list", interview.GetExcludeAnswerLogsV2)
		}

		interviewGPT := new(app.InterviewGPT)
		activity := new(app.Activity)
		user := new(app.User)
		r.POST("/interview/app/u/v1/g/keypoint/record", interviewGPT.RecordAnswerPoint)
		gptv1 := rqu.Group("/v1/g")
		{
			// gptv1.POST("/teacher/save", interviewGPT.TeacherSave)
			gptv1.POST("/user/info", interviewGPT.GetUserInfo)
			gptv1.POST("/question/list", midSign.HandlerFunc, interviewGPT.GetInterviewQuestions)
			gptv1.GET("/question/info", midSign.HandlerFunc, interviewGPT.GetInterviewQuestion)
			gptv1.POST("/random/question/info", midSign.HandlerFunc, interviewGPT.GetRandomInterviewQuestion)
			// gptv1.POST("/question/save", interviewGPT.SaveInterviewQuestion) 面试AI的C端没有保存试题，故注释
			gptv1.POST("/question/status/save", interviewGPT.ChangeInterviewQuestionStatus)

			gptv1.POST("/answer-log/save", interviewGPT.SaveAnswerLog)
			gptv1.POST("/answer-log/list", interviewGPT.GetAnswerLogs)
			gptv1.GET("/answer-log/info", midSign.HandlerFunc, interviewGPT.GetAnswerLog)
			gptv1.GET("/answer-log/del", interviewGPT.DelAnswerLog)
			gptv1.GET("/ws", interviewGPT.SendGPT)
			gptv1.GET("/question/category/list", interviewGPT.GetQuestionCategory)
			gptv1.POST("/user/feedback", new(app.User).UserFeedback)

			gptv1.GET("/exam/category/list", category.ExamCategory)
			gptv1.POST("/question/category/list", category.QuestionCategory)
			gptv1.GET("/invite-user", activity.SaveInviteInfo)
			gptv1.GET("/query-can-use-count", activity.QueryGPTUseCount)
			gptv1.GET("/query-invited-user", activity.UserInviteList)
			gptv1.POST("/submit-custom-question", activity.SubmitCustomQuestion)
			gptv1.POST("/submit-custom-question-list", activity.CustomQuestionList)
			gptv1.GET("/submit-custom-question-info", activity.CustomQuestionInfo)
			gptv1.POST("/need-baipiao", activity.NeedBaipiao)
			gptv1.GET("/is-today-have-baipiao", activity.IsTodayHaveBaipiao)
			gptv1.GET("/need-buy-count", activity.NeedBuyCount)
			gptv1.POST("transfer-id-length", activity.TransferID)
			gptv1.POST("save-gpt-thought-index", interviewGPT.SaveGPTThoughtIndex)

			gptv1.POST("/user/choice/status", user.SaveUserChoiceStatus) // 用户选择状态
			gptv1.GET("/user/choice/status", user.GetUserChoiceStatus)   // 查询用户选择状态
			gptv1.GET("/is-show-class", user.IsShowClass)                // 面试AI小程序中当前用户是否能看到班级

			gptv1.POST("/paper/list", paper.PaperList)                 // 试卷列表
			gptv1.GET("/paper/info", paper.PaperInfo)                  // 试卷详情
			gptv1.GET("/paper/info/questions", paper.QuestionsInPaper) // 试卷下包含的试题
			gptv1.POST("/mockexam/create", amockexam.RoomCreate)
			gptv1.GET("/mockexam/room-conf", amockexam.RoomConf)
			gptv1.POST("/mockexam/list", amockexam.List)
			gptv1.GET("/mockexam/cancel", amockexam.Cancel)                                // 取消
			gptv1.POST("/mockexam/sign-up", amockexam.SignUp)                              // 报名
			gptv1.GET("/mockexam/sign-in", amockexam.SignIn)                               // 签到
			gptv1.POST("/mockexam/subscribe", amockexam.Subscribe)                         // 订阅
			gptv1.GET("/mockexam/end", amockexam.End)                                      // 结束
			gptv1.POST("/mockexam/teacher-control", amockexam.TeacherControl)              // 设置老师控制
			gptv1.GET("/mockexam/base-info", amockexam.BaseInfo)                           // 详情
			gptv1.GET("/mockexam/info", amockexam.Info)                                    // 详情
			gptv1.POST("/mockexam/teacher-mark", amockexam.TeacherMark)                    // 老师打分
			gptv1.POST("/mockexam/set-room-status", amockexam.SetRoomStatus)               // room-status
			gptv1.GET("/mockexam/room-status", amockexam.RoomStatus)                       // room-status
			gptv1.GET("/mockexam/room-list", amockexam.RoomList)                           // room-list
			gptv1.GET("/mockexam/log-detail", amockexam.LogDetail)                         // log-detail
			gptv1.GET("/mockexam/start-vd", amockexam.StartVideo)                          // 开启云端录制
			gptv1.GET("/mockexam/end-vd", amockexam.EndVideo)                              // 关闭云端录制
			gptv1.GET("/mockexam/video-start-midway-join", amockexam.VideoStartMidwayJoin) // 有录制视频的情况下， 学员中途加入
			gptv1.POST("/mockexam/room-member/change", amockexam.RoomMemberChange)         // 成员变化

			gptv1.POST("/review/classes", questionset.HasReviewClass)                       // 获取有测评的班级
			gptv1.GET("/review/info", midSign.HandlerFunc, questionset.ReviewInfo)          // 获取测评
			gptv1.GET("/review/work_info", midSign.HandlerFunc, questionset.ReviewWorkInfo) // 获取班级作业测评
			gptv1.GET("/review/list", questionset.ReviewList)                               // 测评列表
			gptv1.GET("/review/log/list", questionset.ReviewLogList)                        // 测评记录列表
			gptv1.POST("/review/comment/save", questionset.SaveAnswerComment)               // 老师点评
			gptv1.GET("/user/upload/file", upload.UploadFile2)                              // 学员上传试题文件

			gptv1.GET("/morning_read/info", aMorningRead.Info)        // 当前可用的晨读
			gptv1.GET("/morning_read/info/v2", aMorningRead.InfoV2)   // 当前可用的晨读
			gptv1.POST("/morning_read/report", aMorningRead.Report)   // 晨读上报
			gptv1.GET("/morning_read/log_list", aMorningRead.LogList) // 晨读历史
			gptv1.GET("/morning_read/result", aMorningRead.Result)    // 晨读报告
		}
	}
	activity := new(app.Activity)
	rqn := r.Group("/interview/app/n", new(middlewares.App).HandlerFunc())
	{
		v1 := rqn.Group("/v1")
		{
			v1.POST("/upload/file", upload.UploadFile)
			v1.POST("/upload/file/btye", upload.UploadFile)
			v1.POST("/room_token", amockexam.RoomToken)

			v1.GET("/morning_read/tags/list", mMorningReadTag.List) // tag list
			v1.GET("/morning_read/info/v2", aMorningRead.InfoV2)    // 当前可用的晨读
		}
		v2 := rqn.Group("/v2")
		{
			v2.POST("/upload/file", upload.UploadFile2)
		}
	}
	statistics := new(manager.Statistics)
	interviewGPT := new(app.InterviewGPT)
	gQuestion := new(manager.Question)
	aUser := new(app.User)
	aStudy := new(app.Study)
	video := new(manager.VideoKeypoint)
	grqnv1 := rqn.Group("/v1/g")
	{
		grqnv1.POST("/web/cache/save", interviewGPT.SetWebRelayCache)
		grqnv1.GET("/web/cache/info", interviewGPT.GetWebRelayCache)
		grqnv1.GET("/ws", interviewGPT.SendGPT)
		grqnv1.POST("/question/search/list", interviewGPT.GetInterviewQuestionsWithES)
		grqnv1.POST("/question/list", midSign.HandlerFunc, interviewGPT.GetInterviewQuestions)
		grqnv1.POST("/total-question/list", interviewGPT.GetTotalInterviewQuestions)
		grqnv1.GET("/exam/category/list", category.ExamCategory)
		grqnv1.POST("/question/category/list", category.QuestionCategory)
		grqnv1.POST("transfer-id-length", activity.TransferID)
		grqnv1.POST("/question-source/list", gQuestion.QuestionList)
		grqnv1.POST("/question-source/info", midSign.HandlerFunc, gQuestion.QuestionSourceInfo)
		grqnv1.GET("/question/info", midSign.HandlerFunc, interviewGPT.GetInterviewQuestion)
		grqnv1.POST("/tts-prompt", interviewGPT.GetTTSPrompt)
		grqnv1.GET("/is-show-page", interviewGPT.IsShowPage)
		grqnv1.POST("/mockexam/list", amockexam.List)
		grqnv1.POST("/mockexam/auto-record", amockexam.AutoRecord)
		grqnv1.GET("/mockexam/is_teacher", amockexam.IsTeacher)            // 是否是老师
		grqnv1.GET("/mockexam/mark-content", amockexam.TeacherMarkContent) // 模考-获取老师点评内容
		grqnv1.POST("/wechat/jscode2session", aUser.JSCode2session)
		grqnv1.GET("/review/correct/comment", questionset.GetCorrectComment) // 获取老师点评内容
		grqnv1.GET("/wechat-group", aUser.WechatGroup)                       // 面试微信群
		grqnv1.POST("/paper/list", paper.PaperList)                          // 试卷列表

		grqnv1.GET("/paper/info/questions", paper.QuestionsInPaper)     // 试卷下包含的试题
		grqnv1.POST("/zego_callback", amockexam.ZegoCallback)           // 回调
		grqnv1.GET("/wechat", interviewGPT.WechatSign)                  // 当前可用任务线(不必登录)
		grqnv1.POST("/wechat", interviewGPT.WechatMsg)                  // 当前可用任务线(不必登录)
		grqnv1.POST("/activity/new-year", interviewGPT.ActivityNewYear) // 活动
		grqnv1.GET("/activity/neimeng", interviewGPT.ActivityNeiMeng)   // 活动
		grqnv1.GET("/activity/022902", interviewGPT.Activity022902)     // 活动
		grqnv1.GET("/activity/0229", interviewGPT.Activity0229)         // 活动
		grqnv1.GET("/activity/0301", interviewGPT.Activity0301)         // 统计局调查总队
		grqnv1.GET("/share/img", interviewGPT.ShareImg)                 // 分享图片
		grqnv1.POST("/upload/str", upload.RandomStr)                    // 上传文件时前端先获取，作为加密一部分

		grqnv1.POST("/speechtext/test", interviewGPT.SpeechTextTest)
		grqnv1.POST("/speechtext/callback", interviewGPT.SpeechTextCallback)
		grqnv1.GET("/area", activity.Area) // 地区
		grqnv1.POST("/study/course/list", aStudy.CourseList)
		grqnv1.POST("/study/data-pack/list", aStudy.DataPackList)
		grqnv1.POST("/video-keypoint/media/save", video.SaveVideoKeypoints)
		grqnv1.POST("/video-keypoint/media/get", video.GetVideoKeypoints)
		grqnv1.POST("/video-keypoint/media/callback", video.VideoKeypointCallback)
		grqnv1.POST("/data/funneling", statistics.DataFunneling) // 数据漏斗

	}
	mPaper := new(manager.Paper)
	mQuestion := new(manager.Question)
	mUser := new(manager.User)
	mockexam := new(manager.MockExam)
	mUpload := new(manager.Upload)
	mMorningRead := new(manager.MorningRead)
	mStudy := new(manager.Study)
	rqm := r.Group("/interview/manager", new(middlewares.Manager).HandlerFunc())
	{
		v1 := rqm.Group("/v1")
		{ // 统计
			v1.GET("/statistics/question", statistics.QuestionStatistics)
			v1.GET("/statistics/question_years", statistics.QuestionYearsStatistics)
			v1.GET("/daily/statistics", statistics.DailyStatistics)
		}

		{ // 运营需要的统计数据
			v1.POST("/operation/statistics/info", statistics.GetStatisticsInfo)
			v1.POST("/operation/statistics/user-exercise-count", statistics.GetUserInfoAndExerciseCount)
			v1.POST("/operation/statistics/exercise-count", statistics.GetExerciseRoomCount)
		}

		{
			v1.POST("/question/list", mQuestion.QuestionList)
			v1.POST("/question/search/list", mQuestion.QuestionSearchList)
			v1.POST("/question/save", mQuestion.SaveQuestion)
			v1.POST("/question/is-repeat", mQuestion.QuestionIsRepeat)
			v1.GET("/question/info", mQuestion.QuestionInfo)
			v1.POST("/question/infos", mQuestion.QuestionInfos) // 批量查询试题详情
			v1.POST("/feedback-list", mUser.UserFeedbackList)
			v1.GET("/feedback-detail", mUser.UserFeedbackDetail)
			v1.GET("/exam/category/list", category.ExamCategory)
			v1.POST("/question/category/list", category.QuestionCategory)
			v1.GET("/user_permission/question_category/list", category.CategoryPermissionList) // 试题分类tree

			v1.POST("/question/make/gpt/answer", mQuestion.MakeGPTAnswer)
			v1.POST("/answer/list", mQuestion.AnswerList)
			v1.POST("/answer/search/list", mQuestion.AnswerListWithES) // es搜索
			v1.POST("/custom/question/list", mQuestion.CustomQuestionList)
			v1.POST("/question/standard-answer/prompt", mQuestion.PromptFromCategoryTemp)
			v1.POST("/prompt-from-category", mQuestion.PromptFromCategory)
			v1.POST("/save/prompt", mQuestion.SavePrompt)
			v1.POST("/question/answer-log", mQuestion.QuestionAnswerLog)

			v1.POST("/question/gpt-preview", mQuestion.GPTPreview)
			v1.POST("/question/area-sub", mQuestion.QuestionLogAreas)
			v1.POST("/is-show-class-members/save", mUser.SaveIsShowClassMembers)
			v1.GET("/is-show-class-members/list", mUser.GetIsShowClassMembers)
			// v1.GET("/test", mQuestion.Qc)

			v1.POST("/mockexam/list", mockexam.List) // 模考
			v1.POST("/mockexam/edit", mockexam.Edit)
			v1.GET("/mockexam/slottime", mockexam.SlotTime)
			v1.GET("/mockexam/teacher", mockexam.GetTeacher)

			v1.POST("/review/list", mQuestion.ReviewList)
			v1.POST("/review/edit", mQuestion.ReviewEdit)
			v1.POST("/review/relation/class", mQuestion.ReviewEditClass)
			v1.POST("/review/relation/course", mQuestion.ReviewEditCourse)
			v1.POST("/paper/save", mPaper.SavePaper)                 // 保存试卷
			v1.POST("/paper/list", mPaper.PaperList)                 // 试卷列表
			v1.GET("/paper/info", mPaper.PaperInfo)                  // 试卷详情
			v1.GET("/paper/info/questions", mPaper.QuestionsInPaper) // 试卷下包含的试题
			v1.POST("/html/convert", mUpload.ConvertHtml)            // html文本转通用结构

			v1.POST("/user/upload/file/list", mUpload.UploadFileList)
			v1.GET("/user/upload/file/info", mUpload.UploadFileInfo)
			v1.POST("/user/upload/file/save", mUpload.UploadFileSave)

			v1.GET("/morning_read/tags/list", mMorningReadTag.List)                    // 晨读tag
			v1.POST("/morning_read/tags", mMorningReadTag.EditTag)                     // tag add
			v1.DELETE("/morning_read/tags", mMorningReadTag.DelTag)                    // tag del
			v1.GET("/morning_read/list", mMorningRead.MorningReadList)                 // 晨读
			v1.GET("/morning_read/item_log_list", mMorningRead.MorningReadItemLogList) // 晨读记录
			v1.GET("/morning_read/log_list", mMorningRead.MorningReadLogList)          // 晨读记录
			v1.POST("/morning_read/save", mMorningRead.MorningReadSave)
			v1.GET("/morning_read/del", mMorningRead.MorningReadDel)
			v1.POST("/morning_read/import", mMorningRead.Import)
			v1.POST("/gptcountinfo", mUser.GPTCountInfo)

			v1.POST("/study/course/list", mStudy.CourseList)
			v1.POST("/study/course/edit", mStudy.CourseEdit)
			v1.POST("/study/data-pack/list", mStudy.DataPackList)
			v1.POST("/study/data-pack/edit", mStudy.DataPackEdit)
			v1.POST("/question/category/keypoints/video-info", video.GetVideoFromKeypoints)           // 根据考点查视频信息（嵌套模式）
			v1.POST("/question/category/keypoints/save-relevance-info", video.SaveUpAndDownRelevance) // 保存子级父级考点关联的视频信息
		}
	}
	// 首页改版
	{
		r.GET("/interview/app/u/v1/keypoint/statistics", home.KeypointStatisticsV2)
		r.GET("/interview/app/n/v1/keypoint/statistics", home.KeypointStatisticsV2)
		r.POST("/interview/app/n/v1/keypoint/qid/list", home.QIDList)
		r.POST("/interview/app/n/v1/question-source/areas", home.QuestionAreas)
	}

	return r
}
