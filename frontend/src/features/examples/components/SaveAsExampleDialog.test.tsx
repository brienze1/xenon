import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/examples', () => ({ getExample: vi.fn(), saveExample: vi.fn(), createExampleFromResponse: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const examples = await import('@/platform/services/examples')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useExamplesStore } = await import('@/features/examples/store')
const { SaveAsExampleDialog } = await import('./SaveAsExampleDialog')

describe('SaveAsExampleDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useExamplesStore.getState().reset()
  })

  it('stays closed until a tab asks to save its response', () => {
    render(<SaveAsExampleDialog onCreated={vi.fn()} />)
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('needs a name, then creates the example and hands it over', async () => {
    const onCreated = vi.fn()
    vi.mocked(examples.createExampleFromResponse).mockResolvedValue({ id: 'e9' } as never)
    render(<SaveAsExampleDialog onCreated={onCreated} />)
    act(() => useExamplesStore.getState().openSaveDialog('tab-1', 'r1'))
    expect(screen.getByRole('dialog', { name: 'Save Response as Example' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Save Example' }))
    expect(screen.getByRole('alert')).toHaveTextContent('Enter a name for the example.')
    expect(examples.createExampleFromResponse).not.toHaveBeenCalled()
    await userEvent.type(screen.getByRole('textbox', { name: 'Example name' }), ' Created ')
    await userEvent.click(screen.getByRole('button', { name: 'Save Example' }))
    expect(examples.createExampleFromResponse).toHaveBeenCalledWith('ws', 'r1', 'tab-1', 'Created')
    expect(onCreated).toHaveBeenCalledWith('r1', 'e9')
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('shows why the example could not be created', async () => {
    vi.mocked(examples.createExampleFromResponse).mockRejectedValue(new AppError('EXM-02404', 'the tab has no response to save'))
    render(<SaveAsExampleDialog onCreated={vi.fn()} />)
    act(() => useExamplesStore.getState().openSaveDialog('tab-1', 'r1'))
    await userEvent.type(screen.getByRole('textbox', { name: 'Example name' }), 'Created')
    await userEvent.click(screen.getByRole('button', { name: 'Save Example' }))
    expect(await screen.findByText(/the tab has no response to save/)).toBeInTheDocument()
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(useExamplesStore.getState().saveDialog).toBeNull()
  })
})
