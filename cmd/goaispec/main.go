// Command goaispec generates OpenAPI documents from the scaffold route tree.
package main

import (
	"fmt"
	"os"

	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/goai"
	handler "github.com/yetiz-org/goth-scaffold/app/handlers"
)

func main() {
	if err := os.Setenv("APP_DEBUG", "true"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "goaispec: set APP_DEBUG: %v\n", err)
		os.Exit(1)
	}

	goai.RunCLIFromConfig(
		"docs/openapi/goai.yaml",
		func() ghttp.RouteEntriesProvider { return handler.NewAppRoute() },
		goai.WithArgs(os.Args[1:]),
	)
}
