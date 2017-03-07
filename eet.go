package csob

import (
	"fmt"
)

type eet struct {
	premiseId      int64
	cashRegisterId string
}

func (c *CSOB) EET(premise int64, cashRegister string) {
	c.eet = &eet{
		premiseId:      premise,
		cashRegisterId: cashRegister,
	}
}

type EETExtension struct {
	Extension string            `json:"extension"`
	DTTM      string            `json:"dttm"`
	Data      *EETExtensionData `json:"data"`
	Signature string            `json:"signature"`
}

type EETExtensionData struct {
	PremiseId      int64   `json:"premiseId"`
	CashRegisterId string  `json:"cashRegisterId"`
	TotalPrice     float64 `json:"totalPrice"`
}

func (c *CSOB) EETExtension(totalAmount uint) *EETExtension {
	var total float64 = float64(totalAmount) / 100
	ret := &EETExtension{
		Extension: "eetV3",
		DTTM:      timestamp(),
		Data: &EETExtensionData{
			PremiseId:      c.eet.premiseId,
			CashRegisterId: c.eet.cashRegisterId,
			TotalPrice:     total,
		},
	}
	toSign := fmt.Sprintf("%s|%s|%d|%s|%.2f",
		ret.Extension,
		ret.DTTM,
		ret.Data.PremiseId,
		ret.Data.CashRegisterId,
		ret.Data.TotalPrice,
	)
	//fmt.Println(toSign)

	signature, err := signData(c.key, toSign)
	if err != nil {
		panic(err)
	}
	//fmt.Println(signature)

	ret.Signature = signature

	return ret
}
