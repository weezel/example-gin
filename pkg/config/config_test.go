package config

import (
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfig_Parse(t *testing.T) {
	envLock := sync.Mutex{}

	type fields struct {
		setEnvVars      func()
		config          Config
		exptectedConfig Config
	}
	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "Test basic configuration",
			fields: fields{
				config: Config{
					HTTPServer: HTTPServer{
						Port:     "8080",
						Hostname: "localhost",
					},
					Postgres: Postgres{
						Hostname: "localhost",
						Port:     "5432",
						DBName:   "nothing",
						Username: "testing",
						Password: "aaabbbccc",
					},
				},
				exptectedConfig: Config{
					HTTPServer: HTTPServer{
						Port:     "8080",
						Hostname: "localhost",
					},
					Postgres: Postgres{
						Hostname: "localhost",
						Port:     "5432",
						DBName:   "nothing",
						Username: "testing",
						Password: "aaabbbccc",
					},
				},
				setEnvVars: func() {
					t.Setenv("DB_HOST", "localhost")
					t.Setenv("DB_PORT", "1234")
					t.Setenv("DB_NAME", "")
					t.Setenv("DB_USERNAME", "")
					t.Setenv("DB_PASSWORD", "yadda")
				},
			},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since `t.Run()` is ran in parallel we might end up in race condition when changing
			// environmental variables. Lock the changes with mutex. Not really performance critical
			// so mutex is okay here.
			envLock.Lock()
			defer envLock.Unlock()

			tt.fields.setEnvVars()

			c := &Config{
				HTTPServer: tt.fields.config.HTTPServer,
				Postgres:   tt.fields.config.Postgres,
			}
			if err := c.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.fields.config, tt.fields.exptectedConfig); diff != "" {
				t.Errorf("Differing results with:\n%s", diff)
			}
		})
	}
}
