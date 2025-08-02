package trie

//import (
//	"errors"
//	"frizo-blockchain/common"
//	"github.com/ethereum/go-ethereum/log"
//)
//
//// MPTProof represents a Merkle proof in MPT
//type MPTProof struct {
//	Key      []byte      // key -> to path
//	Value    []byte      // value data
//	Proof    [][]byte    // encode node in mpt path
//	RootHash common.Hash // mpt root hash
//}
//
//// ProofNode represents a decoded node in the proof path
//type MPTProofNode struct {
//	NodeType MPTNodeType
//	Path     []byte
//	Value    []byte
//	Children []common.Hash // for branch node, to store all child hash
//}
//
//// GenerateProof generate MPTProof by key
//func (mpt *ModifiedMerklePatriciaTree) GenerateProof(key []byte) (*MPTProof, error) {
//	if len(key) == 0 {
//		return nil, errors.New("key is empty")
//	}
//
//	nibbles := KeyToHex(key)
//	proof := &MPTProof{
//		Key:      key,
//		RootHash: *mpt.root.Hash,
//		Proof:    make([][]byte, 0),
//	}
//
//	// collect proof nodes ([][]byte) alone the path
//	value, err := mpt.collectProof(nibbles, &proof.Proof)
//	if err != nil {
//		log.Error("Failed to collect proof", "err", err)
//		return nil, errors.New("Failed to collect proof")
//	}
//
//	proof.Value = value
//	return proof, nil
//}
//
//// collectProof recursive call collect mpt proof and store into proof([][]byte) -> Return Leaf or Branch value data
//func (mpt *ModifiedMerklePatriciaTree) collectProof(node *MPTNode, path []byte, proof *[][]byte) ([]byte, error) {
//	if node == nil {
//		return nil, nil
//	}
//
//	// Encode current node to proof
//	encodeVal := mpt.encodeNode(node)
//	*proof = append(*proof, encodeVal)
//
//}
//
//// encodeNodeForProof encodes a node for inclusion in a proof
//// This must match the encoding used in MPT for hash calculation
//func encodeNodeForProof(node *MPTNode) []byte {
//
//}
