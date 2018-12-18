package platform_test

import (
	"testing"

	"github.com/influxdata/platform"
	platformtesting "github.com/influxdata/platform/testing"
)

func TestDBRPMapping_Validate(t *testing.T) {
	type fields struct {
		Cluster         string
		Database        string
		RetentionPolicy string
		Default         bool
		OrganizationID  platform.ID
		BucketID        platform.ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "mapping requires a cluster",
			fields: fields{
				Cluster: "",
			},
			wantErr: true,
		},
		{
			name: "mapping requires a database",
			fields: fields{
				Cluster:  "abc",
				Database: "",
			},
			wantErr: true,
		},
		{
			name: "mapping requires an rp",
			fields: fields{
				Cluster:         "abc",
				Database:        "telegraf",
				RetentionPolicy: "",
			},
			wantErr: true,
		},
		{
			name: "mapping requires an orgid",
			fields: fields{
				Cluster:         "abc",
				Database:        "telegraf",
				RetentionPolicy: "autogen",
			},
			wantErr: true,
		},
		{
			name: "mapping requires a bucket id",
			fields: fields{
				Cluster:         "abc",
				Database:        "telegraf",
				RetentionPolicy: "autogen",
				OrganizationID:  platformtesting.MustIDBase16("debac1e0deadbeef"),
			},
			wantErr: true,
		},
		{
			name: "cluster name cannot have non-printable characters.",
			fields: fields{
				Cluster: string([]byte{0x0D}),
			},
			wantErr: true,
		},
		{
			name: "db cannot have non-letters/numbers/_/./-",
			fields: fields{
				Cluster:  "12345_.",
				Database: string([]byte{0x0D}),
			},
			wantErr: true,
		},
		{
			name: "rp cannot have non-printable characters",
			fields: fields{
				Cluster:         "12345",
				Database:        "telegraf",
				RetentionPolicy: string([]byte{0x0D}),
			},
			wantErr: true,
		},
		{
			name: "dash accepted as valid database",
			fields: fields{
				Cluster:         "12345_.",
				Database:        "howdy-doody",
				RetentionPolicy: "autogen",
				OrganizationID:  platformtesting.MustIDBase16("debac1e0deadbeef"),
				BucketID:        platformtesting.MustIDBase16("5ca1ab1edeadbea7"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := platform.DBRPMapping{
				Cluster:         tt.fields.Cluster,
				Database:        tt.fields.Database,
				RetentionPolicy: tt.fields.RetentionPolicy,
				Default:         tt.fields.Default,
				OrganizationID:  tt.fields.OrganizationID,
				BucketID:        tt.fields.BucketID,
			}

			if err := m.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DBRPMapping.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
