package internal

type GameService struct {
	LibraryService *LibraryService
}

func NewGameService(libService *LibraryService) *GameService {
	return &GameService{LibraryService: libService}
}
