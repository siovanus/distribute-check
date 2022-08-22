package common

const (
	GETREWARDS        = "/api/v1/getrewards"
	ACTION_GETREWARDS = "getrewards"
)

type Response struct {
	Action string      `json:"action"`
	Desc   string      `json:"desc"`
	Error  uint32      `json:"error"`
	Result interface{} `json:"result"`
}

type GetRewardsRequest struct {
	Id        string
	Addresses []string
	EndHeight uint64
}

type GetRewardsResponse struct {
	Id     string
	Amount []string
}
