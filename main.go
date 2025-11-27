package main

import (
	"os"

	_ "github.com/bayhaqi/kv/pkg/cmd/edit"
	"github.com/bayhaqi/kv/pkg/cmd/root"
	_ "github.com/bayhaqi/kv/pkg/cmd/show"
)

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
