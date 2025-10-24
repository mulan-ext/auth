package token

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultFilePrefix 默认文件前缀
	DefaultFilePrefix = "ginx_auth_token_"
	// DefaultFileMode 默认文件权限
	DefaultFileMode = 0644
	// DefaultDirMode 默认目录权限
	DefaultDirMode = 0755
)

var _ Store = (*FsStore)(nil)

type fsData struct {
	Data   Data      `json:"data"`
	Expire time.Time `json:"expire"`
}

type FsStore struct {
	dir    string
	prefix string
	maxAge int
}

func NewFsStore(dir string, maxAge ...int) (*FsStore, error) {
	s := &FsStore{
		prefix: DefaultFilePrefix,
		maxAge: DefaultMaxAge,
		dir:    dir,
	}
	if len(maxAge) > 0 {
		s.maxAge = maxAge[0]
	}
	if err := os.MkdirAll(dir, DefaultDirMode); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FsStore) Clear(ctx context.Context, token string) error {
	err := os.Remove(s.getFilePath(token))
	// 忽略文件不存在的错误
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *FsStore) Get(ctx context.Context, token string) (Data, error) {
	buf, err := os.ReadFile(s.getFilePath(token))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var data fsData
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, err
	}

	// 检查是否过期
	if !data.Expire.IsZero() && data.Expire.Before(time.Now()) {
		s.Clear(ctx, token)
		return nil, ErrTokenExpired
	}

	return data.Data, nil
}

func (s *FsStore) Save(ctx context.Context, v Data, lifetime ...time.Duration) error {
	token := v.Token()
	if token == "" {
		token = v.New()
	}

	data := &fsData{
		Data:   v,
		Expire: s.calculateExpireTime(lifetime...),
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.getFilePath(token), buf, DefaultFileMode)
}

// getFilePath 获取文件完整路径
func (s *FsStore) getFilePath(token string) string {
	return filepath.Join(s.dir, s.prefix+token)
}

// calculateExpireTime 计算过期时间
func (s *FsStore) calculateExpireTime(lifetime ...time.Duration) time.Time {
	if len(lifetime) > 0 && lifetime[0] > 0 {
		return time.Now().Add(lifetime[0])
	}
	if s.maxAge > 0 {
		return time.Now().Add(time.Duration(s.maxAge) * time.Second)
	}
	return time.Time{} // 永不过期
}
