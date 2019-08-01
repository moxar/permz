package permz

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func ExampleResolverFactory() {

	var GetUserProjectPermissions func(context.Context, int, int) ([]int, error)
	var NewResolverFromPermissions func([]int) PermissionResolver

	type Pair struct {
		UserID, ProjectID int
	}

	var factory ResolverFactory
	factory = func(ctx context.Context, scope Scope) (PermissionResolver, error) {
		// fetch the set of permissions of this user on this project
		p, ok := scope.(Pair)
		if !ok {
			return nil, errors.New("invalid scope type")
		}
		perms, err := GetUserProjectPermissions(ctx, p.UserID, p.ProjectID)
		if err != nil {
			return nil, err
		}

		// the resolver is now scoped for the userID and projectID
		resolver := NewResolverFromPermissions(perms)
		return resolver, nil
	}

	resolver, err := factory(context.TODO(), Pair{4, 1})
	if err != nil {
		// ...
	}
	_ = resolver // this is the permission resolver for the user 4 on the project 1
}

func TestPool_GetResolver(t *testing.T) {

	// keep track of number of time slow is called.
	var called int32

	// slow resolver, to test the lock behaviour.
	slow := func(ctx context.Context, scope Scope) (PermissionResolver, error) {
		atomic.AddInt32(&called, 1)
		time.Sleep(time.Millisecond * 10)
		return True, nil
	}

	// modulo resolver returns true r % scope == 0
	modulo := func(ctx context.Context, scope Scope) (PermissionResolver, error) {
		return func(r Right) bool {
			return r.(int)%scope.(int) == 0
		}, nil
	}

	fail := func(ctx context.Context, scope Scope) (PermissionResolver, error) {
		return nil, errors.New("boom")
	}

	type Given struct {
		Factory ResolverFactory
		Scope   Scope
		Right   Right
	}

	type Want struct {
		Bool bool
		Err  error
	}

	type Case struct {
		Sentence string
		Given
		Want
	}

	var cases = []Case{
		{
			Sentence: "the factory fails",
			Given: Given{
				Factory: fail,
				Scope:   nil,
				Right:   nil,
			},
			Want: Want{
				Bool: false,
				Err:  errors.New("boom"),
			},
		},
		{
			Sentence: "the factory is modulo",
			Given: Given{
				Factory: modulo,
				Scope:   3,
				Right:   9,
			},
			Want: Want{
				Bool: true,
				Err:  nil,
			},
		},
		{
			Sentence: "the factory is slow",
			Given: Given{
				Factory: slow,
				Scope:   "foo",
				Right:   nil,
			},
			Want: Want{
				Bool: true,
				Err:  nil,
			},
		},
		{
			Sentence: "the factory is slow",
			Given: Given{
				Factory: slow,
				Scope:   "bar",
				Right:   nil,
			},
			Want: Want{
				Bool: true,
				Err:  nil,
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("As a pool, when %s and the scope is %v, given %v", c.Sentence, c.Scope, c.Right), func(t *testing.T) {
			p := NewPool(c.Given.Factory)
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r, err := p.GetResolver(context.TODO(), c.Given.Scope)
					if (c.Want.Err != nil) != (err != nil) {
						t.Error("errors should match")
						t.Error("want:", c.Want.Err)
						t.Error("got: ", err)
						return
					}

					if err != nil {
						return
					}
					if r(c.Given.Right) != c.Want.Bool {
						t.Errorf("permission should be %v", c.Want.Bool)
						return
					}
				}()
			}
			wg.Wait()
		})
	}

	if total := atomic.LoadInt32(&called); total != 2 {
		t.Errorf("slow factory should be 2 time, total of %d calls", total)
	}
}
