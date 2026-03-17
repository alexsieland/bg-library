package api

import "testing"

// --- AddPlayToWinGameByGameId ---

func TestAddPlayToWinGame(t *testing.T) {
	//	t.Run("Should return 204 No Content when game is successfully marked as Play to Win", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//		ptwID := uuid.New()
	//
	//		mockTx := setupAddPlayToWinMocksForSuccess(mockDB, gameID, ptwID)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		server.AddPlayToWinGameByGameId(c, gameID)
	//
	//		assert.Equal(t, http.StatusNoContent, w.Code)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
	//
	//	t.Run("Should return 404 Not Found when game does not exist (foreign key violation)", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//
	//		mockTx := new(MockTx)
	//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
	//
	//		// Fail at GetGame with FK violation
	//		mockRow := new(MockRow)
	//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
	//			Return(&pgconn.PgError{Code: "23503"})
	//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
	//		mockTx.On("Rollback", mock.Anything).Return(nil)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		server.AddPlayToWinGameByGameId(c, gameID)
	//
	//		assert.Equal(t, http.StatusNotFound, w.Code)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
	//
	//	t.Run("Should return 204 No Content when game is already marked as Play to Win (idempotent)", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//
	//		mockTx := setupAddPlayToWinMocksForIdempotent(mockDB, gameID)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		server.AddPlayToWinGameByGameId(c, gameID)
	//
	//		assert.Equal(t, http.StatusNoContent, w.Code)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
	//
	//	t.Run("Should return 500 Internal Server Error when DB returns an unexpected error", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//
	//		mockTx := new(MockTx)
	//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
	//
	//		// Fail at GetGame with unexpected error
	//		mockRow := new(MockRow)
	//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
	//			Return(errors.New("unexpected db error"))
	//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
	//		mockTx.On("Rollback", mock.Anything).Return(nil)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		server.AddPlayToWinGameByGameId(c, gameID)
	//
	//		assert.Equal(t, http.StatusInternalServerError, w.Code)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
}

func TestAddPlayToWin(t *testing.T) {
	//	t.Run("Should return nil when game is successfully marked as Play to Win", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//		ptwID := uuid.New()
	//
	//		mockTx := setupAddPlayToWinMocksForSuccess(mockDB, gameID, ptwID)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		err := server.addPlayToWinByGameId(c, gameID, nil)
	//
	//		assert.NoError(t, err)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
	//
	//	t.Run("Should return nil when game is already marked as Play to Win (idempotent)", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//
	//		mockTx := setupAddPlayToWinMocksForIdempotent(mockDB, gameID)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		err := server.addPlayToWinByGameId(c, gameID, nil)
	//
	//		assert.NoError(t, err)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
	//
	//	t.Run("Should return error when DB returns an unexpected error", func(t *testing.T) {
	//		server, mockDB := setupTestServer()
	//		gameID := uuid.New()
	//		expectedErr := errors.New("unexpected db error")
	//
	//		mockTx := new(MockTx)
	//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
	//
	//		// Fail at GetGame
	//		mockRow := new(MockRow)
	//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
	//			Return(expectedErr)
	//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
	//		mockTx.On("Rollback", mock.Anything).Return(nil)
	//
	//		w := httptest.NewRecorder()
	//		c, _ := gin.CreateTestContext(w)
	//		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)
	//
	//		err := server.addPlayToWinByGameId(c, gameID, nil)
	//
	//		assert.ErrorIs(t, err, expectedErr)
	//		mockDB.AssertExpectations(t)
	//		mockTx.AssertExpectations(t)
	//	})
}

//
//// --- RemovePlayToWinGameByGameId ---
//
//func mockGetGameRow(mockDB *MockDatabase, gameID uuid.UUID, ptwGameID uuid.UUID) {
//	mockRow := new(MockRow)
//	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = "Catan" // display_title
//			*args.Get(2).(*string) = "Catan" // title
//			*args.Get(3).(*string) = "catan"
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameID, Valid: true}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//	mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//}
//
//// setupAddPlayToWinMocksForSuccess mocks the full happy-path flow of addPlayToWinByGameId.
//// It sets up BeginTx, GetGame, two getOrCreatePlayToWinGroup calls (with savepoints),
//// and CreatePlayToWinGame. Returns mockTx for AssertExpectations.
//func setupAddPlayToWinMocksForSuccess(mockDB *MockDatabase, gameID uuid.UUID, ptwID uuid.UUID) *MockTx {
//	groupID := uuid.New()
//	mockTx := new(MockTx)
//	mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil).Once()
//
//	// GetGame via tx (7 scan args: ID, DisplayTitle, Title, SanitizedTitle, Barcode, PlayToWinGameID, CreatedAt)
//	mockGetGameRow := new(MockRow)
//	mockGetGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = "Catan" // display_title
//			*args.Get(2).(*string) = "Catan" // title
//			*args.Get(3).(*string) = "catan"
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//	mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockGetGameRow).Once()
//
//	// First getOrCreatePlayToWinGroup: group does not exist, create it
//	sp1 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
//	mockGroupNotFoundRow := new(MockRow)
//	mockGroupNotFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
//	sp1.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockGroupNotFoundRow)
//	sp1.On("Rollback", mock.Anything).Return(nil)
//	mockCreateGroupRow := new(MockRow)
//	mockCreateGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = "Catan"
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//		}).Return(nil)
//	mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockCreateGroupRow).Once()
//
//	// Second getOrCreatePlayToWinGroup (debug call): group now exists
//	sp2 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
//	mockGroupFoundRow := new(MockRow)
//	mockGroupFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = "Catan"
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//	sp2.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockGroupFoundRow)
//	sp2.On("Commit", mock.Anything).Return(nil)
//
//	// CreatePlayToWinGame savepoint
//	sp3 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp3, nil).Once()
//	mockPtwRow := new(MockRow)
//	mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false} // winner_id
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
//			*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//	sp3.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.UUID{Bytes: groupID, Valid: true}}).Return(mockPtwRow)
//	sp3.On("Commit", mock.Anything).Return(nil)
//
//	mockTx.On("Commit", mock.Anything).Return(nil)
//	return mockTx
//}
//
//// setupAddPlayToWinMocksForIdempotent mocks addPlayToWinByGameId when the PTW game already exists
//// (CreatePlayToWinGame returns 23505), triggering a restore instead.
//func setupAddPlayToWinMocksForIdempotent(mockDB *MockDatabase, gameID uuid.UUID) *MockTx {
//	groupID := uuid.New()
//	mockTx := new(MockTx)
//	mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil).Once()
//
//	mockGetGameRow := new(MockRow)
//	mockGetGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = "Catan"
//			*args.Get(2).(*string) = "Catan"
//			*args.Get(3).(*string) = "catan"
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//	mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockGetGameRow).Once()
//
//	sp1 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
//	mockGroupNotFoundRow := new(MockRow)
//	mockGroupNotFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
//	sp1.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockGroupNotFoundRow)
//	sp1.On("Rollback", mock.Anything).Return(nil)
//	mockCreateGroupRow := new(MockRow)
//	mockCreateGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = "Catan"
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//		}).Return(nil)
//	mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockCreateGroupRow).Once()
//
//	sp2 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
//	mockGroupFoundRow := new(MockRow)
//	mockGroupFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = "Catan"
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//	sp2.On("QueryRow", mock.Anything, mock.Anything, []any{"Catan"}).Return(mockGroupFoundRow)
//	sp2.On("Commit", mock.Anything).Return(nil)
//
//	// CreatePlayToWinGame returns unique constraint violation
//	sp3 := new(MockTx)
//	mockTx.On("Begin", mock.Anything).Return(sp3, nil).Once()
//	mockPtwRow := new(MockRow)
//	mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Return(&pgconn.PgError{Code: "23505"})
//	sp3.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.UUID{Bytes: groupID, Valid: true}}).Return(mockPtwRow)
//	sp3.On("Rollback", mock.Anything).Return(nil)
//
//	// RestorePlayToWinGameByLibraryGameId via tx
//	mockTx.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, nil)
//	mockTx.On("Commit", mock.Anything).Return(nil)
//	return mockTx
//}
//
//func mockGetPlayToWinGameOverviewRow(mockDB *MockDatabase, ptwID uuid.UUID, gameID uuid.UUID) {
//	mockRow := new(MockRow)
//	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true} // PtwGroupID (reuse ptwID for AddPlayToWinSession compatibility)
//			*args.Get(3).(*string) = "Catan"
//			*args.Get(4).(*string) = "catan"
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(6).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//	mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}}).Return(mockRow)
//}
//
//func validRemoveBody(t *testing.T, reason RemovePlayToWinGameRequestRemovalReason, comment *string) *bytes.Buffer {
//	t.Helper()
//	body, _ := json.Marshal(RemovePlayToWinGameRequest{
//		RemovalReason:  reason,
//		RemovalComment: comment,
//	})
//	return bytes.NewBuffer(body)
//}
//
//func TestRemovePlayToWinGame(t *testing.T) {
//	t.Run("Should return 204 No Content when game is successfully removed from Play to Win", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwGameID := uuid.New()
//
//		mockGetGameRow(mockDB, gameID, ptwGameID)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameID, Valid: true},
//			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed, Valid: true},
//			pgtype.Text{Valid: false},
//		}).Return(pgconn.CommandTag{}, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "claimed", nil))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should pass the removal comment to the DB when provided", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwGameID := uuid.New()
//		comment := "Librarian confirmed the game was claimed"
//
//		mockGetGameRow(mockDB, gameID, ptwGameID)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameID, Valid: true},
//			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed, Valid: true},
//			pgtype.Text{String: comment, Valid: true},
//		}).Return(pgconn.CommandTag{}, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "claimed", &comment))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when request body is malformed JSON", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwGameID := uuid.New()
//
//		mockGetGameRow(mockDB, gameID, ptwGameID)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), bytes.NewBufferString("{invalid json}"))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "JSON body is malformed")
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when removal comment exceeds 500 characters", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwGameID := uuid.New()
//		longComment := string(make([]byte, 501))
//
//		mockGetGameRow(mockDB, gameID, ptwGameID)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "other", &longComment))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 404 Not Found when game does not exist", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Return(pgx.ErrNoRows)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "mistake", nil))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusNotFound, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when GetGame returns an unexpected error", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Return(errors.New("db error"))
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "other", nil))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when DeletePlayToWinGame fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwGameID := uuid.New()
//
//		mockGetGameRow(mockDB, gameID, ptwGameID)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameID, Valid: true},
//			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake, Valid: true},
//			pgtype.Text{Valid: false},
//		}).Return(pgconn.CommandTag{}, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "mistake", nil))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.RemovePlayToWinGameByGameId(c, gameID)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//}
//
//// --- GetPlayToWinGameEntries ---
//
//func TestGetPlayToWinGameEntries(t *testing.T) {
//	t.Run("Should return 200 OK with list of entries when entries exist", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//		sessionID := uuid.New()
//		entryID1 := uuid.New()
//		entryID2 := uuid.New()
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID1, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(3).(*string) = "Alice Smith"
//				*args.Get(4).(*string) = "alice123"
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			}).Return(nil).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID2, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(3).(*string) = "Bob Jones"
//				*args.Get(4).(*string) = "bob456"
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			}).Return(nil).Once()
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwGameId, Valid: true}}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/"+ptwGameId.String()+"/entries", nil)
//
//		server.GetPlayToWinGameEntries(c, ptwGameId)
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response PlayToWinEntryList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Len(t, response.Entries, 2)
//		assert.Equal(t, entryID1, response.Entries[0].EntryId)
//		assert.Equal(t, "Alice Smith", response.Entries[0].EntrantName)
//		assert.Equal(t, "alice123", response.Entries[0].EntrantUniqueId)
//		assert.Equal(t, entryID2, response.Entries[1].EntryId)
//		assert.Equal(t, "Bob Jones", response.Entries[1].EntrantName)
//		assert.Equal(t, "bob456", response.Entries[1].EntrantUniqueId)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 200 OK with empty list when no entries exist", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(false)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwGameId, Valid: true}}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/"+ptwGameId.String()+"/entries", nil)
//
//		server.GetPlayToWinGameEntries(c, ptwGameId)
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response PlayToWinEntryList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Empty(t, response.Entries)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when DB query fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwGameId, Valid: true}}).Return(nil, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/"+ptwGameId.String()+"/entries", nil)
//
//		server.GetPlayToWinGameEntries(c, ptwGameId)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//	})
//}
//
//// --- AddPlayToWinSession ---
//
//func validSessionBody(t *testing.T, ptwGameId uuid.UUID, playtimeMinutes *int32, entries []struct{ name, uniqueId string }) *bytes.Buffer {
//	t.Helper()
//
//	// Build the request using the generated CreatePlayToWinSessionRequest type
//	request := CreatePlayToWinSessionRequest{
//		PlayToWinId:     ptwGameId,
//		PlaytimeMinutes: playtimeMinutes,
//		Entries: make([]struct {
//			EntrantName     string `json:"entrantName"`
//			EntrantUniqueId string `json:"entrantUniqueId"`
//		}, len(entries)),
//	}
//
//	for i, entry := range entries {
//		request.Entries[i].EntrantName = entry.name
//		request.Entries[i].EntrantUniqueId = entry.uniqueId
//	}
//
//	body, _ := json.Marshal(request)
//	return bytes.NewBuffer(body)
//}
//
//func TestAddPlayToWinSession(t *testing.T) {
//	t.Run("Should return 201 Created when session is successfully created with entries", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//		sessionID := uuid.New()
//		entryID1 := uuid.New()
//		entryID2 := uuid.New()
//		playtimeMinutes := int32(45)
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice Smith", uniqueId: "alice123"},
//			{name: "Bob Jones", uniqueId: "bob456"},
//		}
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwGameId, uuid.New())
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// Mock CreatePlayToWinSession - returns 7 columns
//		mockSessionRow := new(MockRow)
//		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Int32: playtimeMinutes, Valid: true}
//				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
//				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			pgtype.Int4{Int32: playtimeMinutes, Valid: true},
//		}).Return(mockSessionRow).Once()
//
//		// Mock CreatePlayToWinEntry for first entry - returns 9 columns
//		mockEntry1Row := new(MockRow)
//		mockEntry1Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID1, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(3).(*string) = entries[0].name
//				*args.Get(4).(*string) = entries[0].uniqueId
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(7).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
//				*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: sessionID, Valid: true},
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			entries[0].name,
//			entries[0].uniqueId,
//		}).Return(mockEntry1Row).Once()
//
//		// Mock CreatePlayToWinEntry for second entry - returns 9 columns
//		mockEntry2Row := new(MockRow)
//		mockEntry2Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID2, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(3).(*string) = entries[1].name
//				*args.Get(4).(*string) = entries[1].uniqueId
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(7).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
//				*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: sessionID, Valid: true},
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			entries[1].name,
//			entries[1].uniqueId,
//		}).Return(mockEntry2Row).Once()
//
//		mockTx.On("Commit", mock.Anything).Return(nil)
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, &playtimeMinutes, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response PlayToWinSession
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, sessionID, response.SessionId)
//		assert.NotNil(t, response.PlaytimeMinutes)
//		assert.Equal(t, playtimeMinutes, *response.PlaytimeMinutes)
//		assert.Len(t, response.PlayToWinEntries, 2)
//		assert.Equal(t, entryID1, response.PlayToWinEntries[0].EntryId)
//		assert.Equal(t, entries[0].name, response.PlayToWinEntries[0].EntrantName)
//		assert.Equal(t, entries[0].uniqueId, response.PlayToWinEntries[0].EntrantUniqueId)
//		assert.Equal(t, entryID2, response.PlayToWinEntries[1].EntryId)
//		assert.Equal(t, entries[1].name, response.PlayToWinEntries[1].EntrantName)
//		assert.Equal(t, entries[1].uniqueId, response.PlayToWinEntries[1].EntrantUniqueId)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 201 Created when session is created without playtime", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//		sessionID := uuid.New()
//		entryID := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice Smith", uniqueId: "alice123"},
//		}
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwGameId, uuid.New())
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// Mock CreatePlayToWinSession without playtime - returns 7 columns
//		mockSessionRow := new(MockRow)
//		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Valid: false}
//				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
//				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			pgtype.Int4{Valid: false},
//		}).Return(mockSessionRow).Once()
//
//		// Mock CreatePlayToWinEntry - returns 9 columns
//		mockEntryRow := new(MockRow)
//		mockEntryRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(3).(*string) = entries[0].name
//				*args.Get(4).(*string) = entries[0].uniqueId
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(7).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
//				*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: sessionID, Valid: true},
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			entries[0].name,
//			entries[0].uniqueId,
//		}).Return(mockEntryRow).Once()
//
//		mockTx.On("Commit", mock.Anything).Return(nil)
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response PlayToWinSession
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Nil(t, response.PlaytimeMinutes)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
//		server, _ := setupTestServer()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", bytes.NewBufferString("{invalid json}"))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "JSON body is malformed")
//	})
//
//	t.Run("Should return 400 Bad Request when playtimeMinutes is negative", func(t *testing.T) {
//		server, _ := setupTestServer()
//		ptwGameId := uuid.New()
//		playtimeMinutes := int32(-5)
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: "alice123"},
//		}
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, &playtimeMinutes, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 400 Bad Request when entrantName is empty", func(t *testing.T) {
//		server, _ := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "", uniqueId: "alice123"},
//		}
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 400 Bad Request when entrantName exceeds 100 characters", func(t *testing.T) {
//		server, _ := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: string(make([]byte, 101)), uniqueId: "alice123"},
//		}
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 400 Bad Request when entrantUniqueId is empty", func(t *testing.T) {
//		server, _ := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: ""},
//		}
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 400 Bad Request when entrantUniqueId exceeds 100 characters", func(t *testing.T) {
//		server, _ := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: string(make([]byte, 101))},
//		}
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 404 Not Found when play to win game does not exist (FK violation)", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: "alice123"},
//		}
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Return(pgx.ErrNoRows)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwGameId, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusNotFound, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when transaction begin fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: "alice123"},
//		}
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwGameId, uuid.New())
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(nil, errors.New("tx error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "tx error")
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when CreatePlayToWinSession fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: "alice123"},
//		}
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwGameId, uuid.New())
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		mockSessionRow := new(MockRow)
//		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Return(errors.New("db error"))
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			pgtype.Int4{Valid: false},
//		}).Return(mockSessionRow).Once()
//
//		mockTx.On("Rollback", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when CreatePlayToWinEntry fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId := uuid.New()
//		sessionID := uuid.New()
//
//		entries := []struct{ name, uniqueId string }{
//			{name: "Alice", uniqueId: "alice123"},
//		}
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwGameId, uuid.New())
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// Mock successful session creation - 7 columns
//		mockSessionRow := new(MockRow)
//		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId, Valid: true}
//				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Valid: false}
//				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
//				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			pgtype.Int4{Valid: false},
//		}).Return(mockSessionRow).Once()
//
//		// Mock failed entry creation - 9 columns
//		mockEntryRow := new(MockRow)
//		mockEntryRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Return(errors.New("entry creation error"))
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: sessionID, Valid: true},
//			pgtype.UUID{Bytes: ptwGameId, Valid: true},
//			entries[0].name,
//			entries[0].uniqueId,
//		}).Return(mockEntryRow).Once()
//
//		mockTx.On("Commit", mock.Anything).Return(nil).Maybe()
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, ptwGameId, nil, entries))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddPlayToWinSession(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "entry creation error")
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//}
//
//func TestListPlayToWinGames(t *testing.T) {
//	t.Run("Should return 200 OK with play to win game list when called without filters", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwGameId1 := uuid.New()
//		ptwGameId2 := uuid.New()
//		gameID1 := uuid.New()
//		gameID2 := uuid.New()
//		winnerEntryID := uuid.New()
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId1, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID1, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false} // ptw_group_id
//				*args.Get(3).(*string) = "Azul"
//				*args.Get(4).(*string) = SanitizeTitle("Azul")
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(6).(*pgtype.UUID) = pgtype.UUID{Bytes: winnerEntryID, Valid: true}
//				*args.Get(7).(*pgtype.Text) = pgtype.Text{String: "Alice", Valid: true}
//				*args.Get(8).(*pgtype.Text) = pgtype.Text{String: "alice123", Valid: true}
//			}).Return(nil).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameId2, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID2, Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Valid: false} // ptw_group_id
//				*args.Get(3).(*string) = "Catan"
//				*args.Get(4).(*string) = SanitizeTitle("Catan")
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//				*args.Get(6).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//				*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
//				*args.Get(8).(*pgtype.Text) = pgtype.Text{Valid: false}
//			}).Return(nil).Once()
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%%", int32(100), int32(0)}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/games", nil)
//
//		server.ListPlayToWinGames(c, ListPlayToWinGamesParams{})
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response PlayToWinGameList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Len(t, response.Games, 2)
//		assert.Equal(t, ptwGameId1, response.Games[0].PlayToWinId)
//		assert.Equal(t, gameID1, response.Games[0].GameId)
//		assert.Equal(t, "Azul", response.Games[0].Title)
//		assert.NotNil(t, response.Games[0].Winner)
//		assert.Equal(t, winnerEntryID, response.Games[0].Winner.EntryId)
//		assert.Equal(t, "Alice", response.Games[0].Winner.EntrantName)
//		assert.Equal(t, "alice123", response.Games[0].Winner.EntrantUniqueId)
//		assert.Equal(t, ptwGameId2, response.Games[1].PlayToWinId)
//		assert.Equal(t, gameID2, response.Games[1].GameId)
//		assert.Equal(t, "Catan", response.Games[1].Title)
//		assert.Nil(t, response.Games[1].Winner)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 200 OK with empty list when title limit and offset filters return no matches", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		title := "Catan: Special Edition"
//		limit := int32(25)
//		offset := int32(10)
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%" + SanitizeTitle(title) + "%", limit, offset}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/games?title=Catan:+Special+Edition&limit=25&offset=10", nil)
//
//		server.ListPlayToWinGames(c, ListPlayToWinGamesParams{Title: &title, Limit: &limit, Offset: &offset})
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response PlayToWinGameList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Empty(t, response.Games)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when limit or offset are invalid", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		limit := int32(0)
//		offset := int32(-1)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/games?limit=0&offset=-1", nil)
//
//		server.ListPlayToWinGames(c, ListPlayToWinGamesParams{Limit: &limit, Offset: &offset})
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//		mockDB.AssertNotCalled(t, "Query", mock.Anything, mock.Anything, mock.Anything)
//	})
//
//	t.Run("Should return 500 Internal Server Error when listing play to win games fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		title := "Azul"
//		limit := int32(5)
//		offset := int32(0)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%" + SanitizeTitle(title) + "%", limit, offset}).Return(nil, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/ptw/games?title=Azul&limit=5&offset=0", nil)
//
//		server.ListPlayToWinGames(c, ListPlayToWinGamesParams{Title: &title, Limit: &limit, Offset: &offset})
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//	})
//}
//
//func TestDeletePlayToWinGameEndpoint(t *testing.T) {
//	t.Run("Should return 400 Bad Request when request body is malformed", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/ptwId/"+ptwID.String(), bytes.NewBufferString("{invalid json}"))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.DeletePlayToWinGame(c, ptwID)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		mockDB.AssertNotCalled(t, "Exec", mock.Anything, mock.Anything, mock.Anything)
//	})
//
//	t.Run("Should return 400 Bad Request when removal comment exceeds 500 characters", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//		gameID := uuid.New()
//		longComment := string(make([]byte, 501))
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwID, gameID)
//
//		body, _ := json.Marshal(RemovePlayToWinGameRequest{
//			RemovalReason:  RemovePlayToWinGameRequestRemovalReason("claimed"),
//			RemovalComment: &longComment,
//		})
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/ptwId/"+ptwID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.DeletePlayToWinGame(c, ptwID)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//		mockDB.AssertNotCalled(t, "Exec", mock.Anything, mock.Anything, mock.Anything)
//	})
//
//	t.Run("Should call delete by play to win ID query when request is valid", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//		gameID := uuid.New()
//		comment := "claimed"
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwID, gameID)
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		body, _ := json.Marshal(RemovePlayToWinGameRequest{
//			RemovalReason:  RemovePlayToWinGameRequestRemovalReason("mistake"),
//			RemovalComment: &comment,
//		})
//
//		mockTx.On("Exec", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwID, Valid: true},
//			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake, Valid: true},
//			pgtype.Text{String: comment, Valid: true},
//		}).Return(pgconn.CommandTag{}, nil)
//		mockTx.On("Commit", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/ptwId/"+ptwID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.DeletePlayToWinGame(c, ptwID)
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when delete by play to win ID query fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//		gameID := uuid.New()
//
//		mockGetPlayToWinGameOverviewRow(mockDB, ptwID, gameID)
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		body, _ := json.Marshal(RemovePlayToWinGameRequest{
//			RemovalReason: RemovePlayToWinGameRequestRemovalReason("mistake"),
//		})
//
//		mockTx.On("Exec", mock.Anything, mock.Anything, []any{
//			pgtype.UUID{Bytes: ptwID, Valid: true},
//			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake, Valid: true},
//			pgtype.Text{Valid: false},
//		}).Return(pgconn.CommandTag{}, errors.New("db error"))
//		mockTx.On("Rollback", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/ptw/game/ptwId/"+ptwID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.DeletePlayToWinGame(c, ptwID)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//}
//
//func TestDrawPlayToWinRaffle(t *testing.T) {
//	t.Run("Should return 200 OK with selected winner when entries exist", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//		entryID := uuid.New()
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//				*args.Get(3).(*string) = "Alice"
//				*args.Get(4).(*string) = "alice123"
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			}).Return(nil).Once()
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}}).Return(mockRows, nil)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}, pgtype.UUID{Bytes: entryID, Valid: true}}).Return(pgconn.CommandTag{}, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/raffle/ptwId/"+ptwID.String(), nil)
//
//		server.DrawPlayToWinRaffle(c, ptwID)
//
//		assert.Equal(t, http.StatusOK, w.Code)
// 		var response PlayToWinEntry
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, entryID, response.EntryId)
//		assert.Equal(t, "Alice", response.EntrantName)
//		assert.Equal(t, "alice123", response.EntrantUniqueId)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when listing entries fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}}).Return(nil, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/raffle/ptwId/"+ptwID.String(), nil)
//
//		server.DrawPlayToWinRaffle(c, ptwID)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when updating winner fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		ptwID := uuid.New()
//		entryID := uuid.New()
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//			Run(func(args mock.Arguments) {
//				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID, Valid: true}
//				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Valid: true}
//				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//				*args.Get(3).(*string) = "Alice"
//				*args.Get(4).(*string) = "alice123"
//				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			}).Return(nil).Once()
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}}).Return(mockRows, nil)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: ptwID, Valid: true}, pgtype.UUID{Bytes: entryID, Valid: true}}).Return(pgconn.CommandTag{}, errors.New("update error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/raffle/ptwId/"+ptwID.String(), nil)
//
//		server.DrawPlayToWinRaffle(c, ptwID)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "update error")
//		mockDB.AssertExpectations(t)
//	})
//}
//
//func TestResetPlayToWinRaffle(t *testing.T) {
//	t.Run("Should return 204 No Content when raffle winners are reset", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//
//		mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/raffle/reset", nil)
//
//		server.ResetPlayToWinRaffle(c)
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when reset query fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//
//		mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("reset error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/ptw/raffle/reset", nil)
//
//		server.ResetPlayToWinRaffle(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "reset error")
//		mockDB.AssertExpectations(t)
//	})
//}
