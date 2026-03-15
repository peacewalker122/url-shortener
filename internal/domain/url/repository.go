package url

import "context"

type Repository interface {
	Save(ctx context.Context, mapping Mapping) error
	FindByShortCode(ctx context.Context, shortCode string) (Mapping, error)
}
