package main

import (
	"fmt"
	"time"

	"github.com/DilemaFixer/Arni/recorder"
)

func main() {
	pl, err := recorder.NewVoiceRecorder()
	if err != nil {
		panic(err)
	}
	defer pl.Close()

	fmt.Println("🎤 Начинаю запись...")
	pl.StartRecording()

	fmt.Println("Говорите что-нибудь... (5 секунд)")
	time.Sleep(5 * time.Second)

	fmt.Println("⏹️  Останавливаю запись...")
	pl.StopRecording()

	fmt.Println("🔊 Воспроизвожу запись...")
	time.Sleep(500 * time.Millisecond)

	err = pl.PlayRecording()
	if err != nil {
		fmt.Printf("Ошибка воспроизведения: %v\n", err)
	}

	fmt.Println("✅ Готово!")
}
