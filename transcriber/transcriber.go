package transcriber

import (
	"fmt"
	"runtime"
	"strings"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// samples — 16kHz, mono, [-1..1] float32 (как у тебя в audioBuffer)
func Transcribe(samples []float32, modelPath string, lang string, translate bool) (string, string, error) {
	// 1) загружаем модель
	model, err := whisper.New(modelPath)
	if err != nil {
		return "", "", fmt.Errorf("load model: %w", err)
	}
	defer model.Close()

	// 2) создаём контекст
	ctx, err := model.NewContext()
	if err != nil {
		return "", "", fmt.Errorf("new context: %w", err)
	}

	// 3) настройки инференса
	// "auto" — автоопределение языка; можно "ru", "en" и т.п.
	if lang == "" {
		lang = "run"
	}
	if err := ctx.SetLanguage(lang); err != nil {
		return "", "", fmt.Errorf("set language: %w", err)
	}
	ctx.SetTranslate(translate) // true — принудительный перевод в английский
	ctx.SetThreads(uint(runtime.NumCPU()))
	// при необходимости: ctx.SetTemperature(0.0), ctx.SetBeamSize(5) и т.п. (опции см. доку)

	// 4) запуск распознавания
	var b strings.Builder
	segCB := func(s whisper.Segment) {
		// Сегменты приходят по мере готовности
		// s.Start, s.End — таймкоды (time.Duration), s.Text — текст
		b.WriteString(s.Text)
	}

	// можно ещё передать прогресс-колбэк (0..100) и begin-колбэк для остановки
	err = ctx.Process(samples, nil, segCB, nil)
	if err != nil {
		return "", "", fmt.Errorf("process: %w", err)
	}

	// 5) язык (если был auto)
	detected := ctx.DetectedLanguage() // может быть пустым, если задан фиксированный
	return b.String(), detected, nil
}
