package main

import (
	"context"
	config "github.com/prodyna/invite-to-org/config"
	"github.com/prodyna/invite-to-org/organization"
	"log/slog"
	"os"
)

func main() {
	c, err := config.New()
	if err != nil {
		slog.Error("Unable to create config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	slog.InfoContext(ctx, "Configuration",
		"enterprise", c.Enterprise,
		"sourceOrganization", c.SourceOrganization,
		"targetOrganization", c.TargetOrganization,
		"dryRun", c.DryRun,
		"githubToken", "***")

	sourceOrganization := organization.New(c.GithubToken, c.SourceOrganization, c.DryRun)
	sourceMembers, err := sourceOrganization.Members(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to load members", "error", err, "organization", c.SourceOrganization)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "Loaded members", "organization", c.SourceOrganization, "members", len(*sourceMembers))

	targetOrganization := organization.New(c.GithubToken, c.TargetOrganization, c.DryRun)
	targetMembers, err := targetOrganization.Members(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to load members", "error", err, "organization", c.TargetOrganization)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "Loaded members", "organization", c.TargetOrganization, "members", len(*targetMembers))

	missingMembers := targetOrganization.MissingMembers(*sourceMembers)
	slog.InfoContext(ctx, "Missing members", "members", len(missingMembers))
	err = targetOrganization.Invite(ctx, c.Enterprise, &missingMembers)

	slog.InfoContext(ctx, "Done", "membersAdded", len(missingMembers))
}
