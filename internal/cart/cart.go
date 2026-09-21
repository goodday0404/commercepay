package cart

import "github.com/google/uuid"

type Cart struct {
	id    uuid.UUID
	items []CartItem
}

type CartItem struct {
	id        uuid.UUID
	productID uuid.UUID
	quantity  int
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
