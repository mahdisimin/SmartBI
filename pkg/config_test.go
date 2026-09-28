package pkg

import (
	"reflect"
	"testing"
)

func TestLoadDBConfig_RequiresCredentials(t *testing.T) {
	t.Setenv(EnvDBUser, "")
	t.Setenv(EnvDBPassword, "")
	if _, err := LoadDBConfig(); err == nil {
		t.Fatal("expected an error when credentials are not set")
	}

	t.Setenv(EnvDBUser, "app")
	t.Setenv(EnvDBPassword, "s3cret")
	t.Setenv(EnvDBHost, "")
	t.Setenv(EnvDBName, "")
	cfg, err := LoadDBConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg != (DBConfig{User: "app", Password: "s3cret", Host: defaultDBHost, Database: defaultDBName}) {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestCORSAllowedOrigins(t *testing.T) {
	t.Setenv(EnvCORSOrigins, "")
	if got := CORSAllowedOrigins(); !reflect.DeepEqual(got, []string{defaultCORSOrigin}) {
		t.Fatalf("default: got %v", got)
	}
	t.Setenv(EnvCORSOrigins, " https://bi.example.com , http://localhost:5173,")
	if got := CORSAllowedOrigins(); !reflect.DeepEqual(got, []string{"https://bi.example.com", "http://localhost:5173"}) {
		t.Fatalf("list: got %v", got)
	}
}

func TestLoadSessionConfig(t *testing.T) {
	t.Setenv(EnvSessionTTL, "")
	t.Setenv(EnvCookieSecure, "")
	cfg, err := LoadSessionConfig()
	if err != nil || cfg != (SessionConfig{TTL: defaultSessionTTL, CookieSecure: false}) {
		t.Fatalf("defaults: %+v, %v", cfg, err)
	}

	t.Setenv(EnvSessionTTL, "30m")
	t.Setenv(EnvCookieSecure, "true")
	cfg, err = LoadSessionConfig()
	if err != nil || cfg.TTL.Minutes() != 30 || !cfg.CookieSecure {
		t.Fatalf("explicit: %+v, %v", cfg, err)
	}

	for _, bad := range []struct{ key, value string }{{EnvSessionTTL, "forever"}, {EnvSessionTTL, "-1h"}, {EnvCookieSecure, "yes please"}} {
		t.Setenv(EnvSessionTTL, "1h")
		t.Setenv(EnvCookieSecure, "false")
		t.Setenv(bad.key, bad.value)
		if _, err := LoadSessionConfig(); err == nil {
			t.Errorf("%s=%q: expected an error", bad.key, bad.value)
		}
	}
}
