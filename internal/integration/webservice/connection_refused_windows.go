//go:build windows

package webservice

import "syscall"

const wsaConnectionRefused = 10061

var connectionRefused error = syscall.Errno(wsaConnectionRefused)
