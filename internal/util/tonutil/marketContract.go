package tonutil

import (
	"encoding/hex"
	"log"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetMarketContractCode() *cell.Cell {
	codeHex := "b5ee9c7241010c0100f0000114ff00f4a413f4bcf2c80b01020120020b02014803080202cb040500abd1b088831c02456f8007434c0cc1c6c244c383c059b084074c7c07000638d60c235c6083e405000fe443ca8f5350c1c08be401d3232c084b281f2fff2741de0063232c15633c59c3e80b2dac4b3333260103ec03816e020273060700153b513434c7f4c7f4ffcc20001700b232c7f2c7f2fff27b5520020148090a0017bb39ced44d0d33f31d70bff80011b8c97ed44d0d70b1f80078f28308d71820d31fd33fd31f02f823bbf263f0165132baf2a15144baf2a204f901541055f910f2a3f8009320d74a96d307d402fb00e83001a402f0176bceb6a6"

	boc, err := hex.DecodeString(codeHex)
	if err != nil {
		log.Fatalln("Error decoding market's hex code")
	}

	code, err := cell.FromBOC(boc)
	if err != nil {
		log.Fatalln("Error while parsing BOC to cell")
	}
	return code
}

func GetMarketContractDeployData(seqno, subwallet int32, publicKey []byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(seqno), 32).
		MustStoreUInt(uint64(subwallet), 32).
		MustStoreBinarySnake(publicKey).
		EndCell()
}
