package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"governance-alerts-cosmos/internal/types"

	"gopkg.in/telebot.v3"
)

// Notifier handles sending notifications to various channels
type Notifier struct {
	telegram       *telebot.Bot
	telegramChatID int64
	slack          types.SlackConfig
}

// NewNotifier creates a new notifier instance
func NewNotifier(config *types.NotificationConfig) (*Notifier, error) {
	notifier := &Notifier{}

	// Initialize Telegram if enabled
	if config.Telegram.Enabled {
		bot, err := telebot.NewBot(telebot.Settings{
			Token:  config.Telegram.BotToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
		}
		notifier.telegram = bot
		notifier.telegramChatID = config.Telegram.ChatID
	}

	// Store Slack config
	notifier.slack = config.Slack

	return notifier, nil
}

// SendNotification sends a notification to all enabled channels
func (n *Notifier) SendNotification(msg types.NotificationMessage) error {
	var errors []error

	// Send to Telegram if enabled
	if n.telegram != nil {
		if err := n.sendTelegramNotification(msg); err != nil {
			errors = append(errors, fmt.Errorf("telegram: %w", err))
		}
	}

	// Send to Slack if enabled
	if n.slack.Enabled {
		if err := n.sendSlackNotification(msg); err != nil {
			errors = append(errors, fmt.Errorf("slack: %w", err))
		}
	}

	// Return first error if any
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// sendTelegramNotification sends a notification to Telegram
func (n *Notifier) sendTelegramNotification(msg types.NotificationMessage) error {
	formattedMsg := formatTelegramMessage(msg)

	// Use the configured chat ID
	chat := &telebot.Chat{ID: n.telegramChatID}

	_, err := n.telegram.Send(chat, formattedMsg, &telebot.SendOptions{
		ParseMode: telebot.ModeHTML,
	})

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// sendSlackNotification sends a notification to Slack
func (n *Notifier) sendSlackNotification(msg types.NotificationMessage) error {
	payload := map[string]interface{}{
		"text": formatSlackMessage(msg),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(n.slack.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// formatTelegramMessage formats a message for Telegram
func formatTelegramMessage(msg types.NotificationMessage) string {
	// For startup notifications, don't include Network, Chain ID, and Proposal ID
	if msg.Network == "Governance Alerts" {
		return fmt.Sprintf(
			"🚀 <b>%s</b>\n\n%s",
			msg.Title,
			msg.Content,
		)
	}

	// For proposal notifications, include all details
	return fmt.Sprintf(
		"🚨 <b>%s</b>\n\n"+
			"<b>Network:</b> %s\n"+
			"<b>Chain ID:</b> %s\n"+
			"<b>Proposal ID:</b> %d\n\n"+
			"%s",
		msg.Title,
		msg.Network,
		msg.ChainID,
		msg.ProposalID,
		msg.Content,
	)
}

// formatSlackMessage formats a message for Slack
func formatSlackMessage(msg types.NotificationMessage) string {
	// For startup notifications, don't include Network, Chain ID, and Proposal ID
	if msg.Network == "Governance Alerts" {
		return fmt.Sprintf(
			"🚀 *%s*\n\n%s",
			msg.Title,
			msg.Content,
		)
	}

	// For proposal notifications, include all details
	return fmt.Sprintf(
		"🚨 *%s*\n\n"+
			"*Network:* %s\n"+
			"*Chain ID:* %s\n"+
			"*Proposal ID:* %d\n\n"+
			"%s",
		msg.Title,
		msg.Network,
		msg.ChainID,
		msg.ProposalID,
		msg.Content,
	)
}
