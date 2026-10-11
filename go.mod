module github.com/brienze1/xenon

go 1.26.0

toolchain go1.27.2

require (
	filippo.io/age v1.3.2
	github.com/Bios-Marcel/wastebasket/v2 v2.0.3
	github.com/andybalholm/brotli v1.2.6
	github.com/aws/aws-sdk-go-v2 v1.47.2
	github.com/brianvoe/gofakeit/v7 v7.17.1
	github.com/cucumber/godog v0.16.0
	github.com/dop251/goja v0.0.0-20261009133625-a4bedf5f6000
	github.com/dop251/goja_nodejs v0.0.0-20260918173711-b481721df8a2
	github.com/fsnotify/fsnotify v1.10.1
	github.com/go-playground/validator/v10 v10.30.5
	github.com/icholy/digest v1.2.0
	github.com/oklog/ulid/v2 v2.1.2
	github.com/wailsapp/wails/v3 v3.0.0-beta.28
	github.com/zalando/go-keyring v0.2.8
	go.yaml.in/yaml/v3 v3.0.5
	golang.org/x/sys v0.48.0
	golang.org/x/text v0.42.0
	mvdan.cc/sh/v3 v3.14.1
)

require (
	filippo.io/hpke v0.4.0 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/aws/smithy-go v1.28.4 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/cucumber/gherkin/go/v42 v42.0.0 // indirect
	github.com/cucumber/messages/go/v34 v34.2.0 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/dlclark/regexp2/v2 v2.8.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/crypto v0.57.0 // indirect
)

ignore (
	./build
	./frontend
	./tools
)
