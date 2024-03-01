package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TG_BOT_API_TOKEN")
	userID, err := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)
	if err != nil {
		userID = 0
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if userID != 0 && update.Message.From.ID != userID {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, this is a private bot.")
			msg.ReplyToMessageID = update.Message.MessageID
			_, err = bot.Send(msg)
			if err != nil {
				log.Println("error handling message:", err)
				continue
			}
		}
		if update.Message.Document != nil || update.Message.Voice != nil || update.Message.Audio != nil {
			err := handleFile(update, bot)
			if err != nil {
				log.Println("error handling message:", err)
				continue
			}
		}
	}
}

func handleFile(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	fileID := ""
	if update.Message.Document != nil {
		fileID = update.Message.Document.FileID
	}
	if update.Message.Voice != nil {
		fileID = update.Message.Voice.FileID
	}
	if update.Message.Audio != nil {
		fileID = update.Message.Audio.FileID
	}
	fileURL, err := bot.GetFileDirectURL(fileID)
	if err != nil {
		return err
	}
	log.Println("fileURL:", fileURL)

	response, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	voiceDir := "/app/data/voice/"
	err = os.MkdirAll(voiceDir, 0777)
	if err != nil {
		return err
	}

	fileName := filepath.Base(fileURL)
	fileName = strings.ReplaceAll(fileName, "%20", " ")
	savedFile := voiceDir + fileName
	outFile, err := os.Create(savedFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return err
	}

	python := "/app/venv/bin/python"
	args := []string{"transcribe.py", savedFile}
	out, err := exec.Command(python, args...).Output()
	if err != nil {
		return err
	}
	output := string(out)
	log.Println("Transcription:", output)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, output)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}
