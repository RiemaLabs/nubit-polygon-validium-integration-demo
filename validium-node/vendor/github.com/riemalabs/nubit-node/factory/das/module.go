package das

import (
	"context"

	"go.uber.org/fx"

	"github.com/riemalabs/nubit-node/da/das"
	"github.com/riemalabs/nubit-node/factory/node"
)

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	var err error
	// do not validate daser config for bridge node as it
	// does not need it
	if tp != node.Bridge {
		err = cfg.Validate()
	}

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(err),
		fx.Provide(
			func(c Config) []das.Option {
				return []das.Option{
					das.WithSamplingRange(c.SamplingRange),
					das.WithConcurrencyLimit(c.ConcurrencyLimit),
					das.WithBackgroundStoreInterval(c.BackgroundStoreInterval),
					das.WithSampleFrom(c.SampleFrom),
					das.WithSampleTimeout(c.SampleTimeout),
				}
			},
		),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"das",
			baseComponents,
			fx.Provide(fx.Annotate(
				newDASer,
				fx.OnStart(func(ctx context.Context, breaker *das.DASer) error {
					return breaker.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, breaker *das.DASer) error {
					return breaker.Stop(ctx)
				}),
			)),
			// Module is needed for the RPC handler
			fx.Provide(func(das *das.DASer) Module {
				return das
			}),
		)
	case node.Bridge:
		return fx.Module(
			"das",
			baseComponents,
			fx.Provide(newDaserStub),
		)
	default:
		panic("invalid node type")
	}
}
