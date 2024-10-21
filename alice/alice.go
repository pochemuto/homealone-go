package alice

import "context"

type Alice struct {
	ctx context.Context
}

func NewAlice() Alice {
	return Alice{}
}

func (a Alice) Start(ctx context.Context) error {
	a.ctx = ctx

	return nil
}
