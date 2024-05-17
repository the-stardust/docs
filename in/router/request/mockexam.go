package request

import "interview/models"

type MockexamEditRequest struct {
	Id                string                `json:"id"`
	Title             string                `json:"title"`
	SubTitle          string                `json:"sub_title"`
	Introduce         string                `json:"introduce"`
	ExamCategory      string                `json:"exam_category"`       // 考试分类
	ExamChildCategory string                `json:"exam_child_category"` //考试分类
	ExamMode          string                `json:"exam_mode"`           // 考试类型  结构化、 结构化小组
	ExamInfo          []models.ExamInfo     `json:"exam_info"`           // 考试详情
	SignUpStartTime   string                `json:"sign_up_start_time"`  // 报名开始时间
	SignUpEndTime     string                `json:"sign_up_end_time"`    // 报名结束时间
	QuestionType      int8                  `json:"question_type"`
	Question          []models.MockQuestion `json:"question"`
	Status            int8                  `json:"status"` // 1 正常 -1 禁用
}

// 报名
type SignUpRequest struct {
	ExamId        string `json:"exam_id"`
	SlotDate      string `json:"slot_date"`
	SlotStartTime string `json:"slot_start_time"`
	SlotEndTime   string `json:"slot_end_time"`
}

type SlotTimeRequest struct {
	ExamTime  int    `json:"exam_time"` // 秒
	RestTime  int    `json:"rest_time"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// 音视频权限
type ZegoTokenRequest struct {
	RoomId string `json:"room_id"`
	Uid    string `json:"uid"`
}

// 老师控制学生音视频
//
//	{
//		"mock_exam_id": "123abc",
//		"action": {
//			"uid1": {
//				"video": -1,
//				"audio": 1
//			},
//			"uid2": {
//				"video": 1,
//				"audio": -1
//			}
//		}
//	}
type TeacherControlRequest struct {
	RoomId string                    `json:"room_id"`
	Action map[string]map[string]int `json:"action"`
}

type RoomStatusRequest struct {
	RoomId string `json:"room_id"`
	Action string `json:"action"`
	Val    string `json:"val"`
}

type TeacherMarkRequest struct {
	RoomId string `json:"room_id"`
	UserId string `json:"user_id"`
	models.TeacherCorrect
}

// 创建自主练习
type MockExamCreateRequest struct {
	Title             string `json:"title"`
	ExamCategory      string `json:"exam_category"`       // 考试分类
	ExamChildCategory string `json:"exam_child_category"` //考试分类
	ExamMode          string `json:"exam_mode"`           // 考试类型  结构化、 结构化小组
	Password          string `json:"password"`            // 密码
	Uid               string `json:"-"`
	MaxUserNum        int    `json:"max_user_num"` // 房间人数
}

type RoomMemberChangeRequest struct {
	Uid    string `json:"uid" form:"uid"`
	RoomId string `json:"room_id" form:"room_id"`
	ExamId string `json:"exam_id" form:"exam_id"`
	Action string `json:"action" form:"action"` // in / out
}
