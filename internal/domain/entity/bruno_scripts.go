package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

const untranslatedMessage = "This line has no pm.* equivalent and was commented out: "

type BrunoScripts struct {
	PreRequest       BrunoScript
	PostResponse     BrunoScript
	Tests            BrunoScript
	ItemID           string
	Path             string
	File             string
	PreRequestVars   []BrunoVariable
	PostResponseVars []BrunoVariable
	Assertions       []BrunoAssertion
}

func (b BrunoScripts) IsEmpty() bool {
	return b.PreRequest.IsEmpty() && b.PostResponse.IsEmpty() && b.Tests.IsEmpty() &&
		len(b.PreRequestVars)+len(b.PostResponseVars)+len(b.Assertions) == 0
}

func (b BrunoScripts) Translated(preRequest, postResponse ScriptTranslation) (Scripts, []ImportWarning) {
	scripts := Scripts{PreRequest: preRequest.Text(), PostResponse: postResponse.Text()}
	warnings := make([]ImportWarning, 0)
	for _, line := range preRequest.Append(postResponse).Untranslated {
		warning := NewImportWarning(warningcode.UntranslatedScript, untranslatedMessage+line.Source, b.ItemID, b.Path)
		warnings = append(warnings, warning.AtLine(b.File+":"+strconv.Itoa(line.Line)))
	}
	return scripts, warnings
}
