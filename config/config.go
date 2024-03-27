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
	keyGithubToken            = "github-token"
	keyGitHubTokenEnvironment = "GITHUB_TOKEN"
	keyEnterprise             = "enterprise"
	keyEnterpriseEnvironment  = "ENTERPRISE"

	keySourceOrganization            = "source-organization"
	keySourceOrganizationEnvironment = "SOURCE_ORGANIZATION"
	keyTargetOrganization            = "target-organization"
	keyTargetOrganizationEnvironment = "TARGET_ORGANIZATION"
	keyVerbose                       = "verbose"
	keyVerboseEnvironment            = "VERBOSE"
	keyDryRun                        = "dry-run"
	keyDryRunEnvironment             = "DRY_RUN"
)

type Config struct {
	GithubToken        string
	Enterprise         string
	SourceOrganization string
	TargetOrganization string
	DryRun             bool
}

func New() (*Config, error) {
	c := Config{}
	flag.StringVar(&c.GithubToken, keyGithubToken, lookupEnvOrString(keyGitHubTokenEnvironment, ""), "The GitHub Token to use for authentication.")
	flag.StringVar(&c.Enterprise, keyEnterprise, lookupEnvOrString(keyEnterpriseEnvironment, ""), "The GitHub Enterprise to query for repositories.")
	flag.StringVar(&c.SourceOrganization, keySourceOrganization, lookupEnvOrString(keySourceOrganizationEnvironment, ""), "The Source organization.")
	flag.StringVar(&c.TargetOrganization, keyTargetOrganization, lookupEnvOrString(keyTargetOrganizationEnvironment, ""), "The Target organization.")
	flag.BoolVar(&c.DryRun, keyDryRun, lookupEnvOrBool(keyDryRunEnvironment, false), "Dry run mode.")
	verbose := flag.Int(keyVerbose, lookupEnvOrInt(keyVerboseEnvironment, 0), "Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE.")

	level := slog.LevelError
	switch *verbose {
	case 0:
		level = slog.LevelError
	case 1:
		level = slog.LevelWarn
	case 2:
		level = slog.LevelInfo
	case 3:
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
	flag.Parse()

	if c.GithubToken == "" {
		return nil, errors.New("GitHub Token is required")
	}
	if c.Enterprise == "" {
		return nil, errors.New("Enterprise is required")
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
