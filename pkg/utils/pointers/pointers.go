package pointers

func Bool(b bool) *bool {
	return &b
}

func String(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
