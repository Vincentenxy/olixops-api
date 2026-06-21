package user

import (
	"context"
	"sync"
	"time"

	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// fakeRepo 是 Repository 的内存实现, 用于单元测试和 HTTP handler 测试。
type fakeRepo struct {
	mu    sync.Mutex
	users map[string]*User // key = id
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: make(map[string]*User)}
}

func (r *fakeRepo) Create(_ context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; ok {
		return errs.AlreadyExists("user")
	}
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) Update(_ context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; !ok {
		return errs.NotFound("user")
	}
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, errs.NotFound("user")
	}
	return u, nil
}

func (r *fakeRepo) FindByUsername(_ context.Context, username string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errs.NotFound("user")
}

func (r *fakeRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errs.NotFound("user")
}

func (r *fakeRepo) FindByExternal(_ context.Context, source, externalID string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Source == source && u.ExternalID == externalID {
			return u, nil
		}
	}
	return nil, errs.NotFound("user")
}

func (r *fakeRepo) UpdateLastLogin(_ context.Context, id, ip string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return errs.NotFound("user")
	}
	u.LastLoginAt = &at
	u.LastLoginIP = ip
	return nil
}

func (r *fakeRepo) List(_ context.Context, q pagination.Query, filter ListFilter) ([]*User, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []*User
	for _, u := range r.users {
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		if filter.Source != "" && u.Source != filter.Source {
			continue
		}
		if q.Keyword != "" {
			if !contains(u.Username, q.Keyword) && !contains(u.Email, q.Keyword) && !contains(u.DisplayName, q.Keyword) {
				continue
			}
		}
		items = append(items, u)
	}
	return items, int64(len(items)), nil
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
