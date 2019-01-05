package journal

type JournalFileService interface {
	Update(content string) error
	Content() string
}

type JournalProvider interface {
	Get(accessToken string) JournalFileService
}
