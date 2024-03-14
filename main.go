package main

import (
	"context"
	config "github.com/prodyna/invite-to-org/config"
	"github.com/prodyna/invite-to-org/organization"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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
	slog.InfoContext("Configuration",
		"sourceOrganization", c.SourceOrganization,
		"targetOrganization", c.TargetOrganization,
		"dryRun", c.DryRun,
		"githubToken", "***")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GithubToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	graphClient := githubv4.NewClient(httpClient)

	sourceOrganization := organization.New(graphClient, c.SourceOrganization, c.DryRun)
	sourceMembers, err := sourceOrganization.Members(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to load members", "error", err, "organization", c.SourceOrganization)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "Loaded members", "organization", c.SourceOrganization, "members", len(sourceMembers))

	targetOrganization := organization.New(graphClient, c.TargetOrganization, c.DryRun)
	targetMembers, err := targetOrganization.Members(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to load members", "error", err, "organization", c.TargetOrganization)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "Loaded members", "organization", c.TargetOrganization, "members", len(targetMembers))

	missingMembers := targetOrganization.MissingMembers(sourceMembers)
	for _, m := range missingMembers {
		err := targetOrganization.Invite(ctx, m)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to invite member", "error", err, "organization", c.TargetOrganization, "member", m)
		}
	}

	slog.InfoContext(ctx, "Done", "membersAdded", len(missingMembers))
}
