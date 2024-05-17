package params

type (
	CurriculaListRequestParam struct {
		Id       string `json:"id" form:"id"`
		Status   int    `json:"status" form:"status"`
		Page     int64  `json:"page" form:"page" `
		PageSize int64  `json:"page_size" form:"page_size"`
		UserId   string
	}
	CurriculaRequestParam struct {
		Id             string `json:"id"  form:"id"`
		CurriculaTitle string `json:"curricula_title" form:"curricula_title"`
		Sort           int    `json:"sort" form:"sort"`
		Status         int    `json:"status" form:"status"`
		IsDelete       int8   `json:"is_delete" form:"is_delete"`
		Name           string `json:"name" form:"name"`
		CreateUserId   string
		UpdatedTime    string
	}
	CurriculaInviteCodeRequestParam struct {
		Id     string `json:"id"  form:"id"`
		UserId string
	}
	CurriculaInviteCodeUseRequestParam struct {
		InviteCode string `json:"invite_code" form:"invite_code"`
		Name       string `json:"name" form:"name"`
		UserId     string
	}
	CurriculaAdminUserIdRequestParam struct {
		Id           string `json:"id" form:"id"`
		UserId       string `json:"user_id" form:"user_id"`
		ActionUserId string
	}
	CurriculaAdminListRequestParam struct {
		Id           string `json:"id" form:"id"`
		ActionUserId string
		AppCode      string
	}
	AdminClassListRequestParam struct {
		StartTime   string `json:"start_time" form:"start_time"`
		EndTime     string `json:"end_time" form:"end_time"`
		CurriculaId string `json:"curricula_id" form:"curricula_id"`
		Page        int64  `json:"page" form:"page"`
		PageSize    int64  `json:"page_size" form:"page_size"`
	}
	AdminSaveInterviewQuestionRequestParam struct {
		CurriculaId      string   `json:"curricula_id"`
		ClassIdList      []string `json:"class_id_list" binding:"required"` // 班级id 逗号分隔
		Name             string   `json:"name" binding:"required"`          // 试题名称
		Desc             string   `json:"desc"`                             // 试题描述
		IsTeacherComment int8     `json:"is_teacher_comment"`               //老师点评 0不可以 1可以
		IsStudentComment int8     `json:"is_student_comment"`               //学生点评 0不可以 1可以
		Status           int32    `json:"status"`                           // 试题状态
		AnswerStyle      int8     `json:"answer_style"`                     //0默认都可以看见试题 1.排队模式只有老师放行学生 学生才能看到题
		UserId           string   `json:"user_id"`
	}
)

type (
	CurriculaListResponseParam struct {
		Total int64            `json:"total"`
		Data  []CurriculaParam `json:"data"`
	}
	CurriculaInviteCodeResponseParam struct {
		InviteCode string `json:"invite_code"`
	}
	CurriculaInviteCodeUseResponseParam struct {
		Tips string `json:"tips"`
	}
)

type (
	CurriculaParam struct {
		Id             string           `json:"id" bson:"_id"`
		CurriculaTitle string           `json:"curricula_title" bson:"curricula_title"`
		Sort           int              `json:"sort" bson:"sort"`
		Status         int              `json:"status" bson:"status"`
		CreateUserId   string           `json:"create_user_id" bson:"create_user_id"`
		AdminList      []AdminListParam `json:"admin_list" bson:"admin_list"`
		Name           string           `json:"name" bson:"name"`
		CreatedTime    string           `json:"created_time" bson:"created_time"`
	}
	AdminListParam struct {
		AdminId     string `json:"admin_id" bson:"admin_id"`
		Type        int    `json:"type" bson:"type"`
		Name        string `json:"name" bson:"name"`
		CreatedTime string `json:"created_time" bson:"created_time"`
		Avatar      string `json:"avatar" bson:"-"` //头像
	}

	GroupClassParam struct {
		ClassId string `json:"_id" bson:"_id"`
		Count   int64  `json:"count" bson:"count"`
	}
)
