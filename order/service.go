package main

import (
	"context"
	"errors"
	"log"
	"sync"

	// "github.com/alabarjasteh/deployment/order/repo/product/"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"repo/product"
)

type inmemService struct {
	// use this lock to prevent concurrent access to shared obj
	mux sync.RWMutex

	// this map acts as the database of the service
	cart map[UserID][]CartProduct

	// use this gRPC client in service methods
	productClient product.ProductServiceClient
}

func NewInmemservice(addr string) *inmemService {
	prodConn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer prodConn.Close()

	prodClient := product.NewProductServiceClient(prodConn)

	return &inmemService{productClient: prodClient}
}

type UserID = string
type CartProduct struct {
	ID    int    `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Units int    `json:"units,omitempty"`
}

var (
	ErrAlreadyExists   = errors.New("already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrProductNotFound = errors.New("product not found")
)

func (s *inmemService) GetCartProducts(ctx context.Context, userID UserID) ([]CartProduct, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	products, ok := s.cart[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return products, nil
}

func (s *inmemService) AddItemToCart(ctx context.Context, userID UserID, PID, numberOfUnits int) error {
	s.mux.RLocker()
	defer s.mux.RUnlock()

	cart, ok := s.cart[userID]
	if !ok {
		return ErrUserNotFound
	}
	p := CartProduct{
		ID:    PID,
		Units: numberOfUnits,
	}
	cart = append(cart, p)
	s.cart[userID] = cart
	return nil
}

func (s *inmemService) ModifyCart(ctx context.Context, userID string, PID, offset int) error {
	s.mux.RLocker()
	defer s.mux.RUnlock()

	cart, ok := s.cart[userID]
	if !ok {
		return ErrUserNotFound
	}
	var err error = ErrProductNotFound
	for _, v := range cart {
		if v.ID == PID {
			v.Units += offset
			err = nil
			break
		}
	}
	return err
}
