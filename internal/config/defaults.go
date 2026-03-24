package config

// DefaultEnvironments returns the built-in Zuora environment definitions.
func DefaultEnvironments() map[string]*Environment {
	return map[string]*Environment{
		"sandbox": {
			BaseURL: "https://rest.apisandbox.zuora.com",
		},
		"us-production": {
			BaseURL: "https://rest.na.zuora.com",
		},
		"us-production-cloud2": {
			BaseURL: "https://rest.zuora.com",
		},
		"eu-production": {
			BaseURL: "https://rest.eu.zuora.com",
		},
		"apac-production": {
			BaseURL: "https://rest.ap.zuora.com",
		},
	}
}

const (
	defaultActiveEnvironment = "sandbox"
	defaultZuoraVersion      = "2025-08-12"
	defaultOutput            = "table"
)
