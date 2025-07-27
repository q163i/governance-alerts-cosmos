package types

import (
	"time"
)

// Proposal represents a governance proposal
type Proposal struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	VotingStart time.Time `json:"voting_start"`
	VotingEnd   time.Time `json:"voting_end"`
	Network     string    `json:"network"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Name         string `mapstructure:"name"`
	RestEndpoint string `mapstructure:"rest_endpoint"`
	ChainID      string `mapstructure:"chain_id"`
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	HoursBeforeStart     int  `mapstructure:"hours_before_start"`
	HoursBeforeEnd       int  `mapstructure:"hours_before_end"`
	CheckIntervalMinutes int  `mapstructure:"check_interval_minutes"`
	NotifyOnStartup      bool `mapstructure:"notify_on_startup"`
}

// NotificationConfig represents notification settings
type NotificationConfig struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Slack    SlackConfig    `mapstructure:"slack"`
}

// TelegramConfig represents Telegram notification settings
type TelegramConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BotToken string `mapstructure:"bot_token"`
	ChatID   int64  `mapstructure:"chat_id"`
}

// SlackConfig represents Slack notification settings
type SlackConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	WebhookURL string `mapstructure:"webhook_url"`
}

// LoggingConfig represents logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Config represents the main configuration structure
type Config struct {
	Alerts        AlertConfig              `mapstructure:"alerts"`
	Networks      map[string]NetworkConfig `mapstructure:"networks"`
	Notifications NotificationConfig       `mapstructure:"notifications"`
	Logging       LoggingConfig            `mapstructure:"logging"`
}

// NotificationMessage represents a notification message
type NotificationMessage struct {
	Title       string
	Content     string
	Network     string
	ChainID     string
	ProposalID  uint64
	ExplorerURL string
}
