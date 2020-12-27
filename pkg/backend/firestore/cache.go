package firestore

import "github.com/SunSince90/kube-scraper-telegram-bot/pkg/backend"

func (f *fsBackend) getChatFromCache(id int64) *backend.Chat {
	f.lock.Lock()
	defer f.lock.Unlock()

	c, exists := f.cache[id]
	if exists && c != nil {
		return c
	}

	return nil
}

func (f *fsBackend) insertChatIntoCache(c *backend.Chat) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.cache[c.ChatID] = c
}

func (f *fsBackend) deleteChatFromCache(id int64) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.cache, id)
}

func (f *fsBackend) getAllChatsFromCache() []*backend.Chat {
	f.lock.Lock()
	defer f.lock.Unlock()

	if len(f.cache) == 0 {
		return []*backend.Chat{}
	}

	list := make([]*backend.Chat, len(f.cache))
	i := 0

	for _, c := range f.cache {
		list[i] = c
		i++
	}

	return list
}
