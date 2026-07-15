package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"mira/internal/core"
)

type Job struct {
	NoteID string
}

type Queue struct {
	jobs chan Job
}

func NewQueue(buffer int) *Queue {
	return &Queue{jobs: make(chan Job, buffer)}
}

func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) Close() {
	close(q.jobs)
}

type WorkerPool struct {
	queue       *Queue
	processor   *Processor
	logger      *slog.Logger
	workers     int
	taskTimeout time.Duration
	wg          sync.WaitGroup
}

func NewWorkerPool(queue *Queue, processor *Processor, logger *slog.Logger, workers int, taskTimeout time.Duration) *WorkerPool {
	return &WorkerPool{
		queue:       queue,
		processor:   processor,
		logger:      logger,
		workers:     workers,
		taskTimeout: taskTimeout,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.run(ctx, i)
	}
}

func (p *WorkerPool) run(ctx context.Context, workerID int) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue.Jobs():
			if !ok {
				return
			}
			taskCtx, cancel := context.WithTimeout(ctx, p.taskTimeout)
			err := p.processor.Process(taskCtx, job.NoteID)
			cancel()
			if err != nil {
				if !errors.Is(err, core.ErrNotFound) {
					_ = p.processor.MarkFailed(ctx, job.NoteID)
				}
				p.logger.Warn("enrichment failed", "worker", workerID, "note_id", job.NoteID, "error", err)
			}
		}
	}
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
