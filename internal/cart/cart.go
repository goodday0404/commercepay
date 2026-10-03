package cart

import (
	"strings"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/google/uuid"
)

type Cart struct {
	id    uuid.UUID
	items []CartItem
}

type CartItem struct {
	id             uuid.UUID
	productID      uuid.UUID
	quantity       int
	unitPriceMinor int64
	currency       string
}

type CartDetails struct {
	ID    uuid.UUID
	Items []CartItemDetails
}

type CartItemDetails struct {
	Item    CartItem
	Product catalog.Product
}

func NewCart() Cart {
	return Cart{
		id: uuid.New(),
	}
}

func (c Cart) ID() uuid.UUID {
	return c.id
}

func (c Cart) IsEmpty() bool {
	return len(c.items) == 0
}

func (c Cart) Items() []CartItem {
	items := make([]CartItem, len(c.items))
	copy(items, c.items)
	return items
}

func newCartItem(productID uuid.UUID, quantity int, unitPriceMinor int64, currency string) (CartItem, error) {
	if productID == uuid.Nil {
		return CartItem{}, ErrInvalidProductID
	}

	if quantity <= 0 {
		return CartItem{}, ErrInvalidQuantity
	}

	if unitPriceMinor < 0 {
		return CartItem{}, ErrNegativeUnitPrice
	}

	if strings.TrimSpace(currency) == "" {
		return CartItem{}, ErrEmptyCurrency
	}

	return CartItem{
		id:             uuid.New(),
		productID:      productID,
		quantity:       quantity,
		unitPriceMinor: unitPriceMinor,
		currency:       currency,
	}, nil
}

func (i CartItem) ID() uuid.UUID {
	return i.id
}

func (i CartItem) ProductID() uuid.UUID {
	return i.productID
}

func (i CartItem) Quantity() int {
	return i.quantity
}

func (i CartItem) UnitPriceMinor() int64 {
	return i.unitPriceMinor
}

func (i CartItem) Currency() string {
	return i.currency
}
