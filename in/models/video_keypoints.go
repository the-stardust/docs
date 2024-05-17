package models

import "go.mongodb.org/mongo-driver/bson"

type KeyPointsInfo struct {
}

type VideoKeypoints struct {
	DefaultField      `bson:",inline"`
	MediaId           string   `json:"media_id" bson:"media_id"` // 节详情里的Vid
	LessonId          int      `json:"lesson_id" bson:"lesson_id"`
	ExamCategory      string   `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory  []string `json:"question_category" bson:"question_category"`
	StartTime         int      `json:"start_time" bson:"start_time"`
	EndTime           int      `json:"end_time" bson:"end_time"`
	FileId            string   `json:"file_id" bson:"file_id"`
	VideoUrl          string   `json:"video_url" bson:"video_url"`
	Desc              string   `json:"desc" bson:"desc"`
	JiangjieKey       string   `json:"jiangjie_key" bson:"jiangjie_key"`
	Sort              int      `json:"sort" bson:"sort"`
	FinalUrl          string   `json:"final_url" bson:"final_url"`
	IsFromChild       bool     `json:"is_from_child" bson:"-"`
}

func (sf *VideoKeypoints) GetKeypointsVideo(ExamCategory, ExamChildCategory, SubjectCategory string, QuestionKeypoints []string) (error, []VideoKeypoints) {
	// filter := bson.M{"question_keypoints": QuestionKeypoints, "video_url": bson.M{"$exists": true, "$ne": ""}}
	filter := bson.M{"question_keypoints": QuestionKeypoints}
	if ExamCategory != "" {
		filter["exam_category"] = ExamCategory
	}
	if ExamChildCategory != "" {
		filter["exam_child_category"] = ExamChildCategory
	}
	if SubjectCategory != "" {
		filter["subject_category"] = SubjectCategory
	}
	vks := make([]VideoKeypoints, 0)
	err := sf.DB().Collection("video_keypoints").Where(filter).Sort("start_time").Find(&vks)
	if err != nil {
		return err, vks
	}
	return nil, vks
}

type VideoQuestions struct {
	DefaultField      `bson:",inline"`
	MediaId           string `json:"media_id" bson:"media_id"` // 节详情里的Vid
	LessonId          int    `json:"lesson_id" bson:"lesson_id"`
	ExamCategory      string `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	StartTime         int    `json:"start_time" bson:"start_time"`
	EndTime           int    `json:"end_time" bson:"end_time"`
	FileId            string `json:"file_id" bson:"file_id"`
	VideoUrl          string `json:"video_url" bson:"video_url"`
	Desc              string `json:"desc" bson:"desc"`
	Sort              int    `json:"sort" bson:"sort"`
	FinalUrl          string `json:"final_url" bson:"final_url"`
	QuestionID        string `json:"question_id" bson:"question_id"`
}

type VideoKeypointsForRelevance struct {
	DefaultField      `bson:",inline"`
	MediaId           string   `json:"media_id" bson:"media_id"` // 节详情里的Vid
	LessonId          int      `json:"lesson_id" bson:"lesson_id"`
	ExamCategory      string   `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory  []string `json:"question_category" bson:"question_category"`
	StartTime         int      `json:"start_time" bson:"start_time"`
	EndTime           int      `json:"end_time" bson:"end_time"`
	FileId            string   `json:"file_id" bson:"file_id"`
	VideoUrl          string   `json:"video_url" bson:"video_url"`
	Desc              string   `json:"desc" bson:"desc"`
	JiangjieKey       string   `json:"jiangjie_key" bson:"jiangjie_key"`
	Sort              int      `json:"sort" bson:"sort"`
	FinalUrl          string   `json:"final_url" bson:"final_url"`
	VKId              string   `json:"-" bson:"vk_id"` // video_keypoints里的id
}

func (v *VideoQuestions) GetQuestionVideoList(qid string) []VideoQuestions {
	var list []VideoQuestions
	filter := bson.M{"question_id": qid}
	_ = v.DB().Collection("video_questions").Where(filter).Find(&list)
	return list
}

func (v *VideoKeypoints) GetKeypointVideoList(examCate, examChildCate string, questionCategory []string) []VideoKeypoints {
	var list []VideoKeypoints
	filter := bson.M{"question_category": questionCategory, "video_url": bson.M{"$ne": ""}}
	filter["exam_category"] = examCate
	if examChildCate != "" {
		filter["exam_child_category"] = examChildCate
	}
	_ = v.DB().Collection("video_keypoints").Where(filter).Find(&list)
	return list
}
