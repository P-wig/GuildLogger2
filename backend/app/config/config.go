// backend/app/config/config.go
package config

import (
	"os"
	"strings"
)

type Config struct {
	MongoURI    string
	MongoDB     string
	CORSOrigins []string
}

func Load() Config {
	origins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(origins) == 0 || (len(origins) == 1 && strings.TrimSpace(origins[0]) == "") {
		origins = []string{"http://localhost:5173"}
	}
	return Config{
		MongoURI:    getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:     getenv("MONGO_DB", "test"),
		CORSOrigins: trimAll(origins),
	}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func trimAll(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
