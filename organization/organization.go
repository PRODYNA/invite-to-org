package organization

import (
	"context"
	"github.com/google/go-github/v60/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log/slog"
)

type Organization struct {
	Name        string
	restClient  *github.Client
	graphClient *githubv4.Client
	loaded      bool
	dryRun      bool
	members     *[]Member
}

type Member struct {
	Login string
	Name  string
}

func New(githubToken string, name string, dryRun bool) (organization *Organization) {
	restclient := github.NewClient(nil).WithAuthToken(githubToken)
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken}))
	graphClient := githubv4.NewClient(httpClient)

	return &Organization{
		restClient:  restclient,
		graphClient: graphClient,
		Name:        name,
		loaded:      false,
		dryRun:      dryRun,
		members:     new([]Member),
	}
}

func (o Organization) Members(ctx context.Context) (members *[]Member, err error) {
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
		err := o.graphClient.Query(ctx, &query, variables)
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

func (o *Organization) MissingMembers(members []Member) (missing []Member) {
	for _, m := range members {
		if !o.HasMember(m.Login) {
			missing = append(missing, m)
		}
	}
	return
}

func (o Organization) Invite(ctx context.Context, enterprise string, members *[]Member) error {
	slog.InfoContext(ctx, "Inviting members", "organization", o.Name, "members", len(*members))
	if o.dryRun {
		slog.InfoContext(ctx, "Dry run - skipping invite", "organization", o.Name, "members", len(*members))
		return nil
	}

	for _, m := range *members {
		slog.InfoContext(ctx, "Invite member", "organization", o.Name, "login", m.Login, "name", m.Name)
		_, _, err := o.restClient.Organizations.CreateOrgInvitation(ctx, o.Name, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to invite member", "organization", o.Name, "login", m.Login, "name", m.Name, "error", err)
		}
		break
	}

	// invite members
	return nil
}
