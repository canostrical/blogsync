package main

import "testing"

func TestValidatePubKeysWithValidHex(t *testing.T) {
	err := validatePubKeys([]string{"b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe", "b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe"})
	if err != nil {
		t.Errorf("unexpected pubKey validation error: %v", err)
	}
}

func TestValidatePubKeysWithNpub(t *testing.T) {
	err := validatePubKeys([]string{"b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe", "npub1hz404lnj7lxsdt5vxdlex9rlvhlz6dxqqed4y6tpywvzgwx0qmlqfpl6sm"})
	if err == nil {
		t.Error("expected npub not supported error")
	}
}
