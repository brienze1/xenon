package enums_test

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/changetype"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/divergence"
	"github.com/brienze1/xenon/internal/domain/enums/gitaccountkind"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
	"github.com/brienze1/xenon/internal/domain/enums/nodekind"
	"github.com/brienze1/xenon/internal/domain/enums/pathaction"
	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
	"github.com/brienze1/xenon/internal/domain/enums/syncstate"
)

func TestStageSevenEnumsParseTheirValues(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{gitaccountkind.GetGitAccountKind(" GITHUB ").String(), "github"},
		{gitaccountkind.GetGitAccountKind(" TOKEN ").String(), "token"},
		{gitaccountkind.GetGitAccountKind(" SSH ").String(), "ssh"},
		{gitaccountkind.GetGitAccountKind("not-a-value").String(), "unknown"},
		{sshmode.GetSshMode(" AGENT ").String(), "agent"},
		{sshmode.GetSshMode(" KEY_FILE ").String(), "key_file"},
		{sshmode.GetSshMode(" GENERATED ").String(), "generated"},
		{sshmode.GetSshMode("not-a-value").String(), "unknown"},
		{gitprovider.GetGitProvider(" GITHUB ").String(), "github"},
		{gitprovider.GetGitProvider(" GITLAB ").String(), "gitlab"},
		{gitprovider.GetGitProvider(" BITBUCKET ").String(), "bitbucket"},
		{gitprovider.GetGitProvider(" AZURE_DEVOPS ").String(), "azure_devops"},
		{gitprovider.GetGitProvider(" OTHER ").String(), "other"},
		{gitprovider.GetGitProvider("not-a-value").String(), "unknown"},
		{syncstate.GetSyncState(" UP_TO_DATE ").String(), "up_to_date"},
		{syncstate.GetSyncState(" LOCAL_CHANGES ").String(), "local_changes"},
		{syncstate.GetSyncState(" AHEAD ").String(), "ahead"},
		{syncstate.GetSyncState(" BEHIND ").String(), "behind"},
		{syncstate.GetSyncState(" DIVERGED ").String(), "diverged"},
		{syncstate.GetSyncState(" CONFLICT ").String(), "conflict"},
		{syncstate.GetSyncState(" ERROR ").String(), "error"},
		{syncstate.GetSyncState(" SYNCING ").String(), "syncing"},
		{syncstate.GetSyncState(" NOT_LINKED ").String(), "not_linked"},
		{syncstate.GetSyncState("not-a-value").String(), "unknown"},
		{changetype.GetChangeType(" ADDED ").String(), "added"},
		{changetype.GetChangeType(" MODIFIED ").String(), "modified"},
		{changetype.GetChangeType(" DELETED ").String(), "deleted"},
		{changetype.GetChangeType(" RENAMED ").String(), "renamed"},
		{changetype.GetChangeType("not-a-value").String(), "unknown"},
		{conflictkind.GetConflictKind(" BOTH_MODIFIED ").String(), "both_modified"},
		{conflictkind.GetConflictKind(" DELETE_MODIFY ").String(), "delete_modify"},
		{conflictkind.GetConflictKind(" FIELD ").String(), "field"},
		{conflictkind.GetConflictKind("not-a-value").String(), "unknown"},
		{conflictresolution.GetConflictResolution(" UNRESOLVED ").String(), "unresolved"},
		{conflictresolution.GetConflictResolution(" MINE ").String(), "mine"},
		{conflictresolution.GetConflictResolution(" THEIRS ").String(), "theirs"},
		{conflictresolution.GetConflictResolution(" BOTH ").String(), "both"},
		{conflictresolution.GetConflictResolution(" KEEP ").String(), "keep"},
		{conflictresolution.GetConflictResolution(" DELETE ").String(), "delete"},
		{conflictresolution.GetConflictResolution(" MERGED ").String(), "merged"},
		{conflictresolution.GetConflictResolution("not-a-value").String(), "unknown"},
		{divergence.GetDivergence(" UP_TO_DATE ").String(), "up_to_date"},
		{divergence.GetDivergence(" FAST_FORWARD ").String(), "fast_forward"},
		{divergence.GetDivergence(" AHEAD ").String(), "ahead"},
		{divergence.GetDivergence(" DIVERGED ").String(), "diverged"},
		{divergence.GetDivergence("not-a-value").String(), "unknown"},
		{pathaction.GetPathAction(" KEEP ").String(), "keep"},
		{pathaction.GetPathAction(" TAKE_MINE ").String(), "take_mine"},
		{pathaction.GetPathAction(" TAKE_THEIRS ").String(), "take_theirs"},
		{pathaction.GetPathAction(" MERGE_STRUCTURAL ").String(), "merge_structural"},
		{pathaction.GetPathAction(" BOTH_MODIFIED ").String(), "both_modified"},
		{pathaction.GetPathAction(" DELETE_MODIFY ").String(), "delete_modify"},
		{pathaction.GetPathAction("not-a-value").String(), "unknown"},
		{hostkeystatus.GetHostKeyStatus(" TRUSTED ").String(), "trusted"},
		{hostkeystatus.GetHostKeyStatus(" UNRECOGNIZED ").String(), "unrecognized"},
		{hostkeystatus.GetHostKeyStatus(" CHANGED ").String(), "changed"},
		{hostkeystatus.GetHostKeyStatus("not-a-value").String(), "unknown"},
		{signinstatus.GetSignInStatus(" PENDING ").String(), "pending"},
		{signinstatus.GetSignInStatus(" DONE ").String(), "done"},
		{signinstatus.GetSignInStatus(" EXPIRED ").String(), "expired"},
		{signinstatus.GetSignInStatus(" DENIED ").String(), "denied"},
		{signinstatus.GetSignInStatus("not-a-value").String(), "unknown"},
		{nodekind.GetNodeKind(" MAPPING ").String(), "mapping"},
		{nodekind.GetNodeKind(" SEQUENCE ").String(), "sequence"},
		{nodekind.GetNodeKind(" SCALAR ").String(), "scalar"},
		{nodekind.GetNodeKind("not-a-value").String(), "unknown"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestStageSevenEnumsCompareIgnoringCase(t *testing.T) {
	checks := map[string]bool{
		"gitaccountkind":     gitaccountkind.GitHub.Equals("GITHUB") && !gitaccountkind.GitHub.Equals("token"),
		"sshmode":            sshmode.Agent.Equals("AGENT") && !sshmode.Agent.Equals("key_file"),
		"gitprovider":        gitprovider.GitHub.Equals("GITHUB") && !gitprovider.GitHub.Equals("gitlab"),
		"syncstate":          syncstate.UpToDate.Equals("UP_TO_DATE") && !syncstate.UpToDate.Equals("local_changes"),
		"changetype":         changetype.Added.Equals("ADDED") && !changetype.Added.Equals("modified"),
		"conflictkind":       conflictkind.BothModified.Equals("BOTH_MODIFIED") && !conflictkind.BothModified.Equals("delete_modify"),
		"conflictresolution": conflictresolution.Unresolved.Equals("UNRESOLVED") && !conflictresolution.Unresolved.Equals("mine"),
		"divergence":         divergence.UpToDate.Equals("UP_TO_DATE") && !divergence.UpToDate.Equals("fast_forward"),
		"pathaction":         pathaction.Keep.Equals("KEEP") && !pathaction.Keep.Equals("take_mine"),
		"hostkeystatus":      hostkeystatus.Trusted.Equals("TRUSTED") && !hostkeystatus.Trusted.Equals("unrecognized"),
		"signinstatus":       signinstatus.Pending.Equals("PENDING") && !signinstatus.Pending.Equals("done"),
		"nodekind":           nodekind.Mapping.Equals("MAPPING") && !nodekind.Mapping.Equals("sequence"),
	}
	for name, equal := range checks {
		if !equal {
			t.Errorf("%s: Equals must ignore case and tell values apart", name)
		}
	}
}
