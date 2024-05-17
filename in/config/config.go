package config

import "time"

// 映射 总配置信息
type Server struct {
	MongoDB       MongoDB       `mapstructure:"mongodb" json:"mongodb" yaml:"mongodb"`
	Mysql         Mysql         `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis         Redis         `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log           Log           `mapstructure:"log" json:"log" yaml:"log"`
	ServiceUrls   ServiceUrls   `mapstructure:"service-urls" json:"service-urls" yaml:"service-urls"`
	SignatureInfo SignatureInfo `mapstructure:"signature-info" json:"signature-info" yaml:"signature-info"`
	GPT           GPT           `mapstructure:"gpt" json:"gpt" yaml:"gpt"`
	GPT4          GPT4          `mapstructure:"gpt4" json:"gpt4" yaml:"gpt4"`
	TTS           TTS           `mapstructure:"tts" json:"tts" yaml:"tts"`
	Kafka         Kafka         `mapstructure:"kafka" json:"kafka" yaml:"kafka"`
	ES            ES            `mapstructure:"es" json:"es" yaml:"es"`
	Wechat        Wechat        `mapstructure:"wechat" json:"wechat" yaml:"wechat"`
	Zego          Zego          `mapstructure:"zego" json:"zego" yaml:"zego"`                         //第三方音视频服务
	MockExamOSS   MockExamOSS   `mapstructure:"mockexam-oss" json:"mockexam-oss" yaml:"mockexam-oss"` //面试模考视频OSS配置
	Env           string        `mapstructure:"env" json:"env" yaml:"env"`                            //环境变量
}

// 保存 mongodb 配置信息
type MongoDB struct {
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Path     string `mapstructure:"path" json:"path" yaml:"path"`
	Port     string `mapstructure:"port" json:"port" yaml:"port"`
	Dbname   string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`
}
type Mysql struct {
	Path     string `mapstructure:"path" json:"path" yaml:"path"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DBName   string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`
}

type Redis struct {
	Addr        string        `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password    string        `mapstructure:"password" json:"password" yaml:"password"`
	DB          int           `mapstructure:"db" json:"db" yaml:"db"`
	MaxIdle     int           `mapstructure:"max-idle" json:"maxIdle" yaml:"max-idle"`
	MaxActive   int           `mapstructure:"max-active" json:"maxActive" yaml:"max-active"`
	IdleTimeout time.Duration `mapstructure:"idle-timeout" json:"idleTimeout" yaml:"idle-timeout"`
	Expire      int           `mapstructure:"expire" json:"expire" yaml:"expire"`
	BankDB      int           `mapstructure:"bank-db" json:"bank-db" yaml:"bank-db"`
}
type Log struct {
	Prefix   string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	LogType  string `mapstructure:"log-type" json:"log-type" yaml:"log-type"`
	Filepath string `mapstructure:"file-path" json:"file-path" yaml:"file-path"`
	Debug    bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
}
type ServiceUrls struct {
	ParseUrl               string `mapstructure:"parse-url" json:"parse-url" yaml:"parse-url"`
	UploadUrl              string `mapstructure:"upload-url" json:"upload-url" yaml:"upload-url"`
	UserInfoUrl            string `mapstructure:"user-info-url" json:"user-info-url" yaml:"user-info-url"`
	Omml2ImageUrl          string `mapstructure:"omml-to-image-url" json:"omml-to-image-url" yaml:"omml-to-image-url"`
	RelevanceCourseUrl     string `mapstructure:"relevance-course-url" json:"relevance-course-url" yaml:"relevance-course-url"`
	Mobile2User            string `mapstructure:"mobile-2-user" json:"mobile-2-user" yaml:"mobile-2-user"`
	RegisterUserUrl        string `mapstructure:"register-user-url" json:"register-user-url" yaml:"register-user-url"`
	UserLoginUrl           string `mapstructure:"user-login-url" json:"user-login-url" yaml:"user-login-url"`
	UserMobileUrl          string `mapstructure:"user-mobile-url" json:"user-mobile-url" yaml:"user-mobile-url"`
	GrowUrl                string `mapstructure:"grow-url" json:"grow-url" yaml:"grow-url"`
	TokenUrl               string `mapstructure:"token-url" json:"token-url" yaml:"token-url"`
	CrmDataServiceUrl      string `mapstructure:"crm-data-service-url" json:"crm-data-service-url" yaml:"crm-data-service-url"`
	QuestionBankUrl        string `mapstructure:"question-bank-url" json:"question-bank-url" yaml:"question-bank-url"`
	GPTServiceUrl          string `mapstructure:"gpt-service-url" json:"gpt-service-url" yaml:"gpt-service-url"`
	ClickhouseStatisticUrl string `mapstructure:"clickhouse-statistic-url" json:"clickhouse-statistic-url" yaml:"clickhouse-statistic-url"`
}
type SignatureInfo struct {
	SK string `mapstructure:"sk" json:"sk" yaml:"sk"`
}

type GPT struct {
	ApiKey  string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseUrl string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Engine  string `mapstructure:"engine" json:"engine" yaml:"engine"`
}

type GPT4 struct {
	ApiKey  string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseUrl string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Engine  string `mapstructure:"engine" json:"engine" yaml:"engine"`
}

type TTS struct {
	AzureKey             string `mapstructure:"azure-key" json:"azure-key" yaml:"azure-key"`
	AzureRegion          string `mapstructure:"azure-region" json:"azure-region" yaml:"azure-region"`
	AzureMaleVoiceName   string `mapstructure:"azure-male-voice-name" json:"azure-male-voice-name" yaml:"azure-male-voice-name"`
	AzureFemaleVoiceName string `mapstructure:"azure-female-voice-name" json:"azure-female-voice-name" yaml:"azure-female-voice-name"`
	HeadPrompt           string `mapstructure:"head-prompt" json:"head-prompt" yaml:"head-prompt"`
	TailPrompt           string `mapstructure:"tail-prompt" json:"tail-prompt" yaml:"tail-prompt"`
	Concurrency          int    `mapstructure:"concurrency" json:"concurrency" yaml:"concurrency"`
}

type Kafka struct {
	Address                 string `mapstructure:"address" json:"address" yaml:"address"`
	GroupId                 string `mapstructure:"group-id" json:"group-id" yaml:"group-id"`
	GPTServiceGroupId       string `mapstructure:"gpt-service-group-id" json:"gpt-service-group-id" yaml:"gpt-service-group-id"`
	GPTQuestionPreviewTopic string `mapstructure:"gpt-question-preview-topic" json:"gpt-question-preview-topic" yaml:"gpt-question-preview-topic"`
	QuestionGPTCommonTopic  string `mapstructure:"question-gpt-common-topic" json:"question-gpt-common-topic" yaml:"question-gpt-common-topic"`
}

type ES struct {
	ElasticUrl  string `mapstructure:"elastic_url" json:"elastic_url" yaml:"elastic_url"`
	ElasticName string `mapstructure:"elastic_name" json:"elastic_name" yaml:"elastic_name"`
	ElasticPwd  string `mapstructure:"elastic_pwd" json:"elastic_pwd" yaml:"elastic_pwd"`
}

type Wechat struct {
	AppId  string `mapstructure:"appid" json:"appid" yaml:"appid"`
	Secret string `mapstructure:"secret" json:"secret" yaml:"secret"`
}

type Zego struct {
	AppId          string `mapstructure:"appid" json:"appid" yaml:"appid"`
	Secret         string `mapstructure:"secret" json:"secret" yaml:"secret"`
	AppSign        string `mapstructure:"app_sign" json:"app_sign" yaml:"app_sign"`
	RoomUrl        string `mapstructure:"room_url" json:"room_url" yaml:"room_url"`
	CallbackSecret string `mapstructure:"callback_secret" json:"callback_secret" yaml:"callback_secret"`
}

type MockExamOSS struct {
	AppKeyId string `mapstructure:"appkeyid" json:"appkeyid" yaml:"appkeyid"`
	Secret   string `mapstructure:"secret" json:"secret" yaml:"secret"`
}
