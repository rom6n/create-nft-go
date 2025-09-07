package generalcontractutils

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func CalculateAddress(workchain int32, stateInit *tlb.StateInit) *address.Address {
	address := address.NewAddress(4, 0, cell.BeginCell().
		MustStoreUInt(6, 5).
		MustStoreRef(stateInit.Code).
		MustStoreRef(stateInit.Data).
		EndCell().Hash())
	/*addressSlice := cell.BeginCell().
	MustStoreUInt(4, 3).
	MustStoreInt(int64(workchain), 8).
	MustStoreSlice(stateInit.Hash(), 256).
	EndCell().BeginParse()*/
	return address
}

func PackStateInit(codeCell *cell.Cell, dataCell *cell.Cell) *tlb.StateInit {
	return &tlb.StateInit{
		Code: codeCell,
		Data: dataCell,
	}
}

func PackDefaultMessage(toAddress *address.Address, amount uint64, text ...string) *cell.Cell {
	msgBuilder := cell.BeginCell().
		MustStoreUInt(0x10, 6).
		MustStoreAddr(toAddress).
		MustStoreCoins(amount).
		MustStoreUInt(0, 1+4+4+64+32+1+1)

	if text != nil {
		packOfTextMessage := cell.BeginCell().
			MustStoreUInt(0, 32).
			MustStoreStringSnake(text[0]).
			EndCell()

		msgBuilder.
			MustStoreInt(1, 1).
			MustStoreRef(packOfTextMessage)
	}

	return msgBuilder.EndCell()
}

func PackDeployMessage(toAddress *address.Address, stateInit *tlb.StateInit) *tlb.InternalMessage {
	return &tlb.InternalMessage{
		Bounce:    true,
		Amount:    tlb.MustFromTON("0.05"),
		DstAddr:   toAddress,
		StateInit: stateInit,
		Body:      cell.BeginCell().EndCell(),
	}

}
