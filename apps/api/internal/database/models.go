package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UUIDBase struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *UUIDBase) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type User struct {
	UUIDBase
	Email        string  `gorm:"uniqueIndex"`
	PasswordHash *string
	Name         string
	AvatarURL    *string
	Role         string `gorm:"default:user"`

	OAuthAccounts    []OAuthAccount
	RefreshTokens    []RefreshToken
	Cart             Cart
	Orders           []Order
	PrintJobs        []PrintJob
	PasswordResets   []PasswordReset
	LastLoginAt      *time.Time
	EmailVerifiedAt  *time.Time
	DefaultMaterial  string `gorm:"default:PLA"`
	DefaultPrintQual string `gorm:"default:standard"`
}

type OAuthAccount struct {
	UUIDBase
	UserID          uuid.UUID `gorm:"type:uuid;index"`
	Provider        string    `gorm:"index"`
	ProviderUserID  string
	AccessToken     string
	RefreshToken    string
	ExpiresAt       *time.Time
	Scopes          string
	ProviderPayload string `gorm:"type:jsonb"`
}

type PasswordReset struct {
	UUIDBase
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Token     string    `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type RefreshToken struct {
	UUIDBase
	UserID        uuid.UUID `gorm:"type:uuid;index"`
	TokenHash     string    `gorm:"uniqueIndex"`
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	LastIP        string
	UserAgent     string
	RotatedFromID *uuid.UUID `gorm:"type:uuid"`
}

type Cart struct {
	UUIDBase
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	Items  []CartItem
}

type CartItem struct {
	UUIDBase
	CartID         uuid.UUID `gorm:"type:uuid;index"`
	SKU            string
	DisplayName    string
	Quantity       int
	UnitPriceCents int
	Metadata       map[string]any `gorm:"type:jsonb"`
}

type Order struct {
	UUIDBase
	UserID        uuid.UUID `gorm:"type:uuid;index"`
	Status        string    `gorm:"index"`
	SubtotalCents int
	TaxCents      int
	TotalCents    int
	Currency      string
	Notes         string
	Items         []OrderItem
	PrintJobs     []PrintJob `gorm:"foreignKey:OrderID"`
	PlacedAt      *time.Time
	PaidAt        *time.Time
	FulfilledAt   *time.Time
}

type OrderItem struct {
	UUIDBase
	OrderID        uuid.UUID `gorm:"type:uuid;index"`
	Name           string
	Description    string
	Quantity       int
	UnitPriceCents int
	Metadata       map[string]any `gorm:"type:jsonb"`
}

type PrintJob struct {
	UUIDBase
	UserID            uuid.UUID `gorm:"type:uuid;index"`
	OrderID           *uuid.UUID `gorm:"type:uuid;index"`
	OrderItemID       *uuid.UUID `gorm:"type:uuid;index"`
	FileName          string
	StoragePath       string
	Material          string
	Quality           string
	EstimatedGrams    float64
	EstimatedHours    float64
	EstimatedPrice    int
	Analysis          map[string]any `gorm:"type:jsonb"`
	Status            string         `gorm:"index"`
	LastEstimatedAt   time.Time
	Source            string
	OriginalExt       string
	BoundingBoxMM     map[string]any `gorm:"type:jsonb"`
	SurfaceAreaCM2    float64
	VolumeCM3         float64
	ThumbnailPath     *string
	RequiresApproval  bool
	ApprovalStatus    string
	ApprovedBy        *uuid.UUID `gorm:"type:uuid"`
	ApprovedAt        *time.Time
}

// AllModels returns every struct we need to migrate.
func AllModels() []any {
	return []any{
		&User{},
		&OAuthAccount{},
		&PasswordReset{},
		&RefreshToken{},
		&Cart{},
		&CartItem{},
		&Order{},
		&OrderItem{},
		&PrintJob{},
	}
}
