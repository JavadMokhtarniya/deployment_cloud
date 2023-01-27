package main

import (
	"errors"
	"net/http"
	"strconv"
	// "order/vendor/github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	
)

type handler struct {
	service *inmemService
}

func NewHandler(service *inmemService) *handler {
	return &handler{service}
}

var (
	ErrJWTIsMissing = errors.New("jwt is missing")
	ErrUnathorized  = errors.New("unathorized user")
)

func (h *handler) getBasket(c *gin.Context) {
	uid, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	basket, err := h.service.GetCartProducts(c.Request.Context(), uid)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}

	c.JSON(http.StatusOK, basket)
}

func (h *handler) addProductToBasket(c *gin.Context) {
	uid, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	numberOfUnits, err := strconv.Atoi(c.Query("units"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.service.AddItemToCart(c.Request.Context(), uid, pid, numberOfUnits)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
}

func (h *handler) modifyBasket(c *gin.Context) {
	uid, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	numberOfUnits, err := strconv.Atoi(c.Query("units"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.service.ModifyCart(c.Request.Context(), uid, pid, numberOfUnits)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		if errors.Is(err, ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func GetUserID(c *gin.Context) (string, error) {
	uid := c.Value("userID")
	if uid == nil {
		return "", nil
	}
	usrID, ok := uid.(string)
	if !ok {
		err := errors.New("could not parse user ID")
		return "", err
	}
	return usrID, nil
}
