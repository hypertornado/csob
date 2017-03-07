package csob

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

func (c *CSOB) EETExtension() *EETExtension {
	ret := &EETExtension{
		Extension: "eetV3",
		DTTM:      timestamp(),
		Data: &EETExtensionData{
			PremiseId:      c.eet.premiseId,
			CashRegisterId: c.eet.cashRegisterId,
		},
	}
	return ret
}

//func (e *eet) GetMap() map[string]interface{} {

//}
