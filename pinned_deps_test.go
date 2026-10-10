// Keeps the pinned dependencies in go.mod until feature code imports them.
package xenon_test

import (
	_ "filippo.io/age"
	_ "github.com/Bios-Marcel/wastebasket/v2"
	_ "github.com/andybalholm/brotli"
	_ "github.com/cucumber/godog"
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/oklog/ulid/v2"
	_ "github.com/zalando/go-keyring"
	_ "go.yaml.in/yaml/v3"
)
