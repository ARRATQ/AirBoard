package middleware

import "testing"

func TestCheckSSOProxySecret(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		provided   string
		want       ssoProxyCheck
	}{
		{"no secret configured ignores SSO headers", "", "anything", ssoProxyNotConfigured},
		{"no secret configured and none provided", "", "", ssoProxyNotConfigured},
		{"missing header is rejected", "s3cret-value", "", ssoProxyRejected},
		{"wrong header is rejected", "s3cret-value", "s3cret-valuX", ssoProxyRejected},
		{"prefix of secret is rejected", "s3cret-value", "s3cret", ssoProxyRejected},
		{"matching header is accepted", "s3cret-value", "s3cret-value", ssoProxyTrusted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkSSOProxySecret(tt.configured, tt.provided); got != tt.want {
				t.Errorf("checkSSOProxySecret(%q, %q) = %v, want %v", tt.configured, tt.provided, got, tt.want)
			}
		})
	}
}
