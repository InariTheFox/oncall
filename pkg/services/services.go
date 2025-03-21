package services

import "context"

type BackgroundService interface {
	Run(ctx context.Context) error
}
