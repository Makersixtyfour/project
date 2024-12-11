package models

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// /////////////////////////////////////////////////////
// 6SModel for collection "sixs"
// /////////////////////////////////////////////////////
type SixS struct {
}

type SixSModel struct {
	mgdb *mongo.Database
}

func NewSixSModel(mgdb *mongo.Database) *SixSModel {
	return &SixSModel{mgdb: mgdb}
}

func (m *SixSModel) InsertMany(scoresStrJson string) error {

	return nil
}
