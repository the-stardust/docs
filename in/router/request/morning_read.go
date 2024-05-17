package request

type MorningReadSaveRequest struct {
	Id              string   `form:"id" json:"id"`
	Cover           string   `form:"cover" json:"cover"`
	Name            string   `form:"name" json:"name"`         // 晨读名字
	SubName         string   `form:"sub_name" json:"sub_name"` // 副标题
	Tags            []string `form:"tags" json:"tags"`
	StartDate       string   `form:"start_date" json:"start_date"`
	EndDate         string   `form:"end_date" json:"end_date"`
	QuestionContent []string `form:"question_content" json:"question_content" binding:"required" msg:"invalid question_content"`
	QuestionAnswer  []string `form:"question_answer" json:"question_answer" binding:"required" msg:"invalid question_answer"`
	Keywords        []string `form:"keywords" json:"keywords" binding:"required" msg:"invalid keywords"`
	ReadTimes       int      `form:"read_times" json:"read_times"`
	State           int      `form:"state" json:"state"`

	ManagerId       string   `json:"manager_id" form:"manager_id"`
	Sort            int      `form:"sort" json:"sort"`                                                                            // 排序 数字越大越靠前
	Mode            int      `form:"mode" json:"mode" binding:"required" msg:"invalid mode"`                                      // 1练 2考 3练+考
	OpenSilentRead  int8     `bson:"open_silent_read" json:"open_silent_read"`                                                    // 1开启默读模式
	InteractiveMode int8     `form:"interactive_mode" json:"interactive_mode"  binding:"required" msg:"invalid interactive_mode"` // 1 2 两种交互方式
	ExamCategory    []string `json:"exam_category" form:"exam_category"`                                                          //考试分类
	//ExamChildCategory  string `json:"exam_child_category" form:"exam_child_category"`                                              //考试子分类
	JobTag             []string `json:"job_tag" form:"job_tag"`   //岗位标签
	Province           []string `json:"province" form:"province"` //省份
	PreviewBeforeExam  int      `json:"preview_before_exam" form:"preview_before_exam"`
	ShareImg           string   `json:"share_img" form:"share_img"`
	RelationQuestionId string   `json:"relation_question_id" bson:"relation_question_id"` // 关联的题目ID
	//ClassId         string `form:"class_id" json:"class_id"`
}

type MorningReadReportRequest struct {
	Id string `form:"id" json:"id" binding:"required" msg:"invalid id"`
	//ReadTimes int    `form:"read_times" json:"read_times"` // 次数
	Check bool `form:"check" json:"check"`

	CheckContent []string `json:"check_content" form:"check_content"`
	CheckAnswer  []string `form:"check_answer" json:"check_answer"` // 考核时用户提交的

	ReadContent []string `json:"read_content" form:"read_content"`
	ReadAnswer  []string `form:"read_answer" json:"read_answer"` // 晨读时用户提交的

	CostTime      int  `form:"cost_time" json:"cost_time"`             // 用时（秒）
	AllReadReport bool `form:"all_read_report" json:"all_read_report"` // 是否是完整上报
	SourceType    int
}

type MorningGetKeywordsRequest struct {
	QuestionAnswer string `form:"question_answer" json:"question_answer" binding:"required" msg:"invalid question_answer"`
}

type MorningMultiGetKeywordsRequest struct {
	QuestionAnswer map[string]string `form:"question_answer" json:"question_answer" binding:"required" msg:"invalid question_answer"`
}
