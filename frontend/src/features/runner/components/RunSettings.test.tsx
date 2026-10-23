import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  saveRunnerPreferences: vi.fn(async () => undefined),
  startRun: vi.fn(),
  getRun: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFile: vi.fn(), pickSaveFile: vi.fn(), pickFolder: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { planRow, readyTarget, runConfig, runPlan, runSummaryView } = await import('@/features/runner/test-fixtures')
const { RunSettings } = await import('./RunSettings')

function config() {
  return useRunnerStore.getState().targets.c1!.config!
}

function seed(extra: Parameters<typeof readyTarget>[1] = {}, plan = runPlan()) {
  useRunnerStore.setState({ targets: { c1: readyTarget(plan, extra) } })
}

describe('RunSettings', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows the functional settings with Postman’s labels, order and defaults', () => {
    seed()
    render(<RunSettings itemId="c1" />)
    expect(screen.getByRole('radio', { name: 'Functional' })).toBeChecked()
    expect(screen.getByRole('spinbutton', { name: 'Iterations' })).toHaveValue(1)
    expect(screen.getByRole('spinbutton', { name: /Delay/ })).toHaveValue(0)
    const advanced = screen.getByRole('group', { name: 'Advanced settings' })
    expect(within(advanced).getAllByRole('checkbox').map((box) => [box.parentElement?.textContent, (box as HTMLInputElement).checked])).toEqual([
      ['Persist responses for a session', false],
      ['Turn off logs during run', false],
      ['Stop run if an error occurs', false],
      ['Keep variable values', true],
      ['Run collection without using stored cookies', false],
      ['Save cookies after collection run', false],
    ])
  })

  it('edits the settings, keeping a partly typed number as text', async () => {
    seed()
    render(<RunSettings itemId="c1" />)
    const iterations = screen.getByRole('spinbutton', { name: 'Iterations' })
    await userEvent.clear(iterations)
    expect(iterations).toHaveValue(null)
    await userEvent.type(iterations, '12')
    expect(config().functional.iterations).toBe(12)
    await userEvent.click(screen.getByRole('checkbox', { name: 'Stop run if an error occurs' }))
    expect(config().functional.stopOnError).toBe(true)
  })

  it('shows a RUN-04400 field error inline under its field', () => {
    seed({ fieldErrors: { iterations: 'must be at least 1', data_file: 'not an array of objects' } })
    render(<RunSettings itemId="c1" />)
    expect(screen.getByRole('spinbutton', { name: 'Iterations' })).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getAllByRole('alert').map((alert) => alert.textContent)).toEqual(['must be at least 1', 'not an array of objects'])
  })

  it('picks a CSV or JSON data file, shows its name, and clears it', async () => {
    seed()
    vi.mocked(dialogs.pickFile).mockResolvedValue('/data/users.csv')
    render(<RunSettings itemId="c1" />)
    await userEvent.click(screen.getByRole('button', { name: 'Select file…' }))
    expect(dialogs.pickFile).toHaveBeenCalledWith('Choose Iteration Data', [{ displayName: 'Data files (CSV, JSON)', pattern: '*.csv;*.json' }])
    expect(await screen.findByText('users.csv')).toHaveAttribute('title', '/data/users.csv')
    await userEvent.click(screen.getByRole('button', { name: 'Clear iteration data' }))
    expect(config().functional.dataFile).toBe('')
  })

  it('notes a remembered data file that no longer exists', async () => {
    seed({ dataFileMissing: true })
    render(<RunSettings itemId="c1" />)
    expect(screen.getByRole('status')).toHaveTextContent('no longer exists')
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(screen.queryByText(/no longer exists/)).not.toBeInTheDocument()
  })

  it('picks the run environment: the active one by default, another one or none, flagging a confirm-before-send one', async () => {
    seed()
    render(<RunSettings itemId="c1" />)
    const select = screen.getByRole('combobox', { name: 'Environment' })
    expect(select).toHaveValue('e1')
    expect(within(select).getAllByRole('option').map((option) => option.textContent)).toEqual(['No environment', 'dev (active)', 'prod'])
    await userEvent.selectOptions(select, 'e2')
    expect(config().environmentId).toBe('e2')
    expect(screen.getByText('This environment asks for a confirmation before the run starts.')).toBeInTheDocument()
    await userEvent.selectOptions(select, 'none')
    expect(config().environmentId).toBe('none')
  })

  it('switches to the performance settings', async () => {
    seed({}, runPlan({ config: runConfig() }))
    render(<RunSettings itemId="c1" />)
    await userEvent.click(screen.getByRole('radio', { name: 'Performance' }))
    expect(config().type).toBe('performance')
    expect(screen.getByRole('spinbutton', { name: 'Virtual users' })).toHaveValue(10)
  })

  it('disables Start run without a selection or while a run is active here', () => {
    seed({ rows: [planRow('r1', 1, { selected: false })] })
    const { unmount } = render(<RunSettings itemId="c1" />)
    expect(screen.getByRole('button', { name: 'Start run' })).toBeDisabled()
    unmount()
    seed()
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    render(<RunSettings itemId="c1" />)
    expect(screen.getByRole('button', { name: 'Start run' })).toBeDisabled()
    expect(screen.getByText('A run is already in progress here.')).toBeInTheDocument()
  })
})
