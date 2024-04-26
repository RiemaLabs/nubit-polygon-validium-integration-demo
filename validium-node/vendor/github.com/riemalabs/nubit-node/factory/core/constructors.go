package core

import (
	"github.com/riemalabs/nubit-node/rpc/core"
)

func remote(cfg Config) (core.Client, error) {
	return core.NewRemote(cfg.IP, cfg.RPCPort)
}
