package arp

import "net/http"

// CORSMiddleware adds CORS headers so external UI renderers (CDN, dev server)
// can access the ARP and API endpoints.
func CORSMiddleware(next http.Handler, origin string) http.Handler {
	if origin == "" {
		origin = "*"
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
