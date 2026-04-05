package api

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"io"
	"log"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

type gameService interface {
	ListGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error)
	ListGameStatuses(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error)
	ListCheckedOutGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error)
	InsertGame(ctx context.Context, title string, barcode *string, isPlayToWin bool, optTx pgx.Tx) (db.VwLibraryGame, error)
	UpdateGame(ctx context.Context, gameId pgtype.UUID, title string, barcode *string, optTx pgx.Tx) error
	GetGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGame, error)
	GetGamesByBarcode(ctx context.Context, barcode string, optTx pgx.Tx) ([]db.VwLibraryGame, error)
	GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwGameStatus, error)
	DeleteGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) error
	SetIsPlayToWin(ctx context.Context, gameId pgtype.UUID, isPlayToWin bool, optTx pgx.Tx) error
}

type GameApi struct {
	libraryService internal.LibraryServiceInterface
	service        gameService
}

func NewGamesApi(libService internal.LibraryServiceInterface, gameSrv *internal.GameService) *GameApi {
	return &GameApi{
		libraryService: libService,
		service:        gameSrv,
	}
}

func (api *GameApi) AddGame(ctx context.Context, request CreateGameRequest) (Game, error) {
	var errorDetails ErrorDetails
	isPlayToWin := false
	if request.IsPlayToWin != nil {
		isPlayToWin = *request.IsPlayToWin
	}

	errorDetails.ValidateStringLength("title", request.Title, 1, 100)
	if request.Barcode != nil {
		errorDetails.ValidateStringLength("barcode", *request.Barcode, 1, 48)
	}

	if !errorDetails.Empty() {
		return Game{}, errorDetails
	}

	dbGame, err := api.service.InsertGame(ctx, request.Title, request.Barcode, isPlayToWin, nil)
	if err != nil {
		log.Printf("Error creating game: %v", err)
		return Game{}, err
	}

	return FromVwLibraryGame(dbGame), nil
}

func (api *GameApi) BulkAddGames(ctx context.Context, requestBody io.ReadCloser) (BulkAddResponse, error) {
	decodedReader := base64.NewDecoder(base64.StdEncoding, requestBody)
	csvReader := csv.NewReader(decodedReader)

	// Start a db transaction
	tx, err := api.libraryService.BeginTx(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return BulkAddResponse{}, err
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Process each row
	var errorDetails ErrorDetails
	recordCount := int32(0)
	firstRow := true
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV: %v", err)
			return BulkAddResponse{}, err
		}
		if firstRow {
			firstRow = false
			continue
		}
		if len(record) == 0 {
			continue
		}

		// Validate each row
		title := record[0]
		errorDetails.ValidateStringLength("title", title, 1, 100)

		var barcode *string
		if len(record) > 1 && record[1] != "" {
			barcode = &record[1]
			errorDetails.ValidateStringLength("barcode", *barcode, 1, 48)
		}

		// If there have been any validation errors, check remaining records for validation errors but skip database inserts
		if !errorDetails.Empty() {
			continue
		}

		isPlayToWin := len(record) > 2 && record[2] == "true"

		// If there are no validation errors, attempt to insert patron into database
		_, err = api.service.InsertGame(ctx, title, barcode, isPlayToWin, tx)
		if err != nil {
			log.Printf("Error adding games: %v", err)
			return BulkAddResponse{}, err
		}
		recordCount++
	}

	if !errorDetails.Empty() {
		return BulkAddResponse{}, errorDetails
	}

	//If there are no validation errors, commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return BulkAddResponse{}, err
	}

	tx = nil // Prevent deferred rollback after a successful commit
	return BulkAddResponse{Imported: recordCount}, nil
}

func (api *GameApi) DeleteGame(ctx context.Context, gameId types.UUID) error {
	err := api.service.DeleteGame(ctx, uuidToPgTypeUUID(gameId), nil)
	if err != nil {
		log.Printf("Error deleting game: %v", err)
	}
	return err
}

func (api *GameApi) GetGame(ctx context.Context, gameId types.UUID) (Game, error) {
	dbGame, err := api.service.GetGame(ctx, uuidToPgTypeUUID(gameId), nil)
	if err != nil {
		log.Printf("Error getting game: %v", err)
		return Game{}, err
	}
	return FromVwLibraryGame(dbGame), nil
}

func (api *GameApi) GetGameByBarcode(ctx context.Context, gameBarcode string) (GameList, error) {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("gameBarcode", gameBarcode, 1, 48)
	if !errorDetails.Empty() {
		return GameList{}, errorDetails
	}
	dbGames, err := api.service.GetGamesByBarcode(ctx, gameBarcode, nil)
	if err != nil {
		log.Printf("Error getting game: %v", err)
		return GameList{}, err
	}
	return FromVwLibraryGames(dbGames), nil
}

func (api *GameApi) UpdateGame(ctx context.Context, gameId uuid.UUID, request CreateGameRequest) error {
	// validate field values
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("title", request.Title, 1, 100)
	if request.Barcode != nil {
		errorDetails.ValidateStringLength("barcode", *request.Barcode, 1, 48)
	}
	if !errorDetails.Empty() {
		return errorDetails
	}

	// Start a db transaction
	tx, err := api.libraryService.BeginTx(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return err
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = api.service.UpdateGame(ctx, uuidToPgTypeUUID(gameId), request.Title, request.Barcode, tx)
	if err != nil {
		log.Printf("Error updating game: %v", err)
		return err
	}
	isPlayToWin := request.IsPlayToWin != nil && *request.IsPlayToWin
	err = api.service.SetIsPlayToWin(ctx, uuidToPgTypeUUID(gameId), isPlayToWin, tx)
	if err != nil {
		log.Printf("Error setting game play to win state to %v: %v", isPlayToWin, err)
		return err
	}

	// commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	// prevent deferred rollback after a successful commit
	tx = nil
	return nil
}

func (api *GameApi) ListGames(ctx context.Context, params ListGamesParams) (GameStatusList, error) {
	var (
		dbGameStatusList []db.VwGameStatus
		limit            int32 = 100
		offset           int32 = 0
		errorDetails     ErrorDetails
		err              error
	)

	if params.Title != nil {
		errorDetails.ValidateStringLength("title", *params.Title, 1, 100)
	}
	if !errorDetails.Empty() {
		return GameStatusList{}, errorDetails
	}

	if params.CheckedOut != nil && *params.CheckedOut {
		dbGameStatusList, err = api.service.ListCheckedOutGames(ctx, params.Title, limit, offset, nil)
	} else {
		dbGameStatusList, err = api.service.ListGameStatuses(ctx, params.Title, limit, offset, nil)
	}
	if err != nil {
		log.Printf("Error listing games: %v", err)
		return GameStatusList{}, err
	}

	gameStatusList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameStatusList[i] = FromVwGameStatus(dbGameStatus)
	}

	return GameStatusList{Games: gameStatusList}, nil
}
