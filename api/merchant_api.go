package api

import (
	"kpay/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

type MerchantService interface {
	Auth(username string, psw string) (model.Merchant, error)
	Register(merchant *model.Merchant) error
	Find(id string) (model.Merchant, error)
	Update(merchant model.Merchant) error
	SaleReport(id string) ([]model.Report, error)
}

type MerchantHandler struct {
	merchantService MerchantService
}

func (h *MerchantHandler) registerMerchant(c *gin.Context) {
	var merchant model.Merchant
	err := c.ShouldBindJSON(&merchant)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.merchantService.Register(&merchant)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, merchant)
}

func (h *MerchantHandler) getMerchant(c *gin.Context) {
	id := c.Param("id")
	merchant, err := h.merchantService.Find(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, merchant)
}

func (h *MerchantHandler) updateMerchant(c *gin.Context) {
	id := c.Param("id")
	objID, err := objectid.FromHex(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var merchant model.Merchant
	err = c.ShouldBindJSON(&merchant)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	merchant.ID = objID
	err = h.merchantService.Update(merchant)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, merchant)
}

func (h *MerchantHandler) getSaleReport(c *gin.Context) {
	id := c.Param("id")
	reports, err := h.merchantService.SaleReport(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *MerchantHandler) authUser(c *gin.Context) {
	username := c.GetHeader("username")
	psw := c.GetHeader("password")
	id := c.Param("id")

	merchant, err := h.merchantService.Auth(username, psw)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if merchant.ID.Hex() != id {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Next()
}
