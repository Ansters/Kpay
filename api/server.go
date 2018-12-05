package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/mongo"
	"kpay/model"
)

func StartServer(addr string, client *mongo.Client) error {
	r := gin.Default()
	mHandler := &MerchantHandler {
		merchantService: &model.MerchantServiceImpl{
			Client: client,
		},
	}

	pHandler := &ProductHandler {
		productService: &model.ProductServiceImpl{
			Client: client,
		},
	}

	r.POST("/merchants/register", mHandler.registerMerchant)
	r.POST("/buy/product", pHandler.buyProduct)

	r.Use(mHandler.authUser)
	r.GET("/merchant/:id", mHandler.getMerchant)
	r.PUT("/merchant/:id", mHandler.updateMerchant)
	r.GET("/merchant/:id/report", mHandler.getSaleReport)
	r.GET("/merchant/:id/products", pHandler.findAllProduct)
	r.POST("/merchant/:id/product", pHandler.addProduct)
	r.POST("/merchant/:id/product/:product_id", pHandler.updateProduct)
	r.DELETE("/merchant/:id/product/:product_id", pHandler.removeProduct)

	return r.Run(addr)
}