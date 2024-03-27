package organization

import (
	"context"
	"github.com/google/go-github/v60/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"log/slog"
)

type OrganizationConfig struct {
	GithubToken  string
	DryRun       bool
	Organization string
	Team         string
}

type Organization struct {
	name        string
	restClient  *github.Client
	graphClient *githubv4.Client
	loaded      bool
	dryRun      bool
	members     *[]Member
	team        string
}

type Member struct {
	ID    string
	Login string
	Name  string
}

func New(config OrganizationConfig) *Organization {
	restclient := github.NewClient(nil).WithAuthToken(config.GithubToken)
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GithubToken}))
	graphClient := githubv4.NewClient(httpClient)

	return &Organization{
		restClient:  restclient,
		graphClient: graphClient,
		name:        config.Organization,
		loaded:      false,
		dryRun:      config.DryRun,
		team:        config.Team,
		members:     new([]Member),
	}
}

func (o Organization) Members(ctx context.Context) (members *[]Member, err error) {
	if !o.loaded {
		slog.DebugContext(ctx, "Organization members not loaded", "organization", o.name)
		err = o.loadMembers(ctx)
		if err != nil {
			return nil, err
		}
		slog.DebugContext(ctx, "Organization members loaded", "organization", o.name, "members", len(*o.members))
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
						ID    string
						Login string
						Name  string
					}
				}
			} `graphql:"membersWithRole(first: 100, after: $cursor)"`
		} `graphql:"organization(login: $organization)"`
	}

	variables := map[string]interface{}{
		"organization": githubv4.String(o.name),
		"cursor":       (*githubv4.String)(nil),
	}

	for {
		err := o.graphClient.Query(ctx, &query, variables)
		if err != nil {
			return err
		}
		for _, e := range query.Organization.MembersWithRole.Edges {
			*o.members = append(*o.members, Member{
				ID:    e.Node.ID,
				Login: e.Node.Login,
				Name:  e.Node.Name,
			})
			slog.DebugContext(ctx, "Loaded member", "organization", o.name, "login", e.Node.Login, "name", e.Node.Name, "id", e.Node.ID)
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
	slog.DebugContext(ctx, "Finding team", "organization", o.name, "team", o.team)
	team, _, err := o.restClient.Teams.GetTeamBySlug(ctx, o.name, o.team)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to find team", "organization", o.name, "team", o.team, "error", err)
		return err
	}
	slog.InfoContext(ctx, "Found team", "organization", o.name, "team", o.team, "id", team.GetID())

	slog.DebugContext(ctx, "Checking for already invited members", "organization", o.name, "members", len(*members))
	invitationsMap := make(map[string]bool)
	page := 0
	pageSize := 30
	for {
		invitations, _, err := o.restClient.Organizations.ListPendingOrgInvitations(ctx, o.name, &github.ListOptions{Page: page, PerPage: pageSize})
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list pending invitations", "organization", o.name, "error", err)
			return err
		}
		for _, i := range invitations {
			invitationsMap[i.GetLogin()] = true
		}
		if len(invitations) < pageSize {
			break
		}
		page++
	}

	slog.InfoContext(ctx, "Inviting members", "organization", o.name, "members", len(*members))
	if o.dryRun {
		slog.InfoContext(ctx, "Dry run - skipping invite", "organization", o.name, "members", len(*members))
		return nil
	}

	for _, m := range *members {
		if _, ok := invitationsMap[m.Login]; ok {
			slog.DebugContext(ctx, "Already invited member", "organization", o.name, "login", m.Login, "name", m.Name, "id", m.ID)
			continue
		}
		slog.InfoContext(ctx, "Invite member", "organization", o.name, "login", m.Login, "name", m.Name, "id", m.ID)
		_, _, err := o.restClient.Teams.AddTeamMembershipBySlug(ctx, o.name, team.GetSlug(), m.Login, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to invite member", "organization", o.name, "login", m.Login, "name", m.Name, "error", err)
		}
	}

	// invite members
	return nil
}
