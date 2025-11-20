package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Provider interface {
	Save(ctx context.Context, originalName string, r io.Reader) (string, error)
	Delete(ctx context.Context, path string) error
	Open(path string) (*os.File, error)
	BasePath() string
}

type localProvider struct {
	basePath string
}

func NewLocal(basePath string) (Provider, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, err
	}
	return &localProvider{basePath: basePath}, nil
}

func (l *localProvider) Save(ctx context.Context, originalName string, r io.Reader) (string, error) {
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin"
	}
	name := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixNano(), ext)
	full := filepath.Join(l.basePath, name)

	tmp, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, r); err != nil {
		return "", err
	}
	return name, nil
}

func (l *localProvider) Delete(ctx context.Context, path string) error {
	full := filepath.Join(l.basePath, path)
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *localProvider) Open(path string) (*os.File, error) {
	full := filepath.Join(l.basePath, path)
	return os.Open(full)
}

func (l *localProvider) BasePath() string {
	return l.basePath
}
