package api

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgInt4ToInteger(t *testing.T) {
	t.Run("returns nil for invalid", func(t *testing.T) {
		var in pgtype.Int4 // zero-value, Valid == false
		out := pgInt4ToInteger(in)
		require.Nil(t, out)
	})

	t.Run("returns pointer for valid", func(t *testing.T) {
		in := pgtype.Int4{Int32: 5, Valid: true}
		out := pgInt4ToInteger(in)
		require.NotNil(t, out)
		assert.Equal(t, int32(5), *out)
	})

	t.Run("boundary values", func(t *testing.T) {
		min := pgtype.Int4{Int32: math.MinInt32, Valid: true}
		max := pgtype.Int4{Int32: math.MaxInt32, Valid: true}

		outMin := pgInt4ToInteger(min)
		outMax := pgInt4ToInteger(max)

		require.NotNil(t, outMin)
		require.NotNil(t, outMax)
		assert.Equal(t, int32(math.MinInt32), *outMin)
		assert.Equal(t, int32(math.MaxInt32), *outMax)
	})
}

func TestPgTextToString(t *testing.T) {
	t.Run("returns nil for invalid", func(t *testing.T) {
		var in pgtype.Text
		out := pgTextToString(in)
		require.Nil(t, out)
	})

	t.Run("returns pointer for empty string", func(t *testing.T) {
		in := pgtype.Text{String: "", Valid: true}
		out := pgTextToString(in)
		require.NotNil(t, out)
		assert.Equal(t, "", *out)
	})

	t.Run("returns pointer for normal string", func(t *testing.T) {
		in := pgtype.Text{String: "hello", Valid: true}
		out := pgTextToString(in)
		require.NotNil(t, out)
		assert.Equal(t, "hello", *out)
	})
}

func TestUUIDConversionsRoundTrip(t *testing.T) {
	t.Run("round trip preserves uuid", func(t *testing.T) {
		u := uuid.New()
		pg := uuidToPgTypeUUID(u)
		got := pgUUIDToUUID(pg)
		assert.Equal(t, u, got)
	})

	t.Run("zero-value pgtype.UUID returns zero uuid", func(t *testing.T) {
		var pg pgtype.UUID // Valid == false
		got := pgUUIDToUUID(pg)
		require.Equal(t, uuid.UUID{}, got)
	})
}
