# Permz

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/moxar/permz)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/moxar/permz/master/LICENSE)

`permz` is a Golang permission framework.

## Motivation

The motivation is that permissions can be global (IsAdmin, CanRead, CanValidate)
or scoped (a specific user/group has permissions on specific items).
Thus, we redifine the notion of permission as _a Right on a specific scope_.

The package defines two primitives: `PermissionResolver`, that returns `true` if the permission is granted,
and `ResolverFactory` that fetches a `PermissionResolver` with a scope.

The package provides a threadsafe pool of `ResolverFactory` with internal cache.

See package documentation for details.

## Usage

```go
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
```
