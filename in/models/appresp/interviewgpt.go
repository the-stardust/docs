package appresp

type TTSPromptResp struct {
	MaleHeadPrompt   string  `json:"male_head_prompt" bson:"male_head_prompt" redis:"male_head_prompt"`
	MaleTailPrompt   string  `json:"male_tail_prompt" bson:"male_tail_prompt" redis:"male_tail_prompt"`
	FeMaleHeadPrompt string  `json:"female_head_prompt" bson:"female_head_prompt" redis:"female_head_prompt"`
	FeMaleTailPrompt string  `json:"female_tail_prompt" bson:"female_tail_prompt" redis:"female_tail_prompt"`
	PromptType       int8    `json:"prompt_type" bson:"prompt_type" redis:"prompt_type"`
	HeadVoiceLength  float64 `json:"head_voice_length" bson:"head_voice_length" redis:"head_voice_length"`
	TailVoiceLength  float64 `json:"tail_voice_length" bson:"tail_voice_length" redis:"tail_voice_length"`
}
