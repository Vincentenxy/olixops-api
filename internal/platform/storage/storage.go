// Package storage 抽象文件 / 对象存储后端。
//
// 默认 local 实现,后续可扩展 S3、MinIO、Git 等。
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"olixops/internal/config"
)

var (
	ErrNotFound = errors.New("storage: object not found")
)

// Object 是存储对象的元信息。
type Object struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// Storage 是存储抽象。
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, contentType string) (Object, error)
	Get(ctx context.Context, key string) (io.ReadCloser, *Object, error)
	Delete(ctx context.Context, key string) error
	Stat(ctx context.Context, key string) (*Object, error)
	List(ctx context.Context, prefix string) ([]Object, error)
}

// New 根据配置返回对应实现。当前仅实现 local。
func New(cfg config.StorageConfig) (Storage, error) {
	switch strings.ToLower(cfg.Driver) {
	case "", "local":
		return NewLocal(cfg.LocalDir)
	default:
		return nil, fmt.Errorf("storage driver %q not implemented yet", cfg.Driver)
	}
}

// Local 是本地文件系统存储实现。
type Local struct {
	root string
}

// NewLocal 构造本地存储,目录不存在时自动创建。
func NewLocal(root string) (*Local, error) {
	if root == "" {
		return nil, errors.New("storage local root is empty")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	return &Local{root: root}, nil
}

func (l *Local) full(key string) (string, error) {
	clean := filepath.Clean("/" + key)
	if clean == "/" || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid key: %s", key)
	}
	return filepath.Join(l.root, clean), nil
}

// Put 写入对象,会自动创建父目录。
func (l *Local) Put(_ context.Context, key string, r io.Reader, contentType string) (Object, error) {
	path, err := l.full(key)
	if err != nil {
		return Object{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Object{}, fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return Object{}, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return Object{}, fmt.Errorf("write file: %w", err)
	}
	return Object{
		Key:          key,
		Size:         n,
		ContentType:  contentType,
		LastModified: time.Now(),
	}, nil
}

// Get 读取对象,调用方负责关闭返回的 ReadCloser。
func (l *Local) Get(_ context.Context, key string) (io.ReadCloser, *Object, error) {
	path, err := l.full(key)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("open file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return f, &Object{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}, nil
}

// Delete 删除对象。
func (l *Local) Delete(_ context.Context, key string) error {
	path, err := l.full(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Stat 返回对象元信息。
func (l *Local) Stat(_ context.Context, key string) (*Object, error) {
	path, err := l.full(key)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &Object{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}, nil
}

// List 列出某前缀下的对象。
func (l *Local) List(_ context.Context, prefix string) ([]Object, error) {
	base, err := l.full(prefix)
	if err != nil {
		return nil, err
	}
	var result []Object
	err = filepath.WalkDir(base, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		rel, relErr := filepath.Rel(l.root, p)
		if relErr != nil {
			return relErr
		}
		result = append(result, Object{
			Key:          filepath.ToSlash(rel),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
