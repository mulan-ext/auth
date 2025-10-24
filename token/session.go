package token

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"time"
)

const (
	// DefaultMaxAge 默认过期时间: 7天
	DefaultMaxAge = 7 * 24 * 60 * 60
	// DefaultKeyPrefix 默认key前缀
	DefaultKeyPrefix = "ginx:auth:token:"
)

type Session struct {
	store Store
	ctx   context.Context
	data  Data
	IsNil bool

	keyPrefix string
	maxAge    int
	token     string
	mu        sync.RWMutex // 保护并发访问
	loaded    bool         // 标记是否已从store加载
}

func (s *Session) Token() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

func (s *Session) ID() uint64               { return s.Data().ID() }
func (s *Session) Account() string          { return s.Data().Account() }
func (s *Session) State() uint16            { return s.Data().State() }
func (s *Session) Roles() []string          { return s.Data().Roles() }
func (s *Session) HasRole(role string) bool { return slices.Contains(s.Roles(), role) }

// Data 获取session数据，首次调用时从store加载
func (s *Session) Data() Data {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经加载过，直接返回
	if s.loaded {
		return s.data
	}

	// 尝试从store加载数据
	if s.token != "" {
		if data, err := s.store.Get(s.ctx, s.token); err == nil {
			s.data = data
			s.loaded = true
			return s.data
		}
	}

	// 如果加载失败，标记为已加载（使用当前空数据）
	s.loaded = true
	return s.data
}

// Clear 清空session并生成新token
func (s *Session) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除旧token
	if s.token != "" {
		if err := s.store.Clear(s.ctx, s.token); err != nil {
			return err
		}
	}

	// 生成新token
	s.token = s.data.New()
	s.loaded = false

	// 保存新session
	return s.store.Save(s.ctx, s.data)
}

// Delete 删除指定key
func (s *Session) Delete(key string) error {
	data := s.Data()
	data.Delete(key)
	return s.store.Save(s.ctx, data)
}

func (s *Session) Get(key string) any            { return s.Data().Get(key) }
func (s *Session) Set(key string, val any)       { s.Data().Set(key, val) }
func (s *Session) SetID(val uint64)              { s.Data().SetID(val) }
func (s *Session) SetAccount(val string)         { s.Data().SetAccount(val) }
func (s *Session) SetState(val uint16)           { s.Data().SetState(val) }
func (s *Session) SetRoles(roles []string)       { s.Data().SetRoles(roles) }
func (s *Session) SetValues(key string, val any) { s.Data().SetValues(key, val) }

// Save 保存session数据
func (s *Session) Save(lifetime ...time.Duration) error {
	return s.store.Save(s.ctx, s.Data(), lifetime...)
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Data())
}

func (s *Session) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, s.data); err != nil {
		return err
	}
	s.token = s.data.Token()
	s.loaded = true
	return nil
}

func NewSession(ctx context.Context, store Store, data Data, maxAge ...int) *Session {
	s := &Session{
		ctx:       ctx,
		store:     store,
		keyPrefix: DefaultKeyPrefix,
		maxAge:    DefaultMaxAge,
		data:      data,
		token:     data.Token(),
		loaded:    false,
	}
	if len(maxAge) > 0 {
		s.maxAge = maxAge[0]
	}
	return s
}
