package backend

// Backend is a backend that can be used to store and retrieve
// chats.
type Backend interface {
	// GetChatByID retrieves a chat by the id
	GetChatByID(int64) (*Chat, error)
	// GetChatByUsername retrieves a chat by username
	GetChatByUsername(string) (*Chat, error)
	// GetAllChatIDs gets all chats
	GetAllChats() ([]*Chat, error)
	// StoreChat stores a new chat in the database
	StoreChat(*Chat) error
	// DeleteChat deletes a chat from the database
	DeleteChat(int64) error
	// Close any client
	Close()
}
