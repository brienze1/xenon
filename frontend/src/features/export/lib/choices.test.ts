import { describe, expect, it } from 'vitest'
import type { ExportOptions } from '@/platform/mappers/export'
import type { ExportChoices } from '@/features/export/model'
import { canExport, defaultChoices, exportRequest, includesSecrets, passphraseProblem, showsProgress, suggestedFileName, withFormat } from './choices'

const options: ExportOptions = {
  suggestedFileName: 'shop-api',
  hasSecrets: true,
  hasCurrentValues: true,
  formats: [
    { format: 'xenon_yaml', suggestedFileName: 'shop-api.zip', shapes: ['zip', 'bundle'], includeSecretsAllowed: true },
    { format: 'postman', suggestedFileName: 'shop-api.postman_collection.json', shapes: ['file'], includeSecretsAllowed: false },
  ],
}

const secret: ExportChoices = {
  format: 'xenon_yaml',
  shape: 'bundle',
  includeSecrets: true,
  passphrase: 'correct horse',
  passphraseConfirmation: 'correct horse',
  currentValuesAsInitial: false,
}

describe('export choices', () => {
  it('starts with the first format and shape, without secrets or current values', () => {
    expect(defaultChoices(options)).toEqual({
      format: 'xenon_yaml',
      shape: 'zip',
      includeSecrets: false,
      passphrase: '',
      passphraseConfirmation: '',
      currentValuesAsInitial: false,
    })
    expect(defaultChoices({ ...options, formats: [] })).toMatchObject({ format: 'xenon_yaml', shape: 'file' })
  })

  it('switching to Postman drops the bundle shape and Include secrets; switching back keeps a shape it offers', () => {
    const postman = withFormat(options, secret, 'postman')
    expect(postman).toMatchObject({ format: 'postman', shape: 'file', includeSecrets: false })
    expect(withFormat(options, postman, 'xenon_yaml')).toMatchObject({ format: 'xenon_yaml', shape: 'zip', includeSecrets: false })
    expect(withFormat(options, secret, 'xenon_yaml')).toEqual(secret)
    expect(withFormat({ ...options, formats: [] }, secret, 'postman')).toBe(secret)
  })

  it('needs a passphrase of at least 8 characters typed twice only when secrets are included', () => {
    expect(passphraseProblem({ ...secret, passphrase: 'short', passphraseConfirmation: 'short' })).toBe('too_short')
    expect(passphraseProblem({ ...secret, passphrase: '12345678', passphraseConfirmation: '12345678' })).toBeNull()
    expect(passphraseProblem({ ...secret, passphraseConfirmation: 'correct horsE' })).toBe('mismatch')
    expect(canExport(options, secret)).toBe(true)
    expect(canExport(options, { ...secret, passphraseConfirmation: '' })).toBe(false)
    expect(canExport(options, { ...secret, includeSecrets: false, passphraseConfirmation: '' })).toBe(true)
    expect(canExport(options, { ...secret, format: 'postman', shape: 'file', passphraseConfirmation: '' })).toBe(true)
    expect(canExport(options, { ...secret, format: 'postman', shape: 'zip' })).toBe(false)
    expect(canExport({ ...options, formats: [] }, secret)).toBe(false)
  })

  it('never sends a passphrase for an export without secrets', () => {
    const target = { localId: 'A', kind: 'collection' as const, itemId: 'C', name: 'Shop API' }
    expect(exportRequest(target, options, secret, '/out/shop.yaml')).toEqual({
      localId: 'A',
      kind: 'collection',
      itemId: 'C',
      format: 'xenon_yaml',
      shape: 'bundle',
      includeSecrets: true,
      passphrase: 'correct horse',
      passphraseConfirmation: 'correct horse',
      currentValuesAsInitial: false,
      targetPath: '/out/shop.yaml',
    })
    const postman = exportRequest(target, options, { ...secret, format: 'postman', shape: 'file', currentValuesAsInitial: true }, '/out/shop.json')
    expect(postman).toMatchObject({ includeSecrets: false, passphrase: '', passphraseConfirmation: '', currentValuesAsInitial: true })
    expect(includesSecrets(options, { ...secret, format: 'postman' })).toBe(false)
  })

  it('suggests the chosen format’s file name', () => {
    expect(suggestedFileName(options, secret)).toBe('shop-api.zip')
    expect(suggestedFileName(options, { ...secret, format: 'postman' })).toBe('shop-api.postman_collection.json')
    expect(suggestedFileName({ ...options, formats: [] }, secret)).toBe('shop-api')
  })

  it('shows progress only for exports of more than 200 files', () => {
    expect(showsProgress(null)).toBe(false)
    expect(showsProgress({ done: 0, total: 200 })).toBe(false)
    expect(showsProgress({ done: 50, total: 201 })).toBe(true)
  })
})
