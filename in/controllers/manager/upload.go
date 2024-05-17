package manager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/common/global"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/remeh/sizedwaitgroup"
)

type Upload struct {
	controllers.Controller
}

// html 转通用结构
func (sf *Upload) ConvertHtml(c *gin.Context) {
	var err error
	var param struct {
		HtmlContentList []string `json:"html_content_list"`
		IsConvertInput  bool     `json:"is_convert_input"` //需要将下划线转为input
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if len(param.HtmlContentList) <= 0 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	tempStrs := []string{}
	for _, v := range param.HtmlContentList {
		v = strings.ReplaceAll(v, "<br />", "\n")
		v = strings.ReplaceAll(v, "<br>", "\n")
		url := ""
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(v))
		wg := sizedwaitgroup.New(15)
		doc.Find("img").Each(func(i int, selection *goquery.Selection) {
			wg.Add()
			go func(selection *goquery.Selection, swg *sizedwaitgroup.SizedWaitGroup) {
				defer swg.Done()
				h, isE := selection.Attr("src")
				if isE {
					if !strings.Contains(h, "https://") && !strings.Contains(h, "http://") {
						if strings.Contains(h, ",") {
							strs := strings.Split(h, ",")
							if len(strs) > 1 && len(strs[1]) > 50 {
								h = strs[1]
							}
						}
						byteImage, err := base64.StdEncoding.DecodeString(h)
						if err != nil {
							sf.SLogger().Error(err)
						}
						url, err = new(services.Upload).UploadFile(fmt.Sprintf("question/%v.png", time.Now().Unix()), "", byteImage)
						if err != nil {
							sf.SLogger().Error(err)
						} else {
							h = url
						}
					}
					width := "0"
					height := "0"
					style, exists := selection.Attr("style")
					widthS, widthExists := selection.Attr("width")
					heightS, heightExists := selection.Attr("height")
					if widthExists {
						width = widthS
					}
					if heightExists {
						height = heightS
					}
					if exists {
						expr := `(?<=width:).*?(?=px)`
						reg, _ := regexp2.Compile(expr, 0)
						w, _ := reg.FindStringMatch(style)
						if w != nil && w.String() != "" {
							if width == "0" {
								width = strings.TrimSpace(w.String())
							}

						}
						expr = `(?<=height:).*?(?=px)`
						reg, _ = regexp2.Compile(expr, 0)
						ht, _ := reg.FindStringMatch(style)
						if ht != nil && ht.String() != "" {
							if height == "0" {
								height = strings.TrimSpace(ht.String())
							}
						}
					}
					// 如果是latex的
					latex, latexExists := selection.Attr("data-latex")
					if width == "0" && height == "0" && latexExists {
						// 获取图片宽高
						wInt, hInt, err := new(services.Image).GetImageSize(url)
						if err == nil {
							width = fmt.Sprintf("%d", wInt)
							height = fmt.Sprintf("%d", hInt)
						} else {
							sf.SLogger().Debug(err)
						}
						h = h + "?width=" + width
						h = h + "&latex=" + latex
						h = h + "&height=" + height
						h = h + "$$$"
					} else {
						h = h + "?width=" + width + "&height=" + height + "$$$"
					}
					selection.ReplaceWithHtml(h)
				}
			}(selection, &wg)
		})
		wg1 := sizedwaitgroup.New(15)
		doc.Find("span").Each(func(i int, selection *goquery.Selection) {
			wg1.Add()
			go func(selection *goquery.Selection, swg *sizedwaitgroup.SizedWaitGroup) {
				defer swg.Done()
				style, _ := selection.Attr("style")
				if strings.Contains(style, "underline") {
					h, err := selection.Html()
					if err == nil {
						t := selection.Text()
						if strings.Contains(t, "[style") && strings.Contains(t, ";]") {

							newT := t
							sArr := strings.Split(t, ";]")
							if len(sArr) > 1 {
								newT = sArr[0] + "type=u,val_attr=single" + ";]" + sArr[1]
								h = strings.ReplaceAll(h, t, newT)
							} else {
								sf.SLogger().Error("html2commoncontent fail", t)
							}

						} else {
							h = strings.ReplaceAll(h, t, "[style type=u,val_attr=single;]"+t+"[/style]")
						}

						selection.ReplaceWithHtml(h)
					}
				}
			}(selection, &wg1)
		})
		wg1.Wait()
		wg.Wait()

		tempStrs = append(tempStrs, doc.Text())
	}
	httpRes, err := common.HttpPostJson(fmt.Sprintf("%s/convert-rich-text", global.CONFIG.ServiceUrls.ParseUrl), map[string]interface{}{"rich_text_list": tempStrs, "is_convert_input": param.IsConvertInput})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var res struct {
		Data []models.CommonContent `json:"data"`
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
	}
	err = json.Unmarshal(httpRes, &res)
	if err != nil {
		sf.SLogger().Error(err, string(httpRes))
		sf.Error(common.CodeServerBusy, c)
		return
	}
	if res.Code == 200 {
		sf.Success(res.Data, c)
	} else {
		sf.SLogger().Error(fmt.Errorf(res.Msg))
		sf.Error(common.CodeServerBusy, c)
		return
	}
}

func (sf *Upload) UploadFileList(c *gin.Context) {
	var err error
	var param struct {
		ExamCategory      string `json:"exam_category"`       // 考试分类
		ExamChildCategory string `json:"exam_child_category"` //考试子分类
		CheckStatus       int8   `json:"check_status"`
		StartTime         string `json:"start_time"`
		EndTime           string `json:"end_time"`
		Keywords          string `json:"keywords"`
		PageIndex         int64  `json:"page_index"`
		PageSize          int64  `json:"page_size"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	filter := bson.M{}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.CheckStatus != 0 {
		filter["check_status"] = param.CheckStatus
	}
	if param.StartTime != "" && param.EndTime != "" {
		filter["created_time"] = bson.M{"$gte": param.StartTime, "$lte": param.EndTime}
	}

	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"_id": sf.ObjectID(param.Keywords)},
			bson.M{"user_id": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"file_url": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"file_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}

	respList := make([]models.UserUploadFile, 0)
	err = sf.DB().Collection("user_upload_file").Where(filter).Skip(offset).Limit(limit).Sort("-created_time").Find(&respList)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	totalCount, err := sf.DB().Collection("user_upload_file").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]interface{}{"list": respList, "count": totalCount}, c)

}

func (sf *Upload) UploadFileInfo(c *gin.Context) {
	id := c.DefaultQuery("id", "")
	if id == "" {
		sf.Error(common.CodeServerBusy, c, "id不允许为空")
		return
	}

	var t models.UserUploadFile
	err := sf.DB().Collection("user_upload_file").Where(bson.M{"_id": sf.ObjectID(id)}).Take(&t)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success(t, c)

}

func (sf *Upload) UploadFileSave(c *gin.Context) {
	var err error
	var param struct {
		Id          string `json:"id"`
		CheckStatus int8   `json:"check_status"` // 0 已上传成功未审核，1接受， 2部分接受，3不接受
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var t models.UserUploadFile
	err = sf.DB().Collection("user_upload_file").Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(&t)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	t.CheckStatus = param.CheckStatus
	err = sf.DB().Collection("user_upload_file").Save(&t)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)

}
