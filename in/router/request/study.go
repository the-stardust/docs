package request

type RecommendCourseEditRequest struct {
	Id                string   `json:"id"` // id
	Title             string   `bson:"title" json:"title"`
	CourseType        int      `json:"course_type" bson:"course_type"`                 // 课程类型
	Province          []string `json:"province" bson:"province"`                       //省份
	ProvinceCode      []string `json:"province_code" bson:"province_code"`             //省份代码
	JobTag            string   `json:"job_tag" bson:"job_tag"`                         //岗位标签
	Year              string   `json:"year" bson:"year"`                               // 年份
	ExamCategory      string   `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	OriginData        string   `json:"origin_data" bson:"origin_data"`                 // 原始数据 JSON
	Sort              int      `json:"sort" bson:"sort"`                               // 排序
	Status            int8     `json:"status" bson:"status"`                           // 1 正常 -1 禁用
}

type RecommendDataPackEditRequest struct {
	Id                string `json:"id"` // id
	Title             string `bson:"title" json:"title"`
	CoverImg          string `json:"cover_img" bson:"cover_img"`                     // 封面
	ResUrl            string `json:"res_url" bson:"res_url"`                         //资源地址
	ResId             string `json:"res_id" bson:"res_id"`                           //资源id
	ResTitle          string `json:"res_title" bson:"res_title"`                     //资源title
	JobTag            string `json:"job_tag" bson:"job_tag"`                         //岗位标签
	ExamCategory      string `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	Sort              int    `json:"sort" bson:"sort"`                               // 排序
	Status            int8   `json:"status" bson:"status"`                           // 1 正常 -1 禁用
}
