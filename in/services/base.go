package services

import (
	"interview/common/global"
	"interview/database"
	"interview/models"

	"github.com/garyburd/redigo/redis"
	"gorm.io/gorm"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mgoCfg = global.CONFIG.MongoDB

type ServicesBase struct {
}

func (sf *ServicesBase) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *ServicesBase) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}
func (sf *ServicesBase) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}
func (sf *ServicesBase) Mysql() *gorm.DB {
	return global.Mysql
}

// redis pool
func (sf *ServicesBase) RDBPool(dbName ...string) *redis.Pool {

	return global.REDISPOOL
}
func (sf *ServicesBase) ObjectID(strHex string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(strHex)
	if err != nil {
		return objectId
	}
	return objectId
}

// strings->primitive ids
func (sf *ServicesBase) ObjectIDs(ids []string, noError ...bool) []primitive.ObjectID {
	objectIDs := make([]primitive.ObjectID, 0)
	for i := 0; i < len(ids); i++ {
		objectId, err := primitive.ObjectIDFromHex(ids[i])
		if err != nil {
			if len(noError) > 0 {
				continue
			}
			// sf.HttpError(err)
		}
		objectIDs = append(objectIDs, objectId)
	}
	return objectIDs
}

func (sf *ServicesBase) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}

// 分页
func (sf *ServicesBase) PageLimit(page, size int64) (int64, int64) {
	if size == 0 {
		size = 20
	}
	if page == 0 {
		return 0, size
	}
	return (page - 1) * size, size
}

func (sf *ServicesBase) AreaJson() string {
	return `[{"acronym":"BXDQ","childrenLabels":[],"id":"000000000000000000000000","ntitle":"不限地区","title":"不限地区","code":""},{"acronym":"A","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"H","childrenLabels":[],"title":"合肥市"},{"acronym":"W","childrenLabels":[],"title":"芜湖市"},{"acronym":"H","childrenLabels":[],"title":"黄山市"},{"acronym":"L","childrenLabels":[],"title":"六安市"},{"acronym":"B","childrenLabels":[],"title":"亳州市"},{"acronym":"C","childrenLabels":[],"title":"滁州市"},{"acronym":"C","childrenLabels":[],"title":"池州市"},{"acronym":"H","childrenLabels":[],"title":"淮北市"},{"acronym":"A","childrenLabels":[],"title":"安庆市"},{"acronym":"X","childrenLabels":[],"title":"宣城市"},{"acronym":"S","childrenLabels":[],"title":"宿州市"},{"acronym":"T","childrenLabels":[],"title":"铜陵市"},{"acronym":"H","childrenLabels":[],"title":"淮南市"},{"acronym":"B","childrenLabels":[],"title":"蚌埠市"}],"id":"61af320cfa844faa9ee57fd1","ntitle":"安徽","title":"安徽省","code":"340000"},{"acronym":"B","childrenLabels":[],"id":"61af320cfa844faa9ee57fdb","ntitle":"北京","title":"北京市","code":"110000"},{"acronym":"C","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"Y","childrenLabels":[],"title":"渝北区"},{"acronym":"B","childrenLabels":[],"title":"北碚区"},{"acronym":"D","childrenLabels":[],"title":"大渡口区"},{"acronym":"W","childrenLabels":[],"title":"武隆区"},{"acronym":"F","childrenLabels":[],"title":"涪陵区"},{"acronym":"Z","childrenLabels":[],"title":"长寿区"},{"acronym":"J","childrenLabels":[],"title":"江津区"},{"acronym":"Q","childrenLabels":[],"title":"綦江区"},{"acronym":"D","childrenLabels":[],"title":"大足区"},{"acronym":"K","childrenLabels":[],"title":"开州区"},{"acronym":"H","childrenLabels":[],"title":"合川区"},{"acronym":"N","childrenLabels":[],"title":"南岸区"},{"acronym":"T","childrenLabels":[],"title":"铜梁区"},{"acronym":"Y","childrenLabels":[],"title":"渝中区"},{"acronym":"N","childrenLabels":[],"title":"南川区"},{"acronym":"W","childrenLabels":[],"title":"万州区"},{"acronym":"B","childrenLabels":[],"title":"璧山区"},{"acronym":"B","childrenLabels":[],"title":"巴南区"},{"acronym":"S","childrenLabels":[],"title":"沙坪坝区"},{"acronym":"J","childrenLabels":[],"title":"江北区"},{"acronym":"F","childrenLabels":[],"title":"丰都县"},{"acronym":"Q","childrenLabels":[],"title":"黔江区"},{"acronym":"R","childrenLabels":[],"title":"荣昌区"},{"acronym":"P","childrenLabels":[],"title":"彭水苗族土家族自治县"},{"acronym":"L","childrenLabels":[],"title":"梁平区"},{"acronym":"S","childrenLabels":[],"title":"石柱土家族自治县"},{"acronym":"J","childrenLabels":[],"title":"九龙坡区"},{"acronym":"C","childrenLabels":[],"title":"城口县"},{"acronym":"D","childrenLabels":[],"title":"垫江县"},{"acronym":"W","childrenLabels":[],"title":"巫山县"},{"acronym":"Y","childrenLabels":[],"title":"云阳县"}],"id":"61af320cfa844faa9ee57fcd","ntitle":"重庆","title":"重庆市","code":"500000"},{"acronym":"F","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"F","childrenLabels":[],"title":"福州市"},{"acronym":"N","childrenLabels":[],"title":"南平市"},{"acronym":"P","childrenLabels":[],"title":"莆田市"},{"acronym":"S","childrenLabels":[],"title":"厦门市"}],"id":"61af320cfa844faa9ee57fd4","ntitle":"福建","title":"福建省","code":"350000"},{"acronym":"G","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"L","childrenLabels":[],"title":"兰州市"},{"acronym":"T","childrenLabels":[],"title":"天水市"},{"acronym":"J","childrenLabels":[],"title":"酒泉市"}],"id":"61af320cfa844faa9ee57fc2","ntitle":"甘肃","title":"甘肃省","code":"620000"},{"acronym":"G","childrenLabels":[{"acronym":"G","childrenLabels":[],"title":"贵阳市"},{"acronym":"Z","childrenLabels":[],"title":"遵义市"},{"acronym":"L","childrenLabels":[],"title":"六盘水市"},{"acronym":"B","childrenLabels":[],"title":"毕节市"},{"acronym":"Q","childrenLabels":[],"title":"黔东南苗族侗族自治州"},{"acronym":"Q","childrenLabels":[],"title":"黔南布依族苗族自治州"},{"acronym":"Q","childrenLabels":[],"title":"黔西南布依族苗族自治州"},{"acronym":"A","childrenLabels":[],"title":"安顺市"},{"acronym":"T","childrenLabels":[],"title":"铜仁市"}],"id":"61af320cfa844faa9ee57fc4","ntitle":"贵州","title":"贵州省","code":"520000"},{"acronym":"G","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"G","childrenLabels":[],"title":"桂林市"},{"acronym":"Y","childrenLabels":[],"title":"玉林市"},{"acronym":"H","childrenLabels":[],"title":"河池市"},{"acronym":"Q","childrenLabels":[],"title":"钦州市"}],"id":"61af320cfa844faa9ee57fca","ntitle":"广西","title":"广西壮族自治区","code":"450000"},{"acronym":"G","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"G","childrenLabels":[],"title":"广州市"},{"acronym":"F","childrenLabels":[],"title":"佛山市"},{"acronym":"M","childrenLabels":[],"title":"茂名市"},{"acronym":"M","childrenLabels":[],"title":"梅州市"},{"acronym":"Y","childrenLabels":[],"title":"阳江市"},{"acronym":"Y","childrenLabels":[],"title":"云浮市"},{"acronym":"J","childrenLabels":[],"title":"江门市"},{"acronym":"Z","childrenLabels":[],"title":"肇庆市"},{"acronym":"C","childrenLabels":[],"title":"潮州市"},{"acronym":"Q","childrenLabels":[],"title":"清远市"},{"acronym":"H","childrenLabels":[],"title":"河源市"},{"acronym":"H","childrenLabels":[],"title":"惠州市"},{"acronym":"Z","childrenLabels":[],"title":"中山市"},{"acronym":"S","childrenLabels":[],"title":"汕尾市"},{"acronym":"J","childrenLabels":[],"title":"揭阳市"},{"acronym":"S","childrenLabels":[],"title":"汕头市"}],"id":"61af320cfa844faa9ee57fd8","ntitle":"广东","title":"广东省","code":"440000"},{"acronym":"H","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"L","childrenLabels":[],"title":"漯河市"},{"acronym":"X","childrenLabels":[],"title":"新乡市"},{"acronym":"Z","childrenLabels":[],"title":"周口市"},{"acronym":"Z","childrenLabels":[],"title":"郑州市"},{"acronym":"K","childrenLabels":[],"title":"开封市"},{"acronym":"J","childrenLabels":[],"title":"焦作市"},{"acronym":"X","childrenLabels":[],"title":"许昌市"},{"acronym":"P","childrenLabels":[],"title":"濮阳市"},{"acronym":"A","childrenLabels":[],"title":"安阳市"},{"acronym":"L","childrenLabels":[],"title":"洛阳市"},{"acronym":"H","childrenLabels":[],"title":"鹤壁市"},{"acronym":"S","childrenLabels":[],"title":"商丘市"},{"acronym":"S","childrenLabels":[],"title":"三门峡市"},{"acronym":"P","childrenLabels":[],"title":"平顶山市"},{"acronym":"N","childrenLabels":[],"title":"南阳市"},{"acronym":"Z","childrenLabels":[],"title":"驻马店市"},{"acronym":"X","childrenLabels":[],"title":"信阳市"}],"id":"61af320cfa844faa9ee57fc5","ntitle":"河南","title":"河南省","code":"410000"},{"acronym":"H","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"Z","childrenLabels":[],"title":"长沙市"},{"acronym":"Z","childrenLabels":[],"title":"株洲市"},{"acronym":"X","childrenLabels":[],"title":"湘潭市"},{"acronym":"S","childrenLabels":[],"title":"邵阳市"},{"acronym":"Y","childrenLabels":[],"title":"岳阳市"},{"acronym":"Z","childrenLabels":[],"title":"张家界市"},{"acronym":"C","childrenLabels":[],"title":"常德市"},{"acronym":"C","childrenLabels":[],"title":"郴州市"},{"acronym":"Y","childrenLabels":[],"title":"永州市"},{"acronym":"X","childrenLabels":[],"title":"湘西土家族苗族自治州"},{"acronym":"H","childrenLabels":[],"title":"衡阳市"},{"acronym":"Y","childrenLabels":[],"title":"益阳市"},{"acronym":"L","childrenLabels":[],"title":"娄底市"},{"acronym":"H","childrenLabels":[],"title":"怀化市"}],"id":"61af320cfa844faa9ee57fc8","ntitle":"湖南","title":"湖南省","code":"430000"},{"acronym":"H","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"H","childrenLabels":[],"title":"哈尔滨市"},{"acronym":"J","childrenLabels":[],"title":"佳木斯市"},{"acronym":"H","childrenLabels":[],"title":"黑河市"},{"acronym":"D","childrenLabels":[],"title":"大庆市"},{"acronym":"Q","childrenLabels":[],"title":"七台河市"},{"acronym":"S","childrenLabels":[],"title":"绥化市"},{"acronym":"Q","childrenLabels":[],"title":"齐齐哈尔市"},{"acronym":"J","childrenLabels":[],"title":"鸡西市"},{"acronym":"S","childrenLabels":[],"title":"双鸭山市"}],"id":"61af320cfa844faa9ee57fcb","ntitle":"黑龙江","title":"黑龙江省","code":"230000"},{"acronym":"H","childrenLabels":[{"acronym":"W","childrenLabels":[],"title":"武汉市"},{"acronym":"H","childrenLabels":[],"title":"黄石市"},{"acronym":"S","childrenLabels":[],"title":"十堰市"},{"acronym":"Y","childrenLabels":[],"title":"宜昌市"},{"acronym":"X","childrenLabels":[],"title":"孝感市"},{"acronym":"X","childrenLabels":[],"title":"咸宁市"},{"acronym":"H","childrenLabels":[],"title":"黄冈市"},{"acronym":"X","childrenLabels":[],"title":"襄阳市"}],"id":"61af320cfa844faa9ee57fcf","ntitle":"湖北","title":"湖北省","code":"420000"},{"acronym":"H","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"S","childrenLabels":[],"title":"石家庄市"},{"acronym":"Z","childrenLabels":[],"title":"张家口市"},{"acronym":"Q","childrenLabels":[],"title":"秦皇岛市"},{"acronym":"T","childrenLabels":[],"title":"唐山市"},{"acronym":"C","childrenLabels":[],"title":"承德市"},{"acronym":"C","childrenLabels":[],"title":"沧州市"},{"acronym":"L","childrenLabels":[],"title":"廊坊市"},{"acronym":"H","childrenLabels":[],"title":"衡水市"},{"acronym":"B","childrenLabels":[],"title":"保定市"},{"acronym":"X","childrenLabels":[],"title":"邢台市"}],"id":"61af320cfa844faa9ee57fd2","ntitle":"河北","title":"河北省","code":"130000"},{"acronym":"H","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"S","childrenLabels":[],"title":"三亚市"},{"acronym":"Q","childrenLabels":[],"title":"琼海市"},{"acronym":"H","childrenLabels":[],"title":"海口市"}],"id":"61af320cfa844faa9ee57fda","ntitle":"海南","title":"海南省","code":"460000"},{"acronym":"J","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"N","childrenLabels":[],"title":"南通市"},{"acronym":"L","childrenLabels":[],"title":"连云港市"},{"acronym":"C","childrenLabels":[],"title":"常州市"},{"acronym":"X","childrenLabels":[],"title":"徐州市"},{"acronym":"S","childrenLabels":[],"title":"苏州市"},{"acronym":"N","childrenLabels":[],"title":"南京市"},{"acronym":"Y","childrenLabels":[],"title":"盐城市"},{"acronym":"T","childrenLabels":[],"title":"泰州市"},{"acronym":"Z","childrenLabels":[],"title":"镇江市"},{"acronym":"H","childrenLabels":[],"title":"淮安市"},{"acronym":"S","childrenLabels":[],"title":"宿迁市"}],"id":"61af320cfa844faa9ee57fc3","ntitle":"江苏","title":"江苏省","code":"320000"},{"acronym":"J","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"N","childrenLabels":[],"title":"南昌市"},{"acronym":"G","childrenLabels":[],"title":"赣州市"},{"acronym":"Y","childrenLabels":[],"title":"宜春市"},{"acronym":"J","childrenLabels":[],"title":"吉安市"},{"acronym":"S","childrenLabels":[],"title":"上饶市"},{"acronym":"F","childrenLabels":[],"title":"抚州市"},{"acronym":"J","childrenLabels":[],"title":"九江市"},{"acronym":"J","childrenLabels":[],"title":"景德镇市"},{"acronym":"X","childrenLabels":[],"title":"新余市"},{"acronym":"Y","childrenLabels":[],"title":"鹰潭市"}],"id":"61af320cfa844faa9ee57fc7","ntitle":"江西","title":"江西省","code":"360000"},{"acronym":"J","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"Z","childrenLabels":[],"title":"长春市"},{"acronym":"J","childrenLabels":[],"title":"吉林市"},{"acronym":"T","childrenLabels":[],"title":"通化市"},{"acronym":"B","childrenLabels":[],"title":"白山市"},{"acronym":"Y","childrenLabels":[],"title":"延边朝鲜族自治州"},{"acronym":"S","childrenLabels":[],"title":"松原市"},{"acronym":"S","childrenLabels":[],"title":"四平市"},{"acronym":"B","childrenLabels":[],"title":"白城市"},{"acronym":"L","childrenLabels":[],"title":"辽源市"}],"id":"61af320cfa844faa9ee57fc9","ntitle":"吉林","title":"吉林省","code":"220000"},{"acronym":"L","childrenLabels":[{"acronym":"D","childrenLabels":[],"title":"大连市"},{"acronym":"S","childrenLabels":[],"title":"沈阳市"},{"acronym":"H","childrenLabels":[],"title":"葫芦岛市"},{"acronym":"B","childrenLabels":[],"title":"本溪市"}],"id":"61af320cfa844faa9ee57fd7","ntitle":"辽宁","title":"辽宁省","code":"210000"},{"acronym":"N","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"C","childrenLabels":[],"title":"赤峰市"},{"acronym":"T","childrenLabels":[],"title":"通辽市"},{"acronym":"B","childrenLabels":[],"title":"包头市"},{"acronym":"H","childrenLabels":[],"title":"呼伦贝尔市"},{"acronym":"W","childrenLabels":[],"title":"乌兰察布市"},{"acronym":"A","childrenLabels":[],"title":"阿拉善盟"},{"acronym":"E","childrenLabels":[],"title":"鄂尔多斯市"}],"id":"61af320cfa844faa9ee57fce","ntitle":"内蒙古","title":"内蒙古自治区","code":"150000"},{"acronym":"N","childrenLabels":[],"id":"6322b4eb48434730740ba82b","ntitle":"宁夏","title":"宁夏回族自治区","code":"640000"},{"acronym":"Q","childrenLabels":[],"id":"61af320cfa844faa9ee57fcc","ntitle":"青海","title":"青海省","code":"630000"},{"acronym":"S","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"B","childrenLabels":[],"title":"滨州市"},{"acronym":"J","childrenLabels":[],"title":"济南市"},{"acronym":"Q","childrenLabels":[],"title":"青岛市"},{"acronym":"Z","childrenLabels":[],"title":"淄博市"},{"acronym":"Z","childrenLabels":[],"title":"枣庄市"},{"acronym":"D","childrenLabels":[],"title":"东营市"},{"acronym":"Y","childrenLabels":[],"title":"烟台市"},{"acronym":"W","childrenLabels":[],"title":"潍坊市"},{"acronym":"J","childrenLabels":[],"title":"济宁市"},{"acronym":"T","childrenLabels":[],"title":"泰安市"},{"acronym":"W","childrenLabels":[],"title":"威海市"},{"acronym":"D","childrenLabels":[],"title":"德州市"},{"acronym":"L","childrenLabels":[],"title":"临沂市"},{"acronym":"L","childrenLabels":[],"title":"聊城市"}],"id":"61af320cfa844faa9ee57fc0","ntitle":"山东","title":"山东省","code":"370000"},{"acronym":"S","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"C","childrenLabels":[],"title":"成都市"},{"acronym":"M","childrenLabels":[],"title":"绵阳市"},{"acronym":"Z","childrenLabels":[],"title":"自贡市"},{"acronym":"L","childrenLabels":[],"title":"泸州市"},{"acronym":"D","childrenLabels":[],"title":"德阳市"},{"acronym":"N","childrenLabels":[],"title":"内江市"},{"acronym":"G","childrenLabels":[],"title":"广元市"},{"acronym":"M","childrenLabels":[],"title":"眉山市"},{"acronym":"Y","childrenLabels":[],"title":"宜宾市"},{"acronym":"B","childrenLabels":[],"title":"巴中市"},{"acronym":"Y","childrenLabels":[],"title":"雅安市"},{"acronym":"L","childrenLabels":[],"title":"凉山州"},{"acronym":"G","childrenLabels":[],"title":"甘孜州"},{"acronym":"P","childrenLabels":[],"title":"攀枝花市"}],"id":"61af320cfa844faa9ee57fc1","ntitle":"四川","title":"四川省","code":"510000"},{"acronym":"S","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"X","childrenLabels":[],"title":"西安市"},{"acronym":"H","childrenLabels":[],"title":"汉中市"},{"acronym":"S","childrenLabels":[],"title":"商洛市"},{"acronym":"T","childrenLabels":[],"title":"铜川市"},{"acronym":"Y","childrenLabels":[],"title":"榆林市"},{"acronym":"W","childrenLabels":[],"title":"渭南市"},{"acronym":"Y","childrenLabels":[],"title":"延安市"}],"id":"61af320cfa844faa9ee57fd3","ntitle":"陕西","title":"陕西省","code":"610000"},{"acronym":"S","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"T","childrenLabels":[],"title":"太原市"},{"acronym":"D","childrenLabels":[],"title":"大同市"},{"acronym":"Y","childrenLabels":[],"title":"运城市"},{"acronym":"J","childrenLabels":[],"title":"晋中市"},{"acronym":"Y","childrenLabels":[],"title":"阳泉市"},{"acronym":"L","childrenLabels":[],"title":"临汾市"},{"acronym":"Z","childrenLabels":[],"title":"长治市"},{"acronym":"J","childrenLabels":[],"title":"晋城市"},{"acronym":"L","childrenLabels":[],"title":"吕梁市"},{"acronym":"X","childrenLabels":[],"title":"忻州市"}],"id":"61af320cfa844faa9ee57fd6","ntitle":"山西","title":"山西省","code":"140000"},{"acronym":"S","childrenLabels":[],"id":"61af320cfa844faa9ee57fdc","ntitle":"上海","title":"上海市","code":"310000"},{"acronym":"T","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"H","childrenLabels":[],"title":"河西区"},{"acronym":"B","childrenLabels":[],"title":"滨海新区"},{"acronym":"D","childrenLabels":[],"title":"东丽区"},{"acronym":"N","childrenLabels":[],"title":"南开区"},{"acronym":"X","childrenLabels":[],"title":"西青区"},{"acronym":"J","childrenLabels":[],"title":"静海区"},{"acronym":"H","childrenLabels":[],"title":"河北区"},{"acronym":"J","childrenLabels":[],"title":"津南区"},{"acronym":"B","childrenLabels":[],"title":"北辰区"},{"acronym":"B","childrenLabels":[],"title":"宝坻区"}],"id":"61af320cfa844faa9ee57fd5","ntitle":"天津","title":"天津市","code":"120000"},{"acronym":"X","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"K","childrenLabels":[],"title":"喀什地区"},{"acronym":"Y","childrenLabels":[],"title":"伊犁哈萨克自治州"},{"acronym":"H","childrenLabels":[],"title":"和田地区"},{"acronym":"C","childrenLabels":[],"title":"昌吉回族自治州"},{"acronym":"B","childrenLabels":[],"title":"巴音郭楞蒙古自治州"},{"acronym":"T","childrenLabels":[],"title":"吐鲁番市"},{"acronym":"A","childrenLabels":[],"title":"阿勒泰地区"},{"acronym":"H","childrenLabels":[],"title":"哈密市"}],"id":"61af320cfa844faa9ee57fd9","ntitle":"新疆","title":"新疆维吾尔自治区","code":"650000"},{"acronym":"Y","childrenLabels":[{"acronym":"T","childrenLabels":[],"title":"统考"},{"acronym":"L","childrenLabels":[],"title":"拉萨市"},{"acronym":"R","childrenLabels":[],"title":"日喀则市"},{"acronym":"L","childrenLabels":[],"title":"林芝市"},{"acronym":"S","childrenLabels":[],"title":"山南市"},{"acronym":"A","childrenLabels":[],"title":"阿里地区"}],"id":"6593b071474d5661aed0f9d8","ntitle":"西藏","title":"西藏自治区","code":"540000"},{"acronym":"Y","childrenLabels":[{"acronym":"B","childrenLabels":[],"title":"保山市"},{"acronym":"Z","childrenLabels":[],"title":"昭通市"},{"acronym":"Y","childrenLabels":[],"title":"玉溪市"},{"acronym":"D","childrenLabels":[],"title":"大理白族自治州"},{"acronym":"D","childrenLabels":[],"title":"德宏傣族景颇族自治州"},{"acronym":"Q","childrenLabels":[],"title":"曲靖市"}],"id":"61af320cfa844faa9ee57fc6","ntitle":"云南","title":"云南省","code":"530000"},{"acronym":"Z","childrenLabels":[{"acronym":"H","childrenLabels":[],"title":"杭州市"},{"acronym":"W","childrenLabels":[],"title":"温州市"},{"acronym":"Q","childrenLabels":[],"title":"衢州市"},{"acronym":"N","childrenLabels":[],"title":"宁波市"},{"acronym":"T","childrenLabels":[],"title":"台州市"},{"acronym":"L","childrenLabels":[],"title":"丽水市"},{"acronym":"J","childrenLabels":[],"title":"金华市"},{"acronym":"H","childrenLabels":[],"title":"湖州市"},{"acronym":"S","childrenLabels":[],"title":"绍兴市"},{"acronym":"Z","childrenLabels":[],"title":"舟山市"}],"id":"61af320cfa844faa9ee57fd0","ntitle":"浙江","title":"浙江省","code":"330000"}]`
}

// MockExamModel
func (sf *ServicesBase) MockExamModel() *database.MongoWork {
	return sf.DB().Collection(new(models.MockExam).TableName())
}

// MockExamLogModel
func (sf *ServicesBase) MockExamLogModel() *database.MongoWork {
	return sf.DB().Collection(new(models.MockExamLog).TableName())
}

// MockExamSlotRoomModel
func (sf *ServicesBase) MockExamSlotRoomModel() *database.MongoWork {
	return sf.DB().Collection(new(models.MockExamSlotRoom).TableName())
}

func (sf *ServicesBase) PaperModel() *database.MongoWork {
	return sf.DB().Collection(new(models.Paper).TableName())
}

func (sf *ServicesBase) ReviewModel() *database.MongoWork {
	return sf.DB().Collection(new(models.Review).TableName())
}
func (sf *ServicesBase) ReviewLogModel() *database.MongoWork {
	return sf.DB().Collection(new(models.ReviewLog).TableName())
}

func (sf *ServicesBase) GQuestionModel() *database.MongoWork {
	return sf.DB().Collection(new(models.GQuestion).TableName())
}

func (sf *ServicesBase) GAnswerLogModel() *database.MongoWork {
	return sf.DB().Collection(new(models.GAnswerLog).TableName())
}

func (sf *ServicesBase) GAnswerLogDailyStatisticsModel() *database.MongoWork {
	return sf.DB().Collection(new(models.GAnswerLogDailyStatistics).TableName())
}

func (sf *ServicesBase) RecommendCourseModel() *database.MongoWork {
	return sf.DB().Collection(new(models.RecommendCourse).TableName())
}

func (sf *ServicesBase) RecommendDataPackModel() *database.MongoWork {
	return sf.DB().Collection(new(models.RecommendDataPack).TableName())
}
