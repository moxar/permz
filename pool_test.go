package permz

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func ExampleResolverFactory() {

	var GetUserProjectPermissions func(context.Context, int, int) ([]int, error)
	var NewResolverFromPermissions func([]int) PermissionResolver

	var factory ResolverFactory
	factory = func(ctx context.Context, userID, projectID interface{}) (PermissionResolver, error) {
		// fetch the set of permissions of this user on this project
		perms, err := GetUserProjectPermissions(ctx, userID.(int), projectID.(int))
		if err != nil {
			return nil, err
		}

		// the resolver is now scoped for the userID and projectID
		resolver := NewResolverFromPermissions(perms)
		return resolver, nil
	}

	resolver, err := factory(context.TODO(), 4, 1)
	if err != nil {
		// ...
	}
	_ = resolver // this is the permission resolver for the user 4 on the project 1
}

func TestPool_GetResolver(t *testing.T) {

	// keep track of number of time slow is called.
	var called int32

	// slow resolver, to test the lock behaviour.
	slow := func(context.Context, interface{}, interface{}) (PermissionResolver, error) {
		atomic.AddInt32(&called, 1)
		time.Sleep(time.Millisecond * 10)
		return True, nil
	}

	// isEven's resolver returns true if k1 is even.
	isEven := func(ctx context.Context, key1, key2 interface{}) (PermissionResolver, error) {
		return func(int) bool {
			return key1.(int)%2 == 0
		}, nil
	}

	var dictionnary = map[string]ResolverFactory{
		"true":  func(context.Context, interface{}, interface{}) (PermissionResolver, error) { return True, nil },
		"false": func(context.Context, interface{}, interface{}) (PermissionResolver, error) { return False, nil },
		"error": func(context.Context, interface{}, interface{}) (PermissionResolver, error) {
			return nil, errors.New("boom")
		},
		"slow": slow,
		"even": isEven,
	}

	ctx := context.TODO()
	p := NewPool(dictionnary)

	t.Run("when resolver is in dictionnary", func(t *testing.T) {
		r, err := p.GetResolver(ctx, "true", 0, 0)
		if err != nil {
			t.Error("resolver is in dictionnary, should not fail")
			return
		}
		if !r(0) {
			t.Error("should resolve true")
			return
		}
	})

	t.Run("when resolver is not in dictionnary", func(t *testing.T) {
		_, err := p.GetResolver(ctx, "undefined", 0, 0)
		if err == nil {
			t.Error("resolver is not dictionnary, should fail")
			return
		}
	})

	t.Run("when factory fails", func(t *testing.T) {
		_, err := p.GetResolver(ctx, "error", 0, 0)
		if err == nil {
			t.Error("resolver is not dictionnary, should fail")
			return
		}
	})

	t.Run("when scope is complex", func(t *testing.T) {
		t.Run("with odd value", func(t *testing.T) {
			r, err := p.GetResolver(ctx, "even", 1, 0)
			if err != nil {
				t.Error("resolver is in dictionnary, should not fail")
				return
			}
			if r(0) {
				t.Error("should resolve false with odd key")
				return
			}

		})

		t.Run("with even value", func(t *testing.T) {
			r, err := p.GetResolver(ctx, "even", 2, 0)
			if err != nil {
				t.Error("resolver is in dictionnary, should not fail")
				return
			}
			if !r(0) {
				t.Error("should resolve true with even key")
				return
			}

		})
	})

	t.Run("when factory is slow", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i <= 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				r, err := p.GetResolver(ctx, "slow", i%2, 0)
				if err != nil {
					t.Error("slowFactory call should not return error")
					return
				}
				if !r(0) {
					t.Error("slow should not fail")
					return
				}
			}(i)
		}
		wg.Wait()

		if atomic.LoadInt32(&called) != 2 {
			t.Errorf("slow should be called once for scope 1 and 2, then cached: %d calls", called)
			return
		}
	})
}

func TestContext(t *testing.T) {
	p := NewPool(nil)
	ctx := context.Background()
	ctx = CtxSetPool(ctx, p)
	p = CtxGetPool(ctx)
	if p == nil {
		t.Fail()
	}
}
