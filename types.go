package main

// TelegramChat is a representation of a telegram chat
type TelegramChat struct {
	// ChatID is the ID of the chat
	ChatID int64 `firestore:"ChatID"`
	// Username of the user
	Username string `firestore:"Username"`
	// FirstName of the user
	FirstName string `firestore:"FirstName"`
	// LastName of the user
	LastName string `firestore:"LastName"`
}
