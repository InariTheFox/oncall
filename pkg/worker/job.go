package worker

import (
	"context"

	"github.com/google/uuid"
)

type Job struct {
	Args []string
	ID   uuid.UUID
	Type JobType
}

type JobHandler func(ctx context.Context, job *Job)

type JobType string
