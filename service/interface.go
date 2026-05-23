package service

import (
	"context"

	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/GenkiSugiyama/go_todo_app_2/store"
)

// ハンドラ層でサービス層に実装してもらいたい処理用のインターフェースを宣言したように
// サービス層ではリポジトリ層で実装してもらいたいインターフェースを宣言する
// ビジネスロジックとして複数のRepositoryによる永続化処理を行い、1つでも失敗したらロールバックする必要がある場合
// トランザクションはサービス層で制御する必要がある
// そのために現状のサービス層では永続化ロジックの抽象に加えてDBの抽象に依存しているが、理想としてはtransactionの抽象に依存すること

type TaskAdder interface {
	AddTask(ctx context.Context, db store.Execer, t *entity.Task) error
}

type TaskLister interface {
	ListTasks(ctx context.Context, db store.Queryer, userID entity.UserID) (entity.Tasks, error)
}

type UserRegister interface {
	RegisterUser(ctx context.Context, db store.Execer, u *entity.User) error
}

type UserGetter interface {
	GetUser(ctx context.Context, db store.Queryer, name string) (*entity.User, error)
}

type TokenGenerator interface {
	GenerateToken(ctx context.Context, u entity.User) ([]byte, error)
}
