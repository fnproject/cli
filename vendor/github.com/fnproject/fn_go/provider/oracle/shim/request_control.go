package shim

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringPtrOr(value string, fallback *string) *string {
	if value != "" {
		return &value
	}
	return fallback
}
