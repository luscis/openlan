package config

import "testing"

func TestToForwardsFindBackendMatch(t *testing.T) {
	backends := ToForwards{
		{
			Server: "10.0.0.1",
			Match:  []string{"example.com"},
		},
	}

	tests := []struct {
		host string
		want bool
	}{
		{host: "example.com", want: true},
		{host: "www.example.com", want: true},
		{host: "www.example.com:443", want: true},
		{host: "badexample.com", want: false},
		{host: "www.example.com:https", want: false},
	}

	for _, tt := range tests {
		got := backends.FindBackend(tt.host) != nil
		if got != tt.want {
			t.Fatalf("FindBackend(%q) matched=%v, want %v", tt.host, got, tt.want)
		}
	}
}
