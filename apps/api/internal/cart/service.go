package cart

import (
	"context"
	"errors"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/3dprint-hub/api/internal/database"
)

type Service struct {
	db     *gorm.DB
	logger *slog.Logger
}

type ItemInput struct {
	SKU            string
	DisplayName    string
	Quantity       int
	UnitPriceCents int
	Metadata       map[string]any
}

type CartDTO struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Items     []CartItemDTO
	Subtotal  int
	UpdatedAt time.Time
}

type CartItemDTO struct {
	ID             uuid.UUID
	SKU            string
	DisplayName    string
	Quantity       int
	UnitPriceCents int
	Metadata       map[string]any
}

func New(db *gorm.DB, logger *slog.Logger) *Service {
	return &Service{db: db, logger: logger}
}

func (s *Service) GetByUser(ctx context.Context, userID uuid.UUID) (CartDTO, error) {
	var cart database.Cart
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		First(&cart).Error; err != nil {
		return CartDTO{}, err
	}
	return toDTO(cart), nil
}

func (s *Service) AddItem(ctx context.Context, userID uuid.UUID, input ItemInput) (CartDTO, error) {
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var cart database.Cart
	if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		tx.Rollback()
		return CartDTO{}, err
	}
	var item database.CartItem
	err := tx.Where("cart_id = ? AND sku = ?", cart.ID, input.SKU).First(&item).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		item = database.CartItem{
			CartID:         cart.ID,
			SKU:            input.SKU,
			DisplayName:    input.DisplayName,
			Quantity:       input.Quantity,
			UnitPriceCents: input.UnitPriceCents,
			Metadata:       input.Metadata,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return CartDTO{}, err
		}
	case err != nil:
		tx.Rollback()
		return CartDTO{}, err
	default:
		item.Quantity += input.Quantity
		item.UnitPriceCents = input.UnitPriceCents
		if input.DisplayName != "" {
			item.DisplayName = input.DisplayName
		}
		if input.Metadata != nil {
			item.Metadata = input.Metadata
		}
		if err := tx.Save(&item).Error; err != nil {
			tx.Rollback()
			return CartDTO{}, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return CartDTO{}, err
	}
	return s.GetByUser(ctx, userID)
}

func (s *Service) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) (CartDTO, error) {
	var cart database.Cart
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return CartDTO{}, err
	}
	if err := s.db.WithContext(ctx).Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&database.CartItem{}).Error; err != nil {
		return CartDTO{}, err
	}
	return s.GetByUser(ctx, userID)
}

func (s *Service) Clear(ctx context.Context, userID uuid.UUID) error {
	var cart database.Cart
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Where("cart_id = ?", cart.ID).Delete(&database.CartItem{}).Error
}

func toDTO(cart database.Cart) CartDTO {
	items := make([]CartItemDTO, len(cart.Items))
	subtotal := 0
	for i, item := range cart.Items {
		items[i] = CartItemDTO{
			ID:             item.ID,
			SKU:            item.SKU,
			DisplayName:    item.DisplayName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			Metadata:       item.Metadata,
		}
		subtotal += item.Quantity * item.UnitPriceCents
	}
	return CartDTO{
		ID:        cart.ID,
		UserID:    cart.UserID,
		Items:     items,
		Subtotal:  subtotal,
		UpdatedAt: cart.UpdatedAt,
	}
}
