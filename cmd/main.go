package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DilemaFixer/Arni/embed"
	"github.com/DilemaFixer/Arni/recorder"
	"github.com/DilemaFixer/Arni/transcriber"
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

	t, t2, err := transcriber.Transcribe(pl.AudioBuffer, "/Users/illashisko/Documents/Code/whisper.cpp/models/ggml-base.bin", "ru", false)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
	fmt.Println(t2)
	fmt.Println("✅ Готово!")

	ctx := context.Background()

	// создаём клиент (локальная модель bge-m3)
	cli := embed.New("bge-m3") // при необходимости: cli.BaseURL = "http://localhost:11434"
	cli.Normalize = true       // косинус = dot

	// получаем векторы для любых строк (слово/фраза — без разницы)
	vCreateRu, err := cli.Embed(ctx, "создать пост")
	if err != nil {
		log.Fatal(err)
	}

	vCreateEn, err := cli.Embed(ctx, t)
	if err != nil {
		log.Fatal(err)
	}

	vDelete, err := cli.Embed(ctx, "удалить запись")
	if err != nil {
		log.Fatal(err)
	}

	// сравнение (чем ближе к 1 — тем похожее)
	same := embed.Cosine(vCreateRu, vCreateEn)
	diff := embed.Cosine(vCreateRu, vDelete)

	fmt.Printf("create RU vs EN: %.3f\n", same)
	fmt.Printf("create vs delete: %.3f\n", diff)

	// простая проверка порогом
	const threshold = 0.75
	if same >= threshold {
		fmt.Println("→ это одна и та же команда (матч)")
	} else {
		fmt.Println("→ не похоже (ниже порога)")
	}
}
