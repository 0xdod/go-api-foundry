package postgres

import (
	"encoding/json"
	goerr "errors"

	"time"

	"github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func Text(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func NullText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{
			String: "",
			Valid:  false,
		}
	}
	return Text(value)
}

func Date(t time.Time) pgtype.Date {
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

func DateFromString(s string) pgtype.Date {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return pgtype.Date{
			Valid: false,
		}
	}
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func TimestamptzFromString(s string) pgtype.Timestamptz {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return pgtype.Timestamptz{
			Valid: false,
		}
	}
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func Bool(value bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  value,
		Valid: true,
	}
}

func NullBool(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{
			Valid: false,
		}
	}
	return Bool(*value)
}

func Int8(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: value,
		Valid: true,
	}
}

func Int4(value int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}

func FromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func JSONB(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func NullTextPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{
			Valid: false,
		}
	}
	return Text(*value)
}

func NullUUIDPtr(u *uuid.UUID) uuid.NullUUID {
	if u == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *u, Valid: true}
}

func TextPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{
			Valid: false,
		}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func CheckErrNoRows(err error, msg string) error {
	if err == pgx.ErrNoRows {
		return errors.NewNotFoundError(msg, err)
	}
	return err
}

func CheckErrUniqueViolation(err error, msg string) error {
	var pgErr *pgconn.PgError
	if !goerr.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23505" {
		return errors.NewConflictError(msg, err)
	}

	return err
}
