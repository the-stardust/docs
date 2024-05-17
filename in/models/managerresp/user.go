package managerresp

type FeedBackResp struct {
	Id          string   `json:"id"`
	CreatedTime string   `json:"created_time"`
	UserId      string   `json:"user_id"`
	FastRemark  []string `json:"fast_remark"` // 快捷反馈列表
	Remark      string   `json:"remark"`      // 用户反馈信息
	SourceType  string   `json:"source_type"`
}

type FeedBackList struct {
	List       []FeedBackResp `json:"list"`
	TotalCount int64          `json:"total_count"` //总数量
}
