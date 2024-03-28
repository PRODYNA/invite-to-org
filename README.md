# GitHub Action for automatically inviting usrs

Users from one organization are automatically invited to another organization.

## Usage

```yaml
jobs:
  invite-to-org:
    runs-on: ubuntu-latest
    concurrency:
      group: ${{ github.workflow }}-${{ github.ref }}
      cancel-in-progress: true

    steps:
      - name: Invite to organization
        uses: prodyna/invite-to-org@v1.1
        with:
          # The GitHub Token to use for authentication, it needs permissions to read member in source-org and invite them to the target-org
          github-token: ${{ secrets.INVITE_TOKEN }}
          # The enterprise to which the source and target organizations belong
          enterprise: "prodyna"
          # The source organization from which to invite members
          source-organization: "prodyna"
          # The target organization to which to invite members
          target-organization: "prodyna-yasm"
          # No dry-run
          dry-run: false
          # Info level
          verbose: 3
          # The team to which to invite members
          team: yasm-all
```

In this case all members of the organization `prodyna` are invited to the organization `prodyna-yasm` in the team `yasm-all`.