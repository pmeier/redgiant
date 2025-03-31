package redgiant

func oneOptionalOrDefault[T any](vars []T, defaultFunc func() T) T {
	switch len(vars) {
	case 0:
		return defaultFunc()
	case 1:
		return vars[0]
	default:
		panic("only 0 or 1 optional allowed")
	}
}
