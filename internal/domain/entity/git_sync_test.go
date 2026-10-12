package entity

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/changetype"
	"github.com/brienze1/xenon/internal/domain/enums/divergence"
	"github.com/brienze1/xenon/internal/domain/enums/gitaccountkind"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/pathaction"
	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
)

var accountCreatedAt = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

func TestGitHubAccountTakesTheUserIdentity(t *testing.T) {
	account := NewGitHubAccount("a1", GitHubUser{ID: 7, Login: "luis", AvatarURL: "https://avatars/7"}, accountCreatedAt)
	want := GitAccount{
		ID: "a1", Kind: gitaccountkind.GitHub, Provider: gitprovider.GitHub, DisplayName: "luis @ github.com", Host: "github.com",
		Username: "luis", CommitName: "luis", CommitEmail: "7+luis@users.noreply.github.com", AvatarURL: "https://avatars/7", CreatedAt: accountCreatedAt,
	}
	if !reflect.DeepEqual(account, want) {
		t.Fatalf("account = %+v", account)
	}
	if !account.CanCallGitHub() || account.IsSsh() || !account.StoresSecret() || account.BasicAuthUsername() != "luis" {
		t.Fatalf("predicates of %+v", account)
	}
	named := GitHubUser{ID: 7, Login: "luis", Name: "Luis B", Email: "luis@acme.test"}
	if named.CommitName() != "Luis B" || named.CommitEmail() != "luis@acme.test" {
		t.Fatalf("named user identity %q %q", named.CommitName(), named.CommitEmail())
	}
}

func TestTokenAccountsUseTheHostConventions(t *testing.T) {
	cases := map[string]struct {
		account  GitAccount
		username string
	}{
		"github with login":    {NewTokenAccount("t", gitprovider.GitHub, "GitHub.com ", " luis ", accountCreatedAt), "luis"},
		"github without login": {NewTokenAccount("t", gitprovider.GitHub, "github.com", "", accountCreatedAt), "x-access-token"},
		"gitlab":               {NewTokenAccount("t", gitprovider.GitLab, "gitlab.acme.test", "luis", accountCreatedAt), "oauth2"},
		"bitbucket":            {NewTokenAccount("t", gitprovider.Bitbucket, "bitbucket.org", "luis", accountCreatedAt), "x-bitbucket-api-token-auth"},
		"azure":                {NewTokenAccount("t", gitprovider.AzureDevOps, "dev.azure.com", "luis", accountCreatedAt), ""},
		"other":                {NewTokenAccount("t", gitprovider.Other, "git.acme.test", "deploy", accountCreatedAt), "deploy"},
	}
	for name, c := range cases {
		if got := c.account.BasicAuthUsername(); got != c.username {
			t.Errorf("%s: BasicAuthUsername = %q, want %q", name, got, c.username)
		}
	}
	github := cases["github with login"].account
	if github.Host != "github.com" || github.Username != "luis" || github.DisplayName != "luis @ github.com" || github.Kind != gitaccountkind.Token || !github.CanCallGitHub() {
		t.Fatalf("token account %+v", github)
	}
	anonymous := cases["github without login"].account
	if anonymous.DisplayName != "github.com" || cases["gitlab"].account.CanCallGitHub() {
		t.Fatalf("display name %q", anonymous.DisplayName)
	}
	identified := anonymous.WithGitHubUser(GitHubUser{ID: 9, Login: "octo", Name: "Octo", AvatarURL: "https://a/9"})
	if identified.Username != "octo" || identified.DisplayName != "octo @ github.com" || identified.CommitName != "Octo" || identified.CommitEmail != "9+octo@users.noreply.github.com" || identified.AvatarURL != "https://a/9" {
		t.Fatalf("WithGitHubUser = %+v", identified)
	}
	kept := anonymous.WithIdentity("", "Me", "me@acme.test").WithGitHubUser(GitHubUser{Login: "octo", Name: "Octo"})
	if kept.CommitName != "Me" || kept.CommitEmail != "me@acme.test" {
		t.Fatalf("WithGitHubUser must keep an explicit identity: %+v", kept)
	}
}

func TestAccountIdentityAndHosts(t *testing.T) {
	account := NewTokenAccount("t", gitprovider.GitLab, "gitlab.acme.test", "luis", accountCreatedAt)
	renamed := account.WithIdentity("  Work  ", " Luis ", " luis@acme.test ")
	if renamed.DisplayName != "Work" || renamed.CommitName != "Luis" || renamed.CommitEmail != "luis@acme.test" {
		t.Fatalf("WithIdentity = %+v", renamed)
	}
	if unchanged := account.WithIdentity(" ", "", ""); unchanged.DisplayName != account.DisplayName || unchanged.CommitName != "" {
		t.Fatalf("blank display name must be kept: %+v", unchanged)
	}
	hosts := map[string]bool{"gitlab.acme.test": true, "GITLAB.acme.test": true, "gitlab.acme.test:8443": true, "github.com": false, "evil.test:443": false, "gitlab.acme.test.evil": false}
	for host, want := range hosts {
		if got := account.ServesHost(host); got != want {
			t.Errorf("ServesHost(%q) = %v, want %v", host, got, want)
		}
	}
	ported := NewTokenAccount("t", gitprovider.Other, "127.0.0.1:8080", "", accountCreatedAt)
	if !ported.ServesHost("127.0.0.1:8080") || ported.ServesHost("127.0.0.1:9090") || ported.ServesHost("127.0.0.1") {
		t.Fatal("an account with a port only serves that port")
	}
}

func TestSshAccounts(t *testing.T) {
	key := SshKey{Path: "/keys/id", PublicKey: "ssh-ed25519 AAAA", Fingerprint: "SHA256:x"}
	cases := map[string]struct {
		account                                       GitAccount
		agent, generated, storesSecret, hasPassphrase bool
	}{
		"agent":                  {NewSshAccount("s", sshmode.Agent, " GitHub.com", SshKey{}, true, accountCreatedAt), true, false, false, false},
		"key file":               {NewSshAccount("s", sshmode.KeyFile, "github.com", key, false, accountCreatedAt), false, false, false, false},
		"key file with password": {NewSshAccount("s", sshmode.KeyFile, "github.com", key, true, accountCreatedAt), false, false, true, true},
		"generated":              {NewSshAccount("s", sshmode.Generated, "github.com", key, true, accountCreatedAt), false, true, false, false},
	}
	for name, c := range cases {
		a := c.account
		if !a.IsSsh() || a.UsesAgent() != c.agent || a.IsGenerated() != c.generated || a.StoresSecret() != c.storesSecret || a.HasPassphrase != c.hasPassphrase || a.CanCallGitHub() {
			t.Errorf("%s: predicates of %+v", name, a)
		}
		if a.Host != "github.com" || a.Username != "git" || a.DisplayName != "SSH @ github.com" || a.Provider != gitprovider.Other {
			t.Errorf("%s: fields of %+v", name, a)
		}
	}
	if file := cases["key file"].account; file.KeyPath != "/keys/id" || file.PublicKey != "ssh-ed25519 AAAA" {
		t.Fatalf("key fields %+v", file)
	}
	token := NewTokenAccount("t", gitprovider.Other, "h", "", accountCreatedAt)
	if token.UsesAgent() || token.IsGenerated() {
		t.Fatal("token accounts are not ssh")
	}
	if (GitAccount{Kind: gitaccountkind.Ssh, SshMode: sshmode.Agent}).IsGenerated() || (GitAccount{SshMode: sshmode.Generated}).IsGenerated() {
		t.Fatal("IsGenerated requires a generated ssh account")
	}
}

func TestGitHubDeviceSessionTiming(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	session := NewGitHubDeviceSession("s1", "dc", 900, 5, now)
	if session.IsTooEarly(now.Add(5*time.Second)) || !session.IsTooEarly(now.Add(4*time.Second)) {
		t.Fatal("the first poll waits one interval")
	}
	if session.IsExpired(now.Add(899*time.Second)) || !session.IsExpired(now.Add(900*time.Second)) {
		t.Fatal("the session expires after expires_in")
	}
	polled := session.Polled(now.Add(6 * time.Second))
	if !polled.NextPollAt.Equal(now.Add(11*time.Second)) || polled.IntervalSeconds() != 5 {
		t.Fatalf("Polled = %+v", polled)
	}
	slowed := polled.SlowedDown(now.Add(11 * time.Second))
	if slowed.IntervalSeconds() != 10 || !slowed.NextPollAt.Equal(now.Add(21*time.Second)) {
		t.Fatalf("SlowedDown = %+v", slowed)
	}
	logged := session.LogValue().(map[string]any)
	if _, leaked := logged["deviceCode"]; leaked || logged["id"] != "s1" {
		t.Fatalf("LogValue = %v", logged)
	}
}

func TestGitHubSignInHidesTheToken(t *testing.T) {
	done := GitHubSignIn{Status: signinstatus.Done, Token: "gho_secret", Interval: 5}
	if !done.IsDone() || (GitHubSignIn{Status: signinstatus.Pending}).IsDone() {
		t.Fatal("IsDone")
	}
	if logged := done.LogValue().(map[string]any); logged["hasToken"] != true || logged["status"] != signinstatus.Done || logged["interval"] != 5 {
		t.Fatalf("LogValue = %v", logged)
	}
}

func TestKnownHostsStatus(t *testing.T) {
	trusted := HostKey{Host: "github.com", KeyType: "ssh-ed25519", Key: "AAAA1"}
	known := KnownHosts{Keys: []HostKey{trusted, {Host: "github.com", KeyType: "ecdsa-sha2-nistp256", Key: "AAAA2"}, {Host: "gitlab.com", KeyType: "ssh-ed25519", Key: "AAAA3"}}}
	cases := map[string]struct {
		key  HostKey
		want string
	}{
		"same key":              {trusted, hostkeystatus.Trusted.String()},
		"other type known":      {HostKey{Host: "github.com", KeyType: "ecdsa-sha2-nistp256", Key: "AAAA2"}, hostkeystatus.Trusted.String()},
		"changed key":           {HostKey{Host: "github.com", KeyType: "ssh-ed25519", Key: "OTHER"}, hostkeystatus.Changed.String()},
		"same bytes other type": {HostKey{Host: "github.com", KeyType: "ssh-rsa", Key: "AAAA1"}, hostkeystatus.Changed.String()},
		"unknown host":          {HostKey{Host: "[127.0.0.1]:2222", KeyType: "ssh-ed25519", Key: "AAAA1"}, hostkeystatus.Unrecognized.String()},
	}
	for name, c := range cases {
		if got := known.StatusOf(c.key).String(); got != c.want {
			t.Errorf("%s: StatusOf = %s, want %s", name, got, c.want)
		}
	}
	if keys := known.KeysFor("gitlab.com"); len(keys) != 1 || keys[0].Key != "AAAA3" {
		t.Fatalf("KeysFor = %+v", keys)
	}
}

func TestGitAccessHidesTheSecret(t *testing.T) {
	access := NewGitAccess("https://h/r.git", GitAccount{ID: "a1"}, "s3cret").WithTrustedHostKeys([]HostKey{{Host: "h"}})
	if access.Repository != "https://h/r.git" || access.Secret != "s3cret" || len(access.TrustedHostKeys) != 1 {
		t.Fatalf("access %+v", access)
	}
	logged := access.LogValue().(map[string]any)
	if logged["hasSecret"] != true || logged["accountId"] != "a1" || logged["trustedHostKeys"] != 1 || logged["repository"] != "https://h/r.git" {
		t.Fatalf("LogValue = %v", logged)
	}
}

func TestLocalSyncSettingsAccount(t *testing.T) {
	settings := LocalSyncSettings{AccountID: "a1", ClonePath: "/c", AutoPull: true}
	if !settings.UsesAccount("a1") || settings.UsesAccount("a2") || (LocalSyncSettings{}).UsesAccount("") {
		t.Fatal("UsesAccount")
	}
	if cleared := settings.WithoutAccount(); cleared.AccountID != "" || cleared.ClonePath != "/c" || !cleared.AutoPull {
		t.Fatalf("WithoutAccount = %+v", cleared)
	}
}

func TestChangeItemCombination(t *testing.T) {
	ref := ItemRef{ID: "r1", Name: "Orders", Kind: itemkind.Request}
	deleted := NewChangeItem(ref, PathChange{Path: "old.yaml", Change: changetype.Deleted})
	moved := deleted.Combined(PathChange{Path: "new.yaml", Change: changetype.Added})
	if moved.Change != changetype.Renamed || !reflect.DeepEqual(moved.Paths, []string{"old.yaml", "new.yaml"}) || !reflect.DeepEqual(deleted.Paths, []string{"old.yaml"}) {
		t.Fatalf("moved %+v, original %+v", moved, deleted)
	}
	modified := NewChangeItem(ref, PathChange{Path: "a.yaml", Change: changetype.Modified}).Combined(PathChange{Path: "a.yaml", Change: changetype.Deleted})
	if modified.Change != changetype.Modified {
		t.Fatalf("modified %+v", modified)
	}
}

func TestPathPlanResultAndStructuralKinds(t *testing.T) {
	if (PathPlan{Mine: "m", Theirs: "t", Action: pathaction.TakeTheirs}).Result() != "t" || (PathPlan{Mine: "m", Theirs: "t", Action: pathaction.Keep}).Result() != "m" || (PathPlan{Mine: "m", Theirs: "t", Action: pathaction.TakeMine}).Result() != "m" {
		t.Fatal("Result")
	}
	kinds := map[string]string{
		"workspace.yaml": "workspace", "collections/a/collection.yaml": "collection", "collections/a/b/folder.yaml": "folder",
		"environments/dev.yaml": "environment", "environments/notes.txt": "unknown", "environments/x/y.yaml": "unknown",
		"collections/a/req.yaml": "unknown", "collections/workspace.yaml": "unknown",
	}
	for path, want := range kinds {
		if got := (PathPlan{Path: path}).StructuralKind().String(); got != want {
			t.Errorf("StructuralKind(%s) = %s, want %s", path, got, want)
		}
	}
	if !(PathPlan{Action: pathaction.BothModified}).IsConflict() || !(PathPlan{Action: pathaction.DeleteModify}).IsConflict() || (PathPlan{Action: pathaction.MergeStructural}).IsConflict() {
		t.Fatal("IsConflict")
	}
	if NewPathPlan("environments/a.yaml", "", "", "1").Action != pathaction.TakeTheirs || NewPathPlan("environments/a.yaml", "", "1", "2").Action != pathaction.MergeStructural {
		t.Fatal("added environments")
	}
}

func TestDocumentNodeEquality(t *testing.T) {
	mapping := NewMappingNode().Set("a", NewScalarNode("!!str", "1")).Set("b", NewSequenceNode([]*DocumentNode{NewScalarNode("!!str", "x")}))
	reordered := NewMappingNode().Set("b", NewSequenceNode([]*DocumentNode{NewScalarNode("!!str", "x")})).Set("a", NewScalarNode("!!str", "1"))
	if !mapping.Equals(reordered) || !reflect.DeepEqual(mapping.Keys, []string{"a", "b"}) {
		t.Fatal("mappings compare by content, not order")
	}
	mapping.Set("a", NewScalarNode("!!str", "2"))
	if len(mapping.Keys) != 2 || mapping.Field("a").Value != "2" {
		t.Fatalf("Set must replace without duplicating keys: %v", mapping.Keys)
	}
	var missing *DocumentNode
	cases := map[string]bool{
		"nil equals nil":       missing.Equals(nil),
		"nil differs":          !missing.Equals(NewScalarNode("!!str", "")),
		"node differs nil":     !NewScalarNode("!!str", "").Equals(nil),
		"tag matters":          !NewScalarNode("!!str", "20").Equals(NewScalarNode("!!int", "20")),
		"value matters":        !NewScalarNode("!!str", "a").Equals(NewScalarNode("!!str", "b")),
		"kind matters":         !NewSequenceNode(nil).Equals(NewMappingNode()),
		"field count matters":  !NewMappingNode().Set("a", NewScalarNode("", "1")).Equals(NewMappingNode()),
		"field value matters":  !NewMappingNode().Set("a", NewScalarNode("", "1")).Equals(NewMappingNode().Set("b", NewScalarNode("", "1"))),
		"items length matters": !NewSequenceNode([]*DocumentNode{NewScalarNode("", "1")}).Equals(NewSequenceNode(nil)),
		"items content":        !NewSequenceNode([]*DocumentNode{NewScalarNode("", "1")}).Equals(NewSequenceNode([]*DocumentNode{NewScalarNode("", "2")})),
		"nil field":            missing.Field("a") == nil,
		"nil is not mapping":   !missing.IsMapping(),
	}
	for name, ok := range cases {
		if !ok {
			t.Errorf("%s", name)
		}
	}
}

func TestDocumentNodeListShapes(t *testing.T) {
	scalars := NewSequenceNode([]*DocumentNode{NewScalarNode("", "a"), NewScalarNode("", "b")})
	if values, ok := scalars.ScalarItems(); !ok || !reflect.DeepEqual(values, []string{"a", "b"}) {
		t.Fatalf("ScalarItems = %v %v", values, ok)
	}
	row := func(key string) *DocumentNode {
		return NewMappingNode().Set("key", NewScalarNode("", key))
	}
	if _, ok := NewSequenceNode([]*DocumentNode{row("a")}).ScalarItems(); ok {
		t.Fatal("mappings are not scalar items")
	}
	if _, ok := NewMappingNode().ScalarItems(); ok {
		t.Fatal("a mapping has no scalar items")
	}
	keys, rows, ok := NewSequenceNode([]*DocumentNode{row("a"), row("b")}).KeyedRows()
	if !ok || !reflect.DeepEqual(keys, []string{"a", "b"}) || len(rows) != 2 {
		t.Fatalf("KeyedRows = %v %v %v", keys, rows, ok)
	}
	invalid := map[string]*DocumentNode{
		"duplicate":    NewSequenceNode([]*DocumentNode{row("a"), row("a")}),
		"scalar item":  scalars,
		"missing key":  NewSequenceNode([]*DocumentNode{NewMappingNode()}),
		"mapping key":  NewSequenceNode([]*DocumentNode{NewMappingNode().Set("key", NewMappingNode())}),
		"not sequence": NewMappingNode(),
	}
	for name, node := range invalid {
		if _, _, ok := node.KeyedRows(); ok {
			t.Errorf("%s: KeyedRows must refuse", name)
		}
	}
	if (StructuralMerge{}).HasConflicts() || !(StructuralMerge{ItemConflict: true}).HasConflicts() || !(StructuralMerge{FieldConflicts: []FieldConflict{{}}}).HasConflicts() {
		t.Fatal("HasConflicts")
	}
}

func TestBranchHeadsClassifyTheirDivergence(t *testing.T) {
	cases := map[string]struct {
		heads       BranchHeads
		want        enums.Divergence
		nothing     bool
		fastForward bool
		diverged    bool
	}{
		"same commit":            {BranchHeads{Local: "a", Remote: "a", Base: "a"}, divergence.UpToDate, true, false, false},
		"both empty":             {BranchHeads{}, divergence.UpToDate, true, false, false},
		"remote branch missing":  {BranchHeads{Local: "a"}, divergence.Ahead, true, false, false},
		"remote is the base":     {BranchHeads{Local: "b", Remote: "a", Base: "a"}, divergence.Ahead, true, false, false},
		"nothing local":          {BranchHeads{Remote: "a"}, divergence.FastForward, false, true, false},
		"local is the base":      {BranchHeads{Local: "a", Remote: "b", Base: "a"}, divergence.FastForward, false, true, false},
		"both moved":             {BranchHeads{Local: "b", Remote: "c", Base: "a"}, divergence.Diverged, false, false, true},
		"unrelated without base": {BranchHeads{Local: "b", Remote: "c"}, divergence.Diverged, false, false, true},
	}
	for name, c := range cases {
		if got := c.heads.Divergence(); got != c.want {
			t.Errorf("%s: Divergence = %s, want %s", name, got, c.want)
		}
		if c.heads.HasNothingToPull() != c.nothing || c.heads.IsFastForward() != c.fastForward || c.heads.IsDiverged() != c.diverged {
			t.Errorf("%s: predicates of %+v", name, c.heads)
		}
	}
}

func TestSyncedFilesLocateWhereEachPathIsRead(t *testing.T) {
	worktree := SyncedFiles{"a.yaml": "1", "b.yaml": "2"}
	present, deleted := worktree.LocateChanges([]PathChange{{Path: "a.yaml", Change: changetype.Modified}, {Path: "c.yaml", Change: changetype.Deleted}}, map[string]string{"c.yaml": "9"})
	if !reflect.DeepEqual(present, []string{"a.yaml"}) || !reflect.DeepEqual(deleted, map[string]string{"c.yaml": "9"}) {
		t.Fatalf("changes located as %v %v", present, deleted)
	}
	present, blobs := worktree.LocatePlans([]PathPlan{{Path: "b.yaml", Mine: "2"}, {Path: "d.yaml", Base: "4", Theirs: "5"}})
	if !reflect.DeepEqual(present, []string{"b.yaml"}) || !reflect.DeepEqual(blobs, map[string]string{"d.yaml": "5"}) {
		t.Fatalf("plans located as %v %v", present, blobs)
	}
}

func TestSyncSessionNeedsAReplayCommitOnlyWhenTheResolvedTreeLeavesTheRemote(t *testing.T) {
	remote := map[string]string{"a.yaml": "1"}
	replay := SyncSession{Replay: true}
	if !replay.ReplaysLocalCommits() || replay.NeedsReplayCommit(remote, map[string]string{"a.yaml": "1"}) || !replay.NeedsReplayCommit(map[string]string{"a.yaml": "2"}, remote) {
		t.Fatalf("replaying session predicates")
	}
	if (SyncSession{}).ReplaysLocalCommits() || (SyncSession{}).NeedsReplayCommit(map[string]string{"a.yaml": "2"}, remote) {
		t.Fatalf("fast-forward session predicates")
	}
	worktree := SyncedFiles{"a.yaml": "1"}
	if !worktree.Holds("a.yaml", "1") || worktree.Holds("a.yaml", "2") || worktree.Holds("b.yaml", "1") {
		t.Fatalf("worktree holds")
	}
}

func TestItemIndexUsesTheIndexAndCombinesMoves(t *testing.T) {
	index := ItemIndex{
		"collections/billing/orders.yaml":         {ID: "r1", Name: "Orders", Kind: itemkind.Request},
		"collections/billing/archive/orders.yaml": {ID: "r1", Name: "Orders", Kind: itemkind.Request},
		"collections/billing/collection.yaml":     {ID: "c1", Name: "Billing", Kind: itemkind.Collection},
		"collections/billing/ping.yaml":           {ID: "r2", Name: "Ping", Kind: itemkind.Request},
	}
	changes := []PathChange{
		{Path: "collections/billing/archive/orders.yaml", Change: changetype.Added},
		{Path: "collections/billing/collection.yaml", Change: changetype.Modified},
		{Path: "collections/billing/orders.yaml", Change: changetype.Deleted},
		{Path: "collections/billing/ping.yaml", Change: changetype.Added},
		{Path: "collections/billing/ping.yaml", Change: changetype.Modified},
		{Path: "environments/hml.yaml", Change: changetype.Deleted},
		{Path: "workspace.yaml", Change: changetype.Modified},
		{Path: "collections/sales/collection.yaml", Change: changetype.Added},
		{Path: "collections/sales/leads/folder.yaml", Change: changetype.Added},
		{Path: "collections/sales/leads/new-lead.yaml", Change: changetype.Added},
	}
	got := index.ChangeItems(changes)
	want := []ChangeItem{
		{ItemID: "r1", ItemName: "Orders", ItemKind: itemkind.Request, Change: changetype.Renamed, Paths: []string{"collections/billing/archive/orders.yaml", "collections/billing/orders.yaml"}},
		{ItemID: "c1", ItemName: "Billing", ItemKind: itemkind.Collection, Change: changetype.Modified, Paths: []string{"collections/billing/collection.yaml"}},
		{ItemID: "r2", ItemName: "Ping", ItemKind: itemkind.Request, Change: changetype.Modified, Paths: []string{"collections/billing/ping.yaml", "collections/billing/ping.yaml"}},
		{ItemID: "environments/hml.yaml", ItemName: "hml", ItemKind: itemkind.Environment, Change: changetype.Deleted, Paths: []string{"environments/hml.yaml"}},
		{ItemID: "workspace.yaml", ItemName: "workspace", ItemKind: itemkind.Workspace, Change: changetype.Modified, Paths: []string{"workspace.yaml"}},
		{ItemID: "collections/sales/collection.yaml", ItemName: "sales", ItemKind: itemkind.Collection, Change: changetype.Added, Paths: []string{"collections/sales/collection.yaml"}},
		{ItemID: "collections/sales/leads/folder.yaml", ItemName: "leads", ItemKind: itemkind.Folder, Change: changetype.Added, Paths: []string{"collections/sales/leads/folder.yaml"}},
		{ItemID: "collections/sales/leads/new-lead.yaml", ItemName: "new-lead", ItemKind: itemkind.Request, Change: changetype.Added, Paths: []string{"collections/sales/leads/new-lead.yaml"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChangeItems =\n%+v\nwant\n%+v", got, want)
	}
}

func TestItemIndexJoinsWorkspaceLevelFilesToTheWorkspace(t *testing.T) {
	workspace := ItemRef{ID: "w1", Name: "Acme", Kind: itemkind.Workspace}
	joined := ItemIndex{"workspace.yaml": workspace}.ChangeItems([]PathChange{
		{Path: ".gitignore", Change: changetype.Added},
		{Path: "workspace.yaml", Change: changetype.Added},
	})
	want := []ChangeItem{{ItemID: "w1", ItemName: "Acme", ItemKind: itemkind.Workspace, Change: changetype.Added, Paths: []string{".gitignore", "workspace.yaml"}}}
	if !reflect.DeepEqual(joined, want) {
		t.Fatalf("joined = %+v, want %+v", joined, want)
	}
	alone := ItemIndex{}.ChangeItems([]PathChange{{Path: ".gitignore", Change: changetype.Modified}})
	want = []ChangeItem{{ItemID: ".gitignore", ItemName: ".gitignore", ItemKind: itemkind.Workspace, Change: changetype.Modified, Paths: []string{".gitignore"}}}
	if !reflect.DeepEqual(alone, want) {
		t.Fatalf("alone = %+v, want %+v", alone, want)
	}
}

func TestGitHubDeviceSessionAnswersExpiredAndPendingSignIns(t *testing.T) {
	session := NewGitHubDeviceSession("s1", "dc", 900, 5, accountCreatedAt)
	expired, pending := session.ExpiredSignIn(), session.PendingSignIn()
	if expired.Status != signinstatus.Expired || expired.Interval != 5 || expired.IsPending() || expired.IsDone() {
		t.Fatalf("expired = %+v", expired)
	}
	if pending.Status != signinstatus.Pending || pending.Interval != 5 || !pending.IsPending() {
		t.Fatalf("pending = %+v", pending)
	}
}
