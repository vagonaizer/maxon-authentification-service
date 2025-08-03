package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

func Logging(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			fields := logger.Fields{
				"method":     req.Method,
				"uri":        req.RequestURI,
				"status":     res.Status,
				"latency":    time.Since(start).String(),
				"user_agent": req.UserAgent(),
				"remote_ip":  c.RealIP(),
			}

			if userID := c.Get("user_id"); userID != nil {
				fields["user_id"] = userID
			}

			if err != nil {
				fields["error"] = err.Error()
				log.WithFields(fields).Error("request completed with error")
			} else {
				if res.Status >= 500 {
					log.WithFields(fields).Error("request completed")
				} else if res.Status >= 400 {
					log.WithFields(fields).Warn("request completed")
				} else {
					log.WithFields(fields).Info("request completed")
				}
			}

			return err
		}
	}
}
