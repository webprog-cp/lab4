package repositories

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestHasDependents(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"translated foreign key": {gorm.ErrForeignKeyViolated, true},
		"restrict violation":     {fmt.Errorf("delete: %w", &pgconn.PgError{Code: "23001"}), true},
		"other postgres error":   {&pgconn.PgError{Code: "23505"}, false},
		"plain error":            {errors.New("boom"), false},
	}
	for name, c := range cases {
		if got := hasDependents(c.err); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}
