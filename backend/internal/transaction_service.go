package internal

type TransactionService struct {
	LibraryService *LibraryService
}

func NewTransactionService(libService *LibraryService) *TransactionService {
	return &TransactionService{LibraryService: libService}
}
