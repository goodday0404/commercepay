package catalog

type createProductRequest struct {
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
	Available  *bool  `json:"available"`
}

type productResponse struct {
	ID         string `json:"id"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
	Available  bool   `json:"available"`
}

type listProductsResponse struct {
	Items      []productResponse `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

func newProductResponse(product Product) productResponse {
	return productResponse{
		ID:         product.ID.String(),
		SKU:        product.SKU,
		Name:       product.Name,
		PriceMinor: product.PriceMinor,
		Currency:   product.Currency,
		Available:  product.Available,
	}
}

func newProductResponses(products []Product) []productResponse {
	productResponses := make([]productResponse, 0, len(products))

	for _, product := range products {
		productResponses = append(productResponses, newProductResponse(product))
	}

	return productResponses
}
