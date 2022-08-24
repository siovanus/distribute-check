/**
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/contracts/native/go_abi/node_manager_abi"
	"github.com/ethereum/go-ethereum/contracts/native/governance/node_manager"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/polynetwork/distribute-check/config"
	"github.com/polynetwork/distribute-check/log"
	"github.com/polynetwork/distribute-check/store"
	"github.com/polynetwork/distribute-check/store/models"
	"math/big"
	"os"
	"strings"
	"time"
)

var nmAbi abi.ABI

type Listener struct {
	conf     *config.Config
	client   *ethclient.Client
	db       *store.Client
	contract *node_manager_abi.INodeManager
	chainId  *big.Int
}

func New(conf *config.Config, db *store.Client) *Listener {
	return &Listener{conf: conf, db: db}
}

func (v *Listener) Init() (err error) {
	nmAbi, err = abi.JSON(strings.NewReader(node_manager_abi.INodeManagerABI))
	client, err := ethclient.Dial(v.conf.ZionConfig.RestURL)
	if err != nil {
		return fmt.Errorf("ethclient.Dial error: %s", err)
	}
	v.client = client

	contract, err := node_manager_abi.NewINodeManager(utils.NodeManagerContractAddress, v.client)
	if err != nil {
		return fmt.Errorf("node_manager_abi.NewINodeManager error: %s", err)
	}
	v.contract = contract

	v.chainId, err = client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("client.ChainID error: %s", err)
	}

	// init epoch info
	err = v.db.SaveEpochInfo(&models.EpochInfo{ID: 1, Validators: make([]string, 0)})
	if err != nil {
		return fmt.Errorf("v.db.SaveEpochInfo error: %s", err)
	}

	return nil
}

func (v *Listener) Listen(ctx context.Context) {
	trackHeight, err := v.db.LoadTrackHeight()
	if err != nil {
		log.Fatalf("v.db.LoadTrackHeight error: %s", err)
		os.Exit(1)
	}
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			height, err := GetCurrentHeight(v.conf.ZionConfig.RestURL)
			if err != nil {
				log.Errorf("GetCurrentHeight failed:%v", err)
				continue
			}
			log.Infof("current zion height:%d", height)
			if height < trackHeight {
				continue
			}

			for trackHeight <= height {
				select {
				case <-ctx.Done():
					log.Info("quiting from signal...")
					return
				default:
				}
				log.Infof("handling zion height:%d", trackHeight)
				err = v.ScanAndExecBlock(trackHeight)
				if err != nil {
					log.Errorf("ScanAndExecBlock failed:%v", err)
					sleep()
					continue
				}

				trackHeight = trackHeight + 1
				err = v.db.SaveTrackHeight(trackHeight)
				if err != nil {
					log.Errorf("v.db.SaveTrackHeight failed:%v", err)
				}
			}

		case <-ctx.Done():
			log.Info("quiting from signal...")
			return
		}
	}
}

func (v *Listener) ScanAndExecBlock(height uint64) error {
	client := v.client
	ctx := context.Background()

	// get block
	block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(height))
	if err != nil {
		return fmt.Errorf("ScanAndExecBlock, client.BlockByNumber error: %s", err)
	}
	totalGas := new(big.Int)
	for _, tx := range block.Transactions() {
		// parse tx data
		data := tx.Data()
		methodName, err := nmAbi.MethodById(tx.Data())
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, nmAbi.MethodById error: %s", err)
		}
		from, err := types.Sender(types.LatestSignerForChainID(v.chainId), tx)
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, types.Sender error: %s", err)
		}
		// get transaction by hash
		transaction, _, err := client.TransactionByHash(ctx, tx.Hash())
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, client.TransactionByHash error: %s", err)
		}
		// get receipt by hash
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, client.TransactionReceipt error: %s", err)
		}
		// accumulate gas
		gas := new(big.Int).Mul(transaction.GasPrice(), new(big.Int).SetUint64(receipt.GasUsed))
		totalGas = new(big.Int).Add(totalGas, gas)
		err = v.db.SaveTotalGas(&models.TotalGas{Height: height, TotalGas: models.NewBigInt(totalGas)})
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, v.db.SaveTotalGas error: %s", err)
		}

		// if success
		if receipt.Status == 0 {
			continue
		}
		// if done
		doneTx, err := v.db.LoadDoneTx(tx.Hash().Hex())
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, v.db.LoadDoneTx error: %s", err)
		}
		if len(doneTx) != 0 {
			continue
		}

		// execute
		switch methodName.Name {
		case node_manager_abi.MethodCreateValidator:
			param := new(node_manager.CreateValidatorParam)
			method, _ := nmAbi.Methods[methodName.Name]
			args, err := method.Inputs.Unpack(data[4:])
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Unpack error: %s", err)
			}
			err = method.Inputs.Copy(param, args)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Copy error: %s", err)
			}
			validator := &models.Validator{
				StakeAddress:     from.Hex(),
				ConsensusAddress: param.ConsensusAddress.Hex(),
				Commission:       models.NewBigInt(param.Commission),
				TotalStake:       models.NewBigInt(param.InitStake),
				SelfStake:        models.NewBigInt(param.InitStake),
			}
			err = v.db.SaveValidator(validator)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.SaveValidator %s error: %s", param.ConsensusAddress.Hex(), err)
			}
			err = v.db.AddStakeInfo(from.Hex(), param.ConsensusAddress.Hex(), param.InitStake)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.AddStakeInfo error: %s", err)
			}

		case node_manager_abi.MethodStake:
			param := new(node_manager.StakeParam)
			method, _ := nmAbi.Methods[methodName.Name]
			args, err := method.Inputs.Unpack(data[4:])
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Unpack error: %s", err)
			}
			err = method.Inputs.Copy(param, args)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Copy error: %s", err)
			}
			err = v.db.AddStakeInfo(from.Hex(), param.ConsensusAddress.Hex(), param.Amount)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.AddStakeInfo error: %s", err)
			}
			err = v.db.AddValidatorStake(from.Hex(), param.ConsensusAddress.Hex(), param.Amount)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.AddValidatorStake error: %s", err)
			}

		case node_manager_abi.MethodUnStake:
			param := new(node_manager.UnStakeParam)
			method, _ := nmAbi.Methods[methodName.Name]
			args, err := method.Inputs.Unpack(data[4:])
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Unpack error: %s", err)
			}
			err = method.Inputs.Copy(param, args)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock，method.Inputs.Copy error: %s", err)
			}
			err = v.db.SubStakeInfo(from.Hex(), param.ConsensusAddress.Hex(), param.Amount)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.SubStakeInfo error: %s", err)
			}
			err = v.db.SubValidatorStake(param.ConsensusAddress.Hex(), param.Amount)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.SubValidatorStake error: %s", err)
			}

		case node_manager_abi.MethodEndBlock:
			err = v.CalcRewards(height)
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.CalcRewards error: %s", err)
			}

		case node_manager_abi.MethodChangeEpoch:
			num, err := v.db.LoadValidatorNum()
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.LoadValidatorNum error: %s", err)
			}
			latestEpochInfo, err := v.db.LoadLatestEpochInfo()
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.LoadLatestEpochInfo error: %s", err)
			}
			ID := latestEpochInfo.ID + 1
			var validators models.SQLStringArray
			if num >= 4 {
				newEpochInfo, err := v.GetEpochInfo(new(big.Int).SetUint64(ID))
				if err != nil {
					return fmt.Errorf("ScanAndExecBlock, v.GetEpochInfo error: %s", err)
				}
				for _, v := range newEpochInfo.Validators {
					validators = append(validators, v.Hex())
				}
			}
			err = v.db.SaveEpochInfo(&models.EpochInfo{
				ID:         ID,
				Validators: validators,
			})
			if err != nil {
				return fmt.Errorf("ScanAndExecBlock, v.db.SaveEpochInfo error: %s", err)
			}
		default:
		}
		err = v.db.SaveDoneTx(&models.DoneTx{Hash: tx.Hash().Hex(), Height: height})
		if err != nil {
			return fmt.Errorf("ScanAndExecBlock, v.db.SaveDoneTx %s error: %s", tx.Hash().Hex(), err)
		}
	}
	err = v.db.CleanDoneTx()
	if err != nil {
		return fmt.Errorf("CalcReward, v.db.CleanDoneTx error: %s", err)
	}
	return nil
}

func (v *Listener) CalcRewards(height uint64) error {
	accumulatedRewards, err := v.db.LoadAccumulatedRewards()
	if err != nil {
		return fmt.Errorf("CalcReward, v.db.LoadAccumulatedRewards error: %s", err)
	}
	totalGas, err := v.db.LoadTotalGas(height)
	if err != nil {
		return fmt.Errorf("CalcReward, v.db.LoadTotalGas error: %s", err)
	}
	totalRewards := new(big.Int).Add(new(big.Int).Add(&totalGas.TotalGas.Int, params.ZNT1), accumulatedRewards)
	// get validators in this block
	epochInfo, err := v.db.LoadLatestEpochInfo()
	if err != nil {
		return fmt.Errorf("CalcReward, v.db.LoadLatestEpochInfo error: %s", err)
	}
	validatorList := epochInfo.Validators
	if len(validatorList) == 0 {
		err = v.db.SaveAccumulatedRewards(totalRewards)
		if err != nil {
			return fmt.Errorf("CalcReward, v.db.SaveAccumulatedRewards error: %s", err)
		}
	} else {
		validatorRewards := new(big.Int).Div(totalRewards, new(big.Int).SetUint64(uint64(len(validatorList))))
		if err != nil {
			return fmt.Errorf("CalcReward, Div validatorRewards error: %v", err)
		}
		for _, consensusAddress := range validatorList {
			// get validator
			validator, err := v.db.LoadValidator(consensusAddress)
			if err != nil {
				return fmt.Errorf("CalcReward, v.db.LoadValidator error: %v", err)
			}
			commission := new(big.Int).Div(new(big.Int).Mul(validatorRewards, &validator.Commission.Int), node_manager.PercentDecimal)
			stakeRewards := new(big.Int).Sub(validatorRewards, commission)
			rewardsPerToken := new(big.Int).Div(new(big.Int).Mul(stakeRewards, node_manager.TokenDecimal), &validator.TotalStake.Int)
			allStakeAddress, err := v.db.LoadAllStakeAddress(consensusAddress)
			if err != nil {
				return fmt.Errorf("CalcReward, v.db.LoadAllStakeAddress error: %v", err)
			}
			for _, s := range allStakeAddress {
				stakeInfo, err := v.db.LoadStakeInfo(s, consensusAddress)
				if err != nil {
					return fmt.Errorf("CalcReward, v.db.LoadStakeInfo error: %v", err)
				}
				rewards := new(big.Int).Div(new(big.Int).Mul(&stakeInfo.Amount.Int, rewardsPerToken), node_manager.TokenDecimal)
				if s == validator.StakeAddress {
					rewards = new(big.Int).Add(rewards, commission)
				}
				err = v.db.SaveRewards(&models.Rewards{Address: s, Height: height, Amount: models.NewBigInt(rewards)})
				if err != nil {
					return fmt.Errorf("CalcReward, v.db.SaveRewards error: %v", err)
				}
			}
		}
		err = v.db.SaveAccumulatedRewards(new(big.Int))
		if err != nil {
			return fmt.Errorf("CalcReward, v.db.SaveAccumulatedRewards error: %s", err)
		}
	}
	return nil
}

func sleep() {
	time.Sleep(time.Second)
}
