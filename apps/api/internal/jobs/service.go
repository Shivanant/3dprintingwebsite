package jobs

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/3dprint-hub/api/internal/database"
	"github.com/3dprint-hub/api/internal/pricing"
	"github.com/3dprint-hub/api/internal/storage"
)

type Service struct {
	db       *gorm.DB
	logger   *slog.Logger
	storage  storage.Provider
}

type CreateInput struct {
	UserID   uuid.UUID
	FileName string
	Data     []byte
	Estimate *pricing.Estimate
	Material string
	Quality  string
}

func New(db *gorm.DB, logger *slog.Logger, storage storage.Provider) *Service {
	return &Service{db: db, logger: logger, storage: storage}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*database.PrintJob, error) {
	// Save file to storage
	path, err := s.storage.Save(ctx, input.FileName, bytesReader(input.Data))
	if err != nil {
		return nil, err
	}
	job := &database.PrintJob{
		UserID:          input.UserID,
		FileName:        input.FileName,
		StoragePath:     path,
		Material:        input.Material,
		Quality:         input.Quality,
		EstimatedGrams:  input.Estimate.EstimatedGrams,
		EstimatedHours:  input.Estimate.EstimatedHours,
		EstimatedPrice:  int(input.Estimate.EstimatedPrice * 100),
		Analysis: map[string]any{
			"surfaceAreaCm2": input.Estimate.SurfaceAreaCM2,
			"volumeCm3":      input.Estimate.VolumeCM3,
			"triangleCount":  input.Estimate.TriangleCount,
			"infill":         input.Estimate.RecommendedInfill,
		},
		Status:          "draft",
		LastEstimatedAt: time.Now(),
		Source:          "upload",
		OriginalExt:     filepath.Ext(input.FileName),
		BoundingBoxMM: map[string]any{
			"min": input.Estimate.BoundingBoxMM.Min,
			"max": input.Estimate.BoundingBoxMM.Max,
		},
	}
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]database.PrintJob, error) {
	var jobs []database.PrintJob
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func bytesReader(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}
