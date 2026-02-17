package postgres

import (
	"encoding/json"
	goerr "errors"

	"time"

	"github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func Text(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{
			Valid: false,
		}
	}
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func Date(t time.Time) pgtype.Date {
	if t.IsZero() {
		return pgtype.Date{
			Valid: false,
		}
	}
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
	if t.IsZero() {
		return pgtype.Timestamptz{
			Valid: false,
		}
	}
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
	if value == 0 {
		return pgtype.Int8{
			Valid: false,
		}
	}
	return pgtype.Int8{
		Int64: value,
		Valid: true,
	}
}

func Int4(value int32) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{
			Valid: false,
		}
	}
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

func NullText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{
			Valid: false,
		}
	}
	return Text(*value)
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

func CheckErrCheckViolation(err error, msg string) error {
	var pgErr *pgconn.PgError
	if !goerr.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23514" {
		return errors.NewInvalidRequestError(msg, err)
	}

	return err
}

func IsErrNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func IsErrUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !goerr.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}
