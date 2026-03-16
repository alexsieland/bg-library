package internal

type PlayToWinService struct {
	LibraryService *LibraryService
}

func NewPlayToWinService(libService *LibraryService) *PlayToWinService {
	return &PlayToWinService{LibraryService: libService}
}
