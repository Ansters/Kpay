package model

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/options"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type Product struct {
	ID            objectid.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	MerchantID    string 			`json:"-" bson:"merchant_id"`
	Name          string            `json:"name,omitempty"`
	Amount        float32           `json:"amount,omitempty"`
	AmountChanges []Amount         `json:"amount_changes,omitempty" bson:"amount_changes"`
}

type SaleProductIn struct {
	ID 			  	string	`json:"id,omitempty" bson:"id,omitempty"`
	Name		  	string	`json:"name,omitempty"`
	SellingVolume 	int		`json:"selling_volume" bson:"selling_volume"`
}

type SaleProductOut struct {
	Name		  	string	`json:"name,omitempty"`
	SellingVolume 	int		`json:"selling_volume" bson:"selling_volume"`
}

type Amount struct {
	Amount float32 `json:"amount"`
}

type ProductServiceImpl struct {
	Client *mongo.Client
}

func (s *ProductServiceImpl) All(id string) ([]Product, error) {

	var products []Product
	collection := s.Client.Database("kpay").Collection("product")
	idMerchant := bson.D{{"merchant_id", id}}
	cur, err := collection.Find(context.Background(), idMerchant)
	defer cur.Close(context.Background())
	if err != nil {
		return []Product{}, err
	}
	for cur.Next(context.Background()) {
		var product Product
		err = cur.Decode(&product)
		if err != nil {
			return []Product{}, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (s *ProductServiceImpl) Add(product *Product) error {

	collection := s.Client.Database("kpay").Collection("product")
	if isMaximunProducts(collection, product.MerchantID) {
		return errors.New("product limit exceeded")
	}
	res, err := collection.InsertOne(context.Background(), product)
	if err != nil {
		return err
	}
	product.ID = res.InsertedID.(objectid.ObjectID)
	product.AmountChanges = []Amount{}
	return nil
}

func (s *ProductServiceImpl) Update(product Product) error {
	
	idDoc := bson.D{{"_id", product.ID}, {"merchant_id", product.MerchantID}}
	collection := s.Client.Database("kpay").Collection("product")
	oldAmount, err := getCurrentAmount(idDoc, collection)
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(context.Background(), idDoc, bson.D{{"$push", bson.D{{"amount_changes", bson.D{{"amount", oldAmount}}}}}})
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(context.Background(), idDoc, bson.D{{"$set", bson.D{{"amount", product.Amount}}}})
	return err
}

func (s *ProductServiceImpl) Remove(merchantID string, productID string) error {

	if isUserBoughtProduct(s.Client, productID) {
		return errors.New("cannot remove this product.")
	}

	collection := s.Client.Database("kpay").Collection("product")

	objProductID, err := objectid.FromHex(productID)
	if err != nil {
		return err
	}
	idDoc := bson.D{{"_id", objProductID}, {"merchant_id", merchantID}}
	_, err = collection.DeleteOne(context.Background(), idDoc)
	return nil
}

func (s *ProductServiceImpl) Buy(saleProduct SaleProductIn) error {

	collection := s.Client.Database("kpay").Collection("report")
	today := time.Now().Format("2006-01-02")
	product, err := getProduct(s.Client, saleProduct.ID)
	if err != nil {
		return err
	}
	whereTo := bson.D{{"date", today}, {"merchant_id", product.MerchantID}}
	if !isHasSaleToday(collection, whereTo) {
		err := recordNewSale(collection, product.MerchantID, today)
		if err != nil {
			return err
		}
	}

	whereTo = bson.D{{"date", today}, {"merchant_id", product.MerchantID}, {"products.id", saleProduct.ID}}
	insert := bson.D{
		{
			"$inc", bson.D{
				{
					"products.$.selling_volume", saleProduct.SellingVolume,
				},
			},
		},
	}
	count, err := collection.UpdateOne(context.Background(), whereTo, insert)
	if err != nil {
		return err
	}

	whereTo = bson.D{{"date", today}, {"merchant_id", product.MerchantID}}
	if count.ModifiedCount == 0 {
		insert = bson.D{
				{
					"$push", bson.D{
					{
						"products", SaleProductIn{ID: saleProduct.ID, Name:product.Name, SellingVolume: saleProduct.SellingVolume},
					},
				},
				},
				{
					"$inc", bson.D{
					{
						"accumulate", float32(saleProduct.SellingVolume)*product.Amount,
					},
				},
			},
		}
	} else {
		insert = bson.D{
				{
					"$inc", bson.D{
					{
						"accumulate", float32(saleProduct.SellingVolume)*product.Amount,
					},
				},
			},
		}
	}
	_, err = collection.UpdateOne(context.Background(), whereTo, insert)
	return err
}

func getCurrentAmount(id bson.D, collection *mongo.Collection) (float32, error) {
	var product Product
	err := collection.FindOne(context.Background(), id).Decode(&product)
	return product.Amount, err
}

func isMaximunProducts(collection *mongo.Collection, id string) bool {
	idMerchant := bson.D{{"merchant_id", id}}
	count, err := collection.Count(context.Background(), idMerchant)
	if err != nil {
		return false
	}
	if count >= 5 {
		return true
	} 
	return false
}

func isUserBoughtProduct(client *mongo.Client, id string) bool {
	whereTo := bson.D{{"products.id", id}}
	collection := client.Database("kpay").Collection("report")
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

func isHasSaleToday(collection *mongo.Collection, query bson.D) bool {
	count, err := collection.Count(context.Background(), query)
	if err != nil {
		return false
	}
	if count >= 1 {
		return true
	}
	return false
}

func recordNewSale(collection *mongo.Collection, merchantID string, date string) error {
	report := Report{
		Date: date,
		MerchantID: merchantID,
		Products: []SaleProductOut{},
		Accumulate: 0.0,
	}
	_, err := collection.InsertOne(context.Background(), report)
	return err
}

func getProduct(client *mongo.Client, id string) (Product, error) {
	var product Product
	objID, err := objectid.FromHex(id)
	if err != nil {
		return Product{}, err
	}
	idDoc := bson.D{{"_id", objID}}
	collection := client.Database("kpay").Collection("product")
	err = collection.FindOne(context.Background(), idDoc).Decode(&product)
	return product, err
}