package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/GenkiSugiyama/go_todo_app_2/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

// AddTaskのテストは、sqlmockを使用してDBの挙動をモックして行う
func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := clock.FixedClocker{}
	var wantID int64 = 20
	okTask := &entity.Task{
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
		`INSERT INTO task \(title, status, created, modified\) VALUES \(\?, \?, \?, \?\)`,
	).WithArgs(okTask.Title, okTask.Status, c.Now(), c.Now()).WillReturnResult(sqlmock.NewResult(wantID, 1))

	// sqlmockが作成した*sql.DBをラップして*sqlx.DBを作成する
	xdb := sqlx.NewDb(db, "mysql")
	// テスト用の固定Clockerを持つRepositoryを作成してAddTaskを実行する
	r := &Repository{Clocker: c}
	if err := r.AddTask(ctx, xdb, okTask); err != nil {
		t.Errorf("want no error, but got %v\n", err)
	}
}

// LisstTasksのテストは実際のDBを使用して行う
func TestRepository_ListTasks(t *testing.T) {
	ctx := context.Background()

	// BeginTxxは、sqlxのトランザクション開始用の関数で、テスト終了後に必ずロールバックするようにt.Cleanupで登録している。
	// 他のテストや実際にDBに影響を与えないためこのテスト専用のトランザクションを用意して終了後にはロールバックしている
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)
	t.Cleanup(func() { _ = tx.Rollback() })
	if err != nil {
		t.Fatal(err)
	}
	wants := prepareTasks(ctx, t, tx)

	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s\n", d)
	}
}

func prepareTasks(ctx context.Context, t *testing.T, con Execer) entity.Tasks {
	t.Helper()

	if _, err := con.ExecContext(ctx, "DELETE FROM task;"); err != nil {
		t.Logf("failed to initialize tasks: %v\n", err)
	}
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			Title: "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			Title: "want task 2", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			Title: "want task 3", Status: "done",
			Created: c.Now(), Modified: c.Now(),
		},
	}
	result, err := con.ExecContext(
		ctx,
		`INSERT INTO task (title, status, created, modified)
		 VALUES
		 	(?, ?, ?, ?),
		 	(?, ?, ?, ?),
		 	(?, ?, ?, ?);`,
		wants[0].Title, wants[0].Status, wants[0].Created, wants[0].Modified,
		wants[1].Title, wants[1].Status, wants[1].Created, wants[1].Modified,
		wants[2].Title, wants[2].Status, wants[2].Created, wants[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	// mysqlのlast_insert_idは、複数行挿入した場合は最初の行のIDを返す。
	// 複数行挿入した「最後のID」ではないため注意！
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	wants[0].ID = entity.TaskID(id)
	wants[1].ID = entity.TaskID(id + 1)
	wants[2].ID = entity.TaskID(id + 2)
	return wants
}
