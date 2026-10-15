import { describe, expect, it } from 'vitest'
import { AppError } from '@/platform/errors'
import { isNamePending, nameErrorMessage } from './name-draft'

describe('isNamePending', () => {
  it.each([
    ['no saved name (never saved)', 'Ping', null, false],
    ['no draft', undefined, 'Ping', false],
    ['the saved name', 'Ping', 'Ping', false],
    ['only spaces around the saved name', '  Ping ', 'Ping', false],
    ['another name', 'Pong', 'Ping', true],
    ['an empty name', '', 'Ping', true],
    ['spaces only', '   ', 'Ping', true],
  ] as const)('%s', (_case, draft, saved, pending) => {
    expect(isNamePending(draft, saved)).toBe(pending)
  })
})

describe('nameErrorMessage', () => {
  it.each(['request', 'example', 'collection', 'folder', 'environment'] as const)('asks for a %s name', (subject) => {
    expect(nameErrorMessage({ kind: 'required', subject })).toBe(`Type a name for the ${subject}.`)
  })

  it('says which environment name is taken', () => {
    expect(nameErrorMessage({ kind: 'clash', name: '  dev ' })).toBe('An environment named “dev” already exists. Choose another name.')
  })

  it('shows what the rename returned with the next step', () => {
    expect(nameErrorMessage({ kind: 'failed', error: new AppError('TRE-09400', 'The name is too long') })).toBe(
      'The name is too long. Choose another name, then save.',
    )
    expect(nameErrorMessage({ kind: 'failed', error: new AppError('TRE-05500', '') })).toBe('TRE-05500. Choose another name, then save.')
  })
})
