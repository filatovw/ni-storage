package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/filatovw/ni-storage/logger"
	"github.com/go-chi/chi/middleware"
)

func LevelLogger(l logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				l.Info("Served ",
					fmt.Sprintf("proto: %s; ", r.Proto),
					fmt.Sprintf("method: %s; ", r.Method),
					fmt.Sprintf("path: %s; ", r.URL.Path),
					fmt.Sprintf("latency: %d; ", time.Since(t1)),
					fmt.Sprintf("status: %d; ", ww.Status()),
					fmt.Sprintf("size: %d; ", ww.BytesWritten()),
					fmt.Sprintf("reqId: %s", middleware.GetReqID(r.Context())))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
