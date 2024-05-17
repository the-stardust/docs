package models

// 关联课程信息
type QSetCourse struct {
	CourseId          string `json:"course_id" bson:"course_id"`
	CourseName        string `json:"course_name" bson:"course_name"`
	ChapterId         string `json:"chapter_id" bson:"chapter_id"`
	ChapterName       string `json:"chapter_name" bson:"chapter_name"`
	CourseWorkBtnName string `json:"course_work_btn_name"  bson:"course_work_btn_name"`
}

type QSetClassInfo struct {
	ClassID         string `json:"class_id" bson:"class_id"`                   //关联的班级id
	ClassName       string `json:"class_name" bson:"class_name"`               //关联的班级id
	ClassReviewDate string `json:"class_review_date" bson:"class_review_date"` //班级作业创建的日期
}

type Area struct {
	Province string `json:"province" bson:"province"`
	City     string `json:"city" bson:"city"`
	District string `json:"district" bson:"district"`
}

// 试题集合
type Paper struct {
	DefaultField      `bson:",inline"`
	Title             string   `bson:"title" json:"title"`
	PaperType         int8     `json:"paper_type" bson:"paper_type"` //1真题卷 2模拟卷 3集合卷
	ShortId           string   `json:"short_id" bson:"short_id"`     // 短ID 主要用于分享拼链接参数
	ManagerId         string   `json:"manager_id" bson:"manager_id"`
	ManagerName       string   `json:"manager_name" bson:"manager_name"`
	ExamCategory      string   `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	Year              string   `bson:"year" json:"year"`
	Area              []Area   `bson:"area" json:"area"`
	QuestionIds       []string `bson:"question_ids" json:"question_ids"`
	QuestionCount     int      `json:"question_count" bson:"question_count"`
	Status            int8     `json:"status" bson:"status"` // 1 正常 -1 禁用

	// 冗余字段
	UserNum int `json:"user_num" bson:"-"` // 用户数
	LogNum  int `json:"log_num" bson:"-"`  // 做过的数
}

func (sf *Paper) TableName() string {
	return "paper"
}

// 测评
type Review struct {
	DefaultField      `bson:",inline"`
	Title             string `bson:"title" json:"title"`
	ShortId           string `json:"short_id" bson:"short_id"`
	PaperId           string `json:"paper_id" bson:"paper_id"`
	ManagerId         string `json:"manager_id" bson:"manager_id"`
	ManagerName       string `json:"manager_name" bson:"manager_name"`
	ExamCategory      string `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	Status            int8   `json:"status" bson:"status"`                           // 1 正常 -1 禁用

	ScoreType int           `bson:"score_type" json:"score_type"` // 打分方式 1仅GPT 2仅老师 3先GPT后老师
	Course    QSetCourse    `json:"course" bson:"course"`         // 关联的课程信息
	Class     QSetClassInfo `json:"class" bson:"class"`           // 关联的班级信息

	// 冗余字段
	Questions []GQuestion `json:"questions" bson:"-"` // 题
	//UserNum     int         `json:"user_num" bson:"-"`  // 用户数
	//LogNum      int         `json:"log_num" bson:"-"`   // 做过的数
	QuestionCount     int `json:"question_count" bson:"-"`
	AnswerCount       int `json:"answer_count" bson:"-"`  // 回答的题数量
	CorrectCount      int `json:"correct_count" bson:"-"` // 已批改的题数量
	AnswerPersonCount int `json:"people_count" bson:"-"`
}

type ReviewExerciseTotalCount struct {
	Id    string `json:"id" bson:"_id"`
	Total int    `json:"total" bson:"total"`
}

func (sf *Review) TableName() string {
	return "review"
}

// 试题集合记录
type ReviewLog struct {
	DefaultField      `bson:",inline"`
	ExamCategory      string        `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string        `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	UserId            string        `json:"user_id" bson:"user_id"`
	ReviewId          string        `json:"review_id" bson:"review_id"` //集合id
	Questions         []LogQuestion `json:"questions" bson:"questions"`
	AnswerCount       int           `json:"answer_count" bson:"answer_count"`     // 已回答的题数量
	Status            int8          `json:"status" bson:"status"`                 // 0 未答 5已完成
	CorrectStatus     int8          `json:"correct_status" bson:"correct_status"` // 0 未批改 1部分批改 2已全部批改
	Course            QSetCourse    `json:"course" bson:"course"`                 // 关联的课程信息
	Class             QSetClassInfo `json:"class" bson:"class"`                   // 关联的班级信息

	// 冗余字段 Review
	ScoreType     int    `bson:"-" json:"score_type"` // 打分方式 1仅GPT 2仅老师 3先GPT后老师
	QuestionCount int    `json:"question_count" bson:"-"`
	ReviewStatus  int8   `json:"review_status" bson:"-"`
	ReviewName    string `json:"review_name" bson:"-"` //集合名称
}

func (sf *ReviewLog) TableName() string {
	return "review_log"
}

type LogQuestion struct {
	QuestionId        string           `json:"question_id" bson:"question_id"`
	ExamCategory      string           `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string           `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory  []string         `json:"question_category" bson:"question_category"`     //题分类
	AnswerStatus      int8             `json:"answer_status" bson:"answer_status"`             // 0 未回答 1已回答
	LastAnswerTime    string           `json:"last_answer_time" bson:"last_answer_time"`       // 最近一次回答时间
	CorrectStatus     int8             `json:"correct_status" bson:"correct_status"`           // 0 未批改 1部分批改 2已全部批改
	CorrectNum        int              `json:"correct_num" bson:"-"`                           // 批改数量
	AnswerLogs        []GAnswerLogBase `json:"answer_logs" bson:"answer_logs"`                 // GAnswerLog 表的一些简单数据

	// ---- 题内容 不存表
	Name                string        `json:"name" bson:"-"`                  // 试题名称
	Desc                string        `json:"desc" bson:"-"`                  // 试题描述
	Tags                []string      `json:"tags" bson:"-"`                  // 试题tag
	Answer              string        `json:"answer" bson:"-"`                // 试题答案
	Thinking            string        `json:"thinking" bson:"-"`              // 解题思路（第三方）
	CategoryId          string        `json:"category_id" bson:"-"`           // 试题分类
	JobTag              string        `json:"job_tag" bson:"-"`               // 岗位标签，如海关、税务局等
	QuestionSource      string        `json:"question_source" bson:"-"`       // 试题来源
	TTSUrl              TTSUrl        `json:"tts_url" bson:"-"`               // 合成语音地址
	NameStruct          CommonContent `json:"name_struct" bson:"-" `          // 试题名称
	NameDesc            string        `json:"name_desc" bson:"-"`             // 漫画题的总结性内容
	QuestionContentType int8          `json:"question_content_type" bson:"-"` // 试题类别，0普通题，1漫画题
}

type GAnswerLogBase struct {
	GAnswerLogId  string  `json:"g_answer_log_id" bson:"g_answer_log_id"`
	VoiceLength   float64 `bson:"voice_length" json:"voice_length"`
	VoiceUrl      string  `bson:"voice_url" json:"voice_url"`
	VoiceText     string  `bson:"voice_text" json:"voice_text"`
	CorrectStatus int8    `json:"correct_status" bson:"correct_status"` // 0 未批改 1已批改
}
