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

		// First, GetPlayToWinGroupByPlayToWinGameId is called (outside inner tx)
		ptwGameId := uuid.New()
		groupId := uuid.New()
		ptwGroup := db.VwPlayToWinGroup{ID: pgtype.UUID{Bytes: groupId, Valid: true}, Name: "G", CreatedAt: pgtype.Timestamp{Valid: true}}
		groupRow := new(MockRow)
		groupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = ptwGroup.ID
			*args.Get(1).(*string) = ptwGroup.Name
			*args.Get(2).(*pgtype.Timestamp) = ptwGroup.CreatedAt
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwGameId, Valid: true}}).Return(groupRow).Once()

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

		groupRow.AssertExpectations(t)
		sessionRow.AssertExpectations(t)
		entryRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

// Helpers for scanning used only in this file
// (small, focused scan helpers; longer-lived helpers live in db_mock_test.go)

// ptrInt32 is a helper to create *int32 pointers in tests
func ptrInt32(v int32) *int32 { return &v }
