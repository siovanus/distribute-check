package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/governance/node_manager"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"math/big"
)

func (v *Listener) GetEpochInfo(ID *big.Int) (*node_manager.EpochInfo, error) {
	node_manager.InitABI()
	input := &node_manager.GetEpochInfoParam{ID: ID}
	payload, err := input.Encode()
	if err != nil {
		return nil, fmt.Errorf("GetEpochInfo, input.Encode error: %s", err)
	}
	arg := ethereum.CallMsg{
		From: common.Address{},
		To:   &utils.NodeManagerContractAddress,
		Data: payload,
	}
	r, err := v.client.CallContract(context.Background(), arg, nil)
	if err != nil {
		return nil, fmt.Errorf("GetEpochInfo, v.client.CallContract error: %s", err)
	}
	epochInfo := new(node_manager.EpochInfo)
	err = epochInfo.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("GetEpochInfo, epochInfo.Decode error: %s", err)
	}
	return epochInfo, nil
}

func (v *Listener) getRewards(addresses []string, endHeight uint64) ([]string, error) {
	r := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		ar, err := v.db.LoadAccumulateRewards(addr, endHeight)
		if err != nil {
			return nil, fmt.Errorf("getRewards, v.db.LoadAccumulateRewards error: %s", err)
		}
		r = append(r, ar.String())
	}
	return r, nil
}
