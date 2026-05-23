package store

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/GenkiSugiyama/go_todo_app_2/testutil"
	"github.com/GenkiSugiyama/go_todo_app_2/testutil/fixture"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

func prepareUser(ctx context.Context, t *testing.T, db Execer) entity.UserID {
	t.Helper()
	u := fixture.User(nil)
	result, err := db.ExecContext(ctx,
		`INSERT INTO user (name, password, role, created, modified)
			VALUES (?, ?, ?, ?, ?);`,
		u.Name, u.Password, u.Role, u.Created, u.Modified,
	)
	if err != nil {
		t.Fatalf("insert user: %v\n", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("got user_id: %v\n", err)
	}
	return entity.UserID(id)
}

func prepareTasks(ctx context.Context, t *testing.T, con Execer) (entity.UserID, entity.Tasks) {
	t.Helper()

	userID := prepareUser(ctx, t, con)
	otherUserID := prepareUser(ctx, t, con)
	c := clock.FixedClocker{}

	wants := entity.Tasks{
		{
			UserID:   userID,
			Title:    "want task 1",
			Status:   "todo",
			Created:  c.Now(),
			Modified: c.Now(),
		},
		{
			UserID:   userID,
			Title:    "want task 2",
			Status:   "done",
			Created:  c.Now(),
			Modified: c.Now(),
		},
	}
	tasks := entity.Tasks{
		wants[0],
		{
			UserID:   otherUserID,
			Title:    "not want task",
			Status:   "todo",
			Created:  c.Now(),
			Modified: c.Now(),
		},
		wants[1],
	}
	result, err := con.ExecContext(ctx,
		`INSERT INTO task (user_id, title, status, created, modified)
		VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)`,
		tasks[0].UserID, tasks[0].Title, tasks[0].Status, tasks[0].Created, tasks[0].Modified,
		tasks[1].UserID, tasks[1].Title, tasks[1].Status, tasks[1].Created, tasks[1].Modified,
		tasks[2].UserID, tasks[2].Title, tasks[2].Status, tasks[2].Created, tasks[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	tasks[0].ID = entity.TaskID(id)
	tasks[1].ID = entity.TaskID(id + 1)
	tasks[2].ID = entity.TaskID(id + 2)
	return userID, wants
}

// AddTaskのテストは、sqlmockを使用してDBの挙動をモックして行う
func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := clock.FixedClocker{}
	var wantID int64 = 20
	okTask := &entity.Task{
		UserID:   10,
		Title:    "ok task",
		Status:   "todo",
		Created:  c.Now(),
		Modified: c.Now(),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// テスト対象のメソッド実行時に期待されるSQLクエリと引数をsqlmockに登録する
	// WillReturnResultは、期待に合致するクエリが実行されたときにモックが返す結果を指定するためのメソッドで、ここではLastInsertIdがwantIDであることを指定している
	mock.ExpectExec(
		regexp.QuoteMeta(`INSERT INTO task
			(user_id, title, status, created, modified)
			VALUES (?, ?, ?, ?, ?)`),
	).WithArgs(okTask.UserID, okTask.Title, okTask.Status, c.Now(), c.Now()).WillReturnResult(sqlmock.NewResult(wantID, 1))

	// sqlmockが作成した*sql.DBをラップして*sqlx.DBを作成する
	xdb := sqlx.NewDb(db, "mysql")
	// テスト用の固定Clockerを持つRepositoryを作成してAddTaskを実行する
	r := &Repository{Clocker: c}
	if err := r.AddTask(ctx, xdb, okTask); err != nil {
		t.Errorf("want no error, but got %v\n", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations were not met: %v", err)
	}
}

// LisstTasksのテストは実際のDBを使用して行う
func TestRepository_ListTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// BeginTxxは、sqlxのトランザクション開始用の関数で、テスト終了後に必ずロールバックするようにt.Cleanupで登録している。
	// 他のテストや実際にDBに影響を与えないためこのテスト専用のトランザクションを用意して終了後にはロールバックしている
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)
	t.Cleanup(func() { _ = tx.Rollback() })
	if err != nil {
		t.Fatal(err)
	}
	wantUserID, wantTasks := prepareTasks(ctx, t, tx)

	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx, wantUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if d := cmp.Diff(gots, wantTasks); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s\n", d)
	}
}
