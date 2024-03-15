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
	members *[]Member
}

type Member struct {
	Login string
	Name  string
}

func New(client *githubv4.Client, name string, dryRun bool) (organization *Organization) {
	return &Organization{
		client:  client,
		Name:    name,
		loaded:  false,
		dryRun:  dryRun,
		members: new([]Member),
	}
}

func (o *Organization) Members(ctx context.Context) (members *[]Member, err error) {
	if !o.loaded {
		slog.DebugContext(ctx, "Organization members not loaded", "organization", o.Name)
		err = o.loadMembers(ctx)
		if err != nil {
			return nil, err
		}
		slog.DebugContext(ctx, "Organization members loaded", "organization", o.Name, "members", len(*o.members))
	}
	return o.members, nil
}

func (o *Organization) loadMembers(ctx context.Context) error {
	var query struct {
		Organization struct {
			MembersWithRole struct {
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
				Edges []struct {
					Node struct {
						Login string
						Name  string
					}
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $organization)"`
	}

	variables := map[string]interface{}{
		"organization": githubv4.String(o.Name),
		"cursor":       (*githubv4.String)(nil),
	}

	for {
		err := o.client.Query(ctx, &query, variables)
		if err != nil {
			return err
		}
		for _, e := range query.Organization.MembersWithRole.Edges {
			slog.DebugContext(ctx, "Loaded member", "organization", o.Name, "login", e.Node.Login, "name", e.Node.Name)
			*o.members = append(*o.members, Member{
				Login: e.Node.Login,
				Name:  e.Node.Name,
			})
		}
		if !query.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.String(query.Organization.MembersWithRole.PageInfo.EndCursor)
	}

	return nil
}

func (c *Organization) HasMember(login string) bool {
	for _, m := range *c.members {
		if m.Login == login {
			return true
		}
	}
	return false
}

func (c *Organization) MissingMembers(members []Member) (missing []Member) {
	for _, m := range members {
		if !c.HasMember(m.Login) {
			missing = append(missing, m)
		}
	}
	return
}

func (c *Organization) Invite(ctx context.Context, member Member) error {
	slog.InfoContext(ctx, "Inviting member", "organization", c.Name, "login", member.Login, "name", member.Name)
	if c.dryRun {
		slog.InfoContext(ctx, "Dry run, not inviting member", "organization", c.Name, "member", member)
		return nil
	}
	// TODO: Actually invite member
	return nil
}
