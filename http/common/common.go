package common

const (
	GETREWARDS        = "/api/v1/getrewards"
	ACTION_GETREWARDS = "getrewards"

	GETGASFEE        = "/api/v1/getgasfee"
	ACTION_GETGASFEE = "getgasfee"
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

type GetGasFeeRequest struct {
	Id        string
	Addresses []string
	EndHeight uint64
}

type GetGasFeeResponse struct {
	Id     string
	Amount []string
}
