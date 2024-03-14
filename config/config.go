package config

import (
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"
	"strconv"
)

const (
	keyGithubToken        = "github-token"
	keySourceOrganization = "source-organization"
	keyTargetOrganization = "target-organization"
	keyVerbose            = "verbose"
	keyDryRun             = "dry-run"
)

type Config struct {
	GithubToken        string
	SourceOrganization string
	TargetOrganization string
	DryRun             bool
}

func New() (*Config, error) {
	c := Config{}
	flag.StringVar(&c.GithubToken, keyGithubToken, lookupEnvOrString("GITHUB_TOKEN", ""), "The GitHub Token to use for authentication.")
	flag.StringVar(&c.SourceOrganization, keySourceOrganization, lookupEnvOrString("SOURCE_ORGANIZATION", ""), "The Source organization.")
	flag.StringVar(&c.TargetOrganization, keyTargetOrganization, lookupEnvOrString("TARGET_ORGANIZATION", ""), "The Target organization.")
	flag.BoolVar(&c.DryRun, keyDryRun, lookupEnvOrBool("DRY_RUN", false), "Dry run mode.")
	verbose := flag.Int("verbose", lookupEnvOrInt(keyVerbose, 0), "Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE.")

	level := slog.LevelInfo
	if *verbose > 0 {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
	flag.Parse()

	if c.GithubToken == "" {
		return nil, errors.New("GitHub Token is required")
	}
	if c.SourceOrganization == "" {
		return nil, errors.New("Source Organization is required")
	}
	if c.TargetOrganization == "" {
		return nil, errors.New("Target Organization is required")
	}

	return &c, nil
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("LookupEnvOrBool[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
