package usecase

import (
	"net/http"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	unsupportedFormatMsg = " is not a Xenon, Postman, Bruno or cURL export"
	invalidJsonMsg       = " is not valid JSON"
	mixedBrunoFormatsMsg = " mixes Bruno .bru and YAML files"
	postmanV1Msg         = "Postman v1 collections are not supported; export the collection as v2.1"
	newerXenonMsg        = " was written by a newer Xenon (format "
	updateXenonMsg       = "); update Xenon to import it"
)

func DetectImportFormat(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	switch {
	case probe.IsPastedCurl():
		return importkind.Curl, "", nil
	case probe.IsXenon():
		return detectXenon(probe)
	case probe.IsFolder:
		return detectFolder(probe)
	case probe.IsPostmanV3Request():
		return importkind.PostmanRequest, entity.PostmanV3Version, nil
	case probe.IsMalformed:
		return importkind.Unknown, "", invalidImportError(probe.Name + invalidJsonMsg)
	case probe.IsJson:
		return detectJson(probe)
	case probe.IsYaml:
		return detectPostmanV3(probe)
	default:
		return importkind.Unknown, "", unsupportedImportError(probe.Name + unsupportedFormatMsg)
	}
}

func detectFolder(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	switch {
	case probe.IsPostmanDump():
		return importkind.PostmanDump, "", nil
	case probe.MixesBrunoFormats():
		return importkind.Unknown, "", unsupportedImportError(probe.Name + mixedBrunoFormatsMsg)
	case probe.IsBrunoBru():
		return importkind.BrunoBru, "", nil
	case probe.IsBrunoYaml():
		return importkind.BrunoYaml, "", nil
	case probe.IsPostmanV3Workspace():
		return importkind.PostmanWorkspace, entity.PostmanV3Version, nil
	case probe.IsPostmanV3CollectionFolder():
		return importkind.PostmanCollection, entity.PostmanV3Version, nil
	default:
		return importkind.Folder, "", nil
	}
}

func detectXenon(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	version := probe.XenonVersion()
	switch {
	case probe.IsNewerXenon():
		return importkind.Unknown, "", errs.UnprocessableEntityError(probe.Name+newerXenonMsg+version+updateXenonMsg, "IMP-42422")
	case probe.IsXenonWorkspaceExport():
		return importkind.XenonWorkspaceZip, version, nil
	case probe.IsXenonExport():
		return importkind.XenonZip, version, nil
	case probe.IsXenonBundle():
		return importkind.XenonBundle, version, nil
	case probe.IsXenonWorkspaceFolder():
		return importkind.XenonWorkspaceFolder, version, nil
	case probe.IsXenonRequest():
		return importkind.XenonRequest, version, nil
	case probe.IsXenonEnvironment():
		return importkind.XenonEnvironment, version, nil
	default:
		return importkind.Unknown, "", unsupportedImportError(probe.Name + unsupportedFormatMsg)
	}
}

func detectJson(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	if version, found := probe.PostmanV2Version(); found {
		return importkind.PostmanCollection, version, nil
	}
	switch {
	case probe.IsPostmanV1():
		return importkind.Unknown, "", invalidImportError(postmanV1Msg)
	case probe.IsEnvironment():
		return importkind.PostmanEnvironment, "", nil
	case probe.IsGlobals():
		return importkind.PostmanGlobals, "", nil
	default:
		return detectPostmanV3(probe)
	}
}

func detectPostmanV3(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	switch {
	case probe.IsPostmanV3Collection():
		return importkind.PostmanCollection, entity.PostmanV3Version, nil
	case probe.IsPostmanV3Variables() && probe.IsPostmanV3Globals():
		return importkind.PostmanGlobals, entity.PostmanV3Version, nil
	case probe.IsPostmanV3Variables():
		return importkind.PostmanEnvironment, entity.PostmanV3Version, nil
	default:
		return importkind.Unknown, "", unsupportedImportError(probe.Name + unsupportedFormatMsg)
	}
}

func invalidImportError(description string) error {
	return errs.BadRequestError(description, "IMP-01400")
}

func unsupportedImportError(description string) error {
	return errs.ClientError(description, "IMP-00415", http.StatusUnsupportedMediaType)
}
