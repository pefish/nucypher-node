package command

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pefish/go-commander"
	go_config "github.com/pefish/go-config"
	"github.com/pefish/go-decimal"
	go_logger "github.com/pefish/go-logger"
	go_reflect "github.com/pefish/go-reflect"
	go_coin_eth "github.com/pefish/nucypher-node/pkg/go-coin-eth"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type DefaultCommand struct {
	cache Data
	cacheFs *os.File
}

func NewDefaultCommand() *DefaultCommand {
	return &DefaultCommand{}
}

func (dc *DefaultCommand) DecorateFlagSet(flagSet *flag.FlagSet) error {
	flagSet.String("pkey", "", "private key of your worker address")
	flagSet.String("gas-price", "15", "your accepted gas price")
	flagSet.String("server-url", "", "geth server url")
	flagSet.String("interval", "30", "interval time of minutes")
	flagSet.String("staker-address", "0x476fBB25d56B5dD4f1df03165498C403C4713069", "staker address")
	return nil
}

func (dc *DefaultCommand) OnExited() error {
	b, err := json.Marshal(dc.cache)
	if err != nil {
		return err
	}
	err = dc.cacheFs.Truncate(0)
	if err != nil {
		return err
	}
	_, err = dc.cacheFs.WriteAt(b, 0)
	if err != nil {
		return err
	}
	err = dc.cacheFs.Sync()
	if err != nil {
		return err
	}
	err = dc.cacheFs.Close()
	if err != nil {
		return err
	}
	return nil
}

type Data struct {
	Period string `json:"period"`
	TxHash string `json:"tx_hash"`
	Nonce uint64 `json:"nonce"`
	GasPrice string `json:"gas_price"`
}

var abiStr = `[{"inputs":[{"internalType":"contract NuCypherToken","name":"_token","type":"address"},{"internalType":"uint32","name":"_hoursPerPeriod","type":"uint32"},{"internalType":"uint256","name":"_issuanceDecayCoefficient","type":"uint256"},{"internalType":"uint256","name":"_lockDurationCoefficient1","type":"uint256"},{"internalType":"uint256","name":"_lockDurationCoefficient2","type":"uint256"},{"internalType":"uint16","name":"_maximumRewardedPeriods","type":"uint16"},{"internalType":"uint256","name":"_firstPhaseTotalSupply","type":"uint256"},{"internalType":"uint256","name":"_firstPhaseMaxIssuance","type":"uint256"},{"internalType":"uint16","name":"_minLockedPeriods","type":"uint16"},{"internalType":"uint256","name":"_minAllowableLockedTokens","type":"uint256"},{"internalType":"uint256","name":"_maxAllowableLockedTokens","type":"uint256"},{"internalType":"uint16","name":"_minWorkerPeriods","type":"uint16"},{"internalType":"bool","name":"_isTestContract","type":"bool"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":true,"internalType":"uint16","name":"period","type":"uint16"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"CommitmentMade","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"periods","type":"uint16"}],"name":"Deposited","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"oldValue","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"lastPeriod","type":"uint16"},{"indexed":false,"internalType":"uint256","name":"newValue","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"periods","type":"uint16"}],"name":"Divided","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Donated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"reservedReward","type":"uint256"}],"name":"Initialized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"firstPeriod","type":"uint16"},{"indexed":false,"internalType":"uint16","name":"periods","type":"uint16"}],"name":"Locked","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"value1","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"value2","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"lastPeriod","type":"uint16"}],"name":"Merged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":true,"internalType":"uint16","name":"period","type":"uint16"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Minted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"},{"indexed":false,"internalType":"uint16","name":"lastPeriod","type":"uint16"},{"indexed":false,"internalType":"uint16","name":"periods","type":"uint16"}],"name":"Prolonged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint16","name":"lockUntilPeriod","type":"uint16"}],"name":"ReStakeLocked","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"bool","name":"reStake","type":"bool"}],"name":"ReStakeSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"penalty","type":"uint256"},{"indexed":true,"internalType":"address","name":"investigator","type":"address"},{"indexed":false,"internalType":"uint256","name":"reward","type":"uint256"}],"name":"Slashed","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"bool","name":"snapshotsEnabled","type":"bool"}],"name":"SnapshotSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"testTarget","type":"address"},{"indexed":false,"internalType":"address","name":"sender","type":"address"}],"name":"StateVerified","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"target","type":"address"},{"indexed":false,"internalType":"address","name":"sender","type":"address"}],"name":"UpgradeFinished","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"bool","name":"windDown","type":"bool"}],"name":"WindDownSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Withdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":false,"internalType":"bool","name":"measureWork","type":"bool"}],"name":"WorkMeasurementSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"staker","type":"address"},{"indexed":true,"internalType":"address","name":"worker","type":"address"},{"indexed":true,"internalType":"uint16","name":"startPeriod","type":"uint16"}],"name":"WorkerBonded","type":"event"},{"inputs":[],"name":"MAX_SUB_STAKES","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"adjudicator","outputs":[{"internalType":"contract AdjudicatorInterface","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"balanceHistory","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address[]","name":"_stakers","type":"address[]"},{"internalType":"uint256[]","name":"_numberOfSubStakes","type":"uint256[]"},{"internalType":"uint256[]","name":"_values","type":"uint256[]"},{"internalType":"uint16[]","name":"_periods","type":"uint16[]"}],"name":"batchDeposit","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_worker","type":"address"}],"name":"bondWorker","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"commitToNextPeriod","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"currentMintingPeriod","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"currentPeriodSupply","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_value","type":"uint256"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"deposit","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_index","type":"uint256"},{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"depositAndIncrease","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_value","type":"uint256"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"depositFromWorkLock","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_index","type":"uint256"},{"internalType":"uint256","name":"_newValue","type":"uint256"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"divideStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"donate","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint16","name":"_period","type":"uint16"}],"name":"findIndexOfPastDowntime","outputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"}],"name":"finishUpgrade","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"firstPhaseMaxIssuance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"firstPhaseTotalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint16","name":"_periods","type":"uint16"},{"internalType":"uint256","name":"_startIndex","type":"uint256"},{"internalType":"uint256","name":"_maxStakers","type":"uint256"}],"name":"getActiveStakers","outputs":[{"internalType":"uint256","name":"allLockedTokens","type":"uint256"},{"internalType":"uint256[2][]","name":"activeStakers","type":"uint256[2][]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getAllTokens","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getCompletedWork","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getCurrentPeriod","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getFlags","outputs":[{"internalType":"bool","name":"windDown","type":"bool"},{"internalType":"bool","name":"reStake","type":"bool"},{"internalType":"bool","name":"measureWork","type":"bool"},{"internalType":"bool","name":"snapshots","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getLastCommittedPeriod","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_index","type":"uint256"}],"name":"getLastPeriodOfSubStake","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"getLockedTokens","outputs":[{"internalType":"uint256","name":"lockedValue","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_index","type":"uint256"}],"name":"getPastDowntime","outputs":[{"internalType":"uint16","name":"startPeriod","type":"uint16"},{"internalType":"uint16","name":"endPeriod","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getPastDowntimeLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getReservedReward","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getStakersLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_index","type":"uint256"}],"name":"getSubStakeInfo","outputs":[{"internalType":"uint16","name":"firstPeriod","type":"uint16"},{"internalType":"uint16","name":"lastPeriod","type":"uint16"},{"internalType":"uint16","name":"periods","type":"uint16"},{"internalType":"uint128","name":"lockedValue","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getSubStakesLength","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"getWorkerFromStaker","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_reservedReward","type":"uint256"},{"internalType":"address","name":"_sourceOfFunds","type":"address"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"isOwner","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"}],"name":"isReStakeLocked","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"isTestContract","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"isUpgrade","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"lockAndCreate","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_index","type":"uint256"},{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"lockAndIncrease","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"lockDurationCoefficient1","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lockDurationCoefficient2","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint16","name":"_lockReStakeUntilPeriod","type":"uint16"}],"name":"lockReStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"","type":"uint16"}],"name":"lockedPerPeriod","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"maxAllowableLockedTokens","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"maximumRewardedPeriods","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_index1","type":"uint256"},{"internalType":"uint256","name":"_index2","type":"uint256"}],"name":"mergeStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"minAllowableLockedTokens","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"minLockedPeriods","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"minWorkerPeriods","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"mintingCoefficient","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"policyManager","outputs":[{"internalType":"contract PolicyManagerInterface","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"previousPeriodSupply","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"previousTarget","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_index","type":"uint256"},{"internalType":"uint16","name":"_periods","type":"uint16"}],"name":"prolongStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"uint256","name":"_value","type":"uint256"},{"internalType":"address","name":"_tokenContract","type":"address"},{"internalType":"bytes","name":"","type":"bytes"}],"name":"receiveApproval","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"_index","type":"uint16"}],"name":"removeUnusedSubStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"secondsPerPeriod","outputs":[{"internalType":"uint32","name":"","type":"uint32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"contract AdjudicatorInterface","name":"_adjudicator","type":"address"}],"name":"setAdjudicator","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"contract PolicyManagerInterface","name":"_policyManager","type":"address"}],"name":"setPolicyManager","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"_reStake","type":"bool"}],"name":"setReStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"_enableSnapshots","type":"bool"}],"name":"setSnapshots","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"_windDown","type":"bool"}],"name":"setWindDown","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"contract WorkLockInterface","name":"_workLock","type":"address"}],"name":"setWorkLock","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"bool","name":"_measureWork","type":"bool"}],"name":"setWorkMeasurement","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_staker","type":"address"},{"internalType":"uint256","name":"_penalty","type":"uint256"},{"internalType":"address","name":"_investigator","type":"address"},{"internalType":"uint256","name":"_reward","type":"uint256"}],"name":"slashStaker","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"stakerFromWorker","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"stakerInfo","outputs":[{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"uint16","name":"currentCommittedPeriod","type":"uint16"},{"internalType":"uint16","name":"nextCommittedPeriod","type":"uint16"},{"internalType":"uint16","name":"lastCommittedPeriod","type":"uint16"},{"internalType":"uint16","name":"lockReStakeUntilPeriod","type":"uint16"},{"internalType":"uint256","name":"completedWork","type":"uint256"},{"internalType":"uint16","name":"workerStartPeriod","type":"uint16"},{"internalType":"address","name":"worker","type":"address"},{"internalType":"uint256","name":"flags","type":"uint256"},{"internalType":"uint256","name":"reservedSlot1","type":"uint256"},{"internalType":"uint256","name":"reservedSlot2","type":"uint256"},{"internalType":"uint256","name":"reservedSlot3","type":"uint256"},{"internalType":"uint256","name":"reservedSlot4","type":"uint256"},{"internalType":"uint256","name":"reservedSlot5","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"stakers","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"supportsHistory","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"pure","type":"function"},{"inputs":[],"name":"target","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token","outputs":[{"internalType":"contract NuCypherToken","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_blockNumber","type":"uint256"}],"name":"totalStakedAt","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"},{"internalType":"uint256","name":"_blockNumber","type":"uint256"}],"name":"totalStakedForAt","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_testTarget","type":"address"}],"name":"verifyState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"workLock","outputs":[{"internalType":"contract WorkLockInterface","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`
var contractAddress = `0xbbD3C0C794F40c4f993B03F65343aCC6fcfCb2e2`

func (dc *DefaultCommand) Start(data commander.StartData) error {
	stakerAddress, err := go_config.Config.GetString("staker-address")
	if err != nil {
		return err
	}
	serverUrl, err := go_config.Config.GetString("server-url")
	if err != nil {
		return err
	}
	if serverUrl == "" {
		return errors.New("必须指定server-url")
	}
	pkey, err := go_config.Config.GetString("pkey")
	if err != nil {
		return err
	}
	if pkey == "" {
		return errors.New("必须指定pkey")
	}
	gasPrice, err := go_config.Config.GetString("gas-price")
	if err != nil {
		return err
	}
	gasPriceWeiTemp, err := go_decimal.Decimal.Start(gasPrice).ShiftedBy(9)
	if err != nil {
		return err
	}
	gasPriceWei := gasPriceWeiTemp.EndForString()
	intervalStr, err := go_config.Config.GetString("interval")
	if err != nil {
		return err
	}
	interval, err := go_reflect.Reflect.ToInt64(intervalStr)
	if err != nil {
		return err
	}

	fsStat, err := os.Stat(data.DataDir)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") || fsStat == nil || !fsStat.IsDir() {
			err = os.Mkdir(data.DataDir, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	dc.cacheFs, err = os.OpenFile(path.Join(data.DataDir, "data.json"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(dc.cacheFs)
	if err != nil {
		return err
	}
	if len(b) != 0 {
		err := json.Unmarshal(b, &dc.cache)
		if err != nil {
			return err
		}
	}
	go_logger.Logger.InfoF("last info: %v", dc.cache)

	wallet, err := go_coin_eth.NewWallet(serverUrl)
	if err != nil {
		return err
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for range timer.C {
		var getCurrentPeriodResult []interface{}
		var currentPeriod string
		var nextPeriod string
		var stakerInfoResult []interface{}
		var sendedTx *types.Transaction

		getCurrentPeriodResult, err = wallet.CallContractConstant(contractAddress, abiStr,"getCurrentPeriod",  nil)
		if err != nil {
			return err
		}
		currentPeriod = go_reflect.Reflect.ToString(getCurrentPeriodResult[0])
		go_logger.Logger.InfoF("currentPeriod: %s", currentPeriod)

		if dc.cache.Period != "" {  // 说明今天的已经发了
			ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
			_, isPending, err := wallet.RemoteClient.TransactionByHash(ctx, common.HexToHash(dc.cache.TxHash))
			if err != nil {
				go_logger.Logger.Error(err)
				goto continueTimer
			}
			// 未确认交易直接返回
			if isPending {
				if dc.cache.Period != currentPeriod {
					return errors.New("cache中不是当前period，请手动处理")
				}

				if time.Now().UTC().Hour() == 22 && dc.cache.GasPrice == gasPriceWei { // 到时间了而且还是低price的交易，则覆盖
					go_logger.Logger.Info("使用正常gasprice覆盖交易")
					sendedTx, err = wallet.CallMethod(pkey, contractAddress, abiStr, "commitToNextPeriod", &go_coin_eth.CallMethodOpts{
						Nonce: dc.cache.Nonce,
					})
					if err != nil {
						return err
					}
					dc.cache.Period = currentPeriod
					dc.cache.TxHash = sendedTx.Hash().String()
					dc.cache.Nonce = sendedTx.Nonce()
					dc.cache.GasPrice = sendedTx.GasPrice().String()
					go_logger.Logger.InfoF("交易发送成功。txid: %s", dc.cache.TxHash)
					goto continueTimer
				}
				go_logger.Logger.Info("cache sendedTx is pending")
				goto continueTimer
			}
			// 已经确认了，清空cache
			dc.cache = Data{}
			go_logger.Logger.Info("sendedTx confirmed today")
			goto continueTimer
		}


		stakerInfoResult, err = wallet.CallContractConstant(contractAddress, abiStr, "stakerInfo", nil, common.HexToAddress(stakerAddress))
		if err != nil {
			return err
		}
		nextPeriod = go_reflect.Reflect.ToString(stakerInfoResult[2])
		go_logger.Logger.InfoF("my nextPeriod: %s", nextPeriod)

		if go_decimal.Decimal.Start(nextPeriod).Gt(currentPeriod) {
			go_logger.Logger.Info("nextPeriod > currentPeriod, 不处理")
			goto continueTimer
		}

		// 使用指定的price发送一笔交易
		go_logger.Logger.Info("开始发送交易")
		sendedTx, err = wallet.CallMethod(pkey, contractAddress, abiStr, "commitToNextPeriod", &go_coin_eth.CallMethodOpts{
			GasPrice: gasPriceWei,
		})
		if err != nil {
			return err
		}
		dc.cache.Period = currentPeriod
		dc.cache.TxHash = sendedTx.Hash().String()
		dc.cache.Nonce = sendedTx.Nonce()
		dc.cache.GasPrice = sendedTx.GasPrice().String()
		go_logger.Logger.InfoF("交易发送成功。txid: %s", dc.cache.TxHash)


	continueTimer:
		timer.Reset(time.Duration(interval) * time.Minute)
		continue
	}

	return nil
}
