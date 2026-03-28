package miscellaneous

func ToPointer[T any](v T) *T { return &v }

func GetReferenceOrDefault[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}

	return defaultValue
}
