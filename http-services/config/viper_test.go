package config

import "testing"

func TestGetViper(t *testing.T) {
	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	viper := GetViper()
	if viper == nil {
		t.Error("GetViper() returned nil")
	}
	if viper != v {
		t.Error("GetViper() did not return the expected viper instance")
	}
}
