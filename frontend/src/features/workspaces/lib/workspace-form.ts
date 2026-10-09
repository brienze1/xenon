export const NAME_REQUIRED = 'Enter a name for the workspace.'
export const FOLDER_REQUIRED = 'Choose a folder for the workspace.'

export interface WorkspaceFormErrors {
  name: string | null
  folder: string | null
}

export const NO_ERRORS: WorkspaceFormErrors = { name: null, folder: null }

export function validateWorkspaceForm(name: string, folder?: string | null): WorkspaceFormErrors {
  return {
    name: name.trim() === '' ? NAME_REQUIRED : null,
    folder: folder === null ? FOLDER_REQUIRED : null,
  }
}

export function hasErrors(errors: WorkspaceFormErrors): boolean {
  return errors.name !== null || errors.folder !== null
}
