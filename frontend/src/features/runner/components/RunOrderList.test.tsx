import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  saveRunnerPreferences: vi.fn(async () => undefined),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const { useRunnerStore } = await import('@/features/runner/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { planRow, runPlan } = await import('@/features/runner/test-fixtures')
const { RunOrderList } = await import('./RunOrderList')

const ROW_HEIGHT = 32

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: null, url: null, error: null, warning: null, children }
}

function rows() {
  return useRunnerStore.getState().targets.c1!.rows
}

function setRows(planRows = runPlan().rows) {
  const plan = runPlan({ rows: planRows })
  useRunnerStore.setState({
    targets: {
      c1: { ...useRunnerStore.getState().targets.c1!, status: 'ready', target: plan.target, rows: plan.rows, config: plan.config },
    },
  })
}

/** jsdom has no layout: give each row a stacked rectangle so dnd-kit can measure it. */
function layOutRows() {
  within(screen.getByRole('list', { name: 'Requests in run order' }))
    .getAllByRole('listitem')
    .forEach((row, index) => {
      vi.spyOn(row, 'getBoundingClientRect').mockImplementation(() => DOMRect.fromRect({ x: 0, y: index * ROW_HEIGHT, width: 400, height: ROW_HEIGHT }))
    })
}

function announcement() {
  return screen.getByRole('status').textContent
}

describe('RunOrderList', () => {
  beforeEach(() => {
    useRunnerStore.getState().reset()
    useRunnerStore.getState().showView('c1', 'setup')
  })

  it('lists the rows numbered, with folder path, method and name, all checked', () => {
    setRows()
    render(<RunOrderList itemId="c1" />)
    const list = screen.getByRole('list', { name: 'Requests in run order' })
    const items = within(list).getAllByRole('listitem')
    expect(items).toHaveLength(2)
    expect(items[0]).toHaveTextContent('1')
    expect(items[0]).toHaveTextContent('POST')
    expect(items[0]).toHaveTextContent('Auth / Login')
    expect(within(list).getByRole('checkbox', { name: 'Run Login' })).toBeChecked()
    expect(screen.getByText('2 of 2 selected')).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: 'Select all requests' })).toBeChecked()
  })

  it('toggles rows, shows a mixed select-all, and selects all again', async () => {
    setRows()
    render(<RunOrderList itemId="c1" />)
    await userEvent.click(screen.getByRole('checkbox', { name: 'Run Login' }))
    expect(screen.getByText('1 of 2 selected')).toBeInTheDocument()
    const all = screen.getByRole('checkbox', { name: 'Select all requests' })
    expect(all).toHaveAttribute('aria-checked', 'mixed')
    await userEvent.click(all)
    expect(rows().every((row) => row.selected)).toBe(true)
    await userEvent.click(all)
    expect(rows().some((row) => row.selected)).toBe(false)
    expect(screen.getByText('0 of 2 selected')).toBeInTheDocument()
  })

  it('offers an alias select with Default first and remembers the choice', async () => {
    setRows([planRow('r1', 1, { name: 'Login', aliases: [{ id: 'a1', name: 'mx-test-user' }] }), planRow('r2', 2)])
    render(<RunOrderList itemId="c1" />)
    const select = screen.getByRole('combobox', { name: 'Version of Login' })
    expect(within(select).getAllByRole('option').map((option) => option.textContent)).toEqual(['Default', 'mx-test-user'])
    await userEvent.selectOptions(select, 'a1')
    expect(rows()[0]?.aliasId).toBe('a1')
    expect(screen.queryByRole('combobox', { name: 'Version of Request r2' })).not.toBeInTheDocument()
  })

  it('resets to the sidebar order with every row checked', async () => {
    useSidebarStore.setState({ tree: [node('c1', 'collection', [node('r1', 'request'), node('r2', 'request')])] })
    setRows([planRow('r2', 1, { selected: false }), planRow('r1', 2)])
    render(<RunOrderList itemId="c1" />)
    await userEvent.click(screen.getByRole('button', { name: 'Reset' }))
    expect(rows().map((row) => [row.requestId, row.selected])).toEqual([
      ['r1', true],
      ['r2', true],
    ])
  })

  it('reorders from the keyboard with announcements: Space picks up, arrows move, Space drops', async () => {
    setRows()
    render(<RunOrderList itemId="c1" />)
    layOutRows()
    screen.getByRole('button', { name: 'Reorder Get order' }).focus()
    await userEvent.keyboard('[Space]')
    expect(announcement()).toBe('Picked up Get order.')
    await act(() => new Promise((resolve) => setTimeout(resolve)))
    await userEvent.keyboard('[ArrowUp]')
    expect(announcement()).toBe('Get order moved to position 1 of 2.')
    await userEvent.keyboard('[Space]')
    expect(announcement()).toBe('Get order dropped at position 1 of 2.')
    expect(rows().map((row) => [row.requestId, row.number])).toEqual([
      ['r2', 1],
      ['r1', 2],
    ])
  })

  it('cancels a keyboard reorder with Escape', async () => {
    setRows()
    render(<RunOrderList itemId="c1" />)
    layOutRows()
    screen.getByRole('button', { name: 'Reorder Get order' }).focus()
    await userEvent.keyboard('[Space]')
    await act(() => new Promise((resolve) => setTimeout(resolve)))
    await userEvent.keyboard('[ArrowUp]')
    await userEvent.keyboard('[Escape]')
    expect(announcement()).toBe('Reordering cancelled. Get order is back at position 2 of 2.')
    expect(rows().map((row) => row.requestId)).toEqual(['r1', 'r2'])
  })

  it('shows an empty state when the target has no requests', () => {
    setRows([])
    render(<RunOrderList itemId="c1" />)
    expect(screen.getByRole('region', { name: 'No requests to run' })).toBeInTheDocument()
  })
})
