package models

type BatchRecord_datestr struct {
	BatchNo        string  `bson:"batchno"`
	MixingDate     string  `bson:"mixingdate"`
	Volume         float64 `bson:"volume"`
	Operator       string  `bson:"operator"`
	Classification string  `bson:"classification"`
	SOPNo          string  `bson:"sopno"`
	Viscosity      float64 `bson:"viscosity"`
	Nk2            float64 `bson:"nk2"`
	Fordcup4       float64 `bson:"fordcup4"`
	LightDark      float64 `bson:"lightdark"`
	RedGreen       float64 `bson:"redgreen"`
	YellowBlue     float64 `bson:"yellowblue"`
	Status         string  `bson:"status"`
	Supplier       string  `bson:"supplier"`
	IssuedDate     string  `bson:"issueddate"`
	StartUse       string  `bson:"startuse"`
	EndUse         string  `bson:"enduse"`
	Receiver       string  `bson:"receiver"`
	Area           string  `bson:"area"`
	Color          struct {
		Code     string `bson:"code"`
		Name     string `bson:"name"`
		Brand    string `bson:"brand"`
		Supplier string `bson:"supplier"`
	} `bson:"color"`
	Items []struct {
		Code string `bson:"code"`
		Mo   string `bson:"mo"`
	} `bson:"items"`
	Step string `bson:"step"`
}
