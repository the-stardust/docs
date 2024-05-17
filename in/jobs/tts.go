package jobs

import (
	"fmt"
	"interview/common/global"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"sync"
	"time"

	speechCommon "github.com/Microsoft/cognitive-services-speech-sdk-go/common"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ttsFetchInterval = 230
	azureApiTimeout  = 100
	ossTimeout       = 120
	ttsReady2Go      = 5

	ttsTextTable           = "g_interview_questions"
	ttsUrlField            = "tts_url"
	ttsIDField             = "_id"
	ttsStatusField         = "status"
	ttsMaleVoiceUrlField   = "tts_url.male_voice_url"
	ttsFemaleVoiceUrlField = "tts_url.female_voice_url"

	ttsOssKey         = "question_voice/voice.wav"
	ttsOssBucketName  = "xtj-interview"
	ttsFileWavPostfix = ".wav"

	maleVoiceTaskFlag   = "0"
	femaleVoiceTaskFlag = "1"
)

func init() {
	tts := &Text2Speech{
		cap: make(chan struct{}, global.CONFIG.TTS.Concurrency),
	}
	RegisterOnceJob(tts)
}

type Text2Speech struct {
	JobBase
	task sync.Map
	cap  chan struct{}
}

func (sf *Text2Speech) Do() {
	sf.SLogger().Info("启动语音合成任务")
	//ticker := time.NewTicker(ttsFetchInterval * time.Second)
	for {
		var questions []models.GQuestion
		err := sf.DB().Collection(ttsTextTable).Where(
			bson.M{ttsStatusField: ttsReady2Go,
				"question_content_type": 0,
				"$or": []bson.M{
					bson.M{ttsMaleVoiceUrlField: nil},
					bson.M{ttsMaleVoiceUrlField: ""},
					bson.M{ttsFemaleVoiceUrlField: nil},
					bson.M{ttsFemaleVoiceUrlField: ""},
				}}).Limit(100).Find(&questions)
		if err == nil {
			sf.SLogger().Info(fmt.Sprintf("got [%d] tts tasks.", len(questions)))
			for _, q := range questions {
				if q.TTSUrl.MaleVoiceUrl == "" && q.TTSUrl.FemaleVoiceUrl == "" {
					_, err := sf.DB().Collection(ttsTextTable).Where(bson.M{ttsIDField: q.Id}).Update(bson.M{ttsUrlField: models.TTSUrl{}})
					if err != nil {
						sf.SLogger().Error(fmt.Sprintf("[%s] init tts_url field err: %+v", q.Id.Hex(), err))
					}
				}
				if q.TTSUrl.MaleVoiceUrl == "" {
					taskID := q.Id.Hex() + maleVoiceTaskFlag
					if _, exist := sf.task.Load(taskID); !exist {
						sf.cap <- struct{}{}
						sf.SLogger().Info(fmt.Sprintf("dispatch tts task: [%s]", taskID))
						sf.task.Store(taskID, struct{}{})
						go sf.azureTTS(q.Id, q.GetWantedQuestionContent(), maleVoiceTaskFlag)
						time.Sleep(4 * time.Second) // 免费 (F0) 每 60 秒 20 个事务
					}
				}
				if q.TTSUrl.FemaleVoiceUrl == "" {
					taskID := q.Id.Hex() + femaleVoiceTaskFlag
					if _, exist := sf.task.Load(taskID); !exist {
						sf.cap <- struct{}{}
						sf.SLogger().Info(fmt.Sprintf("dispatch tts task: [%s]", taskID))
						sf.task.Store(taskID, struct{}{})
						go sf.azureTTS(q.Id, q.GetWantedQuestionContent(), femaleVoiceTaskFlag)
						time.Sleep(4 * time.Second)
					}
				}
			}
		} else {
			sf.SLogger().Error(fmt.Sprintf("tts get trans list err: %+v", err))
		}
		if len(questions) <= 0 {
			sf.SLogger().Info(fmt.Sprintf("tts get trans list length 0 sleep: %dsecond", ttsFetchInterval))
			time.Sleep(ttsFetchInterval * time.Second)
		}
	}
}

func (sf *Text2Speech) azureTTS(id primitive.ObjectID, text, subTaskFlag string) {
	defer sf.task.Delete(id.Hex() + subTaskFlag)
	defer func() {
		sf.SLogger().Info(fmt.Sprintf("tts task:[%s] released", id.Hex()+subTaskFlag))
		<-sf.cap
	}()

	speechConfig, err := speech.NewSpeechConfigFromSubscription(global.CONFIG.TTS.AzureKey, global.CONFIG.TTS.AzureRegion)
	if err != nil {
		sf.SLogger().Error(fmt.Sprintf("[%s] azure NewSpeechConfigFromSubscription [%s] err: %+v", id.Hex(), subTaskFlag, err))
		return
	}
	defer speechConfig.Close()

	voiceName := global.CONFIG.TTS.AzureMaleVoiceName
	if subTaskFlag == femaleVoiceTaskFlag {
		voiceName = global.CONFIG.TTS.AzureFemaleVoiceName
	}
	speechConfig.SetSpeechSynthesisVoiceName(voiceName)
	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(speechConfig, nil)
	if err != nil {
		sf.SLogger().Error(fmt.Sprintf("[%s] azure NewSpeechSynthesizerFromConfig [%s] err: %+v", id.Hex(), subTaskFlag, err))
		return
	}
	defer speechSynthesizer.Close()

	complete := make(chan struct{})
	speechSynthesizer.SynthesisCompleted(func(event speech.SpeechSynthesisEventArgs) {
		defer event.Close()
		defer func() {
			close(complete)
		}()

		fileDownloadName := id.Hex() + subTaskFlag + ttsFileWavPostfix
		sf.SLogger().Info("准备上传文件到oss")
		url, err := new(services.Upload).UploadFile(ttsOssKey, fileDownloadName, event.Result.AudioData, ttsOssBucketName)
		if err != nil {
			sf.SLogger().Error(fmt.Sprintf("[%s] upload oss [%s] err: %+v", id.Hex(), subTaskFlag, err))
			return
		}
		sf.SLogger().Info("oss上传结束")

		voiceLengthField := "tts_url.male_voice_length"
		voiceUrlField := ttsMaleVoiceUrlField
		if subTaskFlag == femaleVoiceTaskFlag {
			voiceUrlField = ttsFemaleVoiceUrlField
			voiceLengthField = "tts_url.female_voice_length"
		}
		voiceLengthValue := new(controllers.Controller).TransitionFloat64(float64(event.Result.AudioDuration)/float64(time.Second), 2)
		_, err = sf.DB().Collection(ttsTextTable).Where(bson.M{ttsIDField: id}).Update(bson.M{voiceUrlField: url, voiceLengthField: voiceLengthValue})
		if err != nil {
			sf.SLogger().Error(fmt.Sprintf("[%s] update voice url field [%s] err: %+v", id.Hex(), subTaskFlag, err))
		} else {
			sf.SLogger().Info(fmt.Sprintf("[%s] azure tts success [%s] with url: %s", id.Hex(), subTaskFlag, url))
		}
	})
	speechSynthesizer.SynthesisCanceled(func(event speech.SpeechSynthesisEventArgs) {
		defer event.Close()
		close(complete)
		sf.SLogger().Info("azure tts Received a cancellation. text:", text)
	})

	text = global.CONFIG.TTS.HeadPrompt + text + global.CONFIG.TTS.TailPrompt
	task := speechSynthesizer.SpeakTextAsync(text)
	var outcome speech.SpeechSynthesisOutcome
	defer outcome.Close()
	select {
	case outcome = <-task:
	case <-time.After(azureApiTimeout * time.Second):
		sf.SLogger().Error(fmt.Sprintf("[%s] azure tts time out [%s]", id.Hex(), subTaskFlag))
		return
	}

	if outcome.Error != nil {
		sf.SLogger().Error(fmt.Sprintf("[%s] azure tts [%s] err: %+v text:%s", id.Hex(), subTaskFlag, outcome.Error, text))
		return
	}

	if outcome.Result.Reason != speechCommon.SynthesizingAudioCompleted {
		//outcomeStr, _ := ffmt.Puts(outcome)
		cancellation, _ := speech.NewCancellationDetailsFromSpeechSynthesisResult(outcome.Result)
		//sf.SLogger().Error(fmt.Sprintf("azure tts CANCELED: Reason=%d. ", cancellation.Reason), " outcome：", outcomeStr)

		if cancellation.Reason == speechCommon.Error {
			sf.SLogger().Error(fmt.Sprintf("azure tts CANCELED: ErrorCode=%d\nCANCELED: ErrorDetails=[%s]",
				cancellation.ErrorCode,
				cancellation.ErrorDetails))
		}
	}

	select {
	case <-complete:
	case <-time.After(ossTimeout * time.Second):
		sf.SLogger().Error(fmt.Sprintf("[%s] upload oss time out [%s]", id.Hex(), subTaskFlag))
	}
}
