package worker

import "context"

type Worker interface {
	RegisterHandler(JobType, JobHandler, any)

	Stop(ctx context.Context)
}
