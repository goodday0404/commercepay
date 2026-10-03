package cart

import "github.com/google/uuid"

type addItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type createCartResponse struct {
	ID uuid.UUID `json:"id"`
}

type cartItemResponse struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Quantity       int       `json:"quantity"`
	UnitPriceMinor int64     `json:"unit_price_minor"`
	Currency       string    `json:"currency"`
}

type cartResponse struct {
	ID    uuid.UUID          `json:"id"`
	Items []cartItemResponse `json:"items"`
}

func toCartItemResponse(item CartItem) cartItemResponse {
	return cartItemResponse{
		ID:             item.ID(),
		ProductID:      item.ProductID(),
		Quantity:       item.Quantity(),
		UnitPriceMinor: item.UnitPriceMinor(),
		Currency:       item.Currency(),
	}
}

func toCartResponse(cart Cart) cartResponse {
	items := cart.Items()
	responseItems := make([]cartItemResponse, 0, len(items))

	for _, item := range items {
		responseItems = append(responseItems, toCartItemResponse(item))
	}

	return cartResponse{
		ID:    cart.ID(),
		Items: responseItems,
	}
}
