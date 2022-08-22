package listener

import (
	"github.com/polynetwork/distribute-check/http/common"
	"github.com/polynetwork/distribute-check/http/restful"
	"github.com/polynetwork/distribute-check/log"
	"github.com/polynetwork/distribute-check/utils"
)

func (v *Listener) GetRewards(param map[string]interface{}) map[string]interface{} {
	req := &common.GetRewardsRequest{}
	resp := &common.Response{}
	err := utils.ParseParams(req, param)
	if err != nil {
		resp.Error = restful.INVALID_PARAMS
		resp.Desc = err.Error()
		log.Errorf("GetRewards: decode params failed, err: %s", err)
	} else {
		rewards, err := v.getRewards(req.Addresses, req.EndHeight)
		if err != nil {
			resp.Error = restful.INTERNAL_ERROR
			resp.Desc = err.Error()
			log.Errorf("GetRewards error: %s", err)
		} else {
			resp.Error = restful.SUCCESS
			resp.Result = &common.GetRewardsResponse{
				Id:     req.Id,
				Amount: rewards,
			}
			log.Infof("GetRewards success")
		}
	}

	m, err := utils.RefactorResp(resp, resp.Error)
	if err != nil {
		log.Errorf("GetRewards: failed, err: %s", err)
	} else {
		log.Debug("GetRewards: resp success")
	}
	return m
}
