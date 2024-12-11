package models

type ColorRecord_datestr struct {
	Id           string `bson:"_id"`
	PanelNo      string `bson:"panelno"`
	User         string `bson:"user"`
	FinishCode   string `bson:"finishcode"`
	FinishName   string `bson:"finishname"`
	Substrate    string `bson:"substrate"`
	Collection   string `bson:"collection"`
	Brand        string `bson:"brand"`
	FinishSystem string `bson:"chemicalsystem"`
	Texture      string `bson:"texture"`
	Thickness    string `bson:"thickness"`
	Sheen        string `bson:"sheen"`
	Hardness     string `bson:"hardness"`
	Prepared     string `bson:"prepared"`
	Review       string `bson:"review"`
	Approved     string `bson:"approved"`
	ApprovedDate string `bson:"approveddate"`
	ExpiredDate  string `bson:"expireddate"`
	Inspections  []struct {
		Date      string  `bson:"date"`
		Result    string  `bson:"result"`
		Delta     float64 `bson:"delta"`
		Inspector string  `bson:"inspector"`
	} `bson:"inspections"`
	ExpiredColor   string
	NextInspection string
}
