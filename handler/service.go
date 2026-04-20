package handler

import (
	"context"

	"github.com/GenkiSugiyama/go_todo_app_2/entity"
)

// Goではインターフェースの宣言をconsumer側（呼び出し側）で定義することが一般的
// そのため、ハンドラ層で使いたいメソッド用のインターフェースを定義してサービス層でそれを実装する
// C#ではprivider側がインターフェースを公開してconsumer側はそのインターフェースを参照する形が強かったがGoの設計思想では異なる

//go:generate go run github.com/matryer/moq -out moq_test.go . ListTasksService AddTaskService RegisterUserService
type ListTasksService interface {
	ListTasks(ctx context.Context) (entity.Tasks, error)
}

type AddTaskService interface {
	AddTask(ctx context.Context, title string) (*entity.Task, error)
}

type RegisterUserService interface {
	RegisterUser(ctx context.Context, name, password, role string) (*entity.User, error)
}
