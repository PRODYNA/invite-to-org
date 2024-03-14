package organization

import (
	"context"
	"github.com/shurcooL/githubv4"
	"log/slog"
)

type Organization struct {
	Name    string
	client  *githubv4.Client
	loaded  bool
	dryRun  bool
	members []string
}

func New(client *githubv4.Client, name string, dryRun bool) (organization *Organization) {
	return &Organization{
		client: client,
		Name:   name,
		loaded: false,
		dryRun: dryRun,
	}

	// TODO: Load members
}

func (o *Organization) Members(ctx context.Context) (members []string, err error) {
	if !o.loaded {
		slog.DebugContext(ctx, "Organization not loaded", "organization", o.Name)
		// TODO
	}
	return o.members, nil
}

func (c *Organization) HasMember(member string) bool {
	for _, m := range c.members {
		if m == member {
			return true
		}
	}
	return false
}

func (c *Organization) MissingMembers(members []string) (missing []string) {
	for _, m := range members {
		if !c.HasMember(m) {
			missing = append(missing, m)
		}
	}
	return
}

func (c *Organization) Invite(ctx context.Context, member string) error {
	slog.InfoContext(ctx, "Inviting member", "organization", c.Name, "member", member)
	if c.dryRun {
		slog.InfoContext(ctx, "Dry run, not inviting member", "organization", c.Name, "member", member)
		return nil
	}
	// TODO: Actually invite member
	return nil
}
