package models

type MockExam struct {
	DefaultField      `bson:",inline"`
	Title             string         `json:"title" bson:"title"`
	SubTitle          string         `json:"sub_title" bson:"sub_title"`
	Introduce         string         `json:"introduce" bson:"introduce"`
	ExamCategory      string         `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string         `json:"exam_child_category" bson:"exam_child_category"` //考试分类
	ExamMode          string         `json:"exam_mode" bson:"exam_mode"`                     // 考试类型  结构化、 结构化小组
	ExamInfo          []ExamInfo     `json:"exam_info" bson:"exam_info"`                     // 考试详情
	SignUpStartTime   string         `json:"sign_up_start_time" bson:"sign_up_start_time"`   // 报名开始时间
	SignUpEndTime     string         `json:"sign_up_end_time" bson:"sign_up_end_time"`       // 报名结束时间
	QuestionType      int8           `json:"question_type" bson:"question_type"`             // 类型（1，单题时间5分钟，2，套题按总时长））
	Question          []MockQuestion `json:"question" bson:"question"`
	Status            int8           `json:"status" bson:"status"`                     // 1 正常 -1 禁用
	SignUpUserNum     int            `json:"sign_up_user_num" bson:"sign_up_user_num"` // 考试报名人数
	MaxUserNum        int            `json:"max_user_num" bson:"max_user_num"`         // 满员人数

	ExamType  int8   `json:"exam_type" bson:"exam_type"`   // 0 普通考试，1 自主练习
	Password  string `json:"password" bson:"password"`     // 密码
	CreatedBy string `json:"created_by" bson:"created_by"` // 创建人  （ExamType是自主练习时用到  guid）

	// 冗余字段
	Log            *MockExamLog      `json:"log" bson:"-"`             // log
	QuestionDetail []GQuestion       `json:"question_detail" bson:"-"` //
	Room           *MockExamSlotRoom `json:"room" bson:"-"`            //
	QuestionNum    int               `json:"question_num" bson:"-"`    //考试的题数
	RoomStatus     string            `json:"room_status" bson:"-"`     // start / close
	TemplateId     string            `json:"template_id" bson:"-"`     // 微信推送模板ID
	UserIds        []string          `json:"user_ids" bson:"-"`        //
}

type ExamInfo struct {
	StartTime    string           `json:"start_time" bson:"start_time"`
	EndTime      string           `json:"end_time" bson:"end_time"`
	SlotTimeList []ExamSlotTime   `json:"slot_time_list" bson:"slot_time_list"` // 时间段
	SlotUserNum  int              `json:"slot_user_num" bson:"slot_user_num"`   // 每场考试人数
	ExamTime     int              `json:"exam_time" bson:"exam_time"`           // 每场考试时长
	RestTime     int              `json:"rest_time" bson:"rest_time"`           // 每场考试间隔时间
	TeacherTeam  [][]MockExamUser `json:"teacher_team" bson:"teacher_team"`     // 老师团队
	RoomRole     string           `json:"room_role" bson:"-"`
}

type MockExamUser struct {
	UserId    string `json:"user_id" bson:"user_id"`
	UserName  string `json:"user_name" bson:"user_name"`
	Avatar    string `json:"avatar" bson:"avatar"`
	LogStatus int8   `json:"log_status" bson:"-"`
}

type ExamSlotTime struct {
	SlotDate      string `json:"slot_date" bson:"slot_date"`             // 时间段
	SlotStartTime string `json:"slot_start_time" bson:"slot_start_time"` // 时间段
	SlotEndTime   string `json:"slot_end_time" bson:"slot_end_time"`     // 时间段
	//SignUpUsers   []string `json:"sign_up_users" bson:"-"`                 // 报名人
	SignUpUserNum int `json:"sign_up_user_num" bson:"-"` // 报名人数
}

type MockQuestion struct {
	QuestionId            string `json:"question_id" bson:"question_id"`
	RequireAnswerDuration int    `json:"require_answer_duration" bson:"require_answer_duration"` // 规定题答题时间
}

func (me *MockExam) TableName() string {
	return "mock_exam"
}

type RoomVideo struct {
	VideoTaskId string         `json:"video_task_id" bson:"video_task_id"`
	VideoUrl    string         `json:"video_url" bson:"video_url"`
	Users       []MockExamUser `json:"users" bson:"users"`
}

type MockExamSlotRoom struct {
	DefaultField  `bson:",inline"`
	MockExamId    string         `json:"mock_exam_id" bson:"mock_exam_id"`
	SlotDate      string         `json:"slot_date" bson:"slot_date"`             // 时间段
	SlotStartTime string         `json:"slot_start_time" bson:"slot_start_time"` // 时间段
	SlotEndTime   string         `json:"slot_end_time" bson:"slot_end_time"`     // 时间段
	TeacherTeam   []MockExamUser `json:"teacher_team" bson:"teacher_team"`       // 老师团队
	Users         []MockExamUser `json:"users" bson:"users"`
	VideoTaskId   string         `json:"video_task_id" bson:"video_task_id"`
	VideoUrl      string         `json:"video_url" bson:"video_url"`
	Status        int8           `json:"status" bson:"status"`       // 0 考试未开始, 1 考试已结束 2 考试进行中
	ExamType      int8           `json:"exam_type" bson:"exam_type"` // 0 普通考试，1 自主练习
	VideoArr      []RoomVideo    `json:"video_arr" bson:"video_arr"` // 视频数组
	// 冗余字段
	Title             string         `json:"title" bson:"-"`
	SubTitle          string         `json:"sub_title" bson:"-"`
	Introduce         string         `json:"introduce" bson:"-"`
	ExamCategory      string         `json:"exam_category" bson:"-"`       //考试分类
	ExamChildCategory string         `json:"exam_child_category" bson:"-"` //考试分类
	ExamMode          string         `json:"exam_mode" bson:"-"`           // 考试类型  结构化、 结构化小组
	ExamStatus        int8           `json:"exam_status" bson:"-"`         // 1 正常 -1 禁用
	MaxUserNum        int            `json:"max_user_num" bson:"-"`        // room最大人数
	Logs              []MockExamUser `json:"logs"`
	RoomRole          string         `json:"room_role" bson:"-"` // teacher / student
}

func (me *MockExamSlotRoom) TableName() string {
	return "mock_exam_slot_room"
}

type MockExamLog struct {
	DefaultField      `bson:",inline"`
	MockExamId        string         `json:"mock_exam_id" bson:"mock_exam_id"`
	SlotDate          string         `json:"slot_date" bson:"slot_date"`             // 时间段
	SlotStartTime     string         `json:"slot_start_time" bson:"slot_start_time"` // 时间段
	SlotEndTime       string         `json:"slot_end_time" bson:"slot_end_time"`     // 时间段
	UserId            string         `json:"user_id" bson:"user_id"`
	ExamCategory      string         `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string         `json:"exam_child_category" bson:"exam_child_category"` //考试分类
	ExamMode          string         `json:"exam_mode" bson:"exam_mode"`                     // 考试类型  结构化、 结构化小组
	RoomId            string         `json:"room_id" bson:"room_id"`                         // 房间号
	TeacherTeam       []MockExamUser `json:"teacher_team" bson:"teacher_team"`               // 老师团队
	Question          []MockQuestion `json:"question" bson:"question"`

	Status          int8           `json:"status" bson:"status"`                     // -1 取消 0已报名待考试 1考试进行中 2过期 5已考试
	SubscribeStatus int8           `json:"subscribe_status" bson:"subscribe_status"` // 0未订阅 1已订阅 2已推送
	AnswerInfo      []AnswerInfo   `json:"answer_info" bson:"answer_info"`
	TeacherCorrect  TeacherCorrect `json:"teacher_correct" bson:"teacher_correct"` // 得分点
	VideoUrls       []string       `json:"video_urls" bson:"video_urls"`           // 视频地址

	// 冗余字段
	QuestionDetail []GQuestion      `json:"question_detail" bson:"-"` //
	Room           MockExamSlotRoom `json:"room" bson:"-"`            //
	RoomStatus     string           `json:"room_status" bson:"-"`     // start / close
	RoomRole       string           `json:"room_role" bson:"-"`       // teacher / student
}

type AnswerInfo struct {
	QuestionId     string           `json:"question_id" bson:"question_id"`
	AnswerDuration int              `json:"answer_duration" bson:"answer_duration"` // 答题时长
	TeacherCorrect []TeacherCorrect `json:"teacher_correct" bson:"teacher_correct"` // 得分点
}

type TeacherCorrect struct {
	Score           string            `bson:"score" json:"score"`
	MarkContent     map[string]string `bson:"mark_content" json:"mark_content"`
	MarkContentSort []string          `bson:"-" json:"mark_content_sort"`
	CommentText     string            `bson:"comment_text" json:"comment_text"`
	VoiceLength     float64           `bson:"voice_length" json:"voice_length"`
	VoiceUrl        string            `bson:"voice_url" json:"voice_url"`
	VoiceText       string            `bson:"voice_text" json:"voice_text"`
}

func (me *MockExamLog) TableName() string {
	return "mock_exam_log"
}
