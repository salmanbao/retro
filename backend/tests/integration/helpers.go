package integration

// Helper functions for integration tests

func float64Ptr(v float64) *float64 {
	return &v
}

func strPtr(s string) *string {
	return &s
}

func containsField(fields []string, name string) bool {
	for _, f := range fields {
		if f == "all" || f == name {
			return true
		}
	}
	return false
}
