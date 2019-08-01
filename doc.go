// Package permz provides a framework for scoped permission handling.
//
// The motivation is that permissions can be global (IsAdmin, CanRead, CanValidate)
// or scoped (a specific user/group has permissions on specific items).
// Thus, we redifine the notion of permission as: "a Right on a specific scope".
//
// The package defines two primitives: PermissionResolver, that returns true if the permission is granted,
// and ResolverFactory that fetches a permission resolver with a scope.
//
// The package provides a threadsafe pool of ResolverFactory with internal cache.
package permz
