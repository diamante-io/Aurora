package ethereum

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hcnet/go/services/bifrost/common"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
)

func (l *Listener) Start(rpcServer string) error {
	l.log = common.CreateLogger("EthereumListener")
	l.log.Info("EthereumListener starting")

	blockNumber, err := l.Storage.GetEthereumBlockToProcess()
	if err != nil {
		err = errors.Wrap(err, "Error getting ethereum block to process from DB")
		l.log.Error(err)
		return err
	}

	// Check if connected to correct network
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()
	id, err := l.Client.NetworkID(ctx)
	if err != nil {
		err = errors.Wrap(err, "Error getting ethereum network ID")
		l.log.Error(err)
		return err
	}

	if id.String() != l.NetworkID {
		return errors.Errorf("Invalid network ID (have=%s, want=%s)", id.String(), l.NetworkID)
	}

	go l.processBlocks(blockNumber)
	return nil
}

func (l *Listener) processBlocks(blockNumber uint64) {
	if blockNumber == 0 {
		l.log.Info("Starting from the latest block")
	} else {
		l.log.Infof("Starting from block %d", blockNumber)
	}

	// Time when last new block has been seen
	lastBlockSeen := time.Now()
	noBlockWarningLogged := false

	for {
		block, err := l.getBlock(blockNumber)
		if err != nil {
			l.log.WithFields(log.F{"err": err, "blockNumber": blockNumber}).Error("Error getting block")
			time.Sleep(1 * time.Second)
			continue
		}

		// Block doesn't exist yet
		if block == nil {
			if time.Since(lastBlockSeen) > 3*time.Minute && !noBlockWarningLogged {
				l.log.Warn("No new block in more than 3 minutes")
				noBlockWarningLogged = true
			}

			time.Sleep(1 * time.Second)
			continue
		}

		// Reset counter when new block appears
		lastBlockSeen = time.Now()
		noBlockWarningLogged = false

		if block.NumberU64() == 0 {
			l.log.Error("Etheruem node is not synced yet. Unable to process blocks")
			time.Sleep(30 * time.Second)
			continue
		}

		err = l.processBlock(block)
		if err != nil {
			l.log.WithFields(log.F{"err": err, "blockNumber": block.NumberU64()}).Error("Error processing block")
			time.Sleep(1 * time.Second)
			continue
		}

		// Persist block number
		err = l.Storage.SaveLastProcessedEthereumBlock(blockNumber)
		if err != nil {
			l.log.WithField("err", err).Error("Error saving last processed block")
			time.Sleep(1 * time.Second)
			// We continue to the next block
		}

		blockNumber = block.NumberU64() + 1
	}
}

// getBlock returns (nil, nil) if block has not been found (not exists yet)
func (l *Listener) getBlock(blockNumber uint64) (*types.Block, error) {
	var blockNumberInt *big.Int
	if blockNumber > 0 {
		blockNumberInt = big.NewInt(int64(blockNumber))
	}

	d := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	block, err := l.Client.BlockByNumber(ctx, blockNumberInt)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		err = errors.Wrap(err, "Error getting block from geth")
		l.log.WithField("block", blockNumberInt.String()).Error(err)
		return nil, err
	}

	return block, nil
}

func (l *Listener) processBlock(block *types.Block) error {
	transactions := block.Transactions()
	blockTime := time.Unix(block.Time().Int64(), 0)

	localLog := l.log.WithFields(log.F{
		"blockNumber":  block.NumberU64(),
		"blockTime":    blockTime,
		"transactions": len(transactions),
	})
	localLog.Info("Processing block")

	for _, transaction := range transactions {
		to := transaction.To()
		if to == nil {
			// Contract creation
			continue
		}

		tx := Transaction{
			Hash:     transaction.Hash().Hex(),
			ValueWei: transaction.Value(),
			To:       to.Hex(),
		}
		err := l.TransactionHandler(tx)
		if err != nil {
			return errors.Wrap(err, "Error processing transaction")
		}
	}

	localLog.Info("Processed block")

	return nil
}
