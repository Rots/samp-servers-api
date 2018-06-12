package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Southclaws/samp-servers-api/types"
)

func TestApp_GetServers(t *testing.T) {
	type args struct {
		page   string
		sort   string
		by     string
		filter []string
	}
	tests := []struct {
		name        string
		args        args
		wantServers []types.ServerCore
		wantErr     bool
	}{
		{
			"v no sort",
			args{"1", "", "", []string{}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"s4.example.com", "test server 4", 50, 50, "rivershell", "Polish", true, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
				{"s2.example.com", "test server 2", 0, 100, "Grand Larceny", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v desc",
			args{"1", "asc", "", []string{}},
			[]types.ServerCore{
				{"s2.example.com", "test server 2", 0, 100, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
				{"s4.example.com", "test server 4", 50, 50, "rivershell", "Polish", true, "0.3.7-R2"},
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v pass",
			args{"1", "", "", []string{"password"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
				{"s2.example.com", "test server 2", 0, 100, "Grand Larceny", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v empty",
			args{"1", "", "", []string{"empty"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"s4.example.com", "test server 4", 50, 50, "rivershell", "Polish", true, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v full",
			args{"1", "", "", []string{"full"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
				{"s2.example.com", "test server 2", 0, 100, "Grand Larceny", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v pass empty",
			args{"1", "", "", []string{"password", "empty"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v pass full",
			args{"1", "", "", []string{"password", "full"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
				{"s2.example.com", "test server 2", 0, 100, "Grand Larceny", "English", false, "0.3.7-R2"},
			},
			false,
		},
		{
			"v empty full",
			args{"1", "", "", []string{"empty", "full"}},
			[]types.ServerCore{
				{"s3.example.com", "test server 3", 948, 1000, "Grand Larceny", "English", false, "0.3.7-R2"},
				{"ss.southcla.ws", "Scavenge and Survive Official", 4, 32, "Scavenge & Survive by Southclaws", "English", false, "0.3.7-R2"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServers, err := app.GetServers(tt.args.page, tt.args.sort, tt.args.by, tt.args.filter)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantServers, gotServers)
		})
	}
}