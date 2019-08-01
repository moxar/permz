package permz

// Right aliases an integer for documentation purpose;
type Right = int

// PermissionResolver describes a function that returns wheter the Right is granted or not.
type PermissionResolver = func(Right) bool

// And is a resolver that returns true when all resolvers return true.
// nil fns are ignored.
func And(fns ...PermissionResolver) PermissionResolver {
	return func(r Right) bool {
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			if ok := fn(r); !ok {
				return false
			}
		}
		return true
	}
}

// Or is a resolver that returns true when at least one resolver returns true.
// nil fns are ignored.
func Or(fns ...PermissionResolver) PermissionResolver {
	return func(r Right) bool {
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			if ok := fn(r); ok {
				return true
			}
		}
		return false
	}
}

// True is a resolver that always returns true.
func True(Right) bool {
	return true
}

// False is a resolver that always returns false.
func False(Right) bool {
	return false
}
