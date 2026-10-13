package adapter

import wails "github.com/wailsapp/wails/v3/pkg/application"

type Router interface {
	Route() *wails.App
}
