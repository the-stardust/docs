package services

import (
	"bytes"
	"github.com/corona10/goimagehash"
	"github.com/go-ego/gse"
	"image/png"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"
)

var seg gse.Segmenter
var segLoadOnce sync.Once

func init() {
	segLoadOnce.Do(func() {
		env := os.Getenv("GO_ENV")
		if env == "" {
			seg.LoadDict()
		} else {
			seg.LoadDict("/root/s_1.txt")
		}
	})
}

func SpecialcharsReplace(str string) string {
	var replace = strings.NewReplacer("\r", "", "\n", "", "\t", "", "\f", "", "的", "", "了", "", "（", "", "）", "", "(", "", ")", "",
		" ", "", "?", "", "？", "", "：", "", ":", "", "\"", "", "'", "", "“", "", "”", "", "-", "", "~", "", "——", "", "’", "", "‘", "", "。", "", "&nbsp;", "", " ", "")
	str = replace.Replace(str)
	return replace.Replace(str)
}

// 获取中文字符串的simhash值
func StrSimHash(str string) string {
	str = SpecialcharsReplace(str)
	if str == "" {
		return str
	}
	//fmt.Println(str)
	s1s := seg.Cut(str, true)
	//fmt.Println("cut use hmm: ", hmm)
	//x := gojieba.NewJieba()
	//defer x.Free()
	//s1s := x.CutAll(str)
	//fmt.Println(s1s)

	// 计算每个字符在字符数组中出现的次数
	counts1 := make(map[string]int)
	for _, s := range s1s {
		// 如果字符在字符数组中出现过，则计数加1
		if _, ok := counts1[s]; ok {
			counts1[s]++
		} else {
			// 如果字符在字符数组中没出现过，则计数设为1
			counts1[s] = 1
		}
	}
	return IntsToStr(Dimensionality(merge(hashcodeAndAdd(counts1))))
}

// 计算两个simhash值的距离
func HammingDistance(utf8Str1, utf8Str2 string) float64 {
	count := 0

	l1 := utf8.RuneCountInString(utf8Str1)
	max := l1

	l2 := utf8.RuneCountInString(utf8Str2)
	if max < l2 {
		max = l2
	}

	for i, j := 0, 0; i < len(utf8Str1) && j < len(utf8Str2); {
		size := 0
		r1, size := utf8.DecodeRune(StringToBytes(utf8Str1[i:]))
		i += size

		r2, size := utf8.DecodeRune(StringToBytes(utf8Str2[j:]))
		j += size

		if r1 != r2 {
			count++
		}
	}

	return 1 - (float64(count)+math.Abs(float64(l1-l2)))/float64(max)
}

// 图片的hash值
func ImgDifferenceHash(imgUrl string) (string, error) {
	//imgUrl := "https://swimming-ring.oss-cn-zhangjiakou.aliyuncs.com/1/images/xingce/-cc822svejsj8dq55oolg.png"
	resp, err := http.Get(imgUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	img, err := png.Decode(resp.Body)
	if err != nil {
		return "", err
	}
	hash, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return "", err
	}

	str := strconv.FormatUint(hash.GetHash(), 2)
	// 不够64位 补0
	if len(str) < 64 {
		str = StrPad(str, 64, "0", 1)
	}
	return str, nil
}

// 比较文字相同不同的数量得出相似度
func SimilarText(utf8Str1, utf8Str2 string) float64 {
	r1 := []rune(SpecialcharsReplace(utf8Str1))
	r2 := []rune(SpecialcharsReplace(utf8Str2))
	cacheX := make([]int, len(r2))
	if len(r1) == len(r2) {
		return 1
	}

	diagonal := 0
	for y, yLen := 0, len(r1); y < yLen; y++ {
		for x, xLen := 0, len(cacheX); x < xLen; x++ {
			on := x + 1
			left := y + 1
			if x == 0 {
				diagonal = y
			} else if y == 0 {
				diagonal = x
			}
			if y > 0 {
				on = cacheX[x]
			}
			if x-1 >= 0 {
				left = cacheX[x-1]
			}

			same := 0
			if r1[y] != r2[x] {
				same = 1
			}

			oldDiagonal := cacheX[x]
			cacheX[x] = min(min(on+1, left+1), same+diagonal)
			diagonal = oldDiagonal

		}
	}

	//e.mixed = cacheX[len(cacheX)-1]
	return 1.0 - float64(cacheX[len(cacheX)-1])/float64(max(len(r1), len(r2)))
}

// 降维度
func Dimensionality(ins []int) []int {
	for i := 0; i < len(ins); i++ {
		if ins[i] > 0 {
			ins[i] = 1
		} else {
			ins[i] = 0
		}

	}
	return ins
}

// 合并
func merge(ins [][]int) []int {
	res := make([]int, len(ins[0]))
	lens := len(ins)
	for i := 0; i < lens; i++ {
		for j := 0; j < len(ins[i]); j++ {
			res[j] += ins[i][j]
		}
	}
	return res
}

// 计算hashcode并加权
func hashcodeAndAdd(counts map[string]int) [][]int {
	// hashmap
	lens := len(counts)
	h1 := make([][]int, lens)
	// 计算counts1,counts2 中每个字符的hash值, 并且将出现的次数分为5个等级, 将每个字符的hash值与出现的次数等级相乘
	c1 := (lens - 1) * 4.0
	j := 0
	//for j := 0; j < lens; j++ {
	for k, v := range counts {
		////计算每一个字符串的hash
		c := strconv.FormatInt(murmurHash64A([]byte(k)), 2)

		// 将字符串转换为数字数组
		cs := Int64StrToInts(c)
		if v <= c1/5.0 {
			// 加权
			h1[j] = Add(cs, 1)
		} else if v <= c1/5.0*2 {
			// 加权
			h1[j] = Add(cs, 2)
		} else if v <= c1/5.0*3 {
			// 加权
			h1[j] = Add(cs, 3)
		} else if v <= c1/5.0*4 {
			// 加权
			h1[j] = Add(cs, 4)
		} else {
			// 加权
			h1[j] = Add(cs, 5)
		}
		j++
	}

	return h1
}

// Int64StrToInts   将uint64转换成string
func Int64StrToInts(ins string) []int {
	uints := make([]int, 64)

	for i := 0; i < len(ins); i++ {
		if string(ins[i]) == "1" {
			uints[i] = 1
		} else if string(ins[i]) == "0" {
			uints[i] = 0
		}
	}
	return uints

}

// IntsToStr []int 转换成string
func IntsToStr(ins []int) string {
	res := ""
	for _, v := range ins {
		res += strconv.Itoa(v)
	}

	return res
}

// Add 加权
func Add(uint64 []int, int int) []int {
	lens := len(uint64)
	for i := 0; i < 64; i++ {
		if i < lens {
			if uint64[i] == 1 {
				uint64[i] = int
			} else {
				uint64[i] = -int
			}
		} else {
			uint64 = append(uint64, int)
		}

	}
	return uint64
}

func StringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

// 补齐  0补左边 1右边 2两边
func StrPad(Input string, PadLength int, PadString string, PadType int) string {
	var leftPad, rightPad = 0, 0
	numPadChars := PadLength - len(Input)
	if numPadChars <= 0 {
		return Input
	}
	var buffer bytes.Buffer
	buffer.WriteString(Input)
	switch PadType {
	case 0:
		leftPad = numPadChars
		rightPad = 0
	case 1:
		leftPad = 0
		rightPad = numPadChars
	case 2:
		rightPad = numPadChars / 2
		leftPad = numPadChars - rightPad
	}

	var leftBuffer bytes.Buffer
	/* 左填充：循环添加字符*/
	for i := 0; i < leftPad; i++ {
		leftBuffer.WriteString(PadString)
		if leftBuffer.Len() > leftPad {
			leftBuffer.Truncate(leftPad)
			break
		}
	}

	/* 右填充：循环添加字符串*/
	for i := 0; i < rightPad; i++ {
		buffer.WriteString(PadString)
		if buffer.Len() > PadLength {
			buffer.Truncate(PadLength)
			break
		}
	}

	leftBuffer.WriteString(buffer.String())
	return leftBuffer.String()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

const (
	BIG_M = 0xc6a4a7935bd1e995
	BIG_R = 47
	SEED  = 0x1234ABCD
)

func murmurHash64A(data []byte) (h int64) {
	var k int64
	h = SEED ^ int64(uint64(len(data))*BIG_M)

	var ubigm uint64 = BIG_M
	var ibigm = int64(ubigm)
	for l := len(data); l >= 8; l -= 8 {
		k = int64(int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24 |
			int64(data[4])<<32 | int64(data[5])<<40 | int64(data[6])<<48 | int64(data[7])<<56)

		k := k * ibigm
		k ^= int64(uint64(k) >> BIG_R)
		k = k * ibigm

		h = h ^ k
		h = h * ibigm
		data = data[8:]
	}

	switch len(data) {
	case 7:
		h ^= int64(data[6]) << 48
		fallthrough
	case 6:
		h ^= int64(data[5]) << 40
		fallthrough
	case 5:
		h ^= int64(data[4]) << 32
		fallthrough
	case 4:
		h ^= int64(data[3]) << 24
		fallthrough
	case 3:
		h ^= int64(data[2]) << 16
		fallthrough
	case 2:
		h ^= int64(data[1]) << 8
		fallthrough
	case 1:
		h ^= int64(data[0])
		h *= ibigm
	}

	h ^= int64(uint64(h) >> BIG_R)
	h *= ibigm
	h ^= int64(uint64(h) >> BIG_R)
	return
}
