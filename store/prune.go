package store

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"
)

func (bs *BlockStore) flushBlockStore(batch dbm.Batch, base int64) error {
	bs.mtx.Lock()
	bs.base = base
	bs.height = base
	bs.mtx.Unlock()
	bs.saveState()

	err := batch.WriteSync()
	if err != nil {
		return fmt.Errorf("failed to prune up to height %v: %w", base, err)
	}
	return nil
}

func (bs *BlockStore) batchDeleteBlockStore(batch dbm.Batch, height int64) error {
	meta := bs.LoadBlockMeta(height)
	if meta == nil { // assume already deleted
		return fmt.Errorf("can not load last block")
	}
	if err := batch.Delete(calcBlockMetaKey(height)); err != nil {
		return err
	}
	if err := batch.Delete(calcBlockHashKey(meta.BlockID.Hash)); err != nil {
		return err
	}
	if err := batch.Delete(calcBlockCommitKey(height)); err != nil {
		return err
	}
	if err := batch.Delete(calcSeenCommitKey(height)); err != nil {
		return err
	}
	for p := 0; p < int(meta.BlockID.PartSetHeader.Total); p++ {
		if err := batch.Delete(calcBlockPartKey(height, p)); err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockStore) PruneLastBlock() (int64, error) {
	batch := bs.db.NewBatch()
	defer batch.Close()

	height := bs.Height()
	if err := bs.batchDeleteBlockStore(batch, height); err != nil {
		return 0, err
	}
	return height, bs.flushBlockStore(batch, height-1)
}
