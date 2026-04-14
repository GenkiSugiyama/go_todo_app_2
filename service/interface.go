package service

import (
	"context"

	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/GenkiSugiyama/go_todo_app_2/store"
)

type TaskAdder interface {
	AddTask(ctx context.Context, db store.Execer, t *entity.Task) error
}

type TaskLister interface {
	ListTasks(ctx context.Context, db store.Queryer) (entity.Tasks, error)
}
