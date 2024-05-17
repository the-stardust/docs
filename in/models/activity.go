package models

// 邀请助力
type InviteUser struct {
	DefaultField      `bson:",inline"`
	UserId            string `bson:"user_id" json:"user_id"`
	InvitedUserId     string `bson:"invited_user_id" json:"invited_user_id"`
	InvitedUserAvatar string `bson:"-" json:"invited_user_avatar"`
	InvitedUserName   string `bson:"-" json:"invited_user_name"`
}

type GPTCountInfo struct {
	DefaultField     `bson:",inline"`
	UserID           string `json:"user_id" bson:"user_id"`
	TotalCount       int    `json:"total_count" bson:"total_count"`               // 获取到的总次数
	AvailableCount   int    `json:"available_count" bson:"available_count"`       // 可用次数
	TotalInviteCount int    `json:"total_invite_count" bson:"total_invite_count"` // 邀请人数
	SendCount        int    `json:"send_count" bson:"send_count"`                 // 赠送次数
	BaipiaoTimeCount int    `json:"baipiao_time_count" bson:"baipiao_time_count"` // 白嫖了几次
	BaipiaoCount     int    `json:"baipiao_count" bson:"baipiao_count"`           // 白嫖获得的总次数
	BuyTimeCount     int    `json:"buy_time_count" bson:"buy_time_count"`         // 购买了几次
	BuyCount         int    `json:"buy_count" bson:"buy_count"`                   // 购买获得的总次数
}

type BaiPiao struct {
	DefaultField `bson:",inline"`
	UserID       string `json:"user_id" bson:"user_id"`
	Count        int    `json:"count" bson:"count"`   // 单次白嫖获得次数
	Reason       string `json:"reason" bson:"reason"` // 白嫖理由
}
