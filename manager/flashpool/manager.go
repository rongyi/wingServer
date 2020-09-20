package flashpool

import (
	"fmt"
	"github.com/siovanus/wingServer/utils"
	"math"
	"math/big"
	"time"

	sdk "github.com/ontio/ontology-go-sdk"
	ocommon "github.com/ontio/ontology/common"
	"github.com/siovanus/wingServer/config"
	"github.com/siovanus/wingServer/http/common"
	"github.com/siovanus/wingServer/manager/governance"
	"github.com/siovanus/wingServer/store"
)

const (
	BlockPerYear = 60 * 60 * 24 * 365 * 2 / 3
)

type FlashPoolManager struct {
	cfg             *config.Config
	contractAddress ocommon.Address
	oracleAddress   ocommon.Address
	sdk             *sdk.OntologySdk
	store           *store.Client
}

func NewFlashPoolManager(contractAddress, oracleAddress ocommon.Address, sdk *sdk.OntologySdk,
	store *store.Client, cfg *config.Config) *FlashPoolManager {
	manager := &FlashPoolManager{
		cfg:             cfg,
		contractAddress: contractAddress,
		oracleAddress:   oracleAddress,
		sdk:             sdk,
		store:           store,
	}

	return manager
}

func (this *FlashPoolManager) AssetPrice(asset string) (string, error) {
	price, err := this.assetPrice(asset)
	if err != nil {
		return "", fmt.Errorf("AssetPrice, this.assetPrice error: %s", err)
	}
	return utils.ToStringByPrecise(price, this.cfg.TokenDecimal["oracle"]), nil
}

func (this *FlashPoolManager) AssetStoredPrice(asset string) (*big.Int, error) {
	if asset == "USDT" {
		return new(big.Int).SetUint64(uint64(math.Pow10(int(this.cfg.TokenDecimal["oracle"])))), nil
	}
	price, err := this.store.LoadPrice(asset)
	if err != nil {
		return nil, fmt.Errorf("AssetStoredPrice, this.store.LoadPrice error: %s", err)
	}
	return utils.ToIntByPrecise(price.Price, this.cfg.TokenDecimal["oracle"]), nil
}

func (this *FlashPoolManager) FlashPoolMarketDistribution() (*common.FlashPoolMarketDistribution, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolMarketDistribution, this.GetAllMarkets error: %s", err)
	}
	flashPoolMarketDistribution := make([]*common.Distribution, 0)
	for _, address := range allMarkets {
		market, err := this.store.LoadFlashMarket(this.cfg.AssetMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetail, this.store.LoadFlashMarket error: %s", err)
		}
		supplyAmount := market.TotalSupply
		borrowAmount := market.TotalBorrow
		insuranceAmount := market.TotalInsurance

		totalDistribution, err := this.getTotalDistribution(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolMarketDistribution, this.getTotalDistribution error: %s", err)
		}
		distributedDay := (uint64(time.Now().Unix()) - governance.GenesisTime) / governance.DaySecond
		distribution := &common.Distribution{
			Icon: this.cfg.IconMap[this.cfg.AssetMap[address.ToHexString()]],
			Name: this.cfg.AssetMap[address.ToHexString()],
			// totalDistribution / distributedDay
			PerDay: utils.ToStringByPrecise(new(big.Int).Div(totalDistribution,
				new(big.Int).SetUint64(distributedDay)), this.cfg.TokenDecimal["WING"]),
			SupplyAmount:    supplyAmount,
			BorrowAmount:    borrowAmount,
			InsuranceAmount: insuranceAmount,
			Total:           utils.ToStringByPrecise(totalDistribution, this.cfg.TokenDecimal["WING"]),
		}
		flashPoolMarketDistribution = append(flashPoolMarketDistribution, distribution)
	}
	return &common.FlashPoolMarketDistribution{FlashPoolMarketDistribution: flashPoolMarketDistribution}, nil
}

func (this *FlashPoolManager) PoolDistribution() (*common.Distribution, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("PoolDistribution, this.GetAllMarkets error: %s", err)
	}
	distribution := new(common.Distribution)
	s := new(big.Int).SetUint64(0)
	b := new(big.Int).SetUint64(0)
	i := new(big.Int).SetUint64(0)
	d := new(big.Int).SetUint64(0)
	for _, address := range allMarkets {
		market, err := this.store.LoadFlashMarket(this.cfg.AssetMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("PoolDistribution, this.store.LoadFlashMarket error: %s", err)
		}
		supplyAmount := market.TotalSupply
		borrowAmount := market.TotalBorrow
		insuranceAmount := market.TotalInsurance

		totalDistribution, err := this.getTotalDistribution(address)
		if err != nil {
			return nil, fmt.Errorf("PoolDistribution, this.getTotalDistribution error: %s", err)
		}

		// supplyAmount * price
		s = new(big.Int).Add(s, utils.ToIntByPrecise(supplyAmount, this.cfg.TokenDecimal["pUSDT"]))
		// borrowAmount * price
		b = new(big.Int).Add(b, utils.ToIntByPrecise(borrowAmount, this.cfg.TokenDecimal["pUSDT"]))
		// insuranceAmount * price
		i = new(big.Int).Add(i, utils.ToIntByPrecise(insuranceAmount, this.cfg.TokenDecimal["pUSDT"]))
		d = new(big.Int).Add(d, totalDistribution)
	}
	distribution.Name = "Flash"
	distribution.Icon = this.cfg.IconMap[distribution.Name]
	distributedDay := new(big.Int).SetUint64((uint64(time.Now().Unix()) - governance.GenesisTime) / governance.DaySecond)
	distribution.SupplyAmount = utils.ToStringByPrecise(s, this.cfg.TokenDecimal["pUSDT"])
	distribution.BorrowAmount = utils.ToStringByPrecise(b, this.cfg.TokenDecimal["pUSDT"])
	distribution.InsuranceAmount = utils.ToStringByPrecise(i, this.cfg.TokenDecimal["pUSDT"])
	distribution.PerDay = utils.ToStringByPrecise(new(big.Int).Div(d, distributedDay), this.cfg.TokenDecimal["WING"])
	distribution.Total = utils.ToStringByPrecise(d, this.cfg.TokenDecimal["WING"])
	return distribution, nil
}

func (this *FlashPoolManager) FlashPoolBanner() (*common.FlashPoolBanner, error) {
	gap := uint64(time.Now().Unix()) - governance.GenesisTime
	length := len(governance.DailyDistibute)
	epoch := []uint64{0}
	for i := 1; i < length+1; i++ {
		epoch = append(epoch, epoch[i-1]+governance.DistributeTime[i-1])
	}
	if gap > epoch[length] {
		gap = epoch[length]
	}
	index := 0
	for i := 0; i < len(epoch); i++ {
		if gap >= epoch[i] {
			index = i
		}
	}

	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolBanner, this.GetAllMarkets error: %s", err)
	}
	total := new(big.Int).SetUint64(0)
	for _, address := range allMarkets {
		totalDistribution, err := this.getTotalDistribution(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolBanner, this.getTotalDistribution error: %s", err)
		}
		total = new(big.Int).Add(total, totalDistribution)
	}
	today := governance.DailyDistibute[index] * governance.DaySecond
	share := new(big.Int).SetUint64(0)
	if total.Uint64() != 0 {
		t := new(big.Int).Mul(new(big.Int).Mul(new(big.Int).SetUint64(today),
			new(big.Int).SetUint64(uint64(math.Pow10(int(this.cfg.TokenDecimal["WING"]))))),
			new(big.Int).SetUint64(uint64(math.Pow10(int(this.cfg.TokenDecimal["percentage"])))))
		share = new(big.Int).Div(new(big.Int).Div(t, new(big.Int).SetUint64(100)), total)
	}

	return &common.FlashPoolBanner{
		Today: utils.ToStringByPrecise(new(big.Int).SetUint64(today), 2),
		Share: utils.ToStringByPrecise(share, this.cfg.TokenDecimal["percentage"]),
		Total: utils.ToStringByPrecise(total, this.cfg.TokenDecimal["WING"]),
	}, nil
}

func (this *FlashPoolManager) FlashPoolDetail() (*common.FlashPoolDetail, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolDetail, this.GetAllMarkets error: %s", err)
	}
	flashPoolDetail := new(common.FlashPoolDetail)
	s := new(big.Int).SetUint64(0)
	b := new(big.Int).SetUint64(0)
	i := new(big.Int).SetUint64(0)
	for _, address := range allMarkets {
		market, err := this.store.LoadFlashMarket(this.cfg.AssetMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetail, this.store.LoadFlashMarket error: %s", err)
		}
		supplyAmount := market.TotalSupply
		borrowAmount := market.TotalBorrow
		insuranceAmount := market.TotalInsurance

		// supplyAmount * price
		// borrowAmount * price
		// insuranceAmount * price
		supplyDollar := utils.ToIntByPrecise(supplyAmount, this.cfg.TokenDecimal["pUSDT"])
		borrowDollar := utils.ToIntByPrecise(borrowAmount, this.cfg.TokenDecimal["pUSDT"])
		insuranceDollar := utils.ToIntByPrecise(insuranceAmount, this.cfg.TokenDecimal["pUSDT"])
		s = new(big.Int).Add(s, supplyDollar)
		b = new(big.Int).Add(b, borrowDollar)
		i = new(big.Int).Add(i, insuranceDollar)
	}

	flashPoolDetail.TotalSupply = utils.ToStringByPrecise(s, this.cfg.TokenDecimal["pUSDT"])
	flashPoolDetail.TotalBorrow = utils.ToStringByPrecise(b, this.cfg.TokenDecimal["pUSDT"])
	flashPoolDetail.TotalInsurance = utils.ToStringByPrecise(i, this.cfg.TokenDecimal["pUSDT"])

	return flashPoolDetail, nil
}

func (this *FlashPoolManager) FlashPoolDetailForStore() (*store.FlashPoolDetail, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolDetailForStore, this.GetAllMarkets error: %s", err)
	}
	flashPoolDetail := new(store.FlashPoolDetail)
	s := new(big.Int).SetUint64(0)
	b := new(big.Int).SetUint64(0)
	i := new(big.Int).SetUint64(0)
	for _, address := range allMarkets {
		name := this.cfg.AssetMap[address.ToHexString()]
		supplyAmount, err := this.getSupplyAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetailForStore, this.getSupplyAmount error: %s", err)
		}
		borrowAmount, err := this.getBorrowAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetailForStore, this.getSupplyAmount error: %s", err)
		}
		insuranceAmount, err := this.getInsuranceAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetailForStore, this.getSupplyAmount error: %s", err)
		}
		priceStr, err := this.AssetPrice(this.cfg.OracleMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("FlashPoolDetailForStore, this.AssetPrice error: %s", err)
		}
		price := utils.ToIntByPrecise(priceStr, this.cfg.TokenDecimal["oracle"])
		// supplyAmount * price
		// borrowAmount * price
		// insuranceAmount * price
		supplyDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(supplyAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		borrowDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(borrowAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		insuranceDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(insuranceAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		s = new(big.Int).Add(s, supplyDollar)
		b = new(big.Int).Add(b, borrowDollar)
		i = new(big.Int).Add(i, insuranceDollar)
	}
	flashPoolDetail.Timestamp = uint64(time.Now().Unix())
	flashPoolDetail.TotalSupply = utils.ToStringByPrecise(s, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
	flashPoolDetail.TotalBorrow = utils.ToStringByPrecise(b, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
	flashPoolDetail.TotalInsurance = utils.ToStringByPrecise(i, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
	return flashPoolDetail, nil
}

func (this *FlashPoolManager) FlashPoolMarketStore() error {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return fmt.Errorf("FlashPoolMarketStore, this.GetAllMarkets error: %s", err)
	}
	timestamp := uint64(time.Now().Unix())
	for _, address := range allMarkets {
		flashPoolMarket := new(store.FlashPoolMarket)
		supplyAmount, err := this.getSupplyAmount(address)
		if err != nil {
			return fmt.Errorf("FlashPoolMarketStore, this.getSupplyAmount error: %s", err)
		}
		borrowAmount, err := this.getBorrowAmount(address)
		if err != nil {
			return fmt.Errorf("FlashPoolMarketStore, this.getSupplyAmount error: %s", err)
		}
		insuranceAmount, err := this.getInsuranceAmount(address)
		if err != nil {
			return fmt.Errorf("FlashPoolMarketStore, this.getSupplyAmount error: %s", err)
		}
		name := this.cfg.AssetMap[address.ToHexString()]
		price, err := this.AssetStoredPrice(this.cfg.OracleMap[address.ToHexString()])
		if err != nil {
			return fmt.Errorf("FlashPoolMarketStore, this.AssetStoredPrice error: %s", err)
		}
		// supplyAmount * price
		// borrowAmount * price
		// insuranceAmount * price
		supplyDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(supplyAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		borrowDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(borrowAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		insuranceDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(insuranceAmount, price),
			this.cfg.TokenDecimal[name]), this.cfg.TokenDecimal["pUSDT"])
		flashPoolMarket.Name = name
		flashPoolMarket.TotalSupply = utils.ToStringByPrecise(supplyDollar, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
		flashPoolMarket.TotalBorrow = utils.ToStringByPrecise(borrowDollar, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
		flashPoolMarket.TotalInsurance = utils.ToStringByPrecise(insuranceDollar, this.cfg.TokenDecimal["pUSDT"]+this.cfg.TokenDecimal["oracle"])
		flashPoolMarket.Timestamp = timestamp
		err = this.store.SaveFlashPoolMarket(flashPoolMarket)
		if err != nil {
			return fmt.Errorf("FlashPoolMarketStore, this.store.SaveFlashPoolMarket error: %s", err)
		}
	}
	return nil
}

func (this *FlashPoolManager) FlashPoolAllMarket() (*common.FlashPoolAllMarket, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolAllMarket, this.GetAllMarkets error: %s", err)
	}
	flashPoolAllMarket := &common.FlashPoolAllMarket{
		FlashPoolAllMarket: make([]*common.Market, 0),
	}
	for _, address := range allMarkets {
		market, err := this.store.LoadFlashMarket(this.cfg.AssetMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarket, this.store.LoadFlashMarket error: %s", err)
		}
		flashPoolAllMarket.FlashPoolAllMarket = append(flashPoolAllMarket.FlashPoolAllMarket, &market)
	}
	return flashPoolAllMarket, nil
}

func (this *FlashPoolManager) FlashPoolAllMarketForStore() (*common.FlashPoolAllMarket, error) {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.GetAllMarkets error: %s", err)
	}
	flashPoolAllMarket := &common.FlashPoolAllMarket{
		FlashPoolAllMarket: make([]*common.Market, 0),
	}
	for _, address := range allMarkets {
		name := this.cfg.AssetMap[address.ToHexString()]
		supplyAmount, err := this.getSupplyAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getSupplyAmount error: %s", err)
		}
		borrowAmount, err := this.getBorrowAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getBorrowAmount error: %s", err)
		}
		insuranceAmount, err := this.getInsuranceAmount(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getInsuranceAmount error: %s", err)
		}
		price, err := this.AssetStoredPrice(this.cfg.OracleMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.AssetStoredPrice error: %s", err)
		}

		supplyApy, err := this.getSupplyApy(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getSupplyApy error: %s", err)
		}
		borrowApy, err := this.getBorrowApy(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getBorrowApy error: %s", err)
		}
		insuranceApy, err := this.getInsuranceApy(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getInsuranceApy error: %s", err)
		}
		marketMeta, err := this.getMarketMeta(address)
		if err != nil {
			return nil, fmt.Errorf("FlashPoolAllMarketForStore, this.getMarketMeta error: %s", err)
		}

		market := new(common.Market)
		market.Name = this.cfg.AssetMap[address.ToHexString()]
		market.Icon = this.cfg.IconMap[market.Name]

		// supplyAmount * price
		// borrowAmount * price
		// insuranceAmount * price
		market.TotalSupply = utils.ToStringByPrecise(new(big.Int).Mul(supplyAmount, price),
			this.cfg.TokenDecimal[name]+this.cfg.TokenDecimal["oracle"])
		market.TotalBorrow = utils.ToStringByPrecise(new(big.Int).Mul(borrowAmount, price),
			this.cfg.TokenDecimal[name]+this.cfg.TokenDecimal["oracle"])
		market.TotalInsurance = utils.ToStringByPrecise(new(big.Int).Mul(insuranceAmount, price),
			this.cfg.TokenDecimal[name]+this.cfg.TokenDecimal["oracle"])
		market.CollateralFactor = utils.ToStringByPrecise(marketMeta.CollateralFactorMantissa, this.cfg.TokenDecimal["flash"])
		market.SupplyApy = utils.ToStringByPrecise(supplyApy, this.cfg.TokenDecimal["flash"])
		market.BorrowApy = utils.ToStringByPrecise(borrowApy, this.cfg.TokenDecimal["flash"])
		market.InsuranceApy = utils.ToStringByPrecise(insuranceApy, this.cfg.TokenDecimal["flash"])
		flashPoolAllMarket.FlashPoolAllMarket = append(flashPoolAllMarket.FlashPoolAllMarket, market)
	}
	return flashPoolAllMarket, nil
}

func (this *FlashPoolManager) UserFlashPoolOverview(accountStr string) (*common.UserFlashPoolOverview, error) {
	account, err := ocommon.AddressFromBase58(accountStr)
	if err != nil {
		return nil, fmt.Errorf("UserFlashPoolOverview, ocommon.AddressFromBase58 error: %s", err)
	}
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return nil, fmt.Errorf("UserFlashPoolOverview, this.GetAllMarkets error: %s", err)
	}
	userFlashPoolOverview := &common.UserFlashPoolOverview{
		CurrentSupply:    make([]*common.Supply, 0),
		CurrentBorrow:    make([]*common.Borrow, 0),
		CurrentInsurance: make([]*common.Insurance, 0),
		AllMarket:        make([]*common.UserMarket, 0),
	}
	accountLiquidity, err := this.getAccountLiquidity(account)
	if err != nil {
		return nil, fmt.Errorf("UserFlashPoolOverview, this.getAccountLiquidity error: %s", err)
	}
	userFlashPoolOverview.BorrowLimit = utils.ToStringByPrecise(accountLiquidity.Liquidity.ToBigInt(), this.cfg.TokenDecimal["oracle"])

	userBalance, err := this.store.LoadUserBalance(accountStr)
	if err != nil {
		return nil, fmt.Errorf("UserFlashPoolOverview, this.store.LoadUserBalance error: %s", err)
	}
	netApy := new(big.Int).SetUint64(0)
	s := new(big.Int).SetUint64(0)
	b := new(big.Int).SetUint64(0)
	i := new(big.Int).SetUint64(0)
	for _, address := range allMarkets {
		assetName := this.cfg.AssetMap[address.ToHexString()]
		assetApy, err := this.store.LoadAssetApy(assetName)
		if err != nil {
			return nil, fmt.Errorf("UserFlashPoolOverview, this.store.LoadAssetApy error: %s", err)
		}
		price, err := this.AssetStoredPrice(this.cfg.OracleMap[address.ToHexString()])
		if err != nil {
			return nil, fmt.Errorf("UserFlashPoolOverview, this.AssetStoredPrice error: %s", err)
		}
		userAssetBalance := new(store.UserAssetBalance)
		for _, v := range userBalance {
			if v.AssetName == assetName {
				userAssetBalance = &v
			}
		}
		supplyAmount := utils.ToIntByPrecise(userAssetBalance.SupplyBalance, this.cfg.TokenDecimal[assetName])
		borrowAmount := utils.ToIntByPrecise(userAssetBalance.BorrowBalance, this.cfg.TokenDecimal[assetName])
		insuranceAmount := utils.ToIntByPrecise(userAssetBalance.InsuranceBalance, this.cfg.TokenDecimal[assetName])
		// supplyAmount * price
		// borrowAmount * price
		// insuranceAmount * price
		supplyDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(supplyAmount, price),
			this.cfg.TokenDecimal[assetName]), this.cfg.TokenDecimal["pUSDT"])
		borrowDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(borrowAmount, price),
			this.cfg.TokenDecimal[assetName]), this.cfg.TokenDecimal["pUSDT"])
		insuranceDollar := utils.ToIntByPrecise(utils.ToStringByPrecise(new(big.Int).Mul(insuranceAmount, price),
			this.cfg.TokenDecimal[assetName]), this.cfg.TokenDecimal["pUSDT"])
		s = new(big.Int).Add(s, supplyDollar)
		b = new(big.Int).Add(s, borrowDollar)
		i = new(big.Int).Add(i, insuranceDollar)
		supplyApy := utils.ToIntByPrecise(assetApy.SupplyApy, this.cfg.TokenDecimal["flash"])
		borrowApy := utils.ToIntByPrecise(assetApy.BorrowApy, this.cfg.TokenDecimal["flash"])
		insuranceApy := utils.ToIntByPrecise(assetApy.InsuranceApy, this.cfg.TokenDecimal["flash"])
		sa := new(big.Int).Mul(supplyDollar, supplyApy)
		ia := new(big.Int).Mul(insuranceDollar, insuranceApy)
		ba := new(big.Int).Mul(borrowDollar, borrowApy)
		netApy = new(big.Int).Add(netApy, new(big.Int).Sub(new(big.Int).Add(sa, ia), ba))

		if supplyAmount.Uint64() != 0 {
			supply := &common.Supply{
				Name:             this.cfg.AssetMap[address.ToHexString()],
				Icon:             this.cfg.IconMap[this.cfg.AssetMap[address.ToHexString()]],
				SupplyBalance:    utils.ToStringByPrecise(supplyAmount, this.cfg.TokenDecimal[assetName]),
				Apy:              utils.ToStringByPrecise(supplyApy, this.cfg.TokenDecimal["flash"]),
				CollateralFactor: assetApy.CollateralFactor,
				IfCollateral:     userAssetBalance.IfCollateral,
			}
			userFlashPoolOverview.CurrentSupply = append(userFlashPoolOverview.CurrentSupply, supply)
		}
		if borrowAmount.Uint64() != 0 {
			borrow := &common.Borrow{
				Name:             this.cfg.AssetMap[address.ToHexString()],
				Icon:             this.cfg.IconMap[this.cfg.AssetMap[address.ToHexString()]],
				BorrowBalance:    utils.ToStringByPrecise(borrowAmount, this.cfg.TokenDecimal[assetName]),
				Apy:              utils.ToStringByPrecise(borrowApy, this.cfg.TokenDecimal["flash"]),
				CollateralFactor: assetApy.CollateralFactor,
			}
			if b.Uint64() != 0 {
				borrow.Limit = utils.ToStringByPrecise(new(big.Int).Div(new(big.Int).Mul(borrowDollar, new(big.Int).SetUint64(
					uint64(math.Pow10(int(this.cfg.TokenDecimal["percentage"]))))), b), this.cfg.TokenDecimal["percentage"])
			}
			userFlashPoolOverview.CurrentBorrow = append(userFlashPoolOverview.CurrentBorrow, borrow)
		}
		if insuranceAmount.Uint64() != 0 {
			insurance := &common.Insurance{
				Name:             this.cfg.AssetMap[address.ToHexString()],
				Icon:             this.cfg.IconMap[this.cfg.AssetMap[address.ToHexString()]],
				InsuranceBalance: utils.ToStringByPrecise(insuranceAmount, this.cfg.TokenDecimal[assetName]),
				Apy:              utils.ToStringByPrecise(insuranceApy, this.cfg.TokenDecimal["flash"]),
				CollateralFactor: assetApy.CollateralFactor,
			}
			userFlashPoolOverview.CurrentInsurance = append(userFlashPoolOverview.CurrentInsurance, insurance)
		}

		totalBorrowAmount, err := this.getBorrowAmount(address)
		if err != nil {
			return nil, fmt.Errorf("UserFlashPoolOverviewForStore, this.getSupplyAmount error: %s", err)
		}
		totalInsuranceAmount, err := this.getInsuranceAmount(address)
		if err != nil {
			return nil, fmt.Errorf("UserFlashPoolOverviewForStore, this.getInsuranceAmount error: %s", err)
		}
		if supplyAmount.Uint64() == 0 && borrowAmount.Uint64() == 0 && insuranceAmount.Uint64() == 0 {
			userMarket := &common.UserMarket{
				Name:             this.cfg.AssetMap[address.ToHexString()],
				Icon:             this.cfg.IconMap[this.cfg.AssetMap[address.ToHexString()]],
				SupplyApy:        utils.ToStringByPrecise(supplyApy, this.cfg.TokenDecimal["flash"]),
				BorrowApy:        utils.ToStringByPrecise(borrowApy, this.cfg.TokenDecimal["flash"]),
				BorrowLiquidity:  utils.ToStringByPrecise(totalBorrowAmount, this.cfg.TokenDecimal[assetName]),
				InsuranceApy:     utils.ToStringByPrecise(insuranceApy, this.cfg.TokenDecimal["flash"]),
				InsuranceAmount:  utils.ToStringByPrecise(totalInsuranceAmount, this.cfg.TokenDecimal[assetName]),
				CollateralFactor: assetApy.CollateralFactor,
				IfCollateral:     userAssetBalance.IfCollateral,
			}
			userFlashPoolOverview.AllMarket = append(userFlashPoolOverview.AllMarket, userMarket)
		}
	}
	total := new(big.Int).Add(new(big.Int).Add(s, b), i)
	if total.Uint64() != 0 {
		userFlashPoolOverview.NetApy = utils.ToStringByPrecise(new(big.Int).Div(netApy, total), this.cfg.TokenDecimal["flash"])
	}
	return userFlashPoolOverview, nil
}

func (this *FlashPoolManager) UserBalanceForStore(accountStr string) error {
	account, err := ocommon.AddressFromBase58(accountStr)
	if err != nil {
		return fmt.Errorf("UserBalanceForStore, ocommon.AddressFromBase58 error: %s", err)
	}
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return fmt.Errorf("UserBalanceForStore, this.GetAllMarkets error: %s", err)
	}
	assetsIn, _ := this.getAssetsIn(account)
	for _, address := range allMarkets {
		supplyAmount, err := this.getSupplyAmountByAccount(address, account)
		if err != nil {
			return fmt.Errorf("UserBalanceForStore, this.getSupplyAmountByAccount error: %s", err)
		}
		borrowAmount, err := this.getBorrowAmountByAccount(address, account)
		if err != nil {
			return fmt.Errorf("UserBalanceForStore, this.getBorrowAmountByAccount error: %s", err)
		}
		insuranceAmount, err := this.getInsuranceAmountByAccount(address, account)
		if err != nil {
			return fmt.Errorf("UserBalanceForStore, this.getInsuranceAmountByAccount error: %s", err)
		}
		name := this.cfg.AssetMap[address.ToHexString()]
		isAssetIn := false
		for _, a := range assetsIn {
			if address == a {
				isAssetIn = true
				break
			}
		}
		userBalance := &store.UserAssetBalance{
			UserAddress:      accountStr,
			AssetName:        name,
			Icon:             this.cfg.IconMap[name],
			SupplyBalance:    utils.ToStringByPrecise(supplyAmount, this.cfg.TokenDecimal[name]),
			BorrowBalance:    utils.ToStringByPrecise(borrowAmount, this.cfg.TokenDecimal[name]),
			InsuranceBalance: utils.ToStringByPrecise(insuranceAmount, this.cfg.TokenDecimal[name]),
			IfCollateral:     isAssetIn,
		}
		err = this.store.SaveUserAssetBalance(userBalance)
		if err != nil {
			return fmt.Errorf("UserBalanceForStore, this.store.SaveUserAssetBalance error: %s", err)
		}
	}
	return nil
}

func (this *FlashPoolManager) AssetApyForStore() error {
	allMarkets, err := this.GetAllMarkets()
	if err != nil {
		return fmt.Errorf("AssetApyForStore, this.GetAllMarkets error: %s", err)
	}
	for _, address := range allMarkets {
		name := this.cfg.AssetMap[address.ToHexString()]
		marketMeta, err := this.getMarketMeta(address)
		if err != nil {
			return fmt.Errorf("AssetApyForStore, this.getMarketMeta error: %s", err)
		}
		supplyApy, err := this.getSupplyApy(address)
		if err != nil {
			return fmt.Errorf("AssetApyForStore, this.getSupplyApy error: %s", err)
		}
		borrowApy, err := this.getBorrowApy(address)
		if err != nil {
			return fmt.Errorf("AssetApyForStore, this.getBorrowApy error: %s", err)
		}
		insuranceApy, err := this.getInsuranceApy(address)
		if err != nil {
			return fmt.Errorf("AssetApyForStore, this.getInsuranceApy error: %s", err)
		}
		assetApy := &store.AssetApy{
			AssetName:        name,
			SupplyApy:        utils.ToStringByPrecise(supplyApy, this.cfg.TokenDecimal["flash"]),
			BorrowApy:        utils.ToStringByPrecise(borrowApy, this.cfg.TokenDecimal["flash"]),
			InsuranceApy:     utils.ToStringByPrecise(insuranceApy, this.cfg.TokenDecimal["flash"]),
			CollateralFactor: utils.ToStringByPrecise(marketMeta.CollateralFactorMantissa, this.cfg.TokenDecimal["flash"]),
		}
		err = this.store.SaveAssetApy(assetApy)
		if err != nil {
			return fmt.Errorf("AssetApyForStore, this.store.SaveAssetApy error: %s", err)
		}
	}
	return nil
}

func (this *FlashPoolManager) ClaimWing(accountStr string) (string, error) {
	account, err := ocommon.AddressFromBase58(accountStr)
	if err != nil {
		return "", fmt.Errorf("ClaimWing, ocommon.AddressFromBase58 error: %s", err)
	}
	amount, err := this.getClaimWing(account)
	if err != nil {
		return "", fmt.Errorf("ClaimWing, this.getClaimWing error: %s", err)
	}
	return utils.ToStringByPrecise(amount, this.cfg.TokenDecimal["WING"]), nil
}
