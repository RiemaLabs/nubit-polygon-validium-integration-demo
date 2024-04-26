package state

import (
	"github.com/riemalabs/go-header/sync"
	apptypes "github.com/riemalabs/nubit-app/utils/tk/sh/blob/types"

	"github.com/riemalabs/nubit-node/factory/core"
	header "github.com/riemalabs/nubit-node/strucs/eh"
	"github.com/riemalabs/nubit-node/strucs/state"
)

// coreAccessor constructs a new instance of state.Module over
// a nubit-core connection.
func coreAccessor(
	corecfg core.Config,
	signer *apptypes.KeyringSigner,
	sync *sync.Syncer[*header.ExtendedHeader],
) (*state.CoreAccessor, Module) {
	ca := state.NewCoreAccessor(signer, sync, corecfg.IP, corecfg.RPCPort, corecfg.GRPCPort)

	return ca, ca
}
