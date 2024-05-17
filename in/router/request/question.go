package request

type GPTPreviewRequest struct {
	From string   `json:"from" binding:"required"` // 笔试ai 还是 面试ai 或是其他
	Type int      `json:"type" binding:"required"` // 1 试题 2试卷
	Ids  []string `json:"ids" binding:"required"`  // 试题/试卷 id数组
}

type ClickhouseTrackingRequest struct {
	Sign           string `json:"sign"`
	Timestamp      int64  `json:"timestamp"`
	Table          string `json:"table"`
	TableId        int    `json:"table_id"`
	Column         string `json:"column"`
	Condition      string `json:"condition"`
	ConditionExtra string `json:"condition_extra"`
	PageSize       int    `json:"page_size"`
	Page           int    `json:"page"`
	OrderBy        string `json:"order_by"`
	RespType       int    `json:"resp_type"`
}

type DataFunnelingRequest struct {
	Scene         string `json:"scene"`
	StartDate     string `json:"dateStart"`
	EndDate       string `json:"dateEnd"`
	Guids         string `json:"guids"`
	TrackingEvent string
	GuidsArr      []string
}
type DataFunnelingResponse struct {
	UploadCount int      `json:"uploadCount"`
	HitCount    int      `json:"hitCount"`
	MissCount   int      `json:"missCount"`
	Hit         []string `json:"hit"`
	Miss        []string `json:"miss"`
}
