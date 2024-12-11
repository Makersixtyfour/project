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
type ProValRecord struct {
	Date     time.Time `bson:"date"`
	Factory  string    `bson:"factory"`
	ProdType string    `bson:"prodtype"`
	Item     string    `bson:"item"`
	Qty      int       `bson:"qty"`
	Value    float64   `bson:"value"`
	// IdFromOriginPacking
}

type ProValModel struct {
	mgdb *mongo.Database
}

func NewProValModel(mgdb *mongo.Database) *ProValModel {
	return &ProValModel{mgdb: mgdb}
}

func (m *ProValModel) Create(pvr ProValRecord) error {
	_, err := m.mgdb.Collection("prodvalue").InsertOne(context.Background(), bson.M{
		"date":     primitive.NewDateTimeFromTime(pvr.Date),
		"factory":  pvr.Factory,
		"prodtype": pvr.ProdType,
		"item":     pvr.Item,
		"qty":      pvr.Qty,
		"value":    pvr.Value,
	})
	return err
}
