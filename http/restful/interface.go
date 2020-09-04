package restful

type Web interface {
	FlashPoolMarketDistribution(map[string]interface{}) map[string]interface{}
	PoolDistribution(map[string]interface{}) map[string]interface{}
	GovBannerOverview(map[string]interface{}) map[string]interface{}
	GovBanner(map[string]interface{}) map[string]interface{}
	FlashPoolBanner(map[string]interface{}) map[string]interface{}

	FlashPoolDetail(map[string]interface{}) map[string]interface{}
	FlashPoolAllMarket(map[string]interface{}) map[string]interface{}

	AssetPrice(map[string]interface{}) map[string]interface{}
}
