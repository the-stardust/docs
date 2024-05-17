package models

// 晨读
type MorningRead struct {
	DefaultField `bson:",inline"`
	//ClassID         string   `json:"class_id" bson:"class_id"`
	Cover string `json:"cover" bson:"cover"` // 封面

	Name            string   `json:"name" bson:"name"`
	SubName         string   `json:"sub_name" bson:"sub_name"`
	Tags            []string `json:"tags" bson:"tags"`
	StartDate       string   `json:"start_date" bson:"start_date"`
	EndDate         string   `json:"end_date" bson:"end_date"`
	QuestionContent []string `json:"question_content" bson:"question_content"`
	QuestionAnswer  []string `json:"question_answer" bson:"question_answer"`
	Keywords        []string `json:"keywords" bson:"keywords"`
	State           int      `json:"state" bson:"state"`           // 1上架 2下架
	ReadTimes       int      `json:"read_times" bson:"read_times"` // 要求读的次数
	ManagerId       string   `json:"manager_id" bson:"manager_id"` // create/edit user

	Mode            int      `bson:"mode" json:"mode"`                         // 1练 2考 3练+考
	OpenSilentRead  int8     `bson:"open_silent_read" json:"open_silent_read"` // 1开启默读模式
	InteractiveMode int8     `bson:"interactive_mode" json:"interactive_mode"` // 1 2 两种交互方式
	ExamCategory    []string `json:"exam_category" bson:"exam_category"`       //考试分类  支持多选
	//ExamChildCategory  string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	JobTag             []string `json:"job_tag" bson:"job_tag"`                         //岗位标签  支持多选
	Province           []string `json:"province" bson:"province"`                       //省份  支持多选
	PreviewBeforeExam  int      `json:"preview_before_exam" form:"preview_before_exam"` //考前是否可以查看试题和答案 1可以 0不可以
	Sort               int      `json:"sort" form:"sort"`
	ShareImg           string   `json:"share_img" bson:"share_img"`                       // 分享图
	RelationQuestionId string   `json:"relation_question_id" bson:"relation_question_id"` // 关联的题目ID 可以逗号分隔

	// ----------- 冗余MorningReadLog字段 （接口用）start-------
	HasReadTimes          int                 `json:"has_read_times" bson:"-"` // 已读的次数
	Checked               bool                `json:"checked" bson:"-"`        // 是否已考试
	ReadContent           []string            `json:"read_content" bson:"-"`
	ReadAnswer            []string            `json:"read_answer" bson:"-"`
	CheckContent          []string            `json:"check_content" bson:"-"`
	CheckAnswer           []string            `json:"check_answer" bson:"-"`
	CheckAnswerResult     []CheckAnswerResult `json:"check_answer_result" bson:"-"`      // 关键词的对错
	CheckScore            int                 `json:"check_score" bson:"-"`              // 考试总分
	ReadCostTime          int                 `json:"read_cost_time" json:"-"`           // 用时（秒）
	CheckCostTime         int                 `json:"check_cost_time" json:"-"`          // 用时（秒）
	CheckKeywordsNum      int                 `json:"check_keywords_num" bson:"-"`       // 考试的关键词数量
	CheckMatchKeywordsNum int                 `json:"check_match_keywords_num" bson:"-"` // 考试命中的关键词数量
	LatestReport          string              `json:"latest_report" bson:"-"`            // 最新上报是读还是考试
	LogCreatedTime        string              `json:"log_created_time" bson:"-"`         // 晨读记录的时间
	LogCount              int                 `json:"log_count" bson:"-"`                // 对应的晨读记录数量
	Uid                   string              `json:"uid" bson:"-"`
	Nickname              string              `json:"nickname" bson:"-"`
	Avatar                string              `json:"avatar" bson:"-"`
	MobileAffiliation     string              `json:"mobile_affiliation" bson:"-"`
	MobileID              string              `json:"mobile_id" bson:"-"`
	// ----------- 冗余 end ------------
}

func NewMorningRead() *MorningRead {
	return &MorningRead{}
}

func (sf *MorningRead) TableName() string {
	return "morning_read"
}

type MorningReadLog struct {
	DefaultField      `bson:",inline"`
	MorningReadId     string   `json:"morning_read_id" bson:"morning_read_id"`
	MorningReadIdTags []string `json:"morning_read_id_tags" bson:"morning_read_id_tags"`
	//ClassID       string `json:"class_id" bson:"class_id"`
	Uid                   string              `json:"uid" bson:"uid"`
	Date                  string              `json:"date" bson:"date"`
	HasReadTimes          int                 `json:"has_read_times" bson:"has_read_times"`
	Checked               bool                `json:"checked" bson:"checked"`
	ReadContent           []string            `json:"read_content" bson:"read_content"`
	ReadAnswer            []string            `json:"read_answer" bson:"read_answer"`
	CheckContent          []string            `json:"check_content" bson:"check_content"`
	CheckAnswer           []string            `json:"check_answer" bson:"check_answer"`
	CheckAnswerResult     []CheckAnswerResult `json:"check_answer_result" bson:"check_answer_result"`           // 关键词的对错
	CheckScore            int                 `json:"check_score" bson:"check_score"`                           // 考试总分
	CheckKeywordsNum      int                 `json:"check_keywords_num" bson:"check_keywords_num"`             // 考试的关键词数量
	CheckMatchKeywordsNum int                 `json:"check_match_keywords_num" bson:"check_match_keywords_num"` // 考试命中的关键词数量
	ReadCostTime          int                 `bson:"read_cost_time" json:"read_cost_time"`                     // 用时（秒）
	CheckCostTime         int                 `bson:"check_cost_time" json:"check_cost_time"`                   // 用时（秒）
	LatestReport          string              `bson:"latest_report" json:"latest_report"`                       // 最新上报是读还是考试
	SourceType            int                 `bson:"source_type" json:"source_type"`                           // 来源
	Status                int                 `bson:"status" json:"status"`                                     // 状态 1完成
}

type CheckAnswerResult struct {
	Score    int              `json:"score" bson:"score"`
	Keywords []map[string]any `bson:"keywords" json:"keywords"`
	Pass     bool             `json:"pass" bson:"pass"`
}

func (sf *MorningReadLog) TableName() string {
	return "morning_read_log"
}

type MorningReadTag struct {
	DefaultField `bson:",inline"`
	Name         string `json:"name" bson:"name"`
	Cover        string `json:"cover" bson:"cover"`
	IsDeleted    int    `json:"is_deleted" bson:"is_deleted"`
}

const (
	MorningReadTagTable = "morning_read_tags"
)
