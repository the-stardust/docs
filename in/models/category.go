package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// ExamCategory 考试分类
type ExamCategoryItem struct {
	Title string `json:"title" bson:"title"`
	IntId int32  `json:"int_id" bson:"int_id"`
}
type ExamCategory struct {
	Id            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title         string             `json:"title" bson:"title"`
	ChildCategory []ExamCategoryItem `json:"child_category" bson:"child_category"`
	Status        int8               `json:"status" bson:"status"`
	SortNumber    int32              `json:"sortNumber" bson:"sort_number"` // 排序
	IntId         int32              `json:"int_id" bson:"int_id"`
}

// 题分类
type QuestionCategoryItem struct {
	Title                string                 `json:"title" bson:"title"`
	QuestionCount        int64                  `json:"question_count" bson:"question_count"`
	RealQuestionCount    int64                  `json:"real_question_count" bson:"real_question_count"`
	NotRealQuestionCount int64                  `json:"not_real_question_count" bson:"not_real_question_count"`
	ChildCategory        []QuestionCategoryItem `json:"child_category" bson:"child_category"`
	ShortName            string                 `json:"short_name" bson:"short_name"`
	RelationVideo        []VideoKeypoints       `json:"relation_video" bson:"relation_video"`
}
type QuestionCategory struct {
	DefaultField      `bson:",inline"`
	UpdatedTime       string                 `bson:"updated_time"`
	ExamCategory      string                 `json:"exam_category" bson:"exam_category"`
	ExamChildCategory string                 `json:"exam_child_category" bson:"exam_child_category"`
	Categorys         []QuestionCategoryItem `json:"question_category" bson:"question_category"`
}

func (sf *QuestionCategory) TableName() string {
	return "question_category"
}
