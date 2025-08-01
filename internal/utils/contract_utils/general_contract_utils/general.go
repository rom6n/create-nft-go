package generalcontractutils

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func CalculateAddress(workchain int32, stateInit *cell.Cell) *address.Address {
	address := address.NewAddress(4, 0, stateInit.Hash())
	/*addressSlice := cell.BeginCell().
	MustStoreUInt(4, 3).
	MustStoreInt(int64(workchain), 8).
	MustStoreSlice(stateInit.Hash(), 256).
	EndCell().BeginParse()*/
	return address
}

func PackStateInit(codeCell *cell.Cell, dataCell *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(6, 5).
		MustStoreRef(codeCell).
		MustStoreRef(dataCell).
		EndCell()
}

func PackDeployMessage(toAddress *address.Address, stateInit *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0x18, 6).
		MustStoreAddr(toAddress).
		MustStoreCoins(50000000).
		MustStoreUInt(4+2+1, 1+4+4+64+32+1+1+1).
		MustStoreRef(stateInit).
		MustStoreRef(cell.BeginCell().EndCell()).
		EndCell()
}
