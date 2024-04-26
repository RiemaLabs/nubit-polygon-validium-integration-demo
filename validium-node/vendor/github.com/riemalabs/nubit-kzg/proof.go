package kzg

import (
	"errors"
)

// ErrFailedCompletenessCheck indicates that the verification of a namespace proof failed due to the lack of completeness property.
var ErrFailedCompletenessCheck = errors.New("failed completeness check")

// Proof represents a namespace proof of a namespace.ID in an NMT. In case this
// proof proves the absence of a namespace.ID in a tree it also contains the
// leaf hashes of the range where that namespace would be.
// type Proof struct {
// 	// start index of the leaves that match the queried namespace.ID.
// 	start int
// 	// end index (non-inclusive) of the leaves that match the queried
// 	// namespace.ID.
// 	end int
// 	// nodes hold the tree nodes necessary for the Merkle range proof of
// 	// `[start, end)` in the order of an in-order traversal of the tree. in
// 	// specific, nodes contain: 1) the namespaced hash of the left siblings for
// 	// the Merkle inclusion proof of the `start` leaf 2) the namespaced hash of
// 	// the right siblings of the Merkle inclusion proof of  the `end` leaf
// 	nodes [][]byte
// 	// leafHash are nil if the namespace is present in the NMT. In case the
// 	// namespace to be proved is in the min/max range of the tree but absent,
// 	// this will contain the leaf hash necessary to verify the proof of absence.
// 	// leafHash contains a tree leaf that 1) its namespace ID is the smallest
// 	// namespace ID larger than nid and 2) the namespace ID of the leaf to the
// 	// left of it is smaller than the nid.
// 	leafHash []byte
// 	// isMaxNamespaceIDIgnored is set to true if the tree from which this Proof
// 	// was generated from is initialized with Options.IgnoreMaxNamespace ==
// 	// true. The IgnoreMaxNamespace flag influences the calculation of the
// 	// namespace ID range for intermediate nodes in the tree. This flag signals
// 	// that, when determining the upper limit of the namespace ID range for a
// 	// tree node, the maximum possible namespace ID (equivalent to
// 	// "NamespaceIDSize" bytes of 0xFF, or 2^NamespaceIDSize-1) should be
// 	// omitted if feasible. For a more in-depth understanding of this field,
// 	// refer to the "HashNode" method in the "Hasher.
// 	isMaxNamespaceIDIgnored bool
// }

// func (proof Proof) MarshalJSON() ([]byte, error) {
// 	pbProofObj := pb.Proof{
// 		Start:                 int64(proof.start),
// 		End:                   int64(proof.end),
// 		Nodes:                 proof.nodes,
// 		LeafHash:              proof.leafHash,
// 		IsMaxNamespaceIgnored: proof.isMaxNamespaceIDIgnored,
// 	}
// 	return json.Marshal(pbProofObj)
// }

// func (proof *Proof) UnmarshalJSON(data []byte) error {
// 	var pbProof pb.Proof
// 	err := json.Unmarshal(data, &pbProof)
// 	if err != nil {
// 		return err
// 	}
// 	proof.start = int(pbProof.Start)
// 	proof.end = int(pbProof.End)
// 	proof.nodes = pbProof.Nodes
// 	proof.leafHash = pbProof.LeafHash
// 	proof.isMaxNamespaceIDIgnored = pbProof.IsMaxNamespaceIgnored
// 	return nil
// }

// Start index of this proof.
// func (proof Proof) Start() int {
// 	return proof.start
// }

// // End index of this proof, non-inclusive.
// func (proof Proof) End() int {
// 	return proof.end
// }

// // Nodes return the proof nodes that together with the corresponding leaf values
// // can be used to recompute the root and verify this proof.
// func (proof Proof) Nodes() [][]byte {
// 	return proof.nodes
// }

// // IsOfAbsence returns true if this proof proves the absence of leaves of a
// // namespace in the tree.
func (proof NamespaceRangeProof) IsOfAbsence() bool {
	return !proof.InclusionOrAbsence
}

// // LeafHash returns nil if the namespace has leaves in the NMT. In case the
// // namespace.ID to be proved is in the min/max range of the tree but absent,
// // this will contain the leaf hash necessary to verify the proof of absence.
// func (proof Proof) LeafHash() []byte {
// 	return proof.leafHash
// }

// IsNonEmptyRange returns true if this proof contains a valid, non-empty proof
// range.
// func (proof NamespaceRangeProof) IsNonEmptyRange() bool {
// 	return proof.start >= 0 && proof.start < proof.end
// }

// IsMaxNamespaceIDIgnored returns true if the proof has been created under the ignore max namespace logic.
// see ./docs/nmt-lib.md for more details.
// func (proof Proof) IsMaxNamespaceIDIgnored() bool {
// 	return proof.isMaxNamespaceIDIgnored
// }

// NewEmptyRangeProof constructs a proof that proves that a namespace.ID does
// not fall within the range of an NMT.
func NewEmptyRangeProof() NamespaceRangeProof {
	return NamespaceRangeProof{0, 0, 0, 0, KzgOpen{}, KzgOpen{}, KzgOpen{}, KzgOpen{}, false}
}

// NewInclusionProof constructs a proof that proves that a namespace.ID is
// included in an NMT.
func NewInclusionProof(proofStart, proofEnd, preIndex, postIndex int, startProof, endProof, preProof, postProof KzgOpen) NamespaceRangeProof {
	return NamespaceRangeProof{proofStart, proofEnd, preIndex, postIndex, startProof, endProof, preProof, postProof, true}
}

// NewAbsenceProof constructs a proof that proves that a namespace.ID falls
// within the range of an NMT but no leaf with that namespace.ID is included.
func NewAbsenceProof(pre, post int, preProof, postProof KzgOpen) NamespaceRangeProof {
	return NamespaceRangeProof{0, 0, pre, post, KzgOpen{}, KzgOpen{}, preProof, postProof, false}
}

// IsEmptyProof checks whether the proof corresponds to an empty proof as defined in NMT specifications https://github.com/celestiaorg/nmt/blob/master/docs/spec/kzg.md.
func (proof NamespaceRangeProof) IsEmptyProof() bool {
	return proof.start == proof.end && proof.preIndex == proof.postIndex
}

// VerifyNamespace verifies a whole namespace, i.e. 1) it verifies inclusion of
// the provided `leaves` in the tree (or the proof.leafHash in case of
// full/short absence proof) 2) it verifies that the namespace is complete
// i.e., the data items matching the namespace `nID`  are within the range
// [`proof.start`, `proof.end`) and no data of that namespace was left out.
// VerifyNamespace deems an empty `proof` valid if the queried `nID` falls
// outside the namespace  range of the supplied `root` or if the `root` is empty
//
// `h` MUST be the same as the underlying hash function used to generate the
// proof. Otherwise, the verification will fail. `nID` is the namespace ID for
// which the namespace `proof` is generated. `leaves` contains the namespaced
// leaves of the tree in the range of [`proof.start`, `proof.end`).
// For an absence `proof`, the `leaves` is empty.
// `leaves` items MUST be ordered according to their index in the tree,
// with `leaves[0]` corresponding to the namespaced leaf at index `start`,
// and the last element in `leaves` corresponding to the leaf at index `end-1`
// of the tree.
//
// `root` is the root of the NMT against which the `proof` is verified.
// func (proof NamespaceRangeProof) VerifyNamespace(h hash.Hash, nID namespace.ID, leaves [][]byte, root []byte) bool {

// 	// check that all the proof.nodes are valid w.r.t the NMT hasher
// 	for _, node := range proof.nodes {
// 		if err := nth.ValidateNodeFormat(node); err != nil {
// 			return false
// 		}
// 	}

// 	// if the proof is an absence proof, the leafHash must be valid w.r.t the NMT hasher
// 	if proof.IsOfAbsence() {
// 		if err := nth.ValidateNodeFormat(proof.leafHash); err != nil {
// 			return false
// 		}
// 	}

// 	isEmptyRange := proof.start == proof.end
// 	if isEmptyRange {
// 		if proof.IsEmptyProof() && len(leaves) == 0 {
// 			rootMin := namespace.ID(MinNamespace(root, nIDLen))
// 			rootMax := namespace.ID(MaxNamespace(root, nIDLen))
// 			// empty proofs are always rejected unless 1) nID is outside the range of
// 			// namespaces covered by the root 2) the root represents an empty tree, since
// 			// it purports to cover the zero namespace but does not actually include
// 			// any such nodes
// 			if nID.Less(rootMin) || rootMax.Less(nID) {
// 				return true
// 			}
// 			if bytes.Equal(root, nth.EmptyRoot()) {
// 				return true
// 			}
// 			return false
// 		}
// 		// the proof range is empty, and invalid
// 		return false
// 	}

// 	gotLeafHashes := make([][]byte, 0, len(leaves))
// 	if proof.IsOfAbsence() {
// 		gotLeafHashes = append(gotLeafHashes, proof.leafHash)
// 		// conduct some sanity checks:
// 		leafMinNID := namespace.ID(proof.leafHash[:nIDLen])
// 		if !nID.Less(leafMinNID) {
// 			// leafHash.minNID  must be greater than nID
// 			return false
// 		}

// 	} else {
// 		// collect leaf hashes from provided data and do some sanity checks:
// 		hashLeafFunc := nth.HashLeaf
// 		for _, gotLeaf := range leaves {
// 			if nth.ValidateLeaf(gotLeaf) != nil {
// 				return false
// 			}
// 			// check whether the namespace ID of the data matches the queried nID
// 			if gotLeafNid := namespace.ID(gotLeaf[:nIDLen]); !gotLeafNid.Equal(nID) {
// 				// conflicting namespace IDs in data
// 				return false
// 			}
// 			// hash the leaf data
// 			leafHash, err := hashLeafFunc(gotLeaf)
// 			if err != nil { // this can never happen due to the initial validation of the leaf at the beginning of the loop
// 				return false
// 			}
// 			gotLeafHashes = append(gotLeafHashes, leafHash)
// 		}
// 	}
// 	// check whether the number of leaves match the proof range i.e., end-start.
// 	// If not, make an early return.
// 	expectedLeafCount := proof.End() - proof.Start()
// 	if !proof.IsOfAbsence() && len(gotLeafHashes) != expectedLeafCount {
// 		return false
// 	}
// 	// with verifyCompleteness set to true:
// 	res, err := proof.VerifyLeafHashes(nth, true, nID, gotLeafHashes, root)
// 	if err != nil {
// 		return false
// 	}
// 	return res
// }

// ProtoToProof creates a proof from its proto representation.
// func ProtoToProof(protoProof pb.Proof) Proof {
// 	if protoProof.Start == 0 && protoProof.End == 0 {
// 		return NewEmptyRangeProof(protoProof.IsMaxNamespaceIgnored)
// 	}

// 	if len(protoProof.LeafHash) > 0 {
// 		return NewAbsenceProof(
// 			int(protoProof.Start),
// 			int(protoProof.End),
// 			protoProof.Nodes,
// 			protoProof.LeafHash,
// 			protoProof.IsMaxNamespaceIgnored,
// 		)
// 	}

// 	return NewInclusionProof(
// 		int(protoProof.Start),
// 		int(protoProof.End),
// 		protoProof.Nodes,
// 		protoProof.IsMaxNamespaceIgnored,
// 	)
// }
