package types

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	// MaxBlockSizeBytes is the maximum permitted size of the blocks.
	MaxBlockSizeBytes = 104857600 // 100MB

	// BlockPartSizeBytes is the size of one block part.
	BlockPartSizeBytes = 65536 // 64kB
)

// ConsensusParams contains consensus critical parameters that determine the
// validity of blocks.
type ConsensusParams struct {
	BlockSize      `json:"block_size_params"`
	EvidenceParams `json:"evidence_params"`
}

// BlockSize contain limits on the block size.
type BlockSize struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
}

// EvidenceParams determine how we handle evidence of malfeasance
type EvidenceParams struct {
	MaxAge tmtime.DurationPretty `json:"max_age"` // only accept new evidence more recent than this
}

// DefaultConsensusParams returns a default ConsensusParams.
func DefaultConsensusParams() *ConsensusParams {
	return &ConsensusParams{
		DefaultBlockSize(),
		DefaultEvidenceParams(),
	}
}

// DefaultBlockSize returns a default BlockSize.
func DefaultBlockSize() BlockSize {
	return BlockSize{
		MaxBytes: 22020096, // 21MB
		MaxGas:   -1,
	}
}

// DefaultEvidenceParams Params returns a default EvidenceParams.
func DefaultEvidenceParams() EvidenceParams {
	return EvidenceParams{
		MaxAge: tmtime.DurationPretty{48 * time.Hour},
	}
}

// Validate validates the ConsensusParams to ensure all values are within their
// allowed limits, and returns an error if they are not.
func (params *ConsensusParams) Validate() error {
	if params.BlockSize.MaxBytes <= 0 {
		return cmn.NewError("BlockSize.MaxBytes must be greater than 0. Got %d",
			params.BlockSize.MaxBytes)
	}
	if params.BlockSize.MaxBytes > MaxBlockSizeBytes {
		return cmn.NewError("BlockSize.MaxBytes is too big. %d > %d",
			params.BlockSize.MaxBytes, MaxBlockSizeBytes)
	}

	if params.BlockSize.MaxGas < -1 {
		return cmn.NewError("BlockSize.MaxGas must be greater or equal to -1. Got %d",
			params.BlockSize.MaxGas)
	}

	if params.EvidenceParams.MaxAge.Duration <= 0 {
		return cmn.NewError("EvidenceParams.MaxAge must be greater than 0. Got %d",
			params.EvidenceParams.MaxAge.Duration)
	}

	return nil
}

// Hash returns a hash of the parameters to store in the block header
// No Merkle tree here, only three values are hashed here
// thus benefit from saving space < drawbacks from proofs' overhead
// Revisit this function if new fields are added to ConsensusParams
func (params *ConsensusParams) Hash() []byte {
	hasher := tmhash.New()
	bz := cdcEncode(params)
	if bz == nil {
		panic("cannot fail to encode ConsensusParams")
	}
	hasher.Write(bz)
	return hasher.Sum(nil)
}

// Update returns a copy of the params with updates from the non-zero fields of p2.
// NOTE: note: must not modify the original
func (params ConsensusParams) Update(params2 *abci.ConsensusParams) ConsensusParams {
	res := params // explicit copy

	if params2 == nil {
		return res
	}

	// we must defensively consider any structs may be nil
	if params2.BlockSize != nil {
		res.BlockSize.MaxBytes = params2.BlockSize.MaxBytes
		res.BlockSize.MaxGas = params2.BlockSize.MaxGas
	}
	if params2.EvidenceParams != nil {
		res.EvidenceParams.MaxAge = tmtime.DurationPretty{params2.EvidenceParams.MaxAge}
	}
	return res
}
