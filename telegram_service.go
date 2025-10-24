package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramService struct {
	bot          *tgbotapi.BotAPI
	messageCh    chan TelegramMessage
	notifyChatID int64
}

func NewTelegramService(token string, notifyChatID int64) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramService{
		bot:          bot,
		messageCh:    make(chan TelegramMessage, 100),
		notifyChatID: notifyChatID,
	}, nil
}

func (s *TelegramService) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				s.handleCommand(update.Message)
			} else {
				s.handleMessage(update.Message)
			}
		}
	}()
}

func (s *TelegramService) Stop() {
	s.bot.StopReceivingUpdates()
	close(s.messageCh)
}

func (s *TelegramService) GetMessageChannel() chan TelegramMessage {
	return s.messageCh
}

func (s *TelegramService) SendNotification(text string) error {
	msg := tgbotapi.NewMessage(s.notifyChatID, text)
	_, err := s.bot.Send(msg)
	return err
}

func (s *TelegramService) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start":
		s.Reply(message.Chat.ID, "Bot is running")

	case "balance":
		s.messageCh <- TelegramMessage{
			ChatID:  message.Chat.ID,
			Text:    "balance",
			IsReply: true,
		}

	case "positions":
		s.messageCh <- TelegramMessage{
			ChatID:  message.Chat.ID,
			Text:    "positions",
			IsReply: true,
		}

	case "trades":
		limit := 10
		if args != "" {
			if l, err := strconv.Atoi(args); err == nil {
				limit = l
			}
		}
		s.messageCh <- TelegramMessage{
			ChatID:  message.Chat.ID,
			Text:    fmt.Sprintf("trades:%d", limit),
			IsReply: true,
		}

	case "status":
		s.messageCh <- TelegramMessage{
			ChatID:  message.Chat.ID,
			Text:    "status",
			IsReply: true,
		}
	}
}

func (s *TelegramService) handleMessage(message *tgbotapi.Message) {
	text := strings.TrimSpace(message.Text)
	if text == "" {
		return
	}

	s.messageCh <- TelegramMessage{
		ChatID:  message.Chat.ID,
		Text:    text,
		IsReply: false,
	}
}

func (s *TelegramService) Reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := s.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send reply: %v", err)
	}
}
