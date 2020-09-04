package flashpool

type MarketDistribution struct {
	MarketDistribution []*Distribution
}

type PoolDistribution struct {
	PoolDistribution []*Distribution
}

type Distribution struct {
	Icon         string
	Name         string
	PerDay       uint64
	SupplyApy    uint64
	BorrowApy    uint64
	InsuranceApy uint64
	Total        uint64
}

type FlashPoolBanner struct {
	Today uint64
	Share uint64
	Total uint64
}

type FlashPoolDetail struct {
	TotalSupply       uint64
	TotalSupplyRate   uint64
	SupplyMarketRank  []*MarketFund
	SupplyVolumeDaily uint64
	Supplier          uint64

	TotalBorrow       uint64
	TotalBorrowRate   uint64
	BorrowMarketRank  []*MarketFund
	BorrowVolumeDaily uint64
	Borrower          uint64

	TotalInsurance       uint64
	TotalInsuranceRate   uint64
	InsuranceMarketRank  []*MarketFund
	InsuranceVolumeDaily uint64
	Guarantor            uint64
}

type MarketFund struct {
	Icon string
	Name string
	Fund uint64
}
type FlashPoolAllMarket struct {
	FlashPoolAllMarket []*Market
}

type Market struct {
	Icon           string
	Name           string
	TotalSupply    uint64
	SupplyApy      uint64
	TotalBorrow    uint64
	BorrowApy      uint64
	TotalInsurance uint64
	InsuranceApy   uint64
}

func (this *FlashPoolManager) marketDistribution() (*MarketDistribution, error) {
	distribution1 := &Distribution{
		Icon:         "http://106.75.209.209/icon/eth_icon.svg",
		Name:         "oETH",
		PerDay:       234,
		SupplyApy:    6783,
		BorrowApy:    8325,
		InsuranceApy: 9517,
		Total:        121234,
	}
	distribution2 := &Distribution{
		Icon:         "http://106.75.209.209/icon/asset_dai_icon.svg",
		Name:         "oDAI",
		PerDay:       345,
		SupplyApy:    1574,
		BorrowApy:    4576,
		InsuranceApy: 3842,
		Total:        25252,
	}
	return &MarketDistribution{MarketDistribution: []*Distribution{distribution1, distribution2}}, nil
}

func (this *FlashPoolManager) poolDistribution() (*PoolDistribution, error) {
	distribution1 := &Distribution{
		Icon:         "http://106.75.209.209/icon/flash_icon.svg",
		Name:         "Flash",
		PerDay:       231252,
		SupplyApy:    2532,
		BorrowApy:    4547,
		InsuranceApy: 1231,
		Total:        28364,
	}
	distribution2 := &Distribution{
		Icon:         "http://106.75.209.209/icon/if_icon.svg",
		Name:         "IF",
		PerDay:       1431241,
		SupplyApy:    1214,
		BorrowApy:    2525,
		InsuranceApy: 7742,
		Total:        72526,
	}
	return &PoolDistribution{PoolDistribution: []*Distribution{distribution1, distribution2}}, nil
}

func (this *FlashPoolManager) flashPoolBanner() (*FlashPoolBanner, error) {
	return &FlashPoolBanner{
		Today: 8676,
		Share: 7644,
		Total: 3452636,
	}, nil
}

func (this *FlashPoolManager) flashPoolDetail() (*FlashPoolDetail, error) {
	return &FlashPoolDetail{
		TotalSupply:     86544,
		TotalSupplyRate: 8754,
		SupplyMarketRank: []*MarketFund{{Icon: "http://106.75.209.209/icon/eth_icon.svg", Name: "ETH"},
			{Icon: "http://106.75.209.209/icon/asset_dai_icon.svg", Name: "DAI"},
			{Icon: "", Name: "BTC"}},
		SupplyVolumeDaily: 24526,
		Supplier:          125,

		TotalBorrow:     2524,
		TotalBorrowRate: 4252,
		BorrowMarketRank: []*MarketFund{{Icon: "http://106.75.209.209/icon/eth_icon.svg", Name: "ETH"},
			{Icon: "http://106.75.209.209/icon/asset_dai_icon.svg", Name: "DAI"}},
		BorrowVolumeDaily: 3115,
		Borrower:          36,

		TotalInsurance:     6754,
		TotalInsuranceRate: 9632,
		InsuranceMarketRank: []*MarketFund{{Icon: "http://106.75.209.209/icon/eth_icon.svg", Name: "ETH"},
			{Icon: "http://106.75.209.209/icon/asset_dai_icon.svg", Name: "DAI"}},
		InsuranceVolumeDaily: 3277,
		Guarantor:            234,
	}, nil
}

func (this *FlashPoolManager) flashPoolAllMarket() (*FlashPoolAllMarket, error) {
	return &FlashPoolAllMarket{
		FlashPoolAllMarket: []*Market{
			{
				Icon: "http://106.75.209.209/icon/eth_icon.svg",
				Name: "ETH",
				TotalSupply: 2526,
				SupplyApy: 4468,
				TotalBorrow: 25267,
				BorrowApy: 563,
				TotalInsurance: 8265,
				InsuranceApy: 256,
			},
			{
				Icon: "http://106.75.209.209/icon/asset_dai_icon.svg",
				Name: "DAI",
				TotalSupply: 2526,
				SupplyApy: 3526,
				TotalBorrow: 2415,
				BorrowApy: 241,
				TotalInsurance: 3473,
				InsuranceApy: 2541,
			}},
	}, nil
}
