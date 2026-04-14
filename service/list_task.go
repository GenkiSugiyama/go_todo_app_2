package service

import (
	"context"
	"fmt"

	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/GenkiSugiyama/go_todo_app_2/store"
)

type ListTask struct {
	DB   store.Queryer
	Repo TaskLister
}

func (l *ListTask) ListTasks(ctx context.Context) (entity.Tasks, error) {
	tasks, err := l.Repo.ListTasks(ctx, l.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	return tasks, nil
}
