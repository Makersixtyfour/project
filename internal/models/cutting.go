package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// /////////////////////////////////////////////////////
// CuttingModel for collection "cutting"
// /////////////////////////////////////////////////////
type CuttingReport struct {
	ReportId     string    `bson:"_id"`
	Date         time.Time `bson:"date"`
	ProdType     string    `bson:"prodtype"`
	Wrnote       string    `bson:"wrnote"`
	Woodtype     string    `bson:"woodtype"`
	Thickness    float64   `bson:"thickness"`
	Qty          float64   `bson:"qtycbm"`
	Type         string    `bson:"type"`
	Reporter     string    `bson:"reporter"`
	CreatedDate  time.Time `bson:"createddate"`
	LastModified time.Time `bson:"lastmodified"`
}

type CuttingWrnote struct {
	WrnoteId    string    `bson:"_id"`
	WrnoteCode  string    `bson:"wrnotecode"`
	Woodtype    string    `bson:"woodtype"`
	ProdType    string    `bson:"prodtype"`
	Thickness   float64   `bson:"thickness"`
	Qty         float64   `bson:"wrnoteqty"`
	Remain      float64   `bson:"wrremain"`
	Date        time.Time `bson:"date"`
	CreatedDate time.Time `bson:"createat"`
}

type CuttingModel struct {
	mgdb *mongo.Database
}

// Create instance of CuttingModel
func NewCuttingModel(mgdb *mongo.Database) *CuttingModel {
	return &CuttingModel{mgdb: mgdb}
}

func (m *CuttingModel) PartalUpdate(cuttingReport *CuttingReport) error {
	id, _ := primitive.ObjectIDFromHex(cuttingReport.ReportId)
	log.Println(id)
	// m.mgdb.Collection("cutting").UpdateOne(context.Background(), bson.M{"_id": id})
	return nil
}

func (m *CuttingModel) Search(searchWord string) []CuttingReport {
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M

	if err != nil {
		filter = bson.M{"type": "report", "$or": bson.A{
			bson.M{"woodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"wrnote": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"reporter": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"thickness": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"qtycbm": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"type": "report", "date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := m.mgdb.Collection("cutting").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println("failed to access databa cutting at search of model cutting", err)
		return nil
	}
	defer cur.Close(context.Background())

	var results []CuttingReport
	if err = cur.All(context.Background(), &results); err != nil {
		log.Println("faild to decode", err)
		return nil
	}

	return results
}

func (m *CuttingModel) WrnoteSearch(searchWord string) []CuttingWrnote {
	regexWord := ".*" + searchWord + ".*"
	dateSearch, err := time.Parse("2006-01-02", searchWord)
	var filter bson.M

	if err != nil {
		filter = bson.M{"type": "wrnote", "$or": bson.A{
			bson.M{"wrnotecode": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"woodtype": bson.M{"$regex": regexWord, "$options": "i"}},
			bson.M{"prodtype": bson.M{"$regex": regexWord, "$options": "i"}},
		},
		}
	} else {
		filter = bson.M{"type": "wrnote", "date": primitive.NewDateTimeFromTime(dateSearch)}
	}
	cur, err := m.mgdb.Collection("cutting").Find(context.Background(), filter, options.Find().SetSort(bson.M{"date": -1}))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer cur.Close(context.Background())

	var results []CuttingWrnote
	if err = cur.All(context.Background(), &results); err != nil {
		log.Println(err)
		return nil
	}

	return results
}

// có thể bỏ này
func (m *CuttingModel) FindAllWrnotes() ([]CuttingWrnote, error) {
	cur, err := m.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "wrnote"}, options.Find().SetSort(bson.M{"createdat": -1}))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cur.Close(context.Background())

	var results []CuttingWrnote
	if err = cur.All(context.Background(), &results); err != nil {
		log.Println(err)
		return nil, err
	}

	return results, nil
}

func (m *CuttingModel) FindAllReportsSortDateDesc() ([]CuttingReport, error) {
	cur, err := m.mgdb.Collection("cutting").Find(context.Background(), bson.M{"type": "report"}, options.Find().SetSort(bson.M{"occurdate": -1}))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer cur.Close(context.Background())

	var results []CuttingReport
	if err = cur.All(context.Background(), &results); err != nil {
		log.Println(err)
		return nil, err
	}

	return results, nil
}
