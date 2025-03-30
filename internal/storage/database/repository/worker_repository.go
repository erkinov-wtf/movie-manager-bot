package repository

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/google/uuid"
)

type WorkerRepositoryInterface interface {
	GetWorkerState(ctx context.Context, workerID string) (database.WorkerState, error)
	UpsertWorkerState(ctx context.Context, params database.UpsertWorkerStateParams) error
	CreateWorkerTask(ctx context.Context, params database.CreateWorkerTaskParams) (uuid.UUID, error)
	UpdateWorkerTask(ctx context.Context, params database.UpdateWorkerTaskParams) error
	GetRecentTasks(ctx context.Context, params database.GetRecentTasksParams) ([]database.WorkerTask, error)
	GetWorkerTask(ctx context.Context, id uuid.UUID) (database.WorkerTask, error)
}

type WorkerRepository struct {
	q *database.Queries
}

func NewWorkerRepository(db database.DBTX) WorkerRepositoryInterface {
	return &WorkerRepository{
		q: database.New(db),
	}
}

func (r *WorkerRepository) GetWorkerState(ctx context.Context, workerID string) (database.WorkerState, error) {
	return r.q.GetWorkerState(ctx, workerID)
}

func (r *WorkerRepository) UpsertWorkerState(ctx context.Context, params database.UpsertWorkerStateParams) error {
	_, err := r.q.UpsertWorkerState(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *WorkerRepository) CreateWorkerTask(ctx context.Context, params database.CreateWorkerTaskParams) (uuid.UUID, error) {
	return r.q.CreateWorkerTask(ctx, params)
}

func (r *WorkerRepository) UpdateWorkerTask(ctx context.Context, params database.UpdateWorkerTaskParams) error {
	return r.q.UpdateWorkerTask(ctx, params)
}

func (r *WorkerRepository) GetRecentTasks(ctx context.Context, params database.GetRecentTasksParams) ([]database.WorkerTask, error) {
	return r.q.GetRecentTasks(ctx, params)
}

func (r *WorkerRepository) GetWorkerTask(ctx context.Context, id uuid.UUID) (database.WorkerTask, error) {
	return r.q.GetWorkerTask(ctx, id)
}
