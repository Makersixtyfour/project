package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Item struct {
	Id    string `bson:"id" json:"id"`
	Name  string `bson:"name" json:"name"`
	Parts []struct {
		Id   string `bson:"id" json:"id"`
		Name string `bson:"name" json:"name"`
	} `bson:"parts" json:"parts"`
}

type ItemModel struct {
	mgdb *mongo.Database
}

func NewItemModel(mgdb *mongo.Database) *ItemModel {
	return &ItemModel{
		mgdb: mgdb,
	}
}

func (m *ItemModel) UpdateParts(id, partString string) error {
	var parts []interface{}
	if err := bson.UnmarshalExtJSON([]byte(partString), true, &parts); err != nil {
		log.Println("failed to unmarshalExJSON", err)
		return err
	}
	_, err := m.mgdb.Collection("item").UpdateOne(context.Background(), bson.M{"id": id}, bson.M{"$set": bson.M{"parts": parts}})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *ItemModel) InsertByStringJson(strJson string) error {
	var bdoc []interface{}
	err := bson.UnmarshalExtJSON([]byte(strJson), true, &bdoc)
	if err != nil {
		log.Print("failed to unmarshal json string", err)
		return err
	}

	_, err = m.mgdb.Collection("item").InsertMany(context.Background(), bdoc)
	if err != nil {
		log.Println("failed to insert many to item collection", err)
		return err
	}

	return nil
}

func (m *ItemModel) InsertItem(item Item) error {
	_, err := m.mgdb.Collection("item").InsertOne(context.Background(), bson.M{
		"id":   item.Id,
		"name": item.Name,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
