package launcher

import (
	"context"
	"path"
	"fmt"
	"time"


	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"
	"github.com/Fantom-foundation/go-opera/integration"
)


func writeEVMToErigon(ctx *cli.Context) error {

	start := time.Now()
	log.Info("Writing of EVM accounts into Erigon database started")
	// initiate erigon lmdb
	db, tmpDir, err := erigon.SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	log.Info("Initializing raw producer")
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	if gdb.GetHighestLamport() != 0 {
		log.Warn("Attempting genesis export not in a beginning of an epoch. Genesis file output may contain excessive data.")
	}
	defer gdb.Close()

	chaindb := gdb.EvmStore().EvmDb
	root := common.Hash(gdb.GetBlockState().FinalizedStateRoot)
	
	lastBlockIdx := gdb.GetBlockState().LastBlock.Idx
	mptFlag := ctx.String(mptTraversalMode.Name)

	log.Info("Generate Erigon Plain State...")
	if err := erigon.GeneratePlainState(mptFlag, root, chaindb, db, lastBlockIdx); err != nil {
		return err
	}
	log.Info("Generation of Erigon Plain State is complete")
	
	log.Info("Generate Erigon Hash State")
	if err := erigon.GenerateHashedState("HashedState", db, tmpDir, context.Background()); err != nil {
		log.Error("GenerateHashedState error: ", err)
		return err
	}
	log.Info("Generation Hash State is complete")


	log.Info("Generate Intermediate Hashes state and compute State Root")
	trieCfg := erigon.StageTrieCfg(db, true, true, "", nil)
	hash, err := erigon.ComputeStateRoot("Intermediate Hashes", db, trieCfg)
	if err != nil {
		log.Error("GenerateIntermediateHashes error: ", err)
		return err
	}
	log.Info(fmt.Sprintf("[%s] Trie root", "GenerateStateRoot"), "hash", hash.Hex())
	log.Info("Generation of Intermediate Hashes state and computation of State Root Complete")
	
	log.Info("Writing of EVM accounts into Erigon database completed", "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
}





