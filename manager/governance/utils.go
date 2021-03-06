package governance

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"math/big"
)

// get wing total supply
func (this *GovernanceManager) getWingTotalSupply() (*big.Int, error) {
	r, err := this.sdk.GetStorage(this.wingAddress, []byte("TotalSupply"))
	if err != nil {
		return nil, fmt.Errorf("getWingTotalSupply, this.sdk.GetStorage error: %s", err)
	}
	return common.BigIntFromNeoBytes(r), nil
}

func (this *GovernanceManager) getBalanceOf(accountStr string) (uint64, error) {
	account, err := common.AddressFromBase58(accountStr)
	if err != nil {
		return 0, fmt.Errorf("getWingTotalSupply, common.AddressFromBase58 error: %s", err)
	}
	r, err := this.sdk.GetStorage(this.wingAddress, append([]byte{0x01}, account[:]...))
	if err != nil {
		return 0, fmt.Errorf("getWingTotalSupply, this.sdk.GetStorage error: %s", err)
	}
	return common.BigIntFromNeoBytes(r).Uint64(), nil
}

type Pool struct {
	Address common.Address
	Weight  common.I128
	Status  uint8
}

func (this *Pool) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("address deserialization error eof")
	}
	weight, eof := source.NextI128()
	if eof {
		return fmt.Errorf("weight deserialization error eof")
	}
	status, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("status deserialization error eof")
	}
	this.Address = address
	this.Weight = weight
	this.Status = status
	return nil
}

func (this *GovernanceManager) getAllPools() ([]*Pool, error) {
	preExecResult, err := this.sdk.WasmVM.PreExecInvokeWasmVMContract(this.contractAddress,
		"get_product_pools", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("getAllPool, this.sdk.WasmVM.PreExecInvokeWasmVMContract error: %s", err)
	}
	r, err := preExecResult.Result.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("getAllPool, preExecResult.Result.ToByteArray error: %s", err)
	}
	source := common.NewZeroCopySource(r)
	allPools := make([]*Pool, 0)
	l, _, irregular, eof := source.NextVarUint()
	if irregular || eof {
		return nil, fmt.Errorf("getAllPool, source.NextVarUint error")
	}
	for i := 0; uint64(i) < l; i++ {
		pool := new(Pool)
		err := pool.Deserialization(source)
		if err != nil {
			return nil, fmt.Errorf("getAllPool, pool.Deserialization error: %s", err)
		}
		allPools = append(allPools, pool)
	}
	return allPools, nil
}
