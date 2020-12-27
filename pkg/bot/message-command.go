package bot

// MessageCommand contains replies to certain commands
type MessageCommand struct {
	// Command to reply to, i.e. /start
	Command string `yaml:"command"`
	// Reply is the reply that will be sent when user sends that command.
	Reply string `yaml:"reply"`
}
