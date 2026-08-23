/*
Tests the repository-owned GoAI config and route-to-OpenAPI build contract.
The test validates health metadata and schema generation without starting the application.
*/

package goai_test

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/goai"
	handler "github.com/yetiz-org/goth-scaffold/app/handlers"
)

func repositoryRoot(t *testing.T) (root string) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../.."))
}

func TestGoAIConfig_BuildsDocumentedPublicHealthOperation(t *testing.T) {
	root := repositoryRoot(t)
	configDir := filepath.Join(root, "docs", "openapi")
	config, err := goai.LoadConfig(configDir)
	if err != nil {
		t.Fatalf("goai.LoadConfig() error = %v", err)
	}

	if config.Title != "Goth Scaffold API" {
		t.Fatalf("config title = %q, want Goth Scaffold API", config.Title)
	}

	if config.Output["all"] != "openapi.generated.yaml" || config.Output["public"] != "openapi.public.yaml" {
		t.Fatalf("config output = %#v, want all/public outputs", config.Output)
	}

	previous := ghttp.SetSkipHandlerRegister(true)
	t.Cleanup(func() { ghttp.SetSkipHandlerRegister(previous) })

	candidates := goai.Walk(handler.NewAppRoute())
	document := goai.Build(candidates, config.BuildProfile("public"), config.ToBuildOptions())
	path := document.Paths["/api/v1/health"]
	if path == nil || path.Get == nil {
		t.Fatalf("generated public paths = %#v, want GET /api/v1/health", document.Paths)
	}

	operation := path.Get
	if operation.OperationID != "health.get" {
		t.Fatalf("health operationId = %q, want health.get", operation.OperationID)
	}

	if len(operation.Tags) != 1 || operation.Tags[0] != "Public" {
		t.Fatalf("health tags = %#v, want [Public]", operation.Tags)
	}

	if operation.Security == nil || len(*operation.Security) != 0 {
		t.Fatalf("health security = %#v, want explicit no-security array", operation.Security)
	}

	response := operation.Responses["200"]
	if response == nil || response.Content["application/json"] == nil {
		t.Fatalf("health 200 response = %#v, want application/json content", response)
	}

	media := response.Content["application/json"]
	schema := media.Schema
	if schema == nil || schema.Ref != "#/components/schemas/HealthResponse" {
		t.Fatalf("health response schema = %#v, want HealthResponse ref", schema)
	}

	example := media.Examples["healthy"]
	if example == nil || !reflect.DeepEqual(example.Value, map[string]any{"status": "ok"}) {
		t.Fatalf("health response example = %#v, want status ok", example)
	}
}
