package transcriber

import (
	"fmt"
	"runtime"
	"strings"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

func Transcribe(samples []float32, modelPath string, lang string, translate bool) (string, string, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return "", "", fmt.Errorf("load model: %w", err)
	}
	defer model.Close()

	ctx, err := model.NewContext()
	if err != nil {
		return "", "", fmt.Errorf("new context: %w", err)
	}

	if lang == "" {
		lang = "run"
	}
	if err := ctx.SetLanguage(lang); err != nil {
		return "", "", fmt.Errorf("set language: %w", err)
	}
	ctx.SetTranslate(translate) 
	ctx.SetThreads(uint(runtime.NumCPU()))

	var b strings.Builder
	segCB := func(s whisper.Segment) {
		b.WriteString(s.Text)
	}

	err = ctx.Process(samples, nil, segCB, nil)
	if err != nil {
		return "", "", fmt.Errorf("process: %w", err)
	}

	detected := ctx.DetectedLanguage() 	
	return b.String(), detected, nil
}
