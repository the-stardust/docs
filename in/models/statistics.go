package models

type GAnswerLogDailyStatistics struct {
	DefaultField      `bson:",inline"`
	Date              string                                `json:"date" bson:"date"`
	LogType           string                                `json:"log_type" bson:"log_type"` // category    job_tag
	ExamCategory      string                                `json:"exam_category" bson:"exam_category"`
	ExamChildCategory string                                `json:"exam_child_category" bson:"exam_child_category"`
	QuestionCategory  string                                `json:"question_category" bson:"question_category"`
	LogCount          int                                   `json:"log_count" bson:"log_count"`                     // log数
	UserCount         int                                   `json:"user_count" bson:"user_count"`                   // 人数
	CompleteUserCount int                                   `json:"complete_user_count" bson:"complete_user_count"` // 完成人数
	OnlineLogCount    int                                   `json:"online_log_count" bson:"online_log_count"`       // 线上用户log数
	OnlineUserCount   int                                   `json:"online_user_count" bson:"online_user_count"`     // 线上用户人数
	ClassLogCount     int                                   `json:"class_log_count" bson:"class_log_count"`         // 地面班log数
	ClassUserCount    int                                   `json:"class_user_count" bson:"class_user_count"`       // 地面班人数
	SubjectCategories map[string]*GAnswerLogDailyStatistics `json:"subject_categories" bson:"subject_categories"`
}

func (sf *GAnswerLogDailyStatistics) TableName() string {
	return "g_interview_answer_logs_daily_statistics"
}
