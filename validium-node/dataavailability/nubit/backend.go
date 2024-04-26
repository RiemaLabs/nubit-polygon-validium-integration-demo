package nubit

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygondatacommittee"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/riemalabs/nubit-app/da/namespace"
	share "github.com/riemalabs/nubit-node/da"
	client "github.com/riemalabs/nubit-node/rpc/rpc/client"
	nodeBlob "github.com/riemalabs/nubit-node/strucs/btx"
)

// // DABackender is an interface for components that store and retrieve batch data
// type DABackender interface {
// 	SequenceRetriever
// 	SequenceSender
// 	// Init initializes the DABackend
// 	Init() error
// }

// // SequenceSender is used to send provided sequence of batches
// type SequenceSender interface {
// 	// PostSequence sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
// 	// as expected by the contract
// 	PostSequence(ctx context.Context, batchesData [][]byte) ([]byte, error)
// }

// // SequenceRetriever is used to retrieve batch data
// type SequenceRetriever interface {
// 	// GetSequence retrieves the sequence data from the data availability backend
// 	GetSequence(ctx context.Context, batchHashes []common.Hash, dataAvailabilityMessage []byte) ([][]byte, error)
// }

type NubitDABackend struct {
	config              *Config
	attestationContract *polygondatacommittee.Polygondatacommittee
	ns                  namespace.Namespace
	client              *client.Client
}

func NewNubitDABackend(l1RPCURL string, dataCommitteeAddr common.Address) (*NubitDABackend, error) {
	var config Config
	err := config.GetConfig("/app/nubit-config.json")
	if err != nil {
		log.Fatalf("cannot get config:%w", err)
	}

	ethClient, err := ethclient.Dial(l1RPCURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", l1RPCURL, err)
		return nil, err
	}

	log.Infof("⚙️     Nubit config : %#v ", config)

	attestationContract, err := polygondatacommittee.NewPolygondatacommittee(dataCommitteeAddr, ethClient)
	if err != nil {
		return nil, err
	}
	cn, err := client.NewClient(context.TODO(), config.RpcURL, config.AuthKey)
	if err != nil {
		return nil, err
	}
	return &NubitDABackend{
		config:              &config,
		attestationContract: attestationContract,
		ns:                  namespace.MustNewV0([]byte(config.Namespace)),
		client:              cn,
	}, nil
}

func (a *NubitDABackend) Init() error {
	return nil
}

// PostSequence sends the sequence data to the data availability backend, and returns the dataAvailabilityMessage
// as expected by the contract
func (a *NubitDABackend) PostSequence(ctx context.Context, batchesData [][]byte) ([]byte, error) {
	encodedData, err := MarshalBatchData(batchesData)
	if err != nil {
		log.Errorf("🏆    NubitDABackend.MarshalBatchData:%s", err)
		return encodedData, err
	}

	nsp, err := share.NamespaceFromBytes(a.ns.Bytes())
	if nil != err {
		log.Errorf("🏆    NubitDABackend.NamespaceFromBytes:%s", err)
		return nil, err
	}

	body, err := nodeBlob.NewBlobV0(nsp, encodedData)
	if nil != err {
		log.Errorf("🏆    NubitDABackend.NewBlobV0:%s", err)
		return nil, err
	}

	log.Infof("🏆     Nubit send data:%+v", body)

	blockNumber, err := a.client.Blob.Submit(ctx, []*nodeBlob.Blob{body}, 0.01)
	if err != nil {
		log.Errorf("🏆    NubitDABackend.Submit:%s", err)
		return nil, err
	}

	// todo: May be need to sleep
	//dataProof, err := a.client.Blob.GetProof(ctx, uint64(blockNumber), a.ns.Bytes(), body.Commitment)
	//if err != nil {
	//	log.Errorf("🏆    NubitDABackend.GetProof:%s", err)
	//	return nil, err
	//}
	//
	//log.Infof("🏆   Nubit received data proof:%+v", dataProof)

	var batchDAData BatchDAData
	batchDAData.Commitment = body.Commitment

	batchDAData.BlockNumber = big.NewInt(int64(blockNumber))
	log.Infof("🏆  Nubit prepared DA data:%+v", batchDAData)

	// todo: use bridge API data
	returnData, err := batchDAData.Encode()
	if err != nil {
		return nil, fmt.Errorf("🏆  Nubit cannot encode batch data:%w", err)
	}

	log.Infof("🏆  Nubit Data submitted by sequencer:%d bytes against namespace %v sent with height %#x", len(encodedData), a.ns, blockNumber)

	return returnData, nil
}

func (a *NubitDABackend) GetSequence(ctx context.Context, batchHashes []common.Hash, dataAvailabilityMessage []byte) ([][]byte, error) {
	var batchDAData BatchDAData
	err := batchDAData.Decode(dataAvailabilityMessage)
	if err != nil {
		log.Errorf("🏆    NubitDABackend.GetSequence.Decode:%s", err)
		return nil, err
	}
	log.Infof("🏆     Nubit GetSequence batchDAData:%+v", batchDAData)
	blob, err := a.client.Blob.Get(context.TODO(), batchDAData.BlockNumber.Uint64(), a.ns.Bytes(), batchDAData.Commitment)
	if err != nil {
		log.Errorf("🏆    NubitDABackend.GetSequence.Blob.Get:%s", err)
		return nil, err
	}
	log.Infof("🏆     Nubit GetSequence blob.data:%+v", blob.GetData())
	return UnmarshalBatchData(blob.GetData())
}
