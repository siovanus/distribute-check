package restful

type Web interface {
	GetRewards(map[string]interface{}) map[string]interface{}
	GetGasFee(map[string]interface{}) map[string]interface{}
}
