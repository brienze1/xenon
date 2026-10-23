import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { PendingApproval } from '@/platform/mappers/ai-access'

vi.mock('@/platform/services/ai-access', () => ({
  APPROVAL_ALREADY_RESOLVED: 'MCP-05409',
  APPROVAL_UNKNOWN: 'MCP-07400',
  listPendingApprovals: vi.fn(),
  answerApproval: vi.fn(),
}))

const service = await import('@/platform/services/ai-access')
const { useAIAccessStore } = await import('@/features/ai-access/store')
const { ApprovalPrompt } = await import('./ApprovalPrompt')

function approval(id: string, overrides: Partial<PendingApproval> = {}): PendingApproval {
  return {
    id,
    client: 'Claude Code',
    method: 'POST',
    url: '{{base_url}}/orders',
    environment: { id: 'e1', name: 'Production', color: 'red' },
    createdAt: '2026-10-13T10:00:00Z',
    summary: null,
    ...overrides,
  }
}

function store() {
  return useAIAccessStore.getState()
}

describe('ApprovalPrompt', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useAIAccessStore.setState({ pending: [], answering: null, error: null })
  })

  afterEach(() => {
    act(() => useAIAccessStore.setState({ pending: [], answering: null, error: null }))
  })

  it('shows the summary of a run instead of a method and URL', () => {
    act(() => store().add(approval('a1', { method: '', url: '', summary: 'Run Orders API · 6 requests' })))
    render(<ApprovalPrompt />)
    const dialog = screen.getByRole('dialog', { name: 'Approve AI Request?' })
    expect(dialog).toHaveTextContent('Run Orders API · 6 requests')
    expect(dialog).toHaveTextContent('wants to run this with the environment')
    expect(dialog.querySelector('[data-method]')).toBeNull()
  })

  it('shows nothing while no AI send waits', () => {
    render(<ApprovalPrompt />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows the oldest approval: client, method, the URL as written and the environment with its colour', () => {
    act(() => {
      store().add(approval('a1'))
      store().add(approval('a2', { client: 'Cursor', method: 'DELETE', url: 'https://api.test/x' }))
    })
    render(<ApprovalPrompt />)
    const dialog = screen.getByRole('dialog', { name: 'Approve AI Request?' })
    expect(dialog).toHaveTextContent('Claude Code')
    expect(dialog).toHaveTextContent('POST')
    expect(dialog).toHaveTextContent('{{base_url}}/orders')
    expect(dialog).toHaveTextContent('Production')
    expect(dialog.querySelector('[data-color="red"]')).not.toBeNull()
    expect(dialog).toHaveTextContent('+1 more waiting')
    expect(dialog).not.toHaveTextContent('Cursor')
  })

  it('approves through the platform and moves on to the next one', async () => {
    vi.mocked(service.answerApproval).mockResolvedValue()
    act(() => {
      store().add(approval('a1'))
      store().add(approval('a2', { client: 'Cursor' }))
    })
    render(<ApprovalPrompt />)
    await userEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(service.answerApproval).toHaveBeenCalledWith('a1', true)
    const dialog = screen.getByRole('dialog', { name: 'Approve AI Request?' })
    expect(dialog).toHaveTextContent('Cursor')
    expect(dialog).not.toHaveTextContent('more waiting')
  })

  it('denies through the platform, also on Escape', async () => {
    vi.mocked(service.answerApproval).mockResolvedValue()
    act(() => {
      store().add(approval('a1'))
      store().add(approval('a2'))
    })
    render(<ApprovalPrompt />)
    await userEvent.click(screen.getByRole('button', { name: 'Deny' }))
    expect(service.answerApproval).toHaveBeenLastCalledWith('a1', false)
    await userEvent.keyboard('{Escape}')
    expect(service.answerApproval).toHaveBeenLastCalledWith('a2', false)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('disappears when the approval is resolved elsewhere (timeout, quit, client gone)', () => {
    act(() => store().add(approval('a1')))
    render(<ApprovalPrompt />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    act(() => store().resolve('a1'))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('drops an approval that was already resolved and shows other answer failures', async () => {
    act(() => {
      store().add(approval('a1'))
      store().add(approval('a2'))
    })
    render(<ApprovalPrompt />)
    vi.mocked(service.answerApproval).mockRejectedValueOnce(new AppError('MCP-05409', 'Approval already resolved'))
    await userEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(store().pending.map((pending) => pending.id)).toEqual(['a2'])
    vi.mocked(service.answerApproval).mockRejectedValueOnce(new AppError('UNK-00500', 'The app is shutting down'))
    await userEvent.click(screen.getByRole('button', { name: 'Approve' }))
    const dialog = screen.getByRole('dialog')
    expect(within(dialog).getByRole('alert')).toHaveTextContent('The app is shutting down. Try again.')
    expect(store().pending.map((pending) => pending.id)).toEqual(['a2'])
  })

  it('loads the approvals already waiting, without duplicates', async () => {
    act(() => store().add(approval('a1')))
    vi.mocked(service.listPendingApprovals).mockResolvedValue([approval('a1'), approval('a2')])
    await act(() => store().load())
    expect(store().pending.map((pending) => pending.id)).toEqual(['a1', 'a2'])
  })
})
