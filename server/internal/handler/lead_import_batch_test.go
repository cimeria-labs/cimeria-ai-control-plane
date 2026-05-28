package handler

import "testing"

func TestLeadImportBatchProviderSetIncludesApollo(t *testing.T) {
	if !validLeadSourceProviders["apollo"] {
		t.Fatal("lead source provider set must include apollo")
	}
	if !validBatchProviders["apollo"] {
		t.Fatal("lead import batch provider set must include apollo")
	}
}

func TestLeadImportBatchProviderSetCoversLeadSourceProviders(t *testing.T) {
	required := []string{
		"manual",
		"csv",
		"api",
		"form",
		"apollo",
		"hunter",
		"linkedin",
		"referral",
		"website",
		"hubspot",
		"pipedrive",
	}
	for _, provider := range required {
		if !validLeadSourceProviders[provider] {
			t.Fatalf("lead source provider set missing %q", provider)
		}
		if !validBatchProviders[provider] {
			t.Fatalf("lead import batch provider set missing %q", provider)
		}
	}
}
