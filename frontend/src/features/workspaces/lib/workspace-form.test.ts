import { describe, expect, it } from 'vitest'
import { FOLDER_REQUIRED, NAME_REQUIRED, hasErrors, validateWorkspaceForm } from './workspace-form'

describe('validateWorkspaceForm', () => {
  it('requires a non-blank name and, when asked for, a folder', () => {
    expect(validateWorkspaceForm('  ', null)).toEqual({ name: NAME_REQUIRED, folder: FOLDER_REQUIRED })
    expect(validateWorkspaceForm('Acme', '/work/acme')).toEqual({ name: null, folder: null })
    expect(validateWorkspaceForm('Acme')).toEqual({ name: null, folder: null })
  })

  it('reports whether any field is invalid', () => {
    expect(hasErrors(validateWorkspaceForm('', '/work'))).toBe(true)
    expect(hasErrors(validateWorkspaceForm('Acme', null))).toBe(true)
    expect(hasErrors(validateWorkspaceForm('Acme', '/work'))).toBe(false)
  })
})
