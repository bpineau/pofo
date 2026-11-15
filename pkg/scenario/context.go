package scenario

import (
	"context"
	"math/rand/v2"
)

// DrawContext draws a path with the same RNG consumption as Source.Draw and
// discards partial results on cancellation. Built-in sources check between
// periods. Custom sources may implement DrawContext(context.Context,
// *rand.Rand) (Sequence, error); otherwise cancellation is checked before and
// after their Draw call, whose duration remains the custom source's responsibility.
// Prepare a reusable source before calling this repeatedly.
func DrawContext(ctx context.Context, src Source, rng *rand.Rand) (Sequence, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s, ok := src.(interface {
		DrawContext(context.Context, *rand.Rand) (Sequence, error)
	}); ok {
		return s.DrawContext(ctx, rng)
	}
	var seq Sequence
	if s, ok := src.(interface {
		drawContext(context.Context, *rand.Rand) Sequence
	}); ok {
		seq = s.drawContext(ctx, rng)
	} else {
		seq = src.Draw(rng)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return seq, nil
}
