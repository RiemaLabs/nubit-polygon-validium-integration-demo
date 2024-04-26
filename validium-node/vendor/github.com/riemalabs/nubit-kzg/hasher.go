package kzg

import (
	"bytes"
	"errors"
	"fmt"
	"hash"

	"github.com/riemalabs/nubit-kzg/namespace"
)

var _ hash.Hash = (*NmtHasher)(nil)

var (
	ErrUnorderedSiblings         = errors.New("NMT sibling nodes should be ordered lexicographically by namespace IDs")
	ErrInvalidNodeLen            = errors.New("invalid NMT node size")
	ErrInvalidLeafLen            = errors.New("invalid NMT leaf size")
	ErrInvalidNodeNamespaceOrder = errors.New("invalid NMT node namespace order")
)

// Hasher describes the interface nmts use to hash leafs and nodes.
//
// Note: it is not advised to create alternative hashers if following the
// specification is desired. The main reason this exists is to not follow the
// specification for testing purposes.
type Hasher interface {
	NamespaceSize() namespace.IDSize
	HashLeaf(data []byte) ([]byte, error)
	EmptyRoot() []byte

	Commit(data [][]byte) ([]byte, error)
	// Verify returns an error if the proof is invalid.
	Verify(commit []byte, proof KzgOpen) error
	// Open returns the proof for the leaf at the given index.
	Open(data [][]byte, index int) (KzgProof, []byte, error)
}

var _ Hasher = &NmtHasher{}

// NmtHasher is the default hasher. It follows the description of the original
// hashing function described in the LazyLedger white paper.
type NmtHasher struct {
	*KzgCommitter
	NamespaceLen namespace.IDSize
}

func NewNmtHasher(baseHasher hash.Hash, nidLen namespace.IDSize) *NmtHasher {
	committer := NewKzgCommitter(baseHasher)
	return &NmtHasher{
		KzgCommitter: committer,
		NamespaceLen: nidLen,
	}
}

func (n *NmtHasher) NamespaceSize() namespace.IDSize {
	return n.NamespaceLen
}

func (n *NmtHasher) EmptyRoot() []byte {
	res := make([]byte, 2, int(n.NamespaceSize())*2+n.KzgCommitter.Size())
	ccommit := n.KzgCommitter.EmptyCommitment()
	res = append(res, ccommit...)
	return res
}

// ValidateNodeFormat checks whether the supplied node conforms to the
// namespaced hash format and returns ErrInvalidNodeLen if not.
func (n *NmtHasher) ValidateCommitFormat(commit []byte) (err error) {
	expectedNodeLen := n.CommentSize() + int(n.NamespaceSize())*2
	nodeLen := len(commit)
	if nodeLen != expectedNodeLen {
		return fmt.Errorf("%w: got: %v, want %v", ErrInvalidNodeLen, nodeLen, expectedNodeLen)
	}
	// check the namespace order
	minNID := namespace.ID(MinNamespace(commit, n.NamespaceSize()))
	maxNID := namespace.ID(MaxNamespace(commit, n.NamespaceSize()))
	if maxNID.Less(minNID) {
		return fmt.Errorf("%w: max namespace ID %d is less than min namespace ID %d ", ErrInvalidNodeNamespaceOrder, maxNID, minNID)
	}
	return nil
}

// ValidateLeaf verifies if data is namespaced and returns an error if not.
func (n *NmtHasher) ValidateLeaf(data []byte) (err error) {
	nidSize := int(n.NamespaceSize())
	lenData := len(data)
	if lenData < nidSize {
		return fmt.Errorf("%w: got: %v, want >= %v", ErrInvalidLeafLen, lenData, nidSize)
	}
	return nil
}

func (n *NmtHasher) Verify(commit []byte, proof KzgOpen) error {
	if err := n.ValidateCommitFormat(commit); err != nil {
		return err
	}
	commit = commit[2*int(n.NamespaceSize()):]
	if err := n.ValidateLeaf(proof.Value()); err != nil {
		return err
	}
	return n.KzgCommitter.Verify(commit, proof)
}

func max(ns []byte, ns2 []byte) []byte {
	if bytes.Compare(ns, ns2) >= 0 {
		return ns
	}
	return ns2
}

func min(ns []byte, ns2 []byte) []byte {
	if bytes.Compare(ns, ns2) <= 0 {
		return ns
	}
	return ns2
}

// computeNsRange computes the namespace range of the parent node based on the namespace ranges of its left and right children.
func computeNsRange(leftMinNs, leftMaxNs, rightMinNs, rightMaxNs []byte, ignoreMaxNs bool, precomputedMaxNs namespace.ID) (minNs []byte, maxNs []byte) {
	minNs = leftMinNs
	maxNs = rightMaxNs
	if ignoreMaxNs && bytes.Equal(precomputedMaxNs, rightMinNs) {
		maxNs = leftMaxNs
	}
	return minNs, maxNs
}

var _ hash.Hash = (*EmptyKzgHash)(nil)

type EmptyKzgHash struct {
	NamespaceLen namespace.IDSize
}

func (e *EmptyKzgHash) Write(p []byte) (n int, err error) {
	panic("not implement")
}

func (e *EmptyKzgHash) Sum(b []byte) []byte {
	panic("not implement")
}

func (e *EmptyKzgHash) Reset() {

}

func (e *EmptyKzgHash) Size() int {
	return 2*int(e.NamespaceLen) + CommitmentSize
}

func (e *EmptyKzgHash) BlockSize() int {
	panic("not implement")

}
