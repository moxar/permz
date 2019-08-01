package permz

import (
	"context"
	"fmt"
	"sync"
)

// ResolverFactory is a function that takes the parameters of a scope and returns a permission resolver.
// usualy, the resolver is a set of right stored in database. The factory is the function that retreives it.
type ResolverFactory = func(ctx context.Context, bySomeone, onSomething interface{}) (PermissionResolver, error)

// Pool is a metafactory. With a dictionnary - a list of ResolverFactory organized in scopes - it provides.
// the PermissionResolver associated to the given scope and parameters.
// One scope can be used with different parameters. For each combination, a different resolver is returned.
// If the scope is not defined in the dictionnary, Pool.GetResolver returns an error.
// The pool is threadsafe.
type Pool struct {

	// given
	// 	func FetchUserHasRightOnProject(user.id, project.id) UserHasRightOnProject
	// 	func UserHasRightOnProject(Right) (bool, error)

	// the dictionnar maps "user-project" to FetchUserHasRightOnProject
	dictionnary map[string]ResolverFactory

	// one resolver element maps "user-project.4.7" to UserHasRightOnProject
	// with user.id=4 and project.id=7
	resolvers sync.Map // map[string]PermissionResolver

	// onces ensure each resolver is fetched only once per namespace.
	onces sync.Map // map[string]*sync.Once
}

// NewPool creates a pool with a dictionnary. See Pool documentation.
func NewPool(d map[string]ResolverFactory) *Pool {
	return &Pool{
		dictionnary: d,
	}
}

// GetResolver retreives a resolver from the pool or fetches it.
func (p *Pool) GetResolver(ctx context.Context, scope string, bySomeone, onSomething interface{}) (PermissionResolver, error) {

	key := namespace(scope, bySomeone, onSomething)

	if r, ok := p.resolvers.Load(key); ok {
		return r.(PermissionResolver), nil
	}

	o, _ := p.onces.LoadOrStore(key, &sync.Once{})
	once := o.(*sync.Once)

	fn := p.dictionnary[scope]
	if fn == nil {
		return nil, fmt.Errorf("no ResolverFactory in dictionnary for scope %s", scope)
	}

	var err error
	once.Do(func() {
		var r PermissionResolver
		r, err = fn(ctx, bySomeone, onSomething)
		if err != nil {
			p.onces.Store(key, &sync.Once{})
			return
		}
		p.resolvers.Store(key, r)
	})
	if err != nil {
		return nil, err
	}

	return p.GetResolver(ctx, scope, bySomeone, onSomething)
}

func namespace(scope string, bySomeone, onSomething interface{}) string {
	return fmt.Sprintf("%s.%v.%v", scope, bySomeone, onSomething)
}

type ctxKey struct{}

var ctxKeyPool ctxKey

// CtxSetPool is a helper that sets the pool in the context.
func CtxSetPool(ctx context.Context, pool *Pool) context.Context {
	return context.WithValue(ctx, ctxKeyPool, pool)
}

// CtxGetPool is a helper that gets the pool from the context.
func CtxGetPool(ctx context.Context) *Pool {
	p, _ := ctx.Value(ctxKeyPool).(*Pool)
	return p
}
