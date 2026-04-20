package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/jmoiron/sqlx"
)

func TestRepository_RegisterUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := clock.FixedClocker{}
	var wantID int64 = 10
	u := &entity.User{
		Name:     "genki",
		Password: "hashed-password",
		Role:     "user",
		Created:  c.Now(),
		Modified: c.Now(),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(
		`INSERT INTO user \( name, password, role, created, modified \) VALUES \(\?, \?, \?, \?, \?\)`,
	).WithArgs(u.Name, u.Password, u.Role, c.Now(), c.Now()).WillReturnResult(sqlmock.NewResult(wantID, 1))

	xdb := sqlx.NewDb(db, "mysql")
	r := &Repository{Clocker: c}
	if err := r.RegisterUser(ctx, xdb, u); err != nil {
		t.Fatalf("want no error, but got %v", err)
	}
	if got := u.ID; got != entity.UserID(wantID) {
		t.Fatalf("want user id %d, but got %d", wantID, got)
	}
}
