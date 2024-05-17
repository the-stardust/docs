package models

type GatewayUserInfo struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Guid     string `json:"guid"`
}
type MobileInfo struct {
	GUID        string `gorm:"column:guid"`
	MobileID    string `gorm:"column:mobile_id"`
	MobileMask  string `gorm:"column:mobile_mask"`
	Province    string `gorm:"column:province"`
	City        string `gorm:"column:city"`
	NickName    string `gorm:"column:nick_name"`
	Avatar      string `gorm:"column:avatar"`
	AddressCode string `gorm:"column:ad_code"`
	Address     string `gorm:"column:area"`
}

type UserChoiceStatus struct {
	ExamCategory           string `json:"exam_category" redis:"exam_category"`                         // 考试分类
	ExamChildCategory      string `json:"exam_child_category" redis:"exam_child_category"`             //考试子分类
	Province               string `json:"province" redis:"province"`                                   // 省
	City                   string `json:"city" redis:"city"`                                           // 市
	District               string `json:"district" redis:"district"`                                   // 区
	UseNotice              bool   `json:"use_notice" redis:"use_notice"`                               // 用户须知
	UseTip                 bool   `json:"use_tip" redis:"use_tip"`                                     // 用户提示
	PracticeMode           int8   `json:"practice_mode" redis:"practice_mode"`                         // 练习模式， 11是看题-普通模式，12是看题-对镜模式，13是看题-考官模式 21是听题-普通模式，22是听题-对镜模式，23是听题-考官模式
	JobTag                 string `json:"job_tag" redis:"job_tag"`                                     // 岗位标签
	QuestionRealInfo       string `json:"question_real_info" redis:"question_real_info"`               // 记录是否第一次展示试题按钮及看了什么类型的题
	QuestionAnswerAfterTip bool   `json:"question_answer_after_tip" redis:"question_answer_after_tip"` // 点评提示弹窗
	GuidTips               bool   `json:"guid_tips" redis:"guid_tips"`
	NeimengPrediction      bool   `json:"neimeng_prediction" redis:"neimeng_prediction"` //内蒙预测题 弹框
}

type UserUploadFile struct {
	DefaultField      `bson:",inline"`
	ExamCategory      string `json:"exam_category" bson:"exam_category"`             // 考试分类
	ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	UserID            string `json:"user_id" bson:"user_id"`
	FileUrl           string `json:"file_url" bson:"file_url"`
	FileName          string `json:"file_name" bson:"file_name"`
	CheckStatus       int8   `json:"check_status" bson:"check_status"` // 0 已上传成功未审核，1接受， 2部分接受，3不接受
}
