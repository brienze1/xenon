import { describe, expect, it } from 'vitest'
import { AppError } from '@/platform/errors'
import { workspaceErrorMessage } from './error-message'

describe('workspaceErrorMessage', () => {
  it('explains known codes and falls back to the description', () => {
    expect(workspaceErrorMessage(new AppError('WKS-01409', 'x'))).toMatch(/isn’t empty/)
    expect(workspaceErrorMessage(new AppError('WKS-99500', 'disk full'))).toBe(
      'disk full. Try again, or check that the folder exists and you have permission to write to it.',
    )
    expect(workspaceErrorMessage(new AppError('WKS-99500', ''))).toBe(
      'Something went wrong. Try again, or check that the folder exists and you have permission to write to it.',
    )
  })
})
