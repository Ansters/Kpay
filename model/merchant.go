package model

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/options"
	"log"

	"github.com/Pallinder/go-randomdata"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sethvargo/go-password/password"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/bson"
)

type Merchant struct {
	ID          objectid.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string `json:"name,omitempty"`
	BankAccount string `json:"bank_account,omitempty" bson:"bank_account, omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

type MerchantServiceImpl struct {
	Client *mongo.Client
}

func (s *MerchantServiceImpl) Auth(username string, psw string) (Merchant, error) {
	collection := s.Client.Database("kpay").Collection("merchant")
	var merchant Merchant
	whereTo := bson.D{{"username", username}, {"password", psw}}
	err := collection.FindOne(context.Background(), whereTo).Decode(&merchant)
	return merchant, err
}

func (s *MerchantServiceImpl) Register(merchant *Merchant) error {

	collection := s.Client.Database("kpay").Collection("merchant")

	if isDuplicateBankAccount(collection, merchant.BankAccount) {
		return errors.New("duplicate bank account")
	}

	merchant.Username = generateUsername()
	merchant.Password = generatePassword()

	res, err := collection.InsertOne(context.Background(), merchant)
	if err != nil {
		return err
	}
	merchant.ID = res.InsertedID.(objectid.ObjectID)
	return nil
}

func (s *MerchantServiceImpl) Find(id string) (Merchant, error) {

	var merchant Merchant
	objID, err := objectid.FromHex(id)
	if err != nil {
		log.Fatal(err)
		return Merchant{}, err
	}
	idDoc := bson.D{{"_id", objID}}
	collection := s.Client.Database("kpay").Collection("merchant")
	err = collection.FindOne(context.Background(), idDoc).Decode(&merchant)
	return merchant, err
}

func (s *MerchantServiceImpl) Update(merchant Merchant) error {

	idDoc := bson.D{{"_id", merchant.ID}}
	collection := s.Client.Database("kpay").Collection("merchant")
	_, err := collection.UpdateOne(context.Background(), idDoc, bson.D{{"$set", bson.D{{"name", merchant.Name}}}})
	return err
}

func (s *MerchantServiceImpl) SaleReport(id string) ([]Report, error) {
	collection := s.Client.Database("kpay").Collection("report")
	var reports []Report
	idMerchant := bson.D{{"merchant_id", id}}
	cur, err := collection.Find(context.Background(), idMerchant)
	defer cur.Close(context.Background())
	if err != nil {
		return []Report{}, nil
	}
	for cur.Next(context.Background()) {
		var report Report
		err = cur.Decode(&report)
		if err != nil {
			return []Report{}, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func generateUsername() string {
	return randomdata.SillyName() + randomdata.StringNumber(5, "")
}

func generatePassword() string {
	res, err := password.Generate(16, 10, 5, false, false)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func isDuplicateBankAccount(collection *mongo.Collection, bankAccount string) bool {
	whereTo := bson.D{{"bank_account", bankAccount}}
	var limit int64 = 1
	cur, err := collection.Find(context.Background(), whereTo, &options.FindOptions{Limit:&limit})
	if err != nil {
		return false
	}
	if !cur.Next(context.Background()) {
		return false
	}
	return true
}