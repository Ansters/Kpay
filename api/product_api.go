package api

import (
	"kpay/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

type ProductService interface {
	All(id string) ([]model.Product, error)
	Add(product *model.Product) error
	Update(product model.Product) error
	Remove(merchantID string, productID string) error
	Buy(product model.SaleProductIn) error
}

type ProductHandler struct {
	productService ProductService
}

func (h *ProductHandler) findAllProduct(c *gin.Context) {
	id := c.Param("id")
	merchants, err := h.productService.All(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, merchants)
}

func (h *ProductHandler) addProduct(c *gin.Context) {
	var product model.Product
	id := c.Param("id")
	err := c.ShouldBindJSON(&product)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	product.MerchantID = id
	err = h.productService.Add(&product)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) updateProduct(c *gin.Context) {
	var product model.Product
	merchantID := c.Param("id")
	productID := c.Param("product_id")
	err := c.ShouldBindJSON(&product)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	objProductID, err := objectid.FromHex(productID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	product.ID = objProductID
	product.MerchantID = merchantID
	err = h.productService.Update(product)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) removeProduct(c *gin.Context) {
	merchantID := c.Param("id")
	productID := c.Param("product_id")
	err := h.productService.Remove(merchantID, productID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "DELETED Product: " + productID)
}

func (h *ProductHandler) buyProduct(c *gin.Context) {
	var saleProduct model.SaleProductIn
	err := c.ShouldBind(&saleProduct)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.productService.Buy(saleProduct)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "Thank you!")
}
