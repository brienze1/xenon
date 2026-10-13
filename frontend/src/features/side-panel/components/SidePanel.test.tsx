import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({
  inspectVariables: vi.fn(),
  suggestVariables: vi.fn(),
  DUPLICATE_VARIABLE: 'VAR-04409',
  INVALID_NEW_VARIABLE_KEY: 'VAR-13400',
}))

vi.mock('@/platform/services/snippets', () => ({ listSnippetLanguages: vi.fn(async () => []), generateSnippet: vi.fn(), copySnippet: vi.fn() }))

const { SidePanel } = await import('./SidePanel')

describe('SidePanel', () => {
  it('shows the remembered view and reports a switch', async () => {
    const onViewChange = vi.fn()
    const { rerender } = render(<SidePanel tabId="tab-1" view="variables" onViewChange={onViewChange} />)
    expect(screen.getByRole('tab', { name: 'Variables' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('region', { name: 'Variables' })).toBeVisible()
    await userEvent.click(screen.getByRole('tab', { name: 'Info' }))
    expect(onViewChange).toHaveBeenCalledWith('info')
    rerender(<SidePanel tabId="tab-1" view="unknown" onViewChange={onViewChange} />)
    expect(screen.getByRole('tab', { name: 'Info' })).toHaveAttribute('aria-selected', 'true')
    await userEvent.click(screen.getByRole('tab', { name: 'Code' }))
    expect(onViewChange).toHaveBeenCalledWith('code')
  })
})
