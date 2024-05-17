package models

const (
	KeypointStatisticTable = "g_interview_keypoint_statistic"
)

type KeypointStatistic struct {
	DefaultField      `bson:",inline"`
	ExamCategory      string                   `json:"exam_category" bson:"exam_category"`
	ExamChildCategory string                   `json:"exam_child_category" bson:"exam_child_category"`
	JobTag            string                   `json:"job_tag" bson:"job_tag"`
	Provence          string                   `json:"provence" bson:"provence"`
	QuestionReal      int                      `json:"question_real" bson:"question_real"`
	Keypoint          []KeypointStatisticsResp `json:"keypoint" bson:"keypoint"`
}

type KeypointStatisticsResp struct {
	Title                   string                        `json:"title" bson:"title"`
	AllCate                 string                        `json:"-" bson:"all_cate"`
	AnswerQuestionID        []string                      `json:"-"  bson:"-"`
	AnswerCount             int                           `json:"answer_count"  bson:"-"`
	AllQuestionID           []string                      `json:"-" bson:"all_question_id"`
	AllQuestionCount        int                           `json:"all_question_count" bson:"all_question_count"`
	AllQuestionIDMap        map[string]QuestionSimpleItem `json:"-" bson:"-"`
	LastViewQuestionID      string                        `json:"last_view_question_id"  bson:"-"` // 上次浏览的试题 id
	LastViewQuestionIDIndex int                           `json:"-"  bson:"-"`                     // 上次浏览的试题 id
	HasNew                  bool                          `json:"has_new"  bson:"-"`               // 有新题出现,新题的 id
	Child                   []KeypointStatisticsResp      `json:"child" bson:"child"`
}
