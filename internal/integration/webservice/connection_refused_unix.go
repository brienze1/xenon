//go:build !windows

package webservice

import "syscall"

var connectionRefused error = syscall.ECONNREFUSED
