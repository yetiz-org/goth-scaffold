/*
Tests the public Cassandra JSON codecs in app/models/helper.go.
Cases cover direct JSON, legacy quoted JSON, empty payload reset, and malformed input.
*/

package models_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gocql/gocql"
	"github.com/yetiz-org/goth-scaffold/app/models"
)

func TestCassandraJSONCodecs_AcceptDirectAndWrappedJSON(t *testing.T) {
	typeInfo := gocql.NewNativeType(4, gocql.TypeText, "")

	t.Run("metadata direct JSON", func(t *testing.T) {
		var metadata models.Metadata
		if err := metadata.UnmarshalCQL(typeInfo, []byte(`{"enabled":true,"attempts":2}`)); err != nil {
			t.Fatalf("UnmarshalCQL() error = %v", err)
		}

		if enabled, ok := metadata["enabled"].(bool); !ok || !enabled {
			t.Fatalf("metadata enabled = %#v, want true", metadata["enabled"])
		}

		if attempts, ok := metadata["attempts"].(float64); !ok || attempts != 2 {
			t.Fatalf("metadata attempts = %#v, want 2", metadata["attempts"])
		}
	})

	t.Run("scope direct JSON", func(t *testing.T) {
		var scope models.Scope
		if err := scope.UnmarshalCQL(typeInfo, []byte(`["settings:read","settings:write"]`)); err != nil {
			t.Fatalf("UnmarshalCQL() error = %v", err)
		}

		want := models.Scope{"settings:read", "settings:write"}
		if !reflect.DeepEqual(scope, want) {
			t.Fatalf("scope = %#v, want %#v", scope, want)
		}
	})

	t.Run("privileges quoted JSON", func(t *testing.T) {
		direct := `["jobs:read","jobs:write"]`
		wrapped, err := json.Marshal(direct)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var privileges models.Privileges
		if err := privileges.UnmarshalCQL(typeInfo, wrapped); err != nil {
			t.Fatalf("UnmarshalCQL() error = %v", err)
		}

		want := models.Privileges{"jobs:read", "jobs:write"}
		if !reflect.DeepEqual(privileges, want) {
			t.Fatalf("privileges = %#v, want %#v", privileges, want)
		}
	})
}

func TestCassandraJSONCodecs_ResetOnEmptyAndRejectMalformedJSON(t *testing.T) {
	typeInfo := gocql.NewNativeType(4, gocql.TypeText, "")
	metadata := models.Metadata{"stale": true}
	if err := metadata.UnmarshalCQL(typeInfo, nil); err != nil {
		t.Fatalf("empty UnmarshalCQL() error = %v", err)
	}

	if metadata != nil {
		t.Fatalf("metadata after empty payload = %#v, want nil", metadata)
	}

	metadata = models.Metadata{"preserved": true}
	if err := metadata.UnmarshalCQL(typeInfo, []byte(`{"broken":`)); err == nil {
		t.Fatal("malformed JSON returned nil error")
	}

	want := models.Metadata{"preserved": true}
	if !reflect.DeepEqual(metadata, want) {
		t.Fatalf("metadata after malformed JSON = %#v, want %#v", metadata, want)
	}
}
