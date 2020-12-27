package firestore

type chat struct {
	// ChatID is the ID of the chat
	ChatID int64 `firestore:"chat_id"`
	// Title of the chat
	Title string `firestore:"title"`
	// Type is the type of the chat
	Type string `firestore:"type"`
	// Username of the user
	Username string `firestore:"username"`
	// FirstName of the user
	FirstName string `firestore:"first_name"`
	// LastName of the user
	LastName string `firestore:"last_name"`
}
