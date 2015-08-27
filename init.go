package service

import (
	"os"

	"github.com/getsentry/raven-go"
)

func init() {
	raven.SetTagsContext(map[string]string{
		"host":        os.Getenv("HOST"),
		"environment": os.Getenv("APP_ENV"),
		"version":     os.Getenv("APP_VERSION"),
	})
}
