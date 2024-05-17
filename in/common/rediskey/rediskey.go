package rediskey

type RedisKey string

const (
	ProName                   RedisKey = "interview:"
	ManagerId2Name            RedisKey = ProName + "manager_id_2_name"
	InterviewMemberInfo       RedisKey = ProName + "interview_member_info"
	InterviewQuestionId2Name  RedisKey = ProName + "interview_question_id_2_Name"
	InterviewAdminAuthUserIds RedisKey = ProName + "interview_admin_auth_user_ids"

	AnswerVoiceUrlInfo

	InterviewGPTQuestionId2Name RedisKey = ProName + "interview_gpt_question_id_2_name"
	InterviewGPTUserInfo        RedisKey = ProName + "interview_gpt_user_info"
	InterviewGPTAnswerSet       RedisKey = ProName + "interview_gpt_answer_set"
	InterviewGPTCommentSet      RedisKey = ProName + "interview_gpt_comment_set"

	GPTAnswerWaiting   RedisKey = ProName + "interview_gpt_answer:waiting"
	GPTAnswerCompleted RedisKey = ProName + "interview_gpt_answer:completed"
	GPTAnswerFailed    RedisKey = ProName + "interview_gpt_answer:failed"

	GPTCommentWaiting   RedisKey = ProName + "interview_gpt_comment:waiting"
	GPTCommentCompleted RedisKey = ProName + "interview_gpt_comment:completed"
	GPTCommentFailed    RedisKey = ProName + "interview_gpt_comment:failed"
	GPTCommentSuccess   RedisKey = ProName + "interview_gpt_comment:success"

	GPTQuestionAnswerWaiting   RedisKey = ProName + "interview_gpt_question_answer:waiting"
	GPTQuestionAnswerCompleted RedisKey = ProName + "interview_gpt_question_answer:completed"
	GPTQuestionAnswerFailed    RedisKey = ProName + "interview_gpt_question_answer:failed"

	WebRelayCache   RedisKey = ProName + "web_relay_cache"
	UserInfo        RedisKey = ProName + "user_info:"
	UseGPTCountInfo RedisKey = ProName + "user_gpt_count_info:"

	GPTCustomWaiting      RedisKey = ProName + "interview_gpt_custom:waiting"
	GPTCustomCompleted    RedisKey = ProName + "interview_gpt_custom:completed"
	GPTCustomFailed       RedisKey = ProName + "interview_gpt_custom:failed"
	GPTCustomSuccess      RedisKey = ProName + "interview_gpt_custom:success"
	UserChoice                     = ProName + "user_choice:"
	TTSPrompt                      = ProName + "tts_prompt:"
	GPTQuestionPreview             = ProName + "gpt_preview:"
	GPTQuestionPreviewing          = ProName + "gpt_previewing:" // 1生成中2已生成3部分生成4生成失败
	IsShowGPTPage                  = ProName + "is_show_gpt_page"

	CategoryPermissionInterview = "question_bank:category_permission_interview"
	QuestionAreas               = ProName + "question_areas"
	WechatStudyGroup            = ProName + "wechat_study_group"
	// IsShowClass                 = ProName + "is_show_class"
	MockExamStatus    = ProName + "mockexam_status:"
	MockExamSubscribe = ProName + "mockexam_subscribe:"
	WxAccessToken     = ProName + "wx_access_token"
	ActivityPrefix    = ProName + "activity:"

	InterviewCurriculaAdminUserIdHash = ProName + "curricula_admin_user_id:%s"
	InterviewCurriculaCodeString      = ProName + "curricula_code:%s"
	InterviewCurriculaTitleString     = ProName + "curricula_title:%s"

	UserAnswerKeyPointRecord = ProName + "user_keypoint_record:%s_%s_%s_%s_%d"
)
