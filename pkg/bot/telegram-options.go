package bot

// TelegramOptions contains options to run the telegram bot
type TelegramOptions struct {
	// Token for telegram
	Token string `yaml:"token"`
	// Offset where to start from with the messages
	Offset *int `yaml:"offset"`
	// Timeout for waiting for messages
	Timeout *int `yaml:"timeout"`
	// Whether to set debug mode in telegram
	Debug *bool `yaml:"debug,omitempty"`
	// Commands for the bot
	Commands []MessageCommand `yaml:"messages"`
}
