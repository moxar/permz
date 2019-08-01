package permz

import (
	"context"
	"sync"
)

// ResolverFactory is a function that takes the parameters of a scope and returns a permission resolver.
// usualy, the resolver is a set of right stored in database. The factory is the function that retreives it.
type ResolverFactory = func(ctx context.Context, scope Scope) (PermissionResolver, error)

// Scope is an interface of any type that is aliases for documentation purpose.
// The scope must be comparable according to the language specification: https://golang.org/ref/spec#Comparison_operators
type Scope = interface{}

// Pool caches PermissionResolvers that are generated with the same factory. It is threadsafe.
// The cache key is the scope used to create the factory.
type Pool struct {

	// the factory to use when a new scope is given
	factory ResolverFactory

	// one resolver element maps "{4,7}" to factory(4,7)
	resolvers sync.Map // map[scope]PermissionResolver

	// onces ensure each resolver is fetched only once per scope.
	onces sync.Map // map[scope]*sync.Once
}

// NewPool creates a pool with a factory. See Pool documentation.
func NewPool(fn ResolverFactory) *Pool {
	return &Pool{
		factory: fn,
	}
}

// GetResolver retreives a resolver from the pool or fetches it. The scope must be comparable
func (p *Pool) GetResolver(ctx context.Context, scope Scope) (PermissionResolver, error) {

	if r, ok := p.resolvers.Load(scope); ok {
		return r.(PermissionResolver), nil
	}

	o, _ := p.onces.LoadOrStore(scope, &sync.Once{})
	once := o.(*sync.Once)

	var err error
	once.Do(func() {
		var r PermissionResolver
		r, err = p.factory(ctx, scope)
		if err != nil {
			p.onces.Store(scope, &sync.Once{})
			return
		}
		p.resolvers.Store(scope, r)
	})
	if err != nil {
		return nil, err
	}

	return p.GetResolver(ctx, scope)
}
