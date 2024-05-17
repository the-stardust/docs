package managerresp

/**
* @Author XuPEngHao
* @Date 2023/12/7 12:14
 */

type QuestionYearsStatisticsResp struct {
	QuestionCategory string `json:"question_category"`
	List             []QuestionYearsDetail
}

type QuestionYearsDetail struct {
	Years     int64 `json:"years"`
	Total     int64 `json:"total"`
	OnlineCnt int64 `json:"online_cnt"`
}

type QuestionStatisticsResp struct {
	Summary Summary                    `json:"summary"`
	List    []QuestionStatisticsDetail `json:"list"`
}

type QuestionStatisticsDetail struct {
	Title        string                       `json:"title"`
	Total        int64                        `json:"total"`
	OnlineCnt    int64                        `json:"online_cnt"`
	CategoryList []QuestionStatisticsCategory `json:"category_list"`
}

type QuestionStatisticsCategory struct {
	Title     string `json:"title"` // 分类名称
	Total     int64  `json:"total"`
	OnlineCnt int64  `json:"online_cnt"`
	SortNum   int64  `json:"-"`
}

type Summary struct {
	Total      int64 `json:"total"`
	OnlineCnt  int64 `json:"online_cnt"`
	OfflineCnt int64 `json:"offline_cnt"`
}
