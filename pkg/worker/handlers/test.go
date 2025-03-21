package handlers

import (
	"context"
	"fmt"

	"github.com/InariTheFox/oncall/pkg/worker"
)

func Handle(ctx context.Context, job *worker.Job) {
	msg := job.Args[0]
	fmt.Printf("Received message: %s\n", msg)
}
