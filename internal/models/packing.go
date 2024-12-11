package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// /////////////////////////////////////////////////////
// MoModel for collection "mo"
// /////////////////////////////////////////////////////
type PackingRecord struct {
	Date     time.Time `bson:"date" json:"date"`
	Mo       string    `bson:"mo" json:"mo"`
	Factory  string    `bson:"factory" json:"factory"`
	ProdType string    `bson:"prodtype" json:"prodtype"`
	Product  struct {
		Id   string `bson:"id" json:"id"`
		Name string `bson:"name" json:"name"`
	} `bson:"product" json:"product"`
	Parent struct {
		Id      string  `bson:"id" json:"id"`
		Name    string  `bson:"name" json:"name"`
		NoParts int     `bson:"noparts" json:"noparts"`
		Price   float64 `bson:"price" json:"price"`
	} `bson:"parent" json:"parent"`
	Qty        int       `bson:"qty" json:"qty"`
	Value      float64   `bson:"value" json:"value"`
	Reporter   string    `bson:"reporter" json:"reporter"`
	CreatedAt  time.Time `bson:"createdat" json:"createdat"`
	ModifiedAt time.Time `bson:"modifiedat" json:"modifiedat"`
}

type PackingModel struct {
	mgdb *mongo.Database
}

func NewPackingModel(mgdb *mongo.Database) *PackingModel {
	return &PackingModel{mgdb: mgdb}
}

func (m *PackingModel) InsertNewReport(pr PackingRecord) (*mongo.InsertOneResult, error) {
	bdoc := bson.M{
		"date":     primitive.NewDateTimeFromTime(pr.Date),
		"mo":       pr.Mo,
		"factory":  pr.Factory,
		"prodtype": pr.ProdType,
		"product": bson.M{
			"id":   pr.Product.Id,
			"name": pr.Product.Name,
		},
		"parent": bson.M{
			"id":      pr.Parent.Id,
			"name":    pr.Parent.Name,
			"noparts": pr.Parent.NoParts,
			"price":   pr.Parent.Price,
		},
		"qty":       pr.Qty,
		"value":     pr.Value,
		"reporter":  pr.Reporter,
		"createdat": primitive.NewDateTimeFromTime(pr.CreatedAt),
	}
	sresult, err := m.mgdb.Collection("packing").InsertOne(context.Background(), bdoc)
	if err != nil {
		return nil, err
	}
	return sresult, nil
}
