package models

import (
	"context"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Criterion struct {
	Id          string   `bson:"id"`
	Description string   `bson:"description"`
	Point       int      `bson:"point"`
	Kind        string   `bson:"kind"`
	ApplyOn     string   `bson:"applyon"`
	AuthPos     []string `bson:"authpos"`
	EvalPos     []string `bson:"evalpos"`
}

type CriterionModel struct {
	mgdb *mongo.Database
}

func NewCriterionModel(mgdb *mongo.Database) *CriterionModel {
	return &CriterionModel{
		mgdb: mgdb,
	}
}

func (m *CriterionModel) Find() ([]Criterion, error) {
	var criteria []Criterion
	cur, err := m.mgdb.Collection("criterion").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"id": -1}))
	if err != nil {
		log.Println("loi truy xuat database", err)
		return nil, err
	}
	defer cur.Close(context.Background())

	if err = cur.All(context.Background(), &criteria); err != nil {
		log.Println("loi decode criteria", err)
		return nil, err
	}
	return criteria, nil
}

func (m *CriterionModel) Search(searchWord string) ([]Criterion, error) {
	searchRegex := ".*" + searchWord + ".*"
	criterionsearchInt, _ := strconv.Atoi(searchWord)

	filter := bson.M{"$or": bson.A{
		bson.M{"id": bson.M{"$regex": searchRegex}},
		bson.M{"description": bson.M{"$regex": searchRegex, "$options": "i"}},
		bson.M{"kind": bson.M{"$regex": searchRegex, "$options": "i"}},
		bson.M{"point": criterionsearchInt},
	}}

	cur, err := m.mgdb.Collection("criterion").Find(context.Background(), filter)
	if err != nil {
		log.Println("ia_searchcriterion: ", err)
		return nil, err
	}
	defer cur.Close(context.Background())

	var critResults []Criterion
	err = cur.All(context.Background(), &critResults)
	if err != nil {
		log.Println("ia_searchcriterion: ", err)
		return nil, err
	}
	return critResults, nil
}
