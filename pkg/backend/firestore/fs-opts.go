package firestore

// Options contains options for firestore
type Options struct {
	// ProjectName is the name of the firebase project
	ProjectName string `yaml:"projectName"`
	// ChatsCollection is the name of the collections where
	// chats are stored
	ChatsCollection string `yaml:"chatsCollection"`
	// UseCache tells whether to cache chats locally
	UseCache bool `yaml:"useCache"`
}
