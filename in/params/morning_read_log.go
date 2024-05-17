package params

type MorningReadLogParam struct {
	ExamCategory      string `json:"exam_category" form:"exam_category"`
	StartTime         string `json:"start_time" form:"start_time"`
	EndTime           string `json:"end_time" form:"end_time"`
	PageSize          int64  `json:"page_size" form:"page_size"`
	PageIndex         int64  `json:"page_index" form:"page_index"`
	ExamChildCategory string `json:"exam_child_category" form:"exam_child_category"`
	JobTag            string `json:"job_tag" form:"job_tag"`
	Province          string `json:"province" form:"province"`
	Keyword           string `json:"keyword" form:"keyword"`
	Uid               string `json:"uid" form:"uid"`
}

type MorningReadLogResponse struct {
	//Id                string   `json:"id" bson:"_id"`                                // 晨读id
	//Uid               string   `bson:"uid" json:"-"`                                 // 晨读用户id
	//NickName          string   `json:"nickname" bson:"-"`                            // 晨读用户
	//Date              string   `json:"date" bson:"date"`                             // 晨读日期
	//Name              string   `json:"name" bson:"name"`                             // 晨读名称
	//ExamCategory      string   `json:"exam_category" bson:"exam_category"`           // 晨读分类
	//SubjectCategory   string   `json:"subject_category" bson:"subject_category"`     // 科目分类
	//QuestionKeypoints []string `json:"question_keypoints" bson:"question_keypoints"` // 知识点
	//ReadCostTime      int      `bson:"read_cost_time" json:"read_cost_time"`         // 晨读花费时间
	//HasReadTime  int   `json:"has_read_time"`                                        // 晨读次数
	//ReadContent []string `json:"read_content" bson:"read_content"`                   // 晨读内容
	//ReadAnswer  []string `json:"read_answer" bson:"read_answer"`                     //晨读回答
	//ExamChildCategory string  `json:"exam_child_category" bson:"exam_child_category"` //晨读子分类
	//LatestReport string   `bson:"latest_report" json:"latest_report"`
	MorningReadLogResponseParam
	MorningReadServiceParamList
}

type MorningReadLogResponseParam struct {
	Id               string                        `json:"id" bson:"_id"`
	Uid              string                        `bson:"uid" json:"uid"`
	Nickname         string                        `json:"nickname" bson:"-"`
	MorningReadParam []MorningReadServiceParamList `bson:"morning_read" json:"-"`
	ReadTime         string                        `json:"read_time" bson:"updated_time"`
	ReadCostTime     int                           `bson:"read_cost_time" json:"read_cost_time"`
	HasReadTims      int                           `json:"has_read_times" bson:"has_read_times"` // 朗读次数
	ReadContent      []string                      `json:"read_content" bson:"read_content"`
	ReadAnswer       []string                      `json:"read_answer" bson:"read_answer"`
	CheckAnswer      []string                      `json:"check_answer" bson:"check_answer"`
	CheckScore       int32                         `bson:"check_score" json:"check_score"`
	LatestReport     string                        `bson:"latest_report" json:"latest_report"`
	Checked          bool                          `bson:"checked" json:"checked"`
	CheckCostTime    int32                         `bson:"check_cost_time" json:"check_cost_time"`
}

type MorningReadServiceParamList struct {
	Name              string   `json:"name" bson:"name"`
	SubName           string   `json:"sub_name" bson:"sub_name"`
	ExamCategory      []string `json:"exam_category" bson:"exam_category"`
	SubjectCategory   string   `json:"subject_category" bson:"subject_category"`
	JobTag            []string `json:"job_tag" bson:"job_tag"`
	Province          []string `json:"province" bson:"province"`
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"`
	QuestionContent   []string `json:"-" bson:"question_content"`
	QuestionAnswer    []string `json:"-" bson:"question_answer"`
}
