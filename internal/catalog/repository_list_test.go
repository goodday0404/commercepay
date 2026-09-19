package catalog

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRepositoryList_FirstPageReturnsLimitAndNextCursor(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	baseTime := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	oldestID := uuid.New()
	middleID := uuid.New()
	newestID := uuid.New()

	insertProductFixture(t, db, oldestID, "PRODUCT-1", baseTime.Add(1*time.Minute))
	insertProductFixture(t, db, middleID, "PRODUCT-2", baseTime.Add(2*time.Minute))
	insertProductFixture(t, db, newestID, "PRODUCT-3", baseTime.Add(3*time.Minute))

	ctx := testContext(t)

	page, err := repo.List(ctx, 2, nil)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}

	if len(page.Products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(page.Products))
	}

	if page.Products[0].ID != newestID {
		t.Fatalf("expected first product %s, got %s", newestID, page.Products[0].ID)
	}

	if page.Products[1].ID != middleID {
		t.Fatalf("expected second product %s, got %s", middleID, page.Products[1].ID)
	}

	if page.NextCursor == nil {
		t.Fatal("expected next cursor, got nil")
	}

	if page.NextCursor.ID != middleID {
		t.Fatalf("expected cursor ID %s, got %s", middleID, page.NextCursor.ID)
	}

	expectedCursorTime := baseTime.Add(2 * time.Minute)

	if !page.NextCursor.CreatedAt.Equal(expectedCursorTime) {
		t.Fatalf(
			"expected cursor time %s, got %s",
			expectedCursorTime,
			page.NextCursor.CreatedAt,
		)
	}
}

func TestRepositoryList_CursorContinuesWithoutDuplicatesOrMissingProducts(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	baseTime := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	ids := make([]uuid.UUID, 5)

	for i := 0; i < 5; i++ {
		ids[i] = uuid.New()

		insertProductFixture(
			t,
			db,
			ids[i],
			fmt.Sprintf("PRODUCT-%d", i+1),
			baseTime.Add(time.Duration(i+1)*time.Minute),
		)
	}

	var (
		cursor *ProductCursor
		gotIDs []uuid.UUID
	)

	for {
		ctx := testContext(t)
		page, err := repo.List(ctx, 2, cursor)
		if err != nil {
			t.Fatalf("list products: %v", err)
		}

		for _, product := range page.Products {
			gotIDs = append(gotIDs, product.ID)
		}

		if page.NextCursor == nil {
			break
		}

		cursor = page.NextCursor
	}

	if len(gotIDs) != 5 {
		t.Fatalf("expected 5 products across all pages, got %d", len(gotIDs))
	}

	seen := make(map[uuid.UUID]bool)

	for _, id := range gotIDs {
		if seen[id] {
			t.Fatalf("product %s appeared more than once", id)
		}

		seen[id] = true
	}

	for _, expectedID := range ids {
		if !seen[expectedID] {
			t.Fatalf("product %s was missing from pagination results", expectedID)
		}
	}
}

func TestRepositoryList_UsesIDAsTieBreakerForSameTimestamp(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	createdAt := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	id3 := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	insertProductFixture(t, db, id1, "PRODUCT-1", createdAt)
	insertProductFixture(t, db, id2, "PRODUCT-2", createdAt)
	insertProductFixture(t, db, id3, "PRODUCT-3", createdAt)

	ctx := testContext(t)

	firstPage, err := repo.List(ctx, 2, nil)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}

	if len(firstPage.Products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(firstPage.Products))
	}

	if firstPage.Products[0].ID != id3 {
		t.Fatalf("expected first ID %s, got %s", id3, firstPage.Products[0].ID)
	}

	if firstPage.Products[1].ID != id2 {
		t.Fatalf("expected second ID %s, got %s", id2, firstPage.Products[1].ID)
	}

	if firstPage.NextCursor == nil {
		t.Fatal("expected next cursor")
	}

	secondPage, err := repo.List(ctx, 2, firstPage.NextCursor)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}

	if len(secondPage.Products) != 1 {
		t.Fatalf("expected 1 product on second page, got %d", len(secondPage.Products))
	}

	if secondPage.Products[0].ID != id1 {
		t.Fatalf("expected remaining ID %s, got %s", id1, secondPage.Products[0].ID)
	}
}

func TestRepositoryList_NewInsertDoesNotShiftExistingCursor(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	baseTime := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	p1 := uuid.New()
	p2 := uuid.New()
	p3 := uuid.New()
	p4 := uuid.New()
	p5 := uuid.New()

	insertProductFixture(t, db, p1, "P1", baseTime.Add(1*time.Minute))
	insertProductFixture(t, db, p2, "P2", baseTime.Add(2*time.Minute))
	insertProductFixture(t, db, p3, "P3", baseTime.Add(3*time.Minute))
	insertProductFixture(t, db, p4, "P4", baseTime.Add(4*time.Minute))
	insertProductFixture(t, db, p5, "P5", baseTime.Add(5*time.Minute))

	ctx := testContext(t)

	firstPage, err := repo.List(ctx, 2, nil)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}

	if len(firstPage.Products) != 2 {
		t.Fatalf("expected first page size 2, got %d", len(firstPage.Products))
	}

	if firstPage.Products[0].ID != p5 {
		t.Fatalf("expected first product P5, got %s", firstPage.Products[0].ID)
	}

	if firstPage.Products[1].ID != p4 {
		t.Fatalf("expected second product P4, got %s", firstPage.Products[1].ID)
	}

	if firstPage.NextCursor == nil {
		t.Fatal("expected next cursor")
	}

	newProductID := uuid.New()

	insertProductFixture(t, db, newProductID, "NEW", baseTime.Add(6*time.Minute))

	secondPage, err := repo.List(ctx, 2, firstPage.NextCursor)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}

	if len(secondPage.Products) != 2 {
		t.Fatalf("expected second page size 2, got %d", len(secondPage.Products))
	}

	if secondPage.Products[0].ID != p3 {
		t.Fatalf("expected first product on page 2 to be P3, got %s", secondPage.Products[0].ID)
	}

	if secondPage.Products[1].ID != p2 {
		t.Fatalf("expected second product on page 2 to be P2, got %s", secondPage.Products[1].ID)
	}

	for _, product := range secondPage.Products {
		if product.ID == p5 || product.ID == p4 {
			t.Fatalf("page 2 repeated a product from page 1: %s", product.ID)
		}

		if product.ID == newProductID {
			t.Fatalf("page 2 unexpectedly included newly inserted product %s", product.ID)
		}
	}
}
