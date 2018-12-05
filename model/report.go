package model

type Report struct {
	MerchantID 	string        `json:"-" bson:"merchant_id"`
	Date       	string          `json:"date,omitempty"`
	Products   	[]SaleProductOut `json:"products,omitempty"`
	Accumulate 	float32       `json:"accumulate,omitempty"`
}
