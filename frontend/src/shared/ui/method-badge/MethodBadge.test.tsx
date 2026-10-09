import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MethodBadge, methodTone } from './MethodBadge'

describe('MethodBadge', () => {
  it('upper-cases the method and colours standard ones', () => {
    render(<MethodBadge method=" post " />)
    const badge = screen.getByText('POST')
    expect(badge).toHaveAttribute('data-tone', 'post')
    expect(badge.className).toContain('post')
  })

  it('treats custom methods as custom', () => {
    expect(methodTone('PURGE')).toBe('custom')
    expect(methodTone('Delete')).toBe('delete')
  })
})
