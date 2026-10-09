import { describe, expect, it } from 'vitest'
import { AppError } from '@/platform/errors'
import { workspaceErrorMessage } from './error-message'

describe('workspaceErrorMessage', () => {
  it('explains known codes and falls back to the description', () => {
    expect(workspaceErrorMessage(new AppError('WKS-01409', 'x'))).toMatch(/not empty/)
    expect(workspaceErrorMessage(new AppError('WKS-99500', 'disk full'))).toBe('disk full')
    expect(workspaceErrorMessage(new AppError('WKS-99500', ''))).toBe('Something went wrong.')
  })
})
