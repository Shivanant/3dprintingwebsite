package order

import (
	"context"
	"errors"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/3dprint-hub/api/internal/database"
)

var (
	ErrEmptyCart = errors.New("cart is empty")
)

type Service struct {
	db     *gorm.DB
	logger *slog.Logger
}

type CheckoutInput struct {
	Notes string
}

func New(db *gorm.DB, logger *slog.Logger) *Service {
	return &Service{db: db, logger: logger}
}

func (s *Service) Checkout(ctx context.Context, userID uuid.UUID, input CheckoutInput) (*database.Order, error) {
	var cart database.Cart
	if err := s.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}
	order := &database.Order{
		UserID:   userID,
		Status:   "pending",
		Currency: "USD",
		Notes:    input.Notes,
	}
	subtotal := 0
	items := make([]database.OrderItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = database.OrderItem{
			Name:           item.DisplayName,
			Description:    "",
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			Metadata:       item.Metadata,
		}
		subtotal += item.Quantity * item.UnitPriceCents
	}
	order.SubtotalCents = subtotal
	order.TaxCents = int(float64(subtotal) * 0.08)
	order.TotalCents = order.SubtotalCents + order.TaxCents
	now := time.Now()
	order.PlacedAt = &now

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		order.Items = items
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&database.CartItem{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]database.Order, error) {
	var orders []database.Order
	if err := s.db.WithContext(ctx).Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *Service) Get(ctx context.Context, userID, orderID uuid.UUID) (*database.Order, error) {
	var order database.Order
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ? AND id = ?", userID, orderID).
		First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *Service) AdminList(ctx context.Context) ([]database.Order, error) {
	var orders []database.Order
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *Service) UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	return s.db.WithContext(ctx).Model(&database.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}
