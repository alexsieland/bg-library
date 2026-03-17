package internal

type GameService struct {
	libraryService *LibraryService
}

func NewGameService(libService *LibraryService) *GameService {
	return &GameService{libraryService: libService}
}
