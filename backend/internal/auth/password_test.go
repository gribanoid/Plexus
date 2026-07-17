package auth

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		pw    string
		ok    bool
	}{
		{"short1", false},
		{"password", false},
		{"1234567890", false},
		{"password123", true},
		{"LongerPass1", true},
	}
	for _, tc := range cases {
		err := ValidatePassword(tc.pw)
		if tc.ok && err != nil {
			t.Fatalf("%q: unexpected error %v", tc.pw, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: expected error", tc.pw)
		}
	}
}
