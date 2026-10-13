package mcp

import "net/http"

const sessionIdHeader = "Mcp-Session-Id"

// closingSessions runs the session-closed hooks when a client asks to end its session, before the SDK drains the
// session's in-flight calls: a call waiting on the human (an approval) would otherwise keep the session from closing.
func (s *Server) closingSessions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if sessionId := request.Header.Get(sessionIdHeader); request.Method == http.MethodDelete && sessionId != "" {
			for _, hook := range s.hooks() {
				hook(request.Context(), sessionId)
			}
		}
		next.ServeHTTP(writer, request)
	})
}
