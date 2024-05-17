package word

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gitee.com/xintujing/unioffice/color"
	"gitee.com/xintujing/unioffice/measurement"
	"gitee.com/xintujing/unioffice/schema/soo/wml"
	"interview/models"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"interview/common/global"
	"interview/services"

	"github.com/remeh/sizedwaitgroup"

	cm "interview/common"

	"gitee.com/xintujing/unioffice/common"
	"gitee.com/xintujing/unioffice/document"
)

type CT_AbNumILvl struct {
	ilvl   int64
	numFmt string
	text   string
	start  int64
}

type CT_AbNumData struct {
	numId int64
	iLvl  []*CT_AbNumILvl
}

type CT_RowData struct {
	Content     []string
	HtmlContent []string
}

type CT_TableData struct {
	Rows []*CT_RowData
}

type WordResolver struct {
	Uri     string
	Tables  []*CT_TableData
	doc     *document.Document
	mode    int
	modeStr string
	//公式对象 RID:LATEX 的对应关系
	//oles map[string]*string
	oles *sync.Map

	//无法转换的公式转换成图片
	//olesImages map[string]string
	olesImages *sync.Map

	//图片 RID:oss地址 的对应关系
	//images map[string]string
	images *sync.Map

	//自动序号相关
	numIdMapAbNumId map[int64]int64
	numData         []*CT_AbNumData
}

func NewWordResolver() *WordResolver {
	return &WordResolver{mode: -1}
}

// 解析整个word数据
func (sf *WordResolver) GetWordData() ([]string, error) {
	//解析公式和图片
	sf.parseOle()
	sf.parseImage()

	//解析自动序号
	sf.parseOrder()
	//得到word的所有文本信息
	data, err := sf.getPureText()
	if err != nil {
		return []string{}, err
	}
	log.Println("==word内容解析结束")
	return data, nil
}

// 解析word表格数据
func (sf *WordResolver) GetWordTableData() ([]*CT_TableData, error) {
	//解析公式和图片
	sf.parseOle()
	sf.parseImage()
	//读取table数据
	sf.getTableData()

	return sf.Tables, nil
}

// 读取数据
func (sf *WordResolver) Read(fileByte []byte) error {
	doc, err := document.Read(bytes.NewReader(fileByte), int64(len(fileByte)))
	if err != nil {
		return err
	}
	sf.doc = doc
	return nil
}

// 把ole对象文件转为图片
func (sf *WordResolver) parseOle() {
	sf.oles = &sync.Map{}
	sf.olesImages = &sync.Map{}
	wg := sizedwaitgroup.New(15)
	for _, ole := range sf.doc.OleObjectWmfPath {
		wg.Add()
		go func(word *WordResolver, wmfObj document.OleObjectWmfPath) {
			defer wg.Done()
			srcFile, err := os.Open(wmfObj.Path())
			if err != nil {
				log.Printf("wmf opn falied ,err=%s", err)
			}
			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, srcFile)
			if err != nil {
				log.Printf("Write to form file falied: %+v\n", err)
			}
			httpRes, err := cm.HttpPostFile(fmt.Sprintf("%s/wmf-files", global.CONFIG.ServiceUrls.ParseUrl), map[string]interface{}{}, []cm.UploadFile{{FieldName: "file", FileName: "file", Data: buf.Bytes()}})
			if err == nil {
				var res struct {
					Data string `json:"data"`
					Msg  string `json:"msg"`
					Code int    `json:"code"`
				}
				err = json.Unmarshal(httpRes, &res)
				if err != nil {
					log.Printf("wmf to png falied: %+v\n", err)
				}
				if res.Code == 200 {
					word.olesImages.Store(wmfObj.Rid(), res.Data)
				} else {
					log.Printf("wmf to png falied: %+v\n", res.Msg)
				}

			} else {
				log.Printf("wmf to png falied: %+v\n", err)
			}

		}(sf, ole)
	}
	wg.Wait()
}

// 把图片上传到oss
func (sf *WordResolver) parseImage() {
	sf.images = &sync.Map{}
	wg := sizedwaitgroup.New(15)
	for _, img := range sf.doc.Images {
		wg.Add()
		//goroutine 运行
		go func(word *WordResolver, image common.ImageRef) {
			defer wg.Done()
			//调用图片上传
			localFile := image.Path()
			srcFile, err := os.Open(localFile)
			if err != nil {
				log.Printf("Open source file failed: %+v\n", err)
			}
			defer srcFile.Close()
			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, srcFile)
			if err != nil {
				log.Printf("Write to form file falied: %+v\n", err)
			}
			objName := fmt.Sprintf("question/%s.png", image.RelID())
			imageType := cm.GetFileType(buf.Bytes())
			if imageType != "jpg" && imageType != "png" && imageType != "" {
				objName = fmt.Sprintf("question/%s.%s", image.RelID(), imageType)
			}
			uri, err := new(services.Upload).UploadFile(objName, "", buf.Bytes())
			if err != nil {
				log.Printf("upload falied: %+v\n", err)
			}
			word.images.Store(image.RelID(), uri)
		}(sf, img)
	}
	wg.Wait()
	//fmt.Println(sf.doc.Images)
	//fmt.Println(sf.images)
}

// 执行自动序号数据读取
func (sf *WordResolver) parseOrder() {
	if sf.doc.Numbering.X() != nil {
		//读取序号数据
		for _, df := range sf.doc.Numbering.Definitions() {
			abData := &CT_AbNumData{}
			abData.numId = df.AbstractNumberID()
			for _, lv := range df.X().Lvl {
				//todo 有些start为nil？
				start := int64(1)
				if lv.Start != nil {
					start = lv.Start.ValAttr
				}
				abData.iLvl = append(abData.iLvl, &CT_AbNumILvl{
					ilvl:   lv.IlvlAttr,
					numFmt: lv.NumFmt.ValAttr.String(),
					text:   *lv.LvlText.ValAttr,
					start:  start,
				})
			}

			sf.numData = append(sf.numData, abData)
		}

		//numId与abstractNumId的映射关系
		numIdMapAbNumId := make(map[int64]int64)
		for _, nu := range sf.doc.Numbering.X().Num {
			numIdMapAbNumId[nu.NumIdAttr] = nu.AbstractNumId.ValAttr
		}

		sf.numIdMapAbNumId = numIdMapAbNumId
	}
}

type MathUrl struct {
	MathContent string
	Url         string
}

// 得到纯解析的word文本数据
func (sf *WordResolver) getPureText() ([]string, error) {

	// res := bytes.Buffer{}
	resStrs := make([]string, 0)
	//p数据，段落自动编号当前值
	var (
		paragraphSortNum   int8
		paragraphSortNumId int64
	)
	listChans := make(chan MathUrl, 10)
	resChans := make(chan map[string]string, 1)
	go func() {
		r := map[string]string{}
		for obj := range listChans {
			r[obj.MathContent] = obj.Url
		}
		resChans <- r
	}()
	sw := sizedwaitgroup.New(10)
	for _, paragraph := range sf.doc.Paragraphs() {
		var (
			//段落样式
			// paragraphStyle string
			//段落自动编号应该呈现的值
			paragraphSortNumText string
		)
		//读取段落数据
		pString := sf.getParagraphData(paragraph, &sw, listChans)

		// 读取段落样式
		if paragraph.X().PPr != nil {

			// log.Printf("%+v   %+v   %v", pString, paragraph.X().PPr.WordWrap, paragraph.X().PPr.PBdr)
			if pString != "" {
				pString = pString + "\n"
			}

			//段落自动编号样式读取
			//参考文档：http://c-rex.net/projects/samples/ooxml/e1/Part4/OOXML_P4_DOCX_Numbering_topic_ID0EN6IU.html
			if paragraph.X().PPr.NumPr != nil {
				//初始化没有编号ID
				numId := paragraph.X().PPr.NumPr.NumId.ValAttr
				if paragraph.X().PPr.NumPr.NumId.ValAttr != paragraphSortNumId {

					//设置编号ID
					paragraphSortNumId = numId
					//设置当前起始值为1
					if paragraphSortNumId > 0 {
						paragraphSortNum = 1
					} else {
						paragraphSortNum = 0
					}

				} else {
					if numId > 0 {
						//存在当前编号，当前值+1
						paragraphSortNum += 1
					}
				}
			} else {
				//重置整个排序编号值
				paragraphSortNum = 0
			}
			if paragraphSortNum != 0 {
				ivlData, err := sf.readAbNumData(paragraphSortNumId, 0)
				if err != nil {
					return []string{}, err
				}
				numFmt := ivlData.numFmt
				numText := ivlData.text
				var numVal string
				if numFmt == "decimal" {
					numVal = NUM_Decimal(paragraphSortNum).String()
				} else if numFmt == "decimalEnclosedCircle" {
					numVal = NUM_DecimalEnclosedCircle(paragraphSortNum).String()
				} else if numFmt == "japaneseCounting" || numFmt == "chineseCountingThousand" || numFmt == "chineseCounting" {
					numVal = NUM_Counting(paragraphSortNum).String()
				} else if numFmt == "upperLetter" {
					numVal = NUM_UpperLetter(paragraphSortNum).String()
				} else if numFmt == "upperRoman" {
					numVal = NUM_UpperRoman(paragraphSortNum).String()
				} else {
					numVal = fmt.Sprintf("%v", paragraphSortNum)
					log.Printf("暂时不支持的自动序号,numFmt=%s,text=%s", numFmt, numText)
				}
				//替换数据
				paragraphSortNumText = strings.Replace(numText, "%1", numVal, -1)

				//写入自动编号
				pString = fmt.Sprintf("%s %s", paragraphSortNumText, pString)
			}

			//段落缩进
			//https://docs.microsoft.com/zh-cn/dotnet/api/documentformat.openxml.wordprocessing.indentation?view=openxml-2.8.1
			if paragraph.X().PPr.Ind != nil {
				if paragraph.X().PPr.Ind.FirstLineCharsAttr != nil {
					indentNum := int(math.Round(float64(*(paragraph.X().PPr.Ind.FirstLineCharsAttr)) / 50))
					indentNbsp := strings.Repeat("&nbsp;", indentNum)
					//fmt.Println(paragraph.X().PPr.Ind.FirstLineCharsAttr)
					pString = fmt.Sprintf("%s%s", indentNbsp, pString)
				}
			}
			// if paragraph.X().PPr.Jc != nil {
			// 	fmt.Println("^^^^^^^^^^^", pString)
			// 	fmt.Printf("================%#v \n", paragraph.X().PPr.Jc.ValAttr.String())
			// 	//fmt.Println("====================================")
			// 	// paragraphStyle = fmt.Sprintf(" align='%s' ", paragraph.X().PPr.Jc.ValAttr.String())
			// }
			// t := reflect.TypeOf(*paragraph.X().PPr)
			// v := reflect.ValueOf(*paragraph.X().PPr)
			// for k := 0; k < t.NumField(); k++ {
			// 	if !v.Field(k).IsNil() {
			// 		fmt.Printf("%s -- %v \n", t.Field(k).Name, v.Field(k).Interface())

			// 	}
			// }
			// tt := reflect.TypeOf(*paragraph.X().PPr.RPr)
			// vv := reflect.ValueOf(*paragraph.X().PPr.RPr)
			// for k := 0; k < tt.NumField(); k++ {
			// 	if !vv.Field(k).IsNil() {
			// 		// fmt.Printf("%s -- %v ----%v\n", tt.Field(k).Name, vv.Field(k).Interface(), pString)
			// 		if strings.Contains(pString, "单下划线") {
			// 			fmt.Printf("%s -- %v ----%v===%#v\n", tt.Field(k).Name, vv.Field(k).Interface(), pString, paragraph.X().PPr.RPr.U.ValAttr.String())
			// 		}

			// 	}
			// }

		}
		//保存内容
		// res.WriteString(pString)
		resStrs = append(resStrs, pString)
	}

	sw.Wait()
	close(listChans)
	var resMap = map[string]string{}
	resMap = <-resChans
	close(resChans)

	resJson, _ := json.Marshal(resStrs)
	resString := string(resJson)
	for k, v := range resMap {
		resString = strings.ReplaceAll(resString, k, v)
	}
	err := json.Unmarshal([]byte(resString), &resStrs)
	if err != nil {
		log.Println("公式替换异常", err.Error())
	}
	return resStrs, nil
}

// 读取段落数据
func (sf *WordResolver) getParagraphData(paragraph document.Paragraph, sw *sizedwaitgroup.SizedWaitGroup, ch chan MathUrl) string {
	//存储run数据
	paragraphBuffer := bytes.Buffer{}
	//段落下面的每个单元文本数据
	for _, run := range paragraph.Runs() {
		//段落下面的每个单元文本数据
		var text string
		if len(run.DrawingAnchored()) > 0 {
			//图片数据
			text = sf.readAnchoredImage(run.DrawingAnchored())
		} else if len(run.DrawingInline()) > 0 {
			//图片数据
			text = sf.readImage(run.DrawingInline())
		} else if len(run.OleObjects()) > 0 {
			//公式数据
			text = sf.readOles(run.OleObjects())
		} else if len(run.OleObjects()) > 0 {
			//公式数据
			text = sf.readOles(run.OleObjects())
		} else if len(run.Ruby().Rt) > 0 && len(run.Ruby().RubyBase) > 0 {
			//拼音数据
			if len(run.Ruby().Rt) != len(run.Ruby().RubyBase) {
				log.Println("拼音文本数据长度对不上")
			} else {
				log.Println("拼音===========")
				for idx, rt := range run.Ruby().Rt {
					rubyText := run.Ruby().RubyBase[idx]
					// text=fmt.Sprintf("<ruby>%s<rt>%s</rt></ruby>", rubyText, rt)
					text = fmt.Sprintf("%s%s", rubyText, rt)
				}
			}
		} else {
			//	文本数据
			text = run.Text()
			//下划线
			if run.X().RPr != nil && run.X().RPr.U != nil && run.X().RPr.U.ValAttr.String() != "none" {
				valAttr := run.X().RPr.U.ValAttr.String()
				text = "[style type=u,val_attr=" + valAttr + ";]" + text + "[/style]"
			}
			if strings.Contains(text, "oMath") {
				sw.Add()
				go sf.omml2ImageJava(text, sw, ch, 1)
				//统一md5处理 用于公式解析后替换
				text = Md5(text)
			}
			// else {
			// 	//把空格替换成&nbsp;
			// 	if strings.Contains(text, " ") {
			// 		text = strings.Replace(text, " ", "&nbsp;", -1)
			// 	}
			// }
		}
		paragraphBuffer.WriteString(text)
	}
	return paragraphBuffer.String()
}
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	// base64.StdEncoding.EncodeToString(msg)
	return hex.EncodeToString(h.Sum(nil))
}
func (sf *WordResolver) omml2ImageJava(text string, sw *sizedwaitgroup.SizedWaitGroup, ch chan MathUrl, retryTimes int) {
	placeholderUrl := "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/other/3c5ee4f41df97ffb6234b246641e7f1.png?width=60&18$$$"
	type TempParam struct {
		Content  string `json:"xmlBody"`
		FontSize int    `json:"fontSize"`
	}
	pm := TempParam{Content: text, FontSize: 18}
	httpRes, err := cm.HttpPostJson(fmt.Sprintf("%s/omml/v1/convert/single", global.CONFIG.ServiceUrls.Omml2ImageUrl), pm)
	if err == nil {
		var res struct {
			Msg  string `json:"msg"`
			Code int    `json:"code"`
		}
		err = json.Unmarshal(httpRes, &res)
		if err == nil {
			if res.Code == 200 {
				var urlData struct {
					Data struct {
						Url string `json:"url"`
					} `json:"data"`
				}
				err = json.Unmarshal(httpRes, &urlData)
				if err == nil {
					if urlData.Data.Url != "" {
						ch <- MathUrl{MathContent: Md5(text), Url: urlData.Data.Url + "$$$"}
					} else {
						ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
						log.Printf("omath to png url is empty %+v", string(httpRes))
					}
				} else {
					ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
					log.Printf("omath to png parse url failed: %+v\n %+v", err, string(httpRes))
				}

			} else {
				ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
				log.Printf("omath to png parse falied : %+v\n%+v\n%+v", res.Code, res.Msg, text)
			}
		} else {
			ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
			log.Printf("omath to png return base info format invalid: %+v\n httpRes:%+v", err, string(httpRes))
		}
	} else {
		if retryTimes > 0 {
			sf.omml2ImageJava(text, sw, ch, 0)
			sw.Add()
		} else {
			ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
			log.Println("omml2image http falied", err.Error())
		}
	}
	sw.Done()
}

// func (sf *WordResolver) omml2Image(text string, sw *sizedwaitgroup.SizedWaitGroup, ch chan MathUrl, retryTimes int) {
// 	placeholderUrl := "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/other/3c5ee4f41df97ffb6234b246641e7f1.png$$$"
// 	type TempParam struct {
// 		Content string `json:"xml_str"`
// 	}
// 	pm := TempParam{Content: text}
// 	httpRes, err := cm.HttpPostJson(fmt.Sprintf("%s/xml-files", global.CONFIG.ServiceUrls.ParseUrl), pm)
// 	if err == nil {
// 		var res struct {
// 			Msg  string `json:"msg"`
// 			Code int    `json:"code"`
// 		}
// 		err = json.Unmarshal(httpRes, &res)
// 		if err == nil {
// 			if res.Code == 200 {
// 				var pRes struct {
// 					Data string `json:"data"`
// 				}
// 				err = json.Unmarshal(httpRes, &pRes)
// 				if err == nil {
// 					ch <- MathUrl{MathContent: Md5(text), Url: pRes.Data + "$$$"}
// 				} else {
// 					ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
// 					log.Printf("omath to png return data format invalid: %+v\n %+v", err, httpRes)
// 				}

// 			} else {
// 				ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
// 				log.Printf("omath to png parse falied : %+v\n%+v", res.Msg, text)
// 			}
// 		} else {
// 			ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
// 			log.Printf("omath to png return base info format invalid: %+v\n httpRes: %+v", err, httpRes)
// 		}
// 	} else {
// 		if retryTimes > 0 {
// 			sf.omml2Image(text, sw, ch, 0)
// 			sw.Add()
// 		} else {
// 			ch <- MathUrl{MathContent: Md5(text), Url: placeholderUrl}
// 			log.Println("omml2image http falied", err.Error())
// 		}
// 	}
// 	sw.Done()
// }

// 读取图片数据
func (sf *WordResolver) readImage(images []document.InlineDrawing) string {
	var imageUri string
	for _, di := range images {
		imf, _ := di.GetImage()
		w := imf.Size().X
		h := imf.Size().Y
		uri, _ := sf.images.Load(imf.RelID())
		imageUri = fmt.Sprintf("%+v?width=%d&height=%d$$$", uri, w, h)
	}
	return imageUri
}

func (sf *WordResolver) readAnchoredImage(images []document.AnchoredDrawing) string {
	var imageUri string
	for _, di := range images {
		imf, _ := di.GetImage()
		w := imf.Size().X
		h := imf.Size().Y
		uri, _ := sf.images.Load(imf.RelID())
		imageUri = fmt.Sprintf("%+v?width=%d&height=%d$$$", uri, w, h)
	}
	return imageUri
}

// 读取公式数据
func (sf *WordResolver) readOles(ole []document.OleObject) string {
	var latexStr string
	for _, ole := range ole {
		latexPtr, ok := sf.oles.Load(ole.OleRid())
		if ok {
			latexStr = *latexPtr.(*string)
		} else {
			oleImg, ok := sf.olesImages.Load(ole.ImagedataRid())
			if ok {
				// latexStr = fmt.Sprintf("<img src='%s' style='%s' />", oleImg, *ole.Shape().StyleAttr)
				latexStr = oleImg.(string)
			}
		}
	}

	return latexStr
}

// 读取自动序号的数据
func (sf *WordResolver) readAbNumData(numId int64, ilvl int64) (*CT_AbNumILvl, error) {
	abNumId, ok := sf.numIdMapAbNumId[numId]
	if !ok {
		return nil, fmt.Errorf("自动序号解析失败，找不到numId=%d", numId)
	}
	//读取 abstractNum 数据
	var tmpAbData *CT_AbNumData
	for _, abData := range sf.numData {
		abDataNumId := abData.numId

		if abDataNumId == abNumId {
			tmpAbData = abData
		}
	}

	if tmpAbData == nil {
		return nil, fmt.Errorf("找不到AbNum实例数据，abNumId=%d", abNumId)
	}

	//读取 lvl 数据
	var tmpAbLvl *CT_AbNumILvl
	for _, abLvl := range tmpAbData.iLvl {
		abLvlVal := abLvl.ilvl

		if abLvlVal == ilvl {
			tmpAbLvl = abLvl
		}
	}

	if tmpAbLvl == nil {
		return nil, fmt.Errorf("找不到ilvl实例数据，ilvl=%d", ilvl)
	}

	return tmpAbLvl, nil
}

// 读取表格数据
func (sf *WordResolver) getTableData() {
	tables := sf.doc.Tables()
	for _, table := range tables {
		//读取一个表单里面的所有行
		rows := table.Rows()

		//读取行里面的数据
		td := &CT_TableData{}
		for _, row := range rows {
			cells := row.Cells()
			rowData := &CT_RowData{}

			for _, cell := range cells {
				rawText, htmlText := sf.getCellText(&cell)
				rowData.Content = append(rowData.Content, rawText)
				rowData.HtmlContent = append(rowData.HtmlContent, htmlText)
			}

			td.Rows = append(td.Rows, rowData)
		}

		sf.Tables = append(sf.Tables, td)
	}
}

// 读取行里面每一个单元的数据
func (sf *WordResolver) getCellText(cell *document.Cell) (string, string) {
	paragraphs := cell.Paragraphs()

	resText := bytes.Buffer{}
	htmlResText := bytes.Buffer{}
	//循环每一个P标签数据
	for paragIdx, paragraph := range paragraphs {
		runs := paragraph.Runs()

		for _, run := range runs {
			var text string
			if len(run.DrawingAnchored()) > 0 {
				//图片数据
				text = sf.readAnchoredImage(run.DrawingAnchored())
			} else if len(run.DrawingInline()) > 0 {
				//图片数据
				text = sf.readImage(run.DrawingInline())
			} else if len(run.OleObjects()) > 0 {
				//公式数据
				text = sf.readOles(run.OleObjects())
			} else if len(run.OleObjects()) > 0 {
				//公式数据
				text = sf.readOles(run.OleObjects())
			} else if len(run.Ruby().Rt) > 0 && len(run.Ruby().RubyBase) > 0 {
				//拼音数据
				if len(run.Ruby().Rt) != len(run.Ruby().RubyBase) {
					log.Println("拼音文本数据长度对不上")
				} else {
					log.Println("拼音===========")
					for idx, rt := range run.Ruby().Rt {
						rubyText := run.Ruby().RubyBase[idx]
						// text=fmt.Sprintf("<ruby>%s<rt>%s</rt></ruby>", rubyText, rt)
						text = fmt.Sprintf("%s%s", rubyText, rt)
					}
				}
			} else {
				//	文本数据
				text = run.Text()
				// if strings.Contains(text, "oMath") {
				// 	sw.Add()
				// 	go sf.omml2ImageJava(text, sw, ch, 1)
				// 	//统一md5处理 用于公式解析后替换
				// 	text = Md5(text)
				// }
			}
			resText.WriteString(text)
			htmlResText.WriteString(text)
		}

		//新的段落换行
		if paragIdx < len(paragraphs)-1 {
			htmlResText.WriteString("\n")
		}

	}

	return resText.String(), htmlResText.String()
}

// 获取OSS上图片的二进制
func GetOssImgBytes(imgUrl string) ([]byte, error) {
	imgRes, err := cm.HttpGet(imgUrl)
	if err != nil {
		return nil, fmt.Errorf("GetImage fail: ", err)
	}
	return imgRes, nil
}

// 带格式
func (sf *WordResolver) GetDocWithLayout(title string) *document.Document {
	var doc = document.New()

	head := doc.AddHeader()
	para := head.AddParagraph()
	para.Properties().SetAlignment(wml.ST_JcRight)
	run := para.AddRun()
	run.Properties().SetFontFamily("SimHei")
	run.Properties().SetSize(1.5 * measurement.Character)
	run.AddText("易判，易断，易逻辑！")
	doc.BodySection().SetHeader(head, wml.ST_HdrFtrDefault)

	ftr := doc.AddFooter()
	para = ftr.AddParagraph()
	para.Properties().SetAlignment(wml.ST_JcCenter)

	run = para.AddRun()
	run.Properties().SetFontFamily("SimHei")
	run.Properties().SetSize(1.5 * measurement.Character)
	run.AddText("更合理的课程、更严格的课堂        ")
	run.AddFieldWithFormatting(document.FieldCurrentPage, "", false)
	run.AddText("        更专业的老师、更高的上岸率")
	doc.BodySection().SetFooter(ftr, wml.ST_HdrFtrDefault)

	p := doc.AddParagraph()
	p.Properties().SetAlignment(wml.ST_JcCenter)
	p.Properties().AddTabStop(4*measurement.Character, wml.ST_TabJcCenter, wml.ST_TabTlcNone)
	r := p.AddRun()
	r.Properties().SetFontFamily("SimHei")
	r.Properties().SetColor(color.Black)
	r.Properties().SetBold(true)
	r.Properties().SetSize(16 * measurement.Point)
	r.AddText(title)
	p = doc.AddParagraph()
	r = p.AddRun()
	r.Properties().SetColor(color.Red)
	r.Properties().SetSize(10 * measurement.Point)
	r.AddText("部分机型微信内打开不显示图片，可使用其他程序阅读")

	return doc
}

func (sf *WordResolver) AddWordContent(questionContent models.CommonContent, r document.Run, p document.Paragraph, doc *document.Document, questionIndex int) (int, document.Run) {
	addQuestionIndex := false
	if questionContent.DataType == 1 {
		questionContent.Text = strings.TrimFunc(questionContent.Text, func(r rune) bool {
			return r == '\n' || r == '\t' || r == ' '
		})

		textArr := strings.Split(questionContent.Text, "\n")
		for i, text := range textArr {
			if questionIndex > 0 && i == 0 {
				if sf.mode == 999 || sf.mode == 0 {
					text = sf.modeStr + text
				}
				r = p.AddRun()
				r.Properties().SetFontFamily("Calibri")
				r.Properties().SetSize(11 * measurement.Point)
				r.AddText(fmt.Sprintf("%d. ", questionIndex))
				r = p.AddRun()
				r.Properties().SetFontFamily("SimSun")
				r.Properties().SetSize(11 * measurement.Point)
				questionIndex++
			}
			if strings.TrimSpace(text) == "" {
				continue
			}
			r.AddText(text)
			if (i + 1) != len(textArr) {
				r.AddBreak()
			}
		}

	} else if questionContent.DataType == 2 {
		sf.AddWordImg(questionContent, r, doc)
	} else if questionContent.DataType == 3 {
		for _, content := range questionContent.Content {
			if content.DataType == 1 {
				isBreak := false
				if strings.Contains(content.Text, "\n") {
					isBreak = true
				}
				textArr := strings.Split(content.Text, "\n")
				for _, text := range textArr {
					if !addQuestionIndex && questionIndex > 0 {
						if sf.mode == 999 || sf.mode == 0 {
							text = sf.modeStr + text
						}
						r = p.AddRun()
						r.Properties().SetFontFamily("Calibri")
						r.Properties().SetSize(11 * measurement.Point)
						r.AddText(fmt.Sprintf("%d. ", questionIndex))
						r = p.AddRun()
						r.Properties().SetFontFamily("SimSun")
						r.Properties().SetSize(11 * measurement.Point)
						addQuestionIndex = true
						questionIndex++
					}
					if strings.TrimSpace(text) == "" {
						continue
					}
					r.AddText(text)
					if isBreak {
						r.AddBreak()
					}
				}
			} else if content.DataType == 2 {
				err := sf.AddWordImg(content, r, doc)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	} else if questionContent.DataType == 5 {
		r.AddText("________")
	}
	sf.SetWordContentMode(-1, "") // 复原-1
	return questionIndex, r
}

func (sf *WordResolver) AddWordImg(content models.CommonContent, r document.Run, doc *document.Document) error {
	resizeImg := false
	if strings.Index(content.Text, "latex") != -1 {
		resizeImg = true
	}
	imgRes, err := GetOssImgBytes(content.Text)

	img, err := common.ImageFromBytes(imgRes)
	if err != nil {
		return fmt.Errorf("ImageFromFile fail: ", err)
	}

	imgRef, err := doc.AddImage(img)
	if err != nil {
		return fmt.Errorf("doc AddImage fail: ", err)
	}
	var imgW = measurement.Distance(img.Size.X) * measurement.Point
	if img.Size.X > 240 {
		imgW = 240 * measurement.Point
	}
	imgH := imgRef.RelativeHeight(imgW)

	if resizeImg {
		imgH = measurement.Distance(16) * measurement.Point
		imgW = imgRef.RelativeWidth(imgH)
	}

	ild, err := r.AddDrawingInline(imgRef)
	if err != nil {
		return fmt.Errorf("AddDrawingInline fail: ", err)
	}

	ild.SetSize(imgW, imgH)

	err = ild.X().Validate()
	if err != nil {
		return fmt.Errorf("AddDrawingInline Validate fail: ", err)
	}

	return nil
}

func (sf *WordResolver) SetWordContentMode(mode int, modeStr string) {
	sf.mode = mode
	sf.modeStr = modeStr
}
