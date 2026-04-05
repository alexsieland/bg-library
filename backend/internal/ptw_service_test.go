package internal

import (
	"context"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// InsertPlayToWinGame Tests
// ============================================================================

// InsertPlayToWinGame — happy-path: creates cpTx and commits checkpoint
func TestPlayToWinService_InsertPlayToWinGame_Success(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	// wire a real GameService to satisfy GetGame call (it will hit mockDB)
	gs := GameService{libraryService: lib}
	svc.SetGameService(&gs)

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	// Prepare library game returned by GetGame
	gameID := uuid.New()
	expectedGame := makeLibraryGame(gameID, "Catan", nil)
	gameRow := new(MockRow)
	MockVwLibraryGameScan(gameRow, expectedGame, nil)
	mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{expectedGame.ID}).Return(gameRow).Once()

	// Prepare checkpoint txs: one for CreatePlayToWinGroup (sp1) and one for CreatePlayToWinGame (sp2)
	sp1 := new(MockTx)
	sp2 := new(MockTx)
	ptwID := uuid.New()
	ptwRow := new(MockRow)
	// CreatePlayToWinGame returns: id, game_id, ptw_group_id, winner_id, created_at, deleted_at, deletion_reason, deletion_reason_comment
	ptwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(6).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
			*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)

	// First Begin() is for CreatePlayToWinGroup (sp1)
	// CreatePlayToWinGroup returns: id, name, created_at, deleted_at
	groupID := uuid.New()
	createGroupRow := new(MockRow)
	createGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
		*args.Get(1).(*string) = "Catan"
		*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
	}).Return(nil)
	mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
	sp1.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(createGroupRow).Once()
	sp1.On("Commit", mock.Anything).Return(nil).Once()

	// Second Begin() is for CreatePlayToWinGame (sp2)
	mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
	sp2.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(ptwRow).Once()
	sp2.On("Commit", mock.Anything).Return(nil).Once()
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	got, err := svc.InsertPlayToWinGame(ctx, expectedGame.ID, nil)
	assert.NoError(t, err)
	assert.Equal(t, pgtype.UUID{Bytes: ptwID, Valid: true}, got.ID)

	gameRow.AssertExpectations(t)
	ptwRow.AssertExpectations(t)
	sp1.AssertExpectations(t)
	sp2.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// InsertPlayToWinGame — when CreatePlayToWinGame returns unique violation, recover via restore
func TestPlayToWinService_InsertPlayToWinGame_UniqueRecovery(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	gs := GameService{libraryService: lib}
	svc.SetGameService(&gs)

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	gameID := uuid.New()
	expectedGame := makeLibraryGame(gameID, "Catan", nil)
	gameRow := new(MockRow)
	MockVwLibraryGameScan(gameRow, expectedGame, nil)
	mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{expectedGame.ID}).Return(gameRow).Once()

	// Simulate CreatePlayToWinGroup + CreatePlayToWinGame checkpoint txs
	sp1 := new(MockTx)
	sp2 := new(MockTx)
	cpErr := &pgconn.PgError{Code: "23505"}
	rowErr := new(MockRow)
	// CreatePlayToWinGame scans 8 fields: id, game_id, ptw_group_id, winner_id, created_at, deleted_at, deletion_reason, deletion_reason_comment
	MockRowScanError(rowErr, 8, cpErr)

	// Begin for CreatePlayToWinGroup (sp1) -> succeed
	groupID := uuid.New()
	createGroupRow := new(MockRow)
	createGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
		*args.Get(1).(*string) = "Catan"
		*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
	}).Return(nil)
	mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
	sp1.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(createGroupRow).Once()
	sp1.On("Commit", mock.Anything).Return(nil).Once()

	// Begin for CreatePlayToWinGame (sp2) -> fail with unique violation
	mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
	sp2.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(rowErr).Once()
	sp2.On("Rollback", mock.Anything).Return(nil).Once()

	// After unique violation, implementation calls RestorePlayToWinGameByLibraryGameId (exec) and then queries for the restored game
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()

	getRow := new(MockRow)
	ptwID := uuid.New()
	// GetPlayToWinGameByLibraryGame returns: id, game_id, ptw_group_id, group_name, created_at, winner_id (6 fields from VwPlayToWinGame)
	getRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(3).(*pgtype.Text) = pgtype.Text{String: "Catan", Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
		}).Return(nil)
	mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(getRow).Once()
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	got, err := svc.InsertPlayToWinGame(ctx, expectedGame.ID, nil)
	assert.NoError(t, err)
	assert.Equal(t, pgtype.UUID{Bytes: ptwID, Valid: true}, got.ID)

	gameRow.AssertExpectations(t)
	sp1.AssertExpectations(t)
	sp2.AssertExpectations(t)
	getRow.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// ============================================================================
// InsertPlayToWinGroup Tests
// ============================================================================

func TestPlayToWinServiceInsertPlayToWinGroup(t *testing.T) {
	t.Run("Should create play to win group when not exists", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := PlayToWinService{libraryService: lib}

		ctx := context.Background()
		mockTx := MockWithinTx(t)

		// Expect outer tx to begin a checkpoint tx
		cpTx := new(MockTx)

		// When checkpoint CreatePlayToWinGroup runs it will call cpTx.QueryRow
		newID := uuid.New()
		created := db.PlayToWinGroup{ID: pgtype.UUID{Bytes: newID, Valid: true}, Name: "Group", CreatedAt: pgtype.Timestamp{Valid: true}}
		row := new(MockRow)
		// CreatePlayToWinGroup returns the new group row (id, name, created_at, deleted_at)
		row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = created.ID
			*args.Get(1).(*string) = created.Name
			*args.Get(2).(*pgtype.Timestamp) = created.CreatedAt
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
		}).Return(nil)

		// Setup expectations
		mockTx.On("Begin", ctx).Return(cpTx, nil).Once()
		cpTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(row).Once()
		cpTx.On("Commit", ctx).Return(nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		// Call InsertPlayToWinGroup (it uses WithinTx, so will use our MockWithinTx)
		got, err := svc.InsertPlayToWinGroup(ctx, "Group", nil)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)

		row.AssertExpectations(t)
		cpTx.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return existing group when create unique constraint violated", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := PlayToWinService{libraryService: lib}

		ctx := context.Background()
		mockTx := MockWithinTx(t)
		cpTx := new(MockTx)

		// Simulate CreatePlayToWinGroup failing with unique constraint
		pgErr := &pgconn.PgError{Code: "23505"}
		rowErr := new(MockRow)
		MockRowScanError(rowErr, 4, pgErr)

		// When CreatePlayToWinGroup runs in checkpoint tx
		mockTx.On("Begin", ctx).Return(cpTx, nil).Once()
		cpTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(rowErr).Once()
		cpTx.On("Rollback", ctx).Return(nil).Once()

		// Then the implementation should call GetPlayToWinGroupByName on the outer tx
		existingID := uuid.New()
		existing := db.VwPlayToWinGroup{ID: pgtype.UUID{Bytes: existingID, Valid: true}, Name: "Group", CreatedAt: pgtype.Timestamp{Valid: true}}
		getRow := new(MockRow)
		getRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = existing.ID
			*args.Get(1).(*string) = existing.Name
			*args.Get(2).(*pgtype.Timestamp) = existing.CreatedAt
		}).Return(nil)

		// Outer tx QueryRow should be called for GetPlayToWinGroupByName (on the outer tx)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(getRow).Once()

		// Outer tx commit expected
		mockTx.On("Commit", ctx).Return(nil).Once()

		got, err := svc.InsertPlayToWinGroup(ctx, "Group", nil)
		assert.NoError(t, err)
		assert.Equal(t, existing.ID, got.ID)

		rowErr.AssertExpectations(t)
		getRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

func TestPlayToWinServiceRecordPlayToWinSession(t *testing.T) {
	t.Run("Should record session and entries", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := PlayToWinService{libraryService: lib}

		ctx := context.Background()
		mockTx := MockWithinTx(t)

		// Setup play-to-win group and IDs used for session/entries
		groupId := uuid.New()
		ptwGroup := db.VwPlayToWinGroup{ID: pgtype.UUID{Bytes: groupId, Valid: true}, Name: "G", CreatedAt: pgtype.Timestamp{Valid: true}}

		// Now, within the inner tx: CreatePlayToWinSession then CreatePlayToWinEntry for each entry
		sessionId := uuid.New()
		sessionRow := new(MockRow)
		session := db.PlayToWinSession{ID: pgtype.UUID{Bytes: sessionId, Valid: true}, PtwGroupID: ptwGroup.ID, PlaytimeMinutes: pgtype.Int4{Int32: 30, Valid: true}}
		sessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = session.ID
			*args.Get(1).(*pgtype.UUID) = session.PtwGroupID
			*args.Get(2).(*pgtype.Int4) = session.PlaytimeMinutes
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{}
			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)

		// entries
		entryId := uuid.New()
		entryRow := new(MockRow)
		entry := db.PlayToWinEntry{ID: pgtype.UUID{Bytes: entryId, Valid: true}, PtwSessionID: session.ID, PtwGroupID: ptwGroup.ID, EntrantName: "Alice", EntrantUniqueID: "A1", CreatedAt: pgtype.Timestamp{Valid: true}}
		entryRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = entry.ID
			*args.Get(1).(*pgtype.UUID) = entry.PtwSessionID
			*args.Get(2).(*pgtype.UUID) = entry.PtwGroupID
			*args.Get(3).(*string) = entry.EntrantName
			*args.Get(4).(*string) = entry.EntrantUniqueID
			*args.Get(5).(*pgtype.Timestamp) = entry.CreatedAt
			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(7).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{}
			*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)

		// The inner tx will call CreatePlayToWinSession then CreatePlayToWinEntry
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(sessionRow).Once()
		// Next call for CreatePlayToWinEntry
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(entryRow).Once()

		mockTx.On("Commit", ctx).Return(nil).Twice()

		// Call the service methods to exercise the logic
		session, err := svc.InsertPlayToWinSession(ctx, ptwGroup.ID, ptrInt32(30), nil)
		assert.NoError(t, err)

		_, err = svc.InsertPlayToWinEntry(ctx, session.ID, ptwGroup.ID, "Alice", "A1", nil)
		assert.NoError(t, err)

		sessionRow.AssertExpectations(t)
		entryRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

// ============================================================================
// Additional Tests (Delete, Claim, Update, Reset)
// ============================================================================

func TestPlayToWinService_DeletePlayToWinGameByLibraryGameId_Success(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	gameID := uuid.New()

	// Expect outer tx Exec for delete
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	reason := db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake, Valid: true}

	err := svc.DeletePlayToWinGameByLibraryGameId(ctx, pgtype.UUID{Bytes: gameID, Valid: true}, reason, nil, nil)
	assert.NoError(t, err)

	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestPlayToWinService_ClaimPlayToWinGame_NoWinner(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	// Set a GameService even though it won't be used in this test
	gs := GameService{libraryService: lib}
	svc.SetGameService(&gs)

	ctx := context.Background()
	// Prepare a ptw overview row with no winner
	ptwGameID := uuid.New()
	row := new(MockRow)
	// GetPlayToWinGameOverview returns 9 columns: ptw_game_id, game_id, ptw_group_id, game_title, sanitized_title, created_at, winner_id, winner_name, winner_unique_id
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameID, Valid: true}
		*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Valid: false}
		*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false}
		*args.Get(3).(*string) = ""
		*args.Get(4).(*string) = ""
		*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		*args.Get(6).(*pgtype.UUID) = pgtype.UUID{Valid: false}
		*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
		*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
	}).Return(nil)

	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(row).Once()

	err := svc.ClaimPlayToWinGame(ctx, pgtype.UUID{Bytes: ptwGameID, Valid: true}, nil)
	assert.ErrorIs(t, err, ErrClaimUnwonPtwGame)

	row.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestPlayToWinService_UpdatePlayToWinGameWinner(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	ptwGameId := uuid.New()
	entryId := uuid.New()

	// UpdatePlayToWinWinner uses Exec under the tx
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	err := svc.UpdatePlayToWinGameWinner(ctx, pgtype.UUID{Bytes: ptwGameId, Valid: true}, pgtype.UUID{Bytes: entryId, Valid: true}, nil)
	assert.NoError(t, err)

	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestPlayToWinService_ResetPlayToWinGameWinners(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	// ResetPlayToWinGameWinners executes a single Exec
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Once()
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	err := svc.ResetPlayToWinGameWinners(ctx, nil)
	assert.NoError(t, err)

	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestPlayToWinService_ClaimPlayToWinGame_Success(t *testing.T) {
	lib, mockDB := setupTestLibraryService()
	svc := PlayToWinService{libraryService: lib}

	// wire a real GameService
	gs := GameService{libraryService: lib}
	svc.SetGameService(&gs)

	ctx := context.Background()
	mockTx := MockWithinTx(t)

	// Prepare overview row with a winner
	ptwGameID := uuid.New()
	gameID := uuid.New()
	winnerID := uuid.New()
	row := new(MockRow)
	// GetPlayToWinGameOverview returns 9 columns: ptw_game_id, game_id, ptw_group_id, game_title, sanitized_title, created_at, winner_id, winner_name, winner_unique_id
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameID, Valid: true}
		*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
		*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false}
		*args.Get(3).(*string) = "Game"
		*args.Get(4).(*string) = "game"
		*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		*args.Get(6).(*pgtype.UUID) = pgtype.UUID{Bytes: winnerID, Valid: true}
		*args.Get(7).(*pgtype.Text) = pgtype.Text{String: "Winner", Valid: true}
		*args.Get(8).(*pgtype.Text) = pgtype.Text{String: "unique-1", Valid: true}
	}).Return(nil)

	mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(row).Once()

	// Expect sequence of Exec calls: delete ptw game, delete entry, delete library game
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil).Times(3)
	mockTx.On("Commit", mock.Anything).Return(nil).Once()

	err := svc.ClaimPlayToWinGame(ctx, pgtype.UUID{Bytes: ptwGameID, Valid: true}, nil)
	assert.NoError(t, err)

	row.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// ============================================================================
// Helpers
// ============================================================================

// ptrInt32 is a helper to create *int32 pointers in tests
func ptrInt32(v int32) *int32 { return &v }
