package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock of the DB interface
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Close() {
	m.Called()
}

func (m *MockDatabase) Exec(ctx context.Context, s string, i ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, s, i)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockDatabase) Query(ctx context.Context, s string, i ...any) (pgx.Rows, error) {
	args := m.Called(ctx, s, i)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Rows), args.Error(1)
}

func (m *MockDatabase) QueryRow(ctx context.Context, s string, i ...any) pgx.Row {
	args := m.Called(ctx, s, i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(pgx.Row)
}

func (m *MockDatabase) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	args := m.Called(ctx, txOptions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// MockTx is a mock of the pgx.Tx interface
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	// Debug: record that QueryRow was called (helps diagnose test expectation mismatches)
	// Note: tests rely on mock expectations; this log is only for debugging in CI/local runs.
	// fmt.Printf("MockTx.QueryRow called: sql=%s, args=%v\n", sql, args)
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil
	}
	return callArgs.Get(0).(pgx.Row)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*pgx.Conn)
}

// MockRow is a mock of the pgx.Row interface
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

// MockRows is a mock of the pgx.Rows interface
type MockRows struct {
	mock.Mock
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *MockRows) Values() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

func (m *MockRows) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

func MockRowScanError(row *MockRow, argCount int, err error) {
	scanArgs := make([]any, argCount)
	for i := range scanArgs {
		scanArgs[i] = mock.Anything
	}
	row.On("Scan", scanArgs...).Return(err)
}

func MockPatronScan(row *MockRow, patron db.Patron, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = patron.ID
		*args.Get(1).(*string) = patron.FullName
		*args.Get(2).(*pgtype.Timestamp) = patron.CreatedAt
		*args.Get(3).(*pgtype.Timestamp) = patron.DeletedAt
		*args.Get(4).(*pgtype.Text) = patron.Barcode
	}).Return(err)
}

func MockVwLibraryPatronScan(row *MockRow, patron db.VwLibraryPatron, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = patron.ID
		*args.Get(1).(*string) = patron.FullName
		*args.Get(2).(*pgtype.Text) = patron.Barcode
		*args.Get(3).(*pgtype.Timestamp) = patron.CreatedAt
	}).Return(err)
}

func MockVwLibraryPatronRows(rows *MockRows, patrons []db.VwLibraryPatron, err error) {
	for range patrons {
		rows.On("Next").Return(true).Once()
	}
	rows.On("Next").Return(false).Once()

	if len(patrons) > 0 {
		scanIndex := 0
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			patron := patrons[scanIndex]
			scanIndex++

			*args.Get(0).(*pgtype.UUID) = patron.ID
			*args.Get(1).(*string) = patron.FullName
			*args.Get(2).(*pgtype.Text) = patron.Barcode
			*args.Get(3).(*pgtype.Timestamp) = patron.CreatedAt
		}).Return(nil).Times(len(patrons))
	}

	rows.On("Close").Return().Once()
	rows.On("Err").Return(err).Once()
}

// ---- Shared game builders and mocks (used by multiple service tests) ----

func makeLibraryGame(id uuid.UUID, title string, barcode *string) db.VwLibraryGame {
	barcodePg := pgtype.Text{Valid: false}
	if barcode != nil {
		barcodePg = pgtype.Text{String: *barcode, Valid: true}
	}
	return db.VwLibraryGame{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		DisplayTitle:    title,
		Title:           title,
		SanitizedTitle:  SanitizeTitle(title),
		Barcode:         barcodePg,
		PlayToWinGameID: pgtype.UUID{Valid: false},
		CreatedAt:       pgtype.Timestamp{Valid: true},
	}
}

func MockVwLibraryGameScan(row *MockRow, g db.VwLibraryGame, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = g.ID
		*args.Get(1).(*string) = g.DisplayTitle
		*args.Get(2).(*string) = g.Title
		*args.Get(3).(*string) = g.SanitizedTitle
		*args.Get(4).(*pgtype.Text) = g.Barcode
		*args.Get(5).(*pgtype.UUID) = g.PlayToWinGameID
		*args.Get(6).(*pgtype.Timestamp) = g.CreatedAt
	}).Return(err)
}

func MockVwLibraryGameRows(rows *MockRows, items []db.VwLibraryGame, err error) {
	for range items {
		rows.On("Next").Return(true).Once()
	}
	rows.On("Next").Return(false).Once()

	if len(items) > 0 {
		idx := 0
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			item := items[idx]
			idx++
			*args.Get(0).(*pgtype.UUID) = item.ID
			*args.Get(1).(*string) = item.DisplayTitle
			*args.Get(2).(*string) = item.Title
			*args.Get(3).(*string) = item.SanitizedTitle
			*args.Get(4).(*pgtype.Text) = item.Barcode
			*args.Get(5).(*pgtype.UUID) = item.PlayToWinGameID
			*args.Get(6).(*pgtype.Timestamp) = item.CreatedAt
		}).Return(nil).Times(len(items))
	}

	rows.On("Close").Return().Once()
	rows.On("Err").Return(err).Once()
}

func makeGame(id uuid.UUID, title string, barcode *string) db.Game {
	barcodePg := pgtype.Text{Valid: false}
	if barcode != nil {
		barcodePg = pgtype.Text{String: *barcode, Valid: true}
	}
	return db.Game{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		Title:          title,
		DisplayTitle:   pgtype.Text{String: title, Valid: true},
		SanitizedTitle: SanitizeTitle(title),
		CreatedAt:      pgtype.Timestamp{Valid: true},
		DeletedAt:      pgtype.Timestamp{Valid: false},
		Barcode:        barcodePg,
	}
}

func MockGameScan(row *MockRow, g db.Game, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = g.ID
		*args.Get(1).(*string) = g.Title
		*args.Get(2).(*pgtype.Text) = g.DisplayTitle
		*args.Get(3).(*string) = g.SanitizedTitle
		*args.Get(4).(*pgtype.Timestamp) = g.CreatedAt
		*args.Get(5).(*pgtype.Timestamp) = g.DeletedAt
		*args.Get(6).(*pgtype.Text) = g.Barcode
	}).Return(err)
}

// MockPlayToWinService is a mock for the PlayToWinService interface used in
// unit tests where GameService or other services depend on PTW operations.
// Tests can set expectations on methods like InsertPlayToWinGame and
// DeletePlayToWinGameByLibraryGameId.

type MockPlayToWinService struct {
	mock.Mock
}

func (m *MockPlayToWinService) InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}

func (m *MockPlayToWinService) DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, deletionReason, deletionReasonComment, optTx)
	return args.Error(0)
}

func (m *MockPlayToWinService) GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}
