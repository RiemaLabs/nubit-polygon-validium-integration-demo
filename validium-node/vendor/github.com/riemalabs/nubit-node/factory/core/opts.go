package core

import (
	"go.uber.org/fx"

	"github.com/riemalabs/nubit-node/rpc/core"
	"github.com/riemalabs/nubit-node/strucs/eh"
	"github.com/riemalabs/nubit-node/strucs/utils/fxutil"
)

// WithClient sets custom client for core process
func WithClient(client core.Client) fx.Option {
	return fxutil.ReplaceAs(client, new(core.Client))
}

// WithHeaderConstructFn sets custom func that creates extended header
func WithHeaderConstructFn(construct header.ConstructFn) fx.Option {
	return fx.Replace(construct)
}