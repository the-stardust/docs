package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/mozillazg/go-pinyin"
)

type UploadFile struct {
	FieldName string
	FileName  string
	Data      []byte
}

func HttpGet(url string) ([]byte, error) {
	timeout := 15 * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return nil, errReadAll
		}
		return bodyBytes, nil
	}
	return nil, fmt.Errorf("请求状态异常", resp.StatusCode)
}
func HttpPostJson(url string, param interface{}) ([]byte, error) {
	timeout := 60 * time.Second
	httpClient := http.Client{
		Timeout: timeout,
	}
	byteJson, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(byteJson))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return nil, errReadAll
		}
		return bodyBytes, nil
	}
	return nil, err
}

func HttpPost(url string, params map[string]interface{}) ([]byte, error) {
	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	strParams := HttpBuildQuery(params)
	resp, err := client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(strParams))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	}
	return nil, err
}

func HttpPostFile(url string, params map[string]interface{}, files []UploadFile) ([]byte, error) {
	timeout := 5 * time.Second
	httpClient := http.Client{
		Timeout: timeout,
	}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	for key, value := range params {
		err := w.WriteField(key, fmt.Sprint(value))
		if err != nil {
			return nil, err
		}
	}

	for _, uploadFile := range files {
		fw, err := CreateFormFile(uploadFile.FieldName, uploadFile.FileName, uploadFile.Data, w)
		if err != nil {
			return nil, err
		}

		_, err = fw.Write(uploadFile.Data)
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	resp, err := httpClient.Post(url, w.FormDataContentType(), buf)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return nil, e
		}
		return bodyBytes, nil
	}
	return nil, err
}

func CreateFormFile(fieldName, fileName string, data []byte, w *multipart.Writer) (io.Writer, error) {
	contentType := http.DetectContentType(data)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"; filelength=%d`, fieldName, fileName, len(data)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

func HttpBuildQuery(queryData map[string]interface{}) string {
	arrayParams := make([]string, 0)
	for strKey := range queryData {
		arrayParams = append(arrayParams, strKey+"="+fmt.Sprint(queryData[strKey]))
	}
	return strings.Join(arrayParams, "&")
}

func RandomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	byteData := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, byteData[r.Intn(len(byteData))])
	}
	return string(result)
}

func Signature(params ...string) string {
	sort.Strings(params)
	h := sha1.New()
	for _, s := range params {
		_, err := io.WriteString(h, s)
		if err != nil {

		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
func InArr(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

/**
* @des 时间转换函数
* @param atime string 要转换的时间戳（秒）
* @return string
 */
func StrTime(atime int64) string {
	var byTime = []int64{365 * 24 * 60 * 60, 24 * 60 * 60, 60 * 60, 60, 1}
	var unit = []string{"年前", "天前", "小时前", "分钟前", "秒钟前"}
	now := time.Now().Unix()
	ct := now - atime
	if ct < 5 {
		return "刚刚"
	}
	var res string
	for i := 0; i < len(byTime); i++ {
		if ct < byTime[i] {
			continue
		}
		var temp = math.Floor(float64(ct / byTime[i]))
		ct = ct % byTime[i]
		if temp > 0 {
			var tempStr string
			tempStr = strconv.FormatFloat(temp, 'f', -1, 64)
			res = MergeString(tempStr, unit[i]) //此处调用了一个我自己封装的字符串拼接的函数（你也可以自己实现）
		}
		break //我想要的形式是精确到最大单位，即："2天前"这种形式，如果想要"2天12小时36分钟48秒前"这种形式，把此处break去掉，然后把字符串拼接调整下即可（别问我怎么调整，这如果都不会我也是无语）
	}
	return res
}

/**
* @des 拼接字符串
* @param args ...string 要被拼接的字符串序列
* @return string
 */
func MergeString(args ...string) string {
	buffer := bytes.Buffer{}
	for i := 0; i < len(args); i++ {
		buffer.WriteString(args[i])
	}
	return buffer.String()
}

// 获取汉字首字母
func FirstLetterOfPinYin(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)[0]
	result := pinyin.Pinyin(string(r), pinyin.NewArgs())
	if len(result) == 0 {
		return ""
	}
	return string(result[0][0][0])
}

func GetMonthStartEnd(t time.Time) (string, string) {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	monthEndDay := monthStartTime.AddDate(0, 1, -1)
	monthEndTime := time.Date(monthEndDay.Year(), monthEndDay.Month(), monthEndDay.Day(), 23, 59, 59, 0, t.Location()).Format("2006-01-02 15:04:05")
	return monthStartTime.Format("2006-01-02 15:04:05"), monthEndTime
}
func GetMonthStartEndTimestamp(t time.Time) (int64, int64) {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	monthEndDay := monthStartTime.AddDate(0, 1, -1)
	monthEndTime := time.Date(monthEndDay.Year(), monthEndDay.Month(), monthEndDay.Day(), 23, 59, 59, 0, t.Location()).Unix()
	return monthStartTime.Unix(), monthEndTime
}
func SecondTransitionFormat(second int) string {
	if second < 0 {
		second = second * -1
	}
	t := ""
	if second < 60 {
		return fmt.Sprintf("%d秒", second)
	} else if second < 3600 {
		minute := int(second / 60)
		t = fmt.Sprintf("%d分钟", minute)
		second := int(second % 60)
		if second != 0 {
			t = t + fmt.Sprintf("%d秒", second)
		}
		return t
	} else if second < 3600*24 {
		hour := int(second / 3600)
		t = fmt.Sprintf("%d小时", hour)
		residueSecond := int(second % 3600)
		if residueSecond < 60 {
			t = t + fmt.Sprintf("%d秒", residueSecond)
		} else if residueSecond < 3600 {
			minute := int(residueSecond / 60)
			tt := fmt.Sprintf("%d分钟", minute)
			tSecond := int(residueSecond % 60)
			if tSecond != 0 {
				tt = tt + fmt.Sprintf("%d秒", tSecond)
			}
			t = t + tt
		}
	}
	return t
}
func IsMobile(mobile string) bool {
	result, _ := regexp.MatchString(`^(1[2|3|4|5|6|7|8|9][0-9]\d{4,8})$`, mobile)
	if result {
		return true
	} else {
		return false
	}
}

// 求交集
func SliceStringIntersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

// MultiSliceIntersect 传入多个集合求交集
func MultiSliceIntersect(strSlices [][]string) []string {
	if len(strSlices) == 0 {
		return []string{}
	}

	// 统计每个元素在所有切片中出现的次数
	counts := make(map[string]int)
	for _, s := range strSlices {
		seen := make(map[string]bool)
		for _, v := range s {
			if !seen[v] {
				counts[v]++
				seen[v] = true
			}
		}
	}

	var result []string

	// 将出现次数等于切片数量的元素添加到结果切片中
	for k, v := range counts {
		if v == len(strSlices) {
			result = append(result, k)
		}
	}

	return result
}

// 求差集
func Diff(a, b []string) []string {
	result := []string{}
	bMap := make(map[string]bool)

	for _, v := range b {
		bMap[v] = true
	}

	for _, v := range a {
		if !bMap[v] {
			result = append(result, v)
		}
	}

	return result
}

// GetFileFromExcel 接收一个表单文件对象，并返回excelize库中的文件对象
func GetFileFromExcel(file multipart.File, header *multipart.FileHeader) (f *excelize.File, err error) {
	if path.Ext(header.Filename) != ".xlsx" {
		err = errors.New("文件名后缀错误")
		return nil, err
	}

	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, file); err != nil {
		log.Println(err)
		return nil, err
	}

	f, err = excelize.OpenReader(buff)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return f, err

}
func MD5(str string) string {
	data := []byte(str) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}

// slice去重
func RemoveDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func LocalTimeFromDateString(dateStr string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, err
	}
	if len(dateStr) < 11 {
		dateStr = dateStr + " 00:00:00"
	}
	dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, location)
	if err != nil {
		return time.Time{}, err
	}
	return dateTime, err
}

// TodayBeginningWithUTC 获取当天的零时零分零秒，然后转为UTC日期
func TodayBeginningDateWithUTC() time.Time {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local).UTC()
	return today
}

// 浮点数科学计数
func TransitionFloat64(f float64, digit int8) float64 {

	if math.IsInf(f, 1) || math.IsInf(f, -1) || math.IsNaN(f) {
		f = 0
	}
	digitStr := "%f"
	if digit >= 0 {
		digitStr = fmt.Sprintf("%%.%df", digit)
	}
	f, err := strconv.ParseFloat(fmt.Sprintf(digitStr, f), 64)
	if err != nil {

		return 0
	}
	return f
}

// IsAnswerMistake 答题判断是否正确的筛选条件
func IsAnswerMistake(questionType, userAnswerStatus int8) bool {
	if questionType <= 4 && userAnswerStatus != 1 {
		return true
	}
	return false
}

func GetGrade(key string) string {
	gradeMap := map[string]string{"跑题-偏快-无互动-不自信": "C-", "跑题-偏快-无互动-自信": "C-", "跑题-偏快-无互动-过度自信": "C-", "跑题-偏快-有互动-不自信": "C-", "跑题-偏快-有互动-自信": "C-", "跑题-偏快-有互动-过度自信": "C-", "跑题-偏快-互动恰当-不自信": "C-", "跑题-偏快-互动恰当-自信": "C-", "跑题-偏快-互动恰当-过度自信": "C-", "跑题-适中-无互动-不自信": "C-", "跑题-适中-无互动-自信": "C-", "跑题-适中-无互动-过度自信": "C-", "跑题-适中-有互动-不自信": "C-", "跑题-适中-有互动-自信": "C-", "跑题-适中-有互动-过度自信": "C-", "跑题-适中-互动恰当-不自信": "C-", "跑题-适中-互动恰当-自信": "C-", "跑题-适中-互动恰当-过度自信": "C-", "跑题-偏慢-无互动-不自信": "C-", "跑题-偏慢-无互动-自信": "C-", "跑题-偏慢-无互动-过度自信": "C-", "跑题-偏慢-有互动-不自信": "C-", "跑题-偏慢-有互动-自信": "C-", "跑题-偏慢-有互动-过度自信": "C-", "跑题-偏慢-互动恰当-不自信": "C-", "跑题-偏慢-互动恰当-自信": "C-", "跑题-偏慢-互动恰当-过度自信": "C-", "偏题-偏快-无互动-不自信": "C-", "偏题-偏快-无互动-自信": "C-", "偏题-偏快-无互动-过度自信": "C-", "偏题-偏快-有互动-不自信": "C-", "偏题-偏快-有互动-自信": "C", "偏题-偏快-有互动-过度自信": "C", "偏题-偏快-互动恰当-不自信": "C-", "偏题-偏快-互动恰当-自信": "C+", "偏题-偏快-互动恰当-过度自信": "C-", "偏题-适中-无互动-不自信": "C-", "偏题-适中-无互动-自信": "C-", "偏题-适中-无互动-过度自信": "C-", "偏题-适中-有互动-不自信": "C-", "偏题-适中-有互动-自信": "C+", "偏题-适中-有互动-过度自信": "C-", "偏题-适中-互动恰当-不自信": "C-", "偏题-适中-互动恰当-自信": "C+", "偏题-适中-互动恰当-过度自信": "C-", "偏题-偏慢-无互动-不自信": "C-", "偏题-偏慢-无互动-自信": "C-", "偏题-偏慢-无互动-过度自信": "C-", "偏题-偏慢-有互动-不自信": "C-", "偏题-偏慢-有互动-自信": "C", "偏题-偏慢-有互动-过度自信": "C-", "偏题-偏慢-互动恰当-不自信": "C-", "偏题-偏慢-互动恰当-自信": "C+", "偏题-偏慢-互动恰当-过度自信": "C-", "准确-偏快-无互动-不自信": "B-", "准确-偏快-无互动-自信": "B-", "准确-偏快-无互动-过度自信": "B-", "准确-偏快-有互动-不自信": "B-", "准确-偏快-有互动-自信": "B", "准确-偏快-有互动-过度自信": "B", "准确-偏快-互动恰当-不自信": "B-", "准确-偏快-互动恰当-自信": "B+", "准确-偏快-互动恰当-过度自信": "B+", "准确-适中-无互动-不自信": "B-", "准确-适中-无互动-自信": "B-", "准确-适中-无互动-过度自信": "B-", "准确-适中-有互动-不自信": "B-", "准确-适中-有互动-自信": "B", "准确-适中-有互动-过度自信": "B", "准确-适中-互动恰当-不自信": "B-", "准确-适中-互动恰当-自信": "B+", "准确-适中-互动恰当-过度自信": "B+", "准确-偏慢-无互动-不自信": "B-", "准确-偏慢-无互动-自信": "B-", "准确-偏慢-无互动-过度自信": "B-", "准确-偏慢-有互动-不自信": "B-", "准确-偏慢-有互动-自信": "B", "准确-偏慢-有互动-过度自信": "B", "准确-偏慢-互动恰当-不自信": "B-", "准确-偏慢-互动恰当-自信": "B+", "准确-偏慢-互动恰当-过度自信": "B+", "优秀-偏快-无互动-不自信": "A-", "优秀-偏快-无互动-自信": "A-", "优秀-偏快-无互动-过度自信": "A-", "优秀-偏快-有互动-不自信": "A-", "优秀-偏快-有互动-自信": "A", "优秀-偏快-有互动-过度自信": "A", "优秀-偏快-互动恰当-不自信": "A-", "优秀-偏快-互动恰当-自信": "A", "优秀-偏快-互动恰当-过度自信": "A", "优秀-适中-无互动-不自信": "A-", "优秀-适中-无互动-自信": "A-", "优秀-适中-无互动-过度自信": "A-", "优秀-适中-有互动-不自信": "A-", "优秀-适中-有互动-自信": "A", "优秀-适中-有互动-过度自信": "A", "优秀-适中-互动恰当-不自信": "A-", "优秀-适中-互动恰当-自信": "A", "优秀-适中-互动恰当-过度自信": "A", "优秀-偏慢-无互动-不自信": "A-", "优秀-偏慢-无互动-自信": "A-", "优秀-偏慢-无互动-过度自信": "A-", "优秀-偏慢-有互动-不自信": "A-", "优秀-偏慢-有互动-自信": "A", "优秀-偏慢-有互动-过度自信": "A", "优秀-偏慢-互动恰当-不自信": "A-", "优秀-偏慢-互动恰当-自信": "A", "优秀-偏慢-互动恰当-过度自信": "A", "共情-偏快-无互动-不自信": "A+", "共情-偏快-无互动-自信": "A+", "共情-偏快-无互动-过度自信": "A+", "共情-偏快-有互动-不自信": "A+", "共情-偏快-有互动-自信": "A+", "共情-偏快-有互动-过度自信": "A+", "共情-偏快-互动恰当-不自信": "A+", "共情-偏快-互动恰当-自信": "A+", "共情-偏快-互动恰当-过度自信": "A+", "共情-适中-无互动-不自信": "A+", "共情-适中-无互动-自信": "A+", "共情-适中-无互动-过度自信": "A+", "共情-适中-有互动-不自信": "A+", "共情-适中-有互动-自信": "A+", "共情-适中-有互动-过度自信": "A+", "共情-适中-互动恰当-不自信": "A+", "共情-适中-互动恰当-自信": "A+", "共情-适中-互动恰当-过度自信": "A+", "共情-偏慢-无互动-不自信": "A+", "共情-偏慢-无互动-自信": "A+", "共情-偏慢-无互动-过度自信": "A+", "共情-偏慢-有互动-不自信": "A+", "共情-偏慢-有互动-自信": "A+", "共情-偏慢-有互动-过度自信": "A+", "共情-偏慢-互动恰当-不自信": "A+", "共情-偏慢-互动恰当-自信": "A+", "共情-偏慢-互动恰当-过度自信": "A+"}
	value, ok := gradeMap[key]
	if ok {
		return value
	}
	return ""
}

func InArrCommon[T int | int8 | float64 | string](target T, array []T) bool {
	for _, v := range array {
		if v == target {
			return true
		}
	}
	return false
}

// map的下标
func GetMapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

func Struct2Struct(structA, structB interface{}) (interface{}, error) {
	structATemp, err := json.Marshal(&structA)
	if err != nil {
		return nil, err
	}
	//m := make(map[string]interface{})
	err = json.Unmarshal(structATemp, &structB)
	if err != nil {
		return nil, err
	}
	return structB, nil
}

// 根据值 去除arr中的元素
func SliceRemoveByValue[T int | float64 | string | int64](sli []T, value T) []T {
	j := 0
	for _, v := range sli {
		if v != value {
			sli[j] = v
			j++
		}
	}
	return sli[:j]
}

func HmacSha256(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	io.WriteString(h, message)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 进制转换：10进制转为 小于等于32进制的数
func DecimalConversionBelow32(di int, base int) string {
	if di == 0 {
		now := time.Now()
		rand.Seed(now.Unix())
		nowStr := fmt.Sprintf("%d%d", now.Unix(), rand.Intn(10000))
		di, _ = strconv.Atoi(nowStr)
	}
	upperstr := "0123456789abcdefghijklmnopqrstuv"
	substr := upperstr[0:base]
	ret := make([]byte, 0)
	mol := base - 1
	mov := math.Log2(float64(base))
	for di >= base {
		temp := []byte{substr[di&mol]}
		temp = append(temp, ret...)
		ret = temp
		di >>= int(mov)
	}
	temp := []byte{substr[di&mol]}
	return string(append(temp, ret...))
}

// PaginateArray 数组分页
func PaginateArray(arr []string, page int, pageSize int) ([]string, error) {
	// 计算起始索引和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := page * pageSize

	// 检查页数是否超出范围
	if startIndex >= len(arr) {
		return []string{}, errors.New("page number out of range")
	}

	// 根据起始索引和结束索引获取分页结果
	if endIndex > len(arr) {
		endIndex = len(arr)
	}
	paginatedArr := arr[startIndex:endIndex]

	return paginatedArr, nil
}

// 结构体变map   obj传引用进来
func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	value := reflect.ValueOf(obj).Elem() // 获取指针的值
	typ := value.Type()                  // 获取类型信息

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)              // 获取字段信息
		tag := field.Tag.Get("json")       // 获取标签（如果有）
		if tag != "" && !field.Anonymous { // 只处理非匿名字段且有标签的情况
			key := field.Name // 默认使用字段名作为Key
			if tag != "-" {   // 若标签不等于-则使用标签作为Key
				key = tag
			}
			result[key] = value.Field(i).Interface() // 存入Map
		}
	}

	return result
}

// AppendSet 去重 add
func AppendSet[T int | int8 | float64 | string](origin, appendSlice []T) []T {
	if len(origin) == 0 {
		return appendSlice
	}
	if len(appendSlice) == 0 {
		return origin
	}
	for _, v := range appendSlice {
		if InArrCommon(v, origin) {
			continue
		}
		origin = append(origin, v)
	}
	return origin
}
