package cart

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
)

func TestRepositoryInsertCart(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
	repo := NewRepository(db)

	cart := NewCart()
	ctx := testutil.TestContext(t)

	if err := repo.Insert(ctx, cart); err != nil {
		t.Fatalf("insert cart: %v", err)
	}

	var gotID uuid.UUID

	if err := db.QueryRow(ctx, `SELECT id FROM carts WHERE id = $1`, cart.ID()).Scan(&gotID); err != nil {
		t.Fatalf("query inserted cart: %v", err)
	}

	if gotID != cart.ID() {
		t.Fatalf("expected cart ID %s, got %s", cart.ID(), gotID)
	}
}

func TestRepositoryAddItem_FirstAdd(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	// ---------------------------------------------------------
	// Arrange: create a cart and a real Product.
	// ---------------------------------------------------------

	cartRepo, _, cart, product := setTestCartandProduct(t, db)

	// ---------------------------------------------------------
	// Arrange: create the candidate CartItem.
	// ---------------------------------------------------------

	candidate, err := newCartItem(
		product.ID,
		2,
		product.PriceMinor,
		product.Currency,
	)
	if err != nil {
		t.Fatalf("new cart item: %v", err)
	}

	// ---------------------------------------------------------
	// Act.
	// ---------------------------------------------------------

	ctx := testutil.TestContext(t)

	got, err := cartRepo.AddItem(ctx, cart.ID(), candidate)
	if err != nil {
		t.Fatalf("add item: %v", err)
	}

	// ---------------------------------------------------------
	// Assert: the INSERT branch returns the newly-created item.
	// ---------------------------------------------------------

	if got.ID() != candidate.ID() {
		t.Fatalf("expected item ID %s, got %s", candidate.ID(), got.ID())
	}

	if got.ProductID() != product.ID {
		t.Fatalf("expected product ID %s, got %s", product.ID, got.ProductID())
	}

	if got.Quantity() != 2 {
		t.Fatalf("expected quantity 2, got %d", got.Quantity())
	}

	if got.UnitPriceMinor() != 1499 {
		t.Fatalf("expected unit price 1499, got %d", got.UnitPriceMinor())
	}

	if got.Currency() != "CAD" {
		t.Fatalf("expected currency CAD, got %q", got.Currency())
	}

	// ---------------------------------------------------------
	// Assert independently against PostgreSQL.
	//
	// We don't only trust RETURNING.
	// ---------------------------------------------------------

	var (
		dbID             uuid.UUID
		dbCartID         uuid.UUID
		dbProductID      uuid.UUID
		dbQuantity       int
		dbUnitPriceMinor int64
		dbCurrency       string
	)

	err = db.QueryRow(
		ctx,
		`
		SELECT
			id,
			cart_id,
			product_id,
			quantity,
			unit_price_minor,
			currency
		FROM cart_items
		WHERE cart_id = $1
		  AND product_id = $2
		`,
		cart.ID(),
		product.ID,
	).Scan(
		&dbID,
		&dbCartID,
		&dbProductID,
		&dbQuantity,
		&dbUnitPriceMinor,
		&dbCurrency,
	)
	if err != nil {
		t.Fatalf("query persisted cart item: %v", err)
	}

	if dbID != got.ID() {
		t.Fatalf("expected persisted item ID %s, got %s", got.ID(), dbID)
	}

	if dbCartID != cart.ID() {
		t.Fatalf("expected persisted cart ID %s, got %s", cart.ID(), dbCartID)
	}

	if dbProductID != product.ID {
		t.Fatalf("expected persisted product ID %s, got %s", product.ID, dbProductID)
	}

	if dbQuantity != 2 {
		t.Fatalf("expected persisted quantity 2, got %d", dbQuantity)
	}

	if dbUnitPriceMinor != 1499 {
		t.Fatalf("expected persisted unit price 1499, got %d", dbUnitPriceMinor)
	}

	if dbCurrency != "CAD" {
		t.Fatalf("expected persisted currency CAD, got %q", dbCurrency)
	}
}

func TestRepositoryAddItem_SameProductIncrementsQuantity(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	// ---------------------------------------------------------
	// Arrange: create a cart and a real Product.
	// ---------------------------------------------------------

	cartRepo, _, cart, product := setTestCartandProduct(t, db)

	// ---------------------------------------------------------
	// Arrange: create the candidate CartItem.
	// ---------------------------------------------------------

	firstCandidate, err := newCartItem(
		product.ID,
		2,
		1499,
		"CAD",
	)
	if err != nil {
		t.Fatalf("new first cart item: %v", err)
	}

	// ---------------------------------------------------------
	// First add:
	//
	// P1 × 2 @ 1499 CAD
	// ---------------------------------------------------------

	ctx := testutil.TestContext(t)

	firstResult, err := cartRepo.AddItem(ctx, cart.ID(), firstCandidate)
	if err != nil {
		t.Fatalf("add first item: %v", err)
	}

	if firstResult.ID() != firstCandidate.ID() {
		t.Fatalf("expected first item ID %s, got %s", firstCandidate.ID(), firstResult.ID())
	}

	if firstResult.Quantity() != 2 {
		t.Fatalf("expected initial quantity 2, got %d", firstResult.Quantity())
	}

	// Save the original durable CartItem identity.
	// This identity must survive the second AddItem call.
	firstItemID := firstResult.ID()

	// ---------------------------------------------------------
	// Second add:
	//
	// P1 × 1
	//
	// Use a different price to also prove that the temporary
	// Cart snapshot is refreshed during ON CONFLICT.
	// ---------------------------------------------------------

	secondCandidate, err := newCartItem(
		product.ID,
		1,
		1599,
		"CAD",
	)
	if err != nil {
		t.Fatalf("new second cart item: %v", err)
	}

	// The second candidate gets its own new UUID.
	//
	// It must NOT replace the identity of the existing CartItem.
	if secondCandidate.ID() == firstItemID {
		t.Fatal(
			"expected second candidate to have a different item ID",
		)
	}

	// ---------------------------------------------------------
	// Act: add the same Product again.
	// ---------------------------------------------------------

	updated, err := cartRepo.AddItem(ctx, cart.ID(), secondCandidate)
	if err != nil {
		t.Fatalf("add same product again: %v", err)
	}

	// ---------------------------------------------------------
	// Assert: existing identity was preserved.
	// ---------------------------------------------------------

	if updated.ID() != firstItemID {
		t.Fatalf("expected existing item ID %s, got %s", firstItemID, updated.ID())
	}

	if updated.ID() == secondCandidate.ID() {
		t.Fatalf("expected second candidate ID %s not to replace existing item ID", secondCandidate.ID())
	}

	// ---------------------------------------------------------
	// Assert: quantity was incremented.
	//
	// existing 2 + requested 1 = 3
	// ---------------------------------------------------------

	if updated.Quantity() != 3 {
		t.Fatalf("expected quantity 3, got %d", updated.Quantity())
	}

	if updated.ProductID() != product.ID {
		t.Fatalf("expected product ID %s, got %s", product.ID, updated.ProductID())
	}

	// ---------------------------------------------------------
	// Assert: snapshot was refreshed from the new candidate.
	// ---------------------------------------------------------

	if updated.UnitPriceMinor() != 1599 {
		t.Fatalf("expected refreshed unit price 1599, got %d", updated.UnitPriceMinor())
	}

	if updated.Currency() != "CAD" {
		t.Fatalf("expected currency CAD, got %q", updated.Currency())
	}

	// ---------------------------------------------------------
	// Assert independently against PostgreSQL.
	// ---------------------------------------------------------

	var (
		dbID             uuid.UUID
		dbQuantity       int
		dbUnitPriceMinor int64
		dbCurrency       string
	)

	err = db.QueryRow(
		ctx,
		`
		SELECT
			id,
			quantity,
			unit_price_minor,
			currency
		FROM cart_items
		WHERE cart_id = $1
		  AND product_id = $2
		`,
		cart.ID(),
		product.ID,
	).Scan(
		&dbID,
		&dbQuantity,
		&dbUnitPriceMinor,
		&dbCurrency,
	)
	if err != nil {
		t.Fatalf("query updated cart item: %v", err)
	}

	if dbID != firstItemID {
		t.Fatalf("expected persisted item ID %s, got %s", firstItemID, dbID)
	}

	if dbQuantity != 3 {
		t.Fatalf("expected persisted quantity 3, got %d", dbQuantity)
	}

	if dbUnitPriceMinor != 1599 {
		t.Fatalf("expected persisted unit price 1599, got %d", dbUnitPriceMinor)
	}

	if dbCurrency != "CAD" {
		t.Fatalf("expected persisted currency CAD, got %q", dbCurrency)
	}

	// ---------------------------------------------------------
	// Finally prove that the upsert did NOT create a second row.
	// ---------------------------------------------------------

	var count int

	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM cart_items
		WHERE cart_id = $1
		  AND product_id = $2
		`,
		cart.ID(),
		product.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count cart item rows: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected exactly one cart item row, got %d", count)
	}
}

func TestRepositoryAddItem_CartNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	// ---------------------------------------------------------
	// Arrange: create a cart and a real Product.
	// ---------------------------------------------------------

	cartRepo, _, _, product := setTestCartandProduct(t, db)

	// ---------------------------------------------------------
	// Arrange: deliberately use a Cart ID that does not exist.
	// ---------------------------------------------------------

	missingCartID := uuid.New()

	// Sanity-check the fixture.
	//
	// This query belongs in the test, not production code.
	// We are proving that the ID is genuinely absent before
	// exercising Repository.AddItem.
	var cartCount int
	ctx := testutil.TestContext(t)

	err := db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM carts
		WHERE id = $1
		`,
		missingCartID,
	).Scan(&cartCount)
	if err != nil {
		t.Fatalf("verify missing cart: %v", err)
	}

	if cartCount != 0 {
		t.Fatalf("expected cart %s not to exist", missingCartID)
	}

	// ---------------------------------------------------------
	// Arrange: construct an otherwise valid CartItem.
	//
	// We want the ONLY invalid relationship to be cart_id.
	// ---------------------------------------------------------

	item, err := newCartItem(
		product.ID,
		1,
		product.PriceMinor,
		product.Currency,
	)
	if err != nil {
		t.Fatalf("new cart item: %v", err)
	}

	// ---------------------------------------------------------
	// Act.
	//
	// PostgreSQL should reject this because:
	//
	// cart_items.cart_id
	//     → carts.id
	//
	// cannot reference missingCartID.
	// ---------------------------------------------------------

	_, err = cartRepo.AddItem(ctx, missingCartID, item)

	// ---------------------------------------------------------
	// Assert: Repository translated the specific FK failure
	// into Cart's semantic ErrCartNotFound.
	// ---------------------------------------------------------

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}

	// ---------------------------------------------------------
	// Assert: failed INSERT produced no CartItem.
	//
	// PostgreSQL statements are atomic. A failed FK check
	// must not leave a partially inserted row.
	// ---------------------------------------------------------

	var itemCount int

	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM cart_items
		WHERE id = $1
		`,
		item.ID(),
	).Scan(&itemCount)
	if err != nil {
		t.Fatalf("count cart item rows: %v", err)
	}

	if itemCount != 0 {
		t.Fatalf("expected no cart item to be persisted, got %d rows", itemCount)
	}
}

func TestRepositoryAddItem_ConcurrentAdds(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	// ---------------------------------------------------------
	// Arrange: create a cart and a real Product.
	// ---------------------------------------------------------

	cartRepo, _, cart, product := setTestCartandProduct(t, db)

	// ---------------------------------------------------------
	// Arrange: create the candidate CartItem.
	// ---------------------------------------------------------

	initial, err := newCartItem(
		product.ID,
		2,
		1499,
		"CAD",
	)
	if err != nil {
		t.Fatalf("new first cart item: %v", err)
	}

	// ---------------------------------------------------------
	// Act.
	//
	// insert initial cart item:
	//
	// ---------------------------------------------------------

	ctx := testutil.TestContext(t)

	if _, err := cartRepo.AddItem(ctx, cart.ID(), initial); err != nil {
		t.Fatalf("add initial item: %v", err)
	}

	// ---------------------------------------------------------
	// Act.
	//
	// attempt 2 concurrent insertion of a item that has same (cart_id, product_id)
	// with the initial cart item:
	//
	// ---------------------------------------------------------

	start := make(chan struct{})
	errs := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	add := func() {
		defer wg.Done()

		<-start

		item, err := newCartItem(
			product.ID,
			1,
			product.PriceMinor,
			product.Currency,
		)
		if err != nil {
			errs <- err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = cartRepo.AddItem(ctx, cart.ID(), item)

		errs <- err
	}

	go add()
	go add()

	close(start)

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent add: %v", err)
		}
	}

	// ---------------------------------------------------------
	// Assert: correct quantity of the initial CartItem.
	// ---------------------------------------------------------

	var quantity int

	ctx = testutil.TestContext(t)

	err = db.QueryRow(
		ctx,
		`
		SELECT quantity
		FROM cart_items
		WHERE cart_id = $1
		  AND product_id = $2
		`,
		cart.ID(),
		product.ID,
	).Scan(&quantity)
	if err != nil {
		t.Fatalf("query final quantity: %v", err)
	}

	if quantity != 4 {
		t.Fatalf("expected final quantity 4, got %d", quantity)
	}
}
