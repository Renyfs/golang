package service

import (
	"context"
)

type ProductService struct {
}

func (p ProductService) GetProductStock(ctx context.Context, req *ProductRequest) (*ProductResponse, error) {

	return &ProductResponse{ProdStock: req.ProdId}, nil
}
func (p ProductService) mustEmbedUnimplementedProdServiceServer() {}
