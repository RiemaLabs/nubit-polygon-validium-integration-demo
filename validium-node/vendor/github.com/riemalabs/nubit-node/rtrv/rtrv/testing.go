package getters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/riemalabs/nubit-app/da/da"
	"github.com/riemalabs/rsmt2d"

	"github.com/riemalabs/nubit-node/strucs/eh"
	"github.com/riemalabs/nubit-node/strucs/eh/headertest"
	"github.com/riemalabs/nubit-node/da"
	"github.com/riemalabs/nubit-node/rtrv/eds/edstest"
)

// TestGetter provides a testing SingleEDSGetter and the root of the EDS it holds.
func TestGetter(t *testing.T) (share.Getter, *header.ExtendedHeader) {
	eds := edstest.RandEDS(t, 8)
	dah, err := share.NewRoot(eds)
	eh := headertest.RandExtendedHeaderWithRoot(t, dah)
	require.NoError(t, err)
	return &SingleEDSGetter{
		EDS: eds,
	}, eh
}

// SingleEDSGetter contains a single EDS where data is retrieved from.
// Its primary use is testing, and GetSharesByNamespace is not supported.
type SingleEDSGetter struct {
	EDS *rsmt2d.ExtendedDataSquare
}

// GetShare gets a share from a kept EDS if exist and if the correct root is given.
func (seg *SingleEDSGetter) GetShare(
	_ context.Context,
	header *header.ExtendedHeader,
	row, col int,
) (share.Share, error) {
	err := seg.checkRoot(header.DAH)
	if err != nil {
		return nil, err
	}
	return seg.EDS.GetCell(uint(row), uint(col)), nil
}

// GetEDS returns a kept EDS if the correct root is given.
func (seg *SingleEDSGetter) GetEDS(
	_ context.Context,
	header *header.ExtendedHeader,
) (*rsmt2d.ExtendedDataSquare, error) {
	err := seg.checkRoot(header.DAH)
	if err != nil {
		return nil, err
	}
	return seg.EDS, nil
}

// GetSharesByNamespace returns NamespacedShares from a kept EDS if the correct root is given.
func (seg *SingleEDSGetter) GetSharesByNamespace(context.Context, *header.ExtendedHeader, share.Namespace,
) (share.NamespacedShares, error) {
	panic("SingleEDSGetter: GetSharesByNamespace is not implemented")
}

func (seg *SingleEDSGetter) checkRoot(root *share.Root) error {
	dah, err := da.NewDataAvailabilityHeader(seg.EDS)
	if err != nil {
		return err
	}
	if !root.Equals(&dah) {
		return fmt.Errorf("unknown EDS: have %s, asked %s", dah.String(), root.String())
	}
	return nil
}