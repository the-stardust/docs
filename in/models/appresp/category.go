package appresp

import "go.mongodb.org/mongo-driver/bson/primitive"

// ExamCategory 考试分类
type ExamCategoryItem struct {
	Title        string `json:"title" bson:"title"`
	IntId        int32  `json:"int_id" bson:"int_id"`
	RealCount    int64  `json:"real_count"`
	NotRealCount int64  `json:"not_real_count"`
}
type ExamCategoryResp struct {
	Id            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title         string             `json:"title" bson:"title"`
	ChildCategory []ExamCategoryItem `json:"child_category" bson:"child_category"`
	Status        int8               `json:"status" bson:"status"`
	SortNumber    int32              `json:"sortNumber" bson:"sort_number"` // 排序
	IntId         int32              `json:"int_id" bson:"int_id"`
	RealCount     int64              `json:"real_count"`
	NotRealCount  int64              `json:"not_real_count"`
}
