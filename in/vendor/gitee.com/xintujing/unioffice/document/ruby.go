//name: zhexiao(肖哲)
//date: 2019-11-05
//一些与Ruby数据（拼音）相关
//================================start
package document

type CT_Ruby struct {
	Rt       []string
	RubyBase []string
}

func New_CT_Ruby() *CT_Ruby {
	return &CT_Ruby{}
}
