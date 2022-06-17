package configurer

import (
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/containers"
)

type Configurer interface {
	ConfigureChains() error

	ClearResources() error

	GetChainConfig(chainIndex int) ChainConfig

	RunSetup() error

	RunValidators() error

	RunIBC() error

	SendIBC(srcChain *chain.Chain, dstChain *chain.Chain, recipient string, token sdk.Coin)

	CreatePool(chainId string, valIdx int, poolFile string)
}

var (
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "everything",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   0,
			SnapshotKeepRecent: 0,
		},
	}
	validatorConfigsChainB = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
	}
)

// New returns a new Configurer depending on the values of its parameters.
// - If only isIBCEnabled, we want to have 2 chains initialized at the current
// Git branch version of Osmosis codebase.
// - If only isUpgradeEnabled, that is invalid and an error is returned.
// - If both isIBCEnabled and isUpgradeEnabled, we want 2 chains with IBC initialized
// at the previous Osmosis version.
// - If !isIBCEnabled and !isUpgradeEnabled, we only need one chain at the current
// Git branch version of the Osmosis code.
func New(t *testing.T, isIBCEnabled, isUpgradeEnabled bool) (Configurer, error) {
	containerManager, err := containers.NewManager(isUpgradeEnabled)
	if err != nil {
		return nil, err
	}

	if isIBCEnabled && isUpgradeEnabled {
		// skip none - configure two chains via Docker
		// to utilize the older version of osmosis to upgrade from
		return NewUpgradeConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
					skipRunValidatorIndexes: map[int]struct{}{
						3: {}, // skip validator at index 3
					},
				},
				{
					chainId:         chain.ChainBID,
					validatorConfig: validatorConfigsChainB,
				},
			},
			withUpgrade(withIBC(baseSetup)), // base set up with IBC and upgrade
			containerManager,
		), nil
	} else if isIBCEnabled {
		// configure two chains from current Git branch
		return NewCurrentBranchConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
					skipRunValidatorIndexes: map[int]struct{}{
						3: {}, // skip validator at index 3
					},
				},
				{
					chainId:         chain.ChainBID,
					validatorConfig: validatorConfigsChainB,
				},
			},
			withIBC(baseSetup), // base set up with IBC
			containerManager,
		), nil
	} else if isUpgradeEnabled {
		// invalid - IBC tests must be enabled for upgrade
		// to function
		return nil, errors.New("IBC tests must be enabled for upgrade to work")
	} else {
		// configure one chain from current Git branch
		return NewCurrentBranchConfigurer(t,
			[]*ChainConfig{
				{
					chainId:         chain.ChainAID,
					validatorConfig: validatorConfigsChainA,
				},
			},
			baseSetup, // base set up only
			containerManager,
		), nil
	}
}
