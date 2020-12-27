package bot

func getRepliesMap(replies []MessageCommand) map[string]string {
	m := make(map[string]string, len(replies))

	for _, reply := range replies {
		m[reply.Command] = reply.Reply
	}

	return m
}
