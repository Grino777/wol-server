package usecase

import "context"

type ActionsService interface {
	WakeServer(ctx context.Context, id int) error
	OffServer(ctx context.Context, id int) error
}
