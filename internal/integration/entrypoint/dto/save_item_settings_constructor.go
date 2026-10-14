package dto

// NewSaveItemSettings keeps every stored setting the edit leaves out; edited scripts keep the stored inheritance.
func NewSaveItemSettings(localId, id string, stored ItemSettings, edit *MCPTreeEdit) SaveItemSettings {
	return SaveItemSettings{
		LocalId:        localId,
		Id:             id,
		Auth:           overridden(edit.Auth, stored.Auth),
		Headers:        overridden(edit.Headers, stored.Headers),
		Description:    overridden(edit.Description, stored.Description),
		Scripts:        edit.Scripts,
		InheritScripts: editedScriptsInheritance(edit.Scripts, stored.InheritScripts),
	}
}

func editedScriptsInheritance(scripts *Scripts, stored *bool) *bool {
	if scripts == nil {
		return nil
	}
	return stored
}
