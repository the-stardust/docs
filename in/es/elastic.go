package es

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"interview/common/global"
	"interview/database"
	"interview/models"
)

var (
	ElasticClient *elastic.Client
	errs          error
	Ctx           context.Context
)

const (
	QuestionIndex   = "interview_questions"
	QuestionMapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                    "by_smart": {
                        "type": "custom",
                        "tokenizer": "ik_smart",
                        "filter": [
                            "by_tfr",
                            "by_sfr"
                        ]
                    }
                },
                "filter": {
                    "by_tfr": {
                        "type": "stop",
                        "stopwords": [
                            ""
                        ]
                    },
                    "by_sfr": {
                        "type": "synonym",
                        "synonyms_path": "analysis/data_pack_search_synonyms.txt"
                    }
                }
            }
        }
    },
    "mappings": {
        "properties": {
            "name": {
                "type": "text",
                "analyzer": "ik_smart",
                "search_analyzer": "by_smart"
            },
            "name_desc": {
                "type": "text",
                "analyzer": "ik_smart",
                "search_analyzer": "by_smart"
            },
            "name_struct": {
                "properties": {
                    "content": {
                    "type": "nested",
                        "properties": {
                            "text": {
                                "type": "text",
                                "fields": {
                                    "keyword": {
                                        "type": "keyword",
                                        "ignore_above": 256
                                    }
                                }
                            }
                        }
                    }
                }
            },
            "question_category": {
                "type": "keyword"
            },
            "id": {
                "type": "keyword"
            },
            "desc": {
                "type": "text",
                "analyzer": "ik_smart",
                "search_analyzer": "by_smart"
            },
            "tags": {
                "type": "text",
                "analyzer": "ik_smart",
                "search_analyzer": "by_smart"
            },
            "job_tag": {
                "type": "keyword"
            },
            "created_time": {
                "type": "keyword"
            },
            "updated_time": {
                "type": "keyword"
            },
            "creator_user_id": {
                "type": "keyword"
            },
            "answer": {
                "type": "text",
                "analyzer": "ik_smart",
                "search_analyzer": "by_smart"
            },
            "exam_category": {
                "type": "keyword"
            },
            "exam_child_category": {
                "type": "keyword"
            },
            "province": {
                "type": "keyword"
            },
            "city": {
                "type": "keyword"
            },
            "district": {
                "type": "keyword"
            },
            "status": {
                "type": "keyword"
            },
            "question_real": {
                "type": "keyword"
            }
        }
    }
}
`

	AnswerLogIndex   = "interview_answer_logs"
	AnswerLogMapping = `
{
    "settings":{
        "index":{
            "analysis":{
                "analyzer":{
                    "by_smart":{
                        "type":"custom",
                        "tokenizer":"ik_smart",
                        "filter":[
                            "by_tfr",
                            "by_sfr"
                        ]
                    }
                },
                "filter":{
                    "by_tfr":{
                        "type":"stop",
                        "stopwords":[
                            ""
                        ]
                    },
                    "by_sfr":{
                        "type":"synonym",
                        "synonyms_path":"analysis/data_pack_search_synonyms.txt"
                    }
                }
            }
        }
    },
    "mappings":{
        "properties":{
            "question_name":{
                "type":"text",
                "analyzer":"ik_smart",
                "search_analyzer":"by_smart"
            },
            "question_category":{
                "type":"keyword"
            },
           "id":{
                "type":"keyword"
            },
            "gpt_standard_answer":{
                "type":"text",
                "analyzer":"ik_smart",
                "search_analyzer":"by_smart"
            },
            "created_time":{
                "type":"keyword"
            },
            "updated_time":{
                "type":"keyword"
            },
           "gpt_comment.content":{
                "type":"text",
                "analyzer":"ik_smart",
                "search_analyzer":"by_smart"
            },
           "exam_category":{
                "type":"keyword"
            },
           "exam_child_category":{
                "type":"keyword"
            },
           "user_id":{
                "type":"keyword"
            },
           "is_deleted":{
                "type":"keyword"
            },
           "province":{
                "type":"keyword"
            },
           "city":{
                "type":"keyword"
            },
           "district":{
                "type":"keyword"
            },
           "status":{
                "type":"keyword"
            },
            "name_struct": {
                "properties": {
                    "content": {
                    "type": "nested",
                        "properties": {
                            "text": {
                                "type": "text",
                                "fields": {
                                    "keyword": {
                                        "type": "keyword",
                                        "ignore_above": 256
                                    }
                                }
                            }
                        }
                    }
                }
            },
          "question_id":{
                "type":"keyword"
            }

        }
    }
}`
)

type GQuestion models.GQuestion
type GAnswerLog models.GAnswerLog

func CreateEsClient() (*elastic.Client, error) {
	ESCfg := global.CONFIG.ES
	ElasticClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(ESCfg.ElasticUrl),
		elastic.SetBasicAuth(ESCfg.ElasticName, ESCfg.ElasticPwd),
	)
	if err != nil {
		fmt.Println(err)
	}
	return ElasticClient, err
}

func InitElastic() {
	ESCfg := global.CONFIG.ES
	Elastic(ESCfg.ElasticUrl, ESCfg.ElasticName, ESCfg.ElasticPwd)
}

func Elastic(url, name, pwd string) {
	var mgoCfg = global.CONFIG.MongoDB
	var err error
	mongoClient := database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, mgoCfg.Dbname)
	Ctx = context.Background()

	ElasticClient, err = elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(url),
		elastic.SetBasicAuth(name, pwd),
	)
	if err != nil {
		fmt.Println(err)
	}

	// 试题记录
	exists, err := ElasticClient.IndexExists(QuestionIndex).Do(Ctx)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if !exists {
		_, err := ElasticClient.CreateIndex(QuestionIndex).Body(QuestionMapping).Do(Ctx)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		// 查询数据传给es
		res := make([]models.GQuestion, 0)
		err = mongoClient.Collection("g_interview_questions").Find(&res)
		if err != nil {
			fmt.Println(err)
		}
		err = QuestionsAddToEs(ElasticClient, Ctx, res)
		if err != nil {
			fmt.Println(err)
		}
	}

	// 作答记录
	answerLogExists, err := ElasticClient.IndexExists(AnswerLogIndex).Do(Ctx)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if !answerLogExists {
		_, err := ElasticClient.CreateIndex(AnswerLogIndex).Body(AnswerLogMapping).Do(Ctx)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		// 查询数据传给es
		res := make([]models.GAnswerLog, 0)
		err = mongoClient.Collection("g_interview_answer_logs").Find(&res)
		if err != nil {
			fmt.Println(err)
		}
		err = AnswerLogAddToEs(ElasticClient, Ctx, res)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// 试题相关

func QuestionsAddToEs(client *elastic.Client, ctx context.Context, dataPack []models.GQuestion) error {
	req := client.Bulk().Index(QuestionIndex)
	for _, v := range dataPack {
		doc := elastic.NewBulkIndexRequest().Id(v.Id.Hex()).Doc(v)
		req.Add(doc)
	}
	if req.NumberOfActions() < 0 {
		return nil
	}
	res, err := req.Do(ctx)
	if err != nil {
		return err
	}

	if !res.Errors {
		return nil
	}
	for _, it := range res.Failed() {
		if it.Error == nil {
			continue
		}
		return &elastic.Error{
			Status:  it.Status,
			Details: it.Error,
		}
	}
	return nil
}

//func QuestionSearch(client *elastic.Client, ctx context.Context, paramMap map[string]interface{}, uid string) ([]models.GQuestion, int, error) {
//	searchIndex := QuestionIndex
//	pageIndex := float64(0)
//	pageSize := float64(0)
//	var kBoost, pBoost, oBoost float64
//	if len(paramMap["keywords"].(string)) <= 10 {
//		kBoost = 3.0
//		pBoost = 3.5
//		oBoost = 1.2
//	} else {
//		kBoost = 3.5
//		pBoost = 3.5
//		oBoost = 1.2
//	}
//	query := elastic.NewBoolQuery()
//	filterList := make([]elastic.Query, 0)
//	for k, v := range paramMap {
//		switch k {
//		case "keywords":
//			query.Should(elastic.NewMatchQuery("name", v.(string)).Boost(kBoost))
//			query.Should(elastic.NewMatchQuery("desc", v.(string)).Boost(oBoost))
//			query.Should(elastic.NewTermQuery("id", v.(string)).Boost(kBoost))
//		case "question_category":
//			filterList = append(filterList, elastic.NewTermsQuery(k, v.([]interface{})...).Boost(pBoost))
//		case "years":
//			filterList = append(filterList, elastic.NewTermsQuery(k, v.([]interface{})...).Boost(pBoost))
//		case "page_index":
//			pageIndex = v.(float64)
//		case "page_size":
//			pageSize = v.(float64)
//		case "gpt_answer_status":
//			v = v.(float64)
//			if v == 1 {
//				query.MustNot(elastic.NewTermQuery("gpt_answer.content", "").Boost(oBoost))
//			} else if v == 2 {
//				query.Must(elastic.NewTermQuery("gpt_answer.content", "").Boost(oBoost))
//			}
//		case "question_real":
//			if v.(float64) != 0 {
//				filterList = append(filterList, elastic.NewTermQuery(k, v).Boost(oBoost))
//			}
//		case "status":
//			if v.(float64) != 0 {
//				filterList = append(filterList, elastic.NewTermQuery(k, v).Boost(oBoost))
//			}
//		case "open_category_permission":
//		default:
//			if v != "" {
//				filterList = append(filterList, elastic.NewTermQuery(k, v))
//			}
//		}
//	}
//
//	// 试题分类权限控制
//	if questionCategory, ok := paramMap["question_category"]; ok && len(questionCategory.([]interface{})) > 0 {
//		if openCategoryPermission, ok := paramMap["open_category_permission"]; ok {
//			if openCategoryPermission.(float64) == 1 {
//				userKeypoints := new(controllers.Controller).UserCategoryPermissionFilter(uid, paramMap["exam_category"].(string), paramMap["exam_child_category"].(string), questionCategory.([]string)[0], "", 2, "")
//				if _, ok := userKeypoints.(string); ok && (userKeypoints.(string) == "not set" || userKeypoints.(string) == "") {
//					for _, v := range questionCategory.([]interface{}) {
//						query.Must(elastic.NewTermQuery("question_category", v))
//					}
//				} else {
//					query.Must(elastic.NewTermQuery("question_category", userKeypoints))
//				}
//			} else {
//				fmt.Println("2222")
//				for _, v := range questionCategory.([]interface{}) {
//					query.Must(elastic.NewTermQuery("question_category", v))
//				}
//			}
//		}
//	} else {
//		if paramMap["open_category_permission"].(float64) == 1 {
//			for _, v := range questionCategory.([]interface{}) {
//				query.Must(elastic.NewTermQuery("question_category", v))
//			}
//		}
//	}
//	ffmt.Puts(query.Source())
//	query.Filter(filterList...)
//	searchResult, err := ElasticClient.Search().
//		Index(searchIndex).
//		Query(query).
//		SortBy(
//			elastic.NewFieldSort("_score").Desc(),
//		).
//		MinScore(4.00).
//		From((int(pageIndex) - 1) * int(pageSize)).
//		Size(int(pageSize)).
//		Do(ctx)
//	if err != nil {
//		fmt.Println(err)
//		panic(err)
//	}
//	total := searchResult.TotalHits()
//	returnQuestionList := make([]models.GQuestion, 0)
//
//	if total > 0 {
//		for _, v := range searchResult.Hits.Hits {
//			var gQuestion models.GQuestion
//			json.Unmarshal(v.Source, &gQuestion)
//			returnQuestionList = append(returnQuestionList, gQuestion)
//		}
//	} else {
//		if err != nil {
//			return returnQuestionList, 0, err
//		}
//	}
//
//	return returnQuestionList, int(total), err
//}

// 作答记录相关

func AnswerLogAddToEs(client *elastic.Client, ctx context.Context, dataPack []models.GAnswerLog) error {
	req := client.Bulk().Index(AnswerLogIndex)
	for _, v := range dataPack {
		doc := elastic.NewBulkIndexRequest().Id(v.Id.Hex()).Doc(v)
		req.Add(doc)
	}
	if req.NumberOfActions() < 0 {
		return nil
	}
	res, err := req.Do(ctx)
	if err != nil {
		return err
	}

	if !res.Errors {
		return nil
	}
	for _, it := range res.Failed() {
		if it.Error == nil {
			continue
		}
		return &elastic.Error{
			Status:  it.Status,
			Details: it.Error,
		}
	}
	return nil
}
