package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds app configuration from env (and optional .env file).
type Config struct {
	Root                  string        // Filesystem root for files (e.g. /data, /srv)
	DBPath                string        // SQLite database path
	SessionSecret         string        // Secret for signing session cookies
	Port                  string        // HTTP listen port (e.g. 8080, 80)
	DefaultAdminUser      string        // First-run admin username (when no users exist)
	DefaultAdminPassword  string        // First-run admin password (change immediately)
	PUID                  int           // Docker: owner for created files (0 = don't chown)
	PGID                  int           // Docker: group for created files (0 = don't chown)
	AppName               string        // Optional app name in UI (e.g. "Studio Photos")
	// Optimization (docs/optimization-recommendations.md)
	ReadHeaderTimeout     time.Duration // HTTP read header timeout (slow clients)
	ReadTimeout           time.Duration // HTTP read body timeout
	WriteTimeout          time.Duration // HTTP write timeout (e.g. large downloads)
	MaxUploadBytes        int64         // Max size per file upload (0 = use default 100MB)
	MaxRequestBytes        int64         // Max request body (0 = use default 500MB)
	ThumbMaxSizeShare     int           // Thumb max dimension for share page (default 200)
	ThumbMaxSizeAdmin     int           // Thumb max dimension for admin (default 320)
	ThumbConcurrency      int           // Max concurrent thumb generations (0 = use default 4)
	ZipMaxFiles           int           // Max files in one ZIP (0 = use default 500)
	ZipMaxBytes            int64         // Max total bytes in ZIP (0 = use default 2GB)
	SQLiteBusyTimeout      time.Duration // SQLite busy timeout (0 = use default 5s)
	StaticCacheMaxAge      int           // Cache-Control max-age for static assets (seconds, 0 = 86400)
}

const (
	defaultMaxUploadBytes   = 100 << 20  // 100MB
	defaultMaxRequestBytes  = 500 << 20  // 500MB
	defaultThumbMaxShare     = 200
	defaultThumbMaxAdmin     = 320
	defaultThumbConcurrency  = 4
	defaultZipMaxFiles       = 500
	defaultZipMaxBytes       = 2 << 30   // 2GB
	defaultStaticCacheAge   = 86400     // 1 day
)

// Load reads configuration from the environment. For optional .env file,
// source it before running (e.g. export $(cat .env | xargs)) or use env_file in Docker.
func Load() (*Config, error) {
	c := &Config{
		Root:                 getEnv("ROOT", "/data"),
		DBPath:               getEnv("DB_PATH", "/data/db/app.sqlite"),
		SessionSecret:        getEnv("SESSION_SECRET", ""),
		Port:                 getEnv("PORT", "8080"),
		DefaultAdminUser:     getEnv("DEFAULT_ADMIN_USER", "admin"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin"),
		AppName:              getEnv("APP_NAME", "Studio Photos"),
	}
	c.PUID, _ = strconv.Atoi(getEnv("PUID", "0"))
	c.PGID, _ = strconv.Atoi(getEnv("PGID", "0"))

	// Timeouts (optimization: avoid stuck connections on 2GB ARM)
	c.ReadHeaderTimeout = durationEnv("READ_HEADER_TIMEOUT", 10*time.Second)
	c.ReadTimeout = durationEnv("READ_TIMEOUT", 30*time.Second)
	c.WriteTimeout = durationEnv("WRITE_TIMEOUT", 60*time.Second)

	// Upload/request limits (return 413 when exceeded)
	c.MaxUploadBytes = int64Env("MAX_UPLOAD_BYTES", defaultMaxUploadBytes)
	c.MaxRequestBytes = int64Env("MAX_REQUEST_BYTES", defaultMaxRequestBytes)
	if c.MaxUploadBytes <= 0 {
		c.MaxUploadBytes = defaultMaxUploadBytes
	}
	if c.MaxRequestBytes <= 0 {
		c.MaxRequestBytes = defaultMaxRequestBytes
	}

	// Thumbnail limits (memory-safe on 2GB ARM)
	c.ThumbMaxSizeShare = intEnv("THUMB_MAX_SIZE_SHARE", defaultThumbMaxShare)
	c.ThumbMaxSizeAdmin = intEnv("THUMB_MAX_SIZE_ADMIN", defaultThumbMaxAdmin)
	c.ThumbConcurrency = intEnv("THUMB_CONCURRENCY", defaultThumbConcurrency)
	if c.ThumbMaxSizeShare <= 0 {
		c.ThumbMaxSizeShare = defaultThumbMaxShare
	}
	if c.ThumbMaxSizeAdmin <= 0 {
		c.ThumbMaxSizeAdmin = defaultThumbMaxAdmin
	}
	if c.ThumbConcurrency <= 0 {
		c.ThumbConcurrency = defaultThumbConcurrency
	}

	// ZIP limits (avoid long-running requests)
	c.ZipMaxFiles = intEnv("ZIP_MAX_FILES", defaultZipMaxFiles)
	c.ZipMaxBytes = int64Env("ZIP_MAX_BYTES", defaultZipMaxBytes)
	if c.ZipMaxFiles <= 0 {
		c.ZipMaxFiles = defaultZipMaxFiles
	}
	if c.ZipMaxBytes <= 0 {
		c.ZipMaxBytes = defaultZipMaxBytes
	}

	c.SQLiteBusyTimeout = durationEnv("SQLITE_BUSY_TIMEOUT", 5*time.Second)
	c.StaticCacheMaxAge = intEnv("STATIC_CACHE_MAX_AGE", defaultStaticCacheAge)
	if c.StaticCacheMaxAge <= 0 {
		c.StaticCacheMaxAge = defaultStaticCacheAge
	}

	return c, nil
}

func durationEnv(key string, def time.Duration) time.Duration {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

func int64Env(key string, def int64) int64 {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func intEnv(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// EnsureDirs creates ROOT and the parent directory of DBPath if they do not exist.
// Call at startup so storage and DB work without "directory not found" (Phase 1).
func EnsureDirs(c *Config) error {
	if err := os.MkdirAll(c.Root, 0755); err != nil {
		return err
	}
	dbDir := filepath.Dir(c.DBPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	}
	return nil
}
