package api

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func setupTestServer() (Server, *MockDatabase) {
	gin.SetMode(gin.TestMode)
	mockDB := new(MockDatabase)
	server := Server{
		Database: mockDB,
		queries:  db.New(mockDB),
	}
	return server, mockDB
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
