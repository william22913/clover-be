package config

import "testing"

// setBase gives Load everything except the values each case varies, so a case
// fails on its own subject rather than on an unrelated missing variable.
func setBase(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://clover_app:pw@db:5432/clover?sslmode=verify-full")
	t.Setenv("MIGRATE_DATABASE_URL", "postgres://clover_migrator:pw@db:5432/clover?sslmode=verify-full")
	t.Setenv("LISTEN_ADDR", ":8080")
	t.Setenv("METRICS_ADDR", ":9101")
	t.Setenv("PUBLIC_ORIGIN", "")
}

func TestLoad_PublicOrigin(t *testing.T) {
	cases := []struct {
		name    string
		appEnv  string
		origin  string
		wantErr bool
	}{
		{"production requires it", "production", "", true},
		{"production accepts https origin", "production", "https://clover.example", false},
		{"production accepts http origin", "production", "http://localhost:8080", false},
		{"origin without scheme", "production", "clover.example", true},
		{"origin with a path", "production", "https://clover.example/app", true},
		{"development may omit it", "development", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setBase(t)
			t.Setenv("APP_ENV", tc.appEnv)
			t.Setenv("PUBLIC_ORIGIN", tc.origin)

			_, err := Load()
			if tc.wantErr && err == nil {
				t.Fatalf("Load() accepted PUBLIC_ORIGIN %q for APP_ENV=%s", tc.origin, tc.appEnv)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Load() rejected PUBLIC_ORIGIN %q for APP_ENV=%s: %v", tc.origin, tc.appEnv, err)
			}
		})
	}
}

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	setBase(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an empty DATABASE_URL")
	}
}

func TestLoad_MetricsListenerMustDiffer(t *testing.T) {
	setBase(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("METRICS_ADDR", ":8080")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted the metrics listener on the public address")
	}
}
