package entity

import "strconv"

const aiAccessLoopbackName = "localhost"

func IsAILoopbackHost(host string, port int) bool {
	suffix := ":" + strconv.Itoa(port)
	return host == aiAccessHost+suffix || host == aiAccessLoopbackName+suffix
}

func IsAILoopbackOrigin(origin string, port int) bool {
	suffix := ":" + strconv.Itoa(port)
	for _, name := range []string{aiAccessHost, aiAccessLoopbackName} {
		if origin == aiAccessURLScheme+name || origin == aiAccessURLScheme+name+suffix {
			return true
		}
	}
	return false
}
