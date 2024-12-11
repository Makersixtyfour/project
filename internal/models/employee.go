package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// /////////////////////////////////////////////////////
// EmployeeModel for collection "employee"
// /////////////////////////////////////////////////////
type Employee struct {
}

type EmployeeModel struct {
	mgdb *mongo.Database
}

func NewEmployeeModel(mgdb *mongo.Database) *EmployeeModel {
	return &EmployeeModel{mgdb: mgdb}
}

func (m *EmployeeModel) InsertMany(empStrJson string) error {
	var bdoc []interface{}
	err := bson.UnmarshalExtJSON([]byte(empStrJson), true, &bdoc)
	if err != nil {
		log.Print("failed to unmarshal json string", err)
		return err
	}

	_, err = m.mgdb.Collection("employee").InsertMany(context.Background(), bdoc)
	if err != nil {
		log.Println("failed to insert many to employee collection", err)
		return err
	}

	return nil
}
