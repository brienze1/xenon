import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { VariableInspection, VariableUsage } from '@/platform/mappers/inspection'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({
  inspectVariables: vi.fn(),
  suggestVariables: vi.fn(),
  setCurrentValue: vi.fn(),
  addVariable: vi.fn(),
  getVariableTable: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), applyUrl: vi.fn(), applyParams: vi.fn() }))

const variables = await import('@/platform/services/variables')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { VariableCard } = await import('./VariableCard')

const TAB = 'tab-1'

function inspect(scopes: VariableInspection['scopes'], usages: VariableUsage[] = []) {
  useRequestBuilderStore.getState().ensureTab(TAB)
  useRequestBuilderStore.setState((state) => ({ tabs: { ...state.tabs, [TAB]: { ...state.tabs[TAB]!, inspection: { references: [], variables: usages, scopes } } } }))
}

function usage(patch: Partial<VariableUsage>): VariableUsage {
  return { name: 'creatorAdsUrl', status: 'resolved', scope: 'environment', displayValue: 'https://ads.dev', secret: false, usedIn: ['url'], ...patch }
}

// jsdom has no layout and Vitest loads no CSS-module styles: give the value box its stylesheet metrics (18 px lines,
// 4 px vertical padding, 1 px border) and make each line of its text one laid-out line.
const BOX_METRICS = '.value { line-height: 18px; padding: 4px 8px; border: 1px solid transparent; }'
let overrideScrollHeight: number | null = null

function lines(count: number): string {
  return Array.from({ length: count }, (_, index) => `segment-${index + 1}`).join('\n')
}

describe('VariableCard', () => {
  beforeEach(() => {
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments: [{ id: 'e1', name: 'dev', color: 'none', confirmBeforeSend: false, error: null }], activeId: 'e1' })
    vi.mocked(variables.inspectVariables).mockResolvedValue({ references: [], variables: [], scopes: {} })
  })

  afterEach(() => {
    vi.clearAllMocks()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('adds the variable with the typed value to the chosen scope', async () => {
    inspect({ environment: 'env:e1', collection: 'collection:c1', globals: 'globals' })
    vi.mocked(variables.addVariable).mockResolvedValue({} as never)
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={close} />)
    expect(screen.getByText('{{creatorAdsUrl}}')).toBeInTheDocument()
    await userEvent.type(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' }), 'https://ads.dev')
    const targets = screen.getByRole('list', { name: 'Add creatorAdsUrl to' })
    await userEvent.click(within(targets).getByRole('button', { name: /Environment: dev/ }))
    expect(variables.addVariable).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.dev', false)
    expect(variables.inspectVariables).toHaveBeenCalled()
  })

  it('shows a defined variable\'s current value and source, and saves an edit on Enter', async () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    expect(screen.queryByText(/is not defined/)).not.toBeInTheDocument()
    expect(screen.getByText('Environment: dev')).toBeInTheDocument()
    const value = screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' })
    expect(value).toHaveValue('https://ads.dev')
    await userEvent.clear(value)
    await userEvent.type(value, 'https://ads.prd{Enter}')
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.prd')
  })

  it('keeps an edit when the card closes before Enter or blur', async () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    vi.mocked(variables.setCurrentValue).mockResolvedValue({} as never)
    const { unmount } = render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    await userEvent.type(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' }), '/v2')
    unmount()
    expect(variables.setCurrentValue).toHaveBeenCalledTimes(1)
    expect(variables.setCurrentValue).toHaveBeenCalledWith('ws', 'env:e1', 'creatorAdsUrl', 'https://ads.dev/v2')
  })

  it('does not save an untouched value when the card closes', () => {
    inspect({ environment: 'env:e1' }, [usage({})])
    const { unmount } = render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
    unmount()
    expect(variables.setCurrentValue).not.toHaveBeenCalled()
  })

  it('never shows a secret, and opens the request\'s variables in the side panel', async () => {
    inspect({ globals: 'globals' }, [usage({ scope: 'globals', secret: true, displayValue: '•••••' })])
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={close} />)
    expect(screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' })).toHaveValue('')
    await userEvent.click(screen.getByRole('button', { name: 'Variables in request →' }))
    expect(useWorkspacesStore.getState().layout).toMatchObject({ rightPanelVisible: true, rightPanelView: 'variables' })
    expect(close).toHaveBeenCalled()
  })

  it('disables the scopes the request cannot use, like Postman', () => {
    inspect({ globals: 'globals' })
    render(<VariableCard tabId={TAB} name="x" close={vi.fn()} />)
    expect(screen.getByRole('button', { name: /No active environment/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /No parent collection/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: /Globals/ })).toBeEnabled()
  })

  it('keeps the card open with the reason when adding fails', async () => {
    inspect({ globals: 'globals' })
    vi.mocked(variables.addVariable).mockRejectedValue(new AppError('VAR-04409', 'A variable with that name already exists.'))
    const close = vi.fn()
    render(<VariableCard tabId={TAB} name="x" close={close} />)
    await userEvent.click(screen.getByRole('button', { name: /Globals/ }))
    expect(await screen.findByRole('alert')).toHaveTextContent('A variable with that name already exists.')
    expect(close).not.toHaveBeenCalled()
  })

  describe('value box sizing', () => {
    let metrics: HTMLStyleElement

    beforeEach(() => {
      metrics = document.createElement('style')
      metrics.textContent = BOX_METRICS
      document.head.append(metrics)
      Object.defineProperty(HTMLTextAreaElement.prototype, 'scrollHeight', {
        configurable: true,
        get(this: HTMLTextAreaElement) {
          return overrideScrollHeight ?? 8 + 18 * (this.value === '' ? 1 : this.value.split('\n').length)
        },
      })
    })

    afterEach(() => {
      metrics.remove()
      overrideScrollHeight = null
      delete (HTMLTextAreaElement.prototype as { scrollHeight?: number }).scrollHeight
    })

    function newValueBox(): HTMLTextAreaElement {
      inspect({ environment: 'env:e1' })
      render(<VariableCard tabId={TAB} name="authPrvateUrl" close={vi.fn()} />)
      return screen.getByRole('textbox', { name: 'Value of authPrvateUrl' }) as HTMLTextAreaElement
    }

    function definedValueBox(patch: Partial<VariableUsage>): HTMLTextAreaElement {
      inspect({ environment: 'env:e1' }, [usage(patch)])
      render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
      return screen.getByRole('textbox', { name: 'Value of creatorAdsUrl' }) as HTMLTextAreaElement
    }

    it('opens an undefined variable\'s empty box one line tall, with no scrollbars', () => {
      const box = newValueBox()
      expect(box.rows).toBe(1)
      expect(box.style.height).toBe('28px')
      expect(box.style.overflowY).toBe('hidden')
      expect(box.style.overflowX).toBe('hidden')
    })

    it('keeps a one-line value one line tall, with no scrollbars', async () => {
      const box = newValueBox()
      await userEvent.type(box, 'https://auth.dev')
      expect(box.style.height).toBe('28px')
      expect(box.style.overflowY).toBe('hidden')
    })

    it('grows with a value of several lines, with no scrollbars', async () => {
      const box = newValueBox()
      await userEvent.type(box, 'a{Shift>}{Enter}{/Shift}b{Shift>}{Enter}{/Shift}c')
      expect(box.style.height).toBe('64px')
      expect(box.style.overflowY).toBe('hidden')
    })

    it('stops at 6 lines and scrolls only vertically past them', async () => {
      const box = newValueBox()
      await userEvent.click(box)
      await userEvent.paste(lines(9))
      expect(box.style.height).toBe('118px')
      expect(box.style.overflowY).toBe('auto')
      expect(box.style.overflowX).toBe('hidden')
    })

    it('shrinks back to one line when the value is deleted', async () => {
      const box = newValueBox()
      await userEvent.click(box)
      await userEvent.paste(lines(3))
      expect(box.style.height).toBe('64px')
      await userEvent.clear(box)
      expect(box.style.height).toBe('28px')
    })

    it('does not scroll at exactly 6 lines when the measured height rounds one pixel over', async () => {
      const box = newValueBox()
      overrideScrollHeight = 117
      await userEvent.click(box)
      await userEvent.paste(lines(6))
      expect(box.style.height).toBe('119px')
      expect(box.style.overflowY).toBe('hidden')
    })

    it('opens a defined variable\'s one-line value one line tall, with no scrollbars', () => {
      const box = definedValueBox({})
      expect(box.style.height).toBe('28px')
      expect(box.style.overflowY).toBe('hidden')
      expect(box.style.overflowX).toBe('hidden')
    })

    it('opens a defined variable\'s value of several lines already grown to fit', () => {
      const box = definedValueBox({ displayValue: 'a\nb\nc' })
      expect(box.style.height).toBe('64px')
      expect(box.style.overflowY).toBe('hidden')
    })

    it('shows a secret\'s placeholder on one line, with no scrollbars', () => {
      const box = definedValueBox({ secret: true, displayValue: '' })
      expect(box).toHaveAttribute('placeholder', 'Secret — type to replace…')
      expect(box.style.height).toBe('28px')
      expect(box.style.overflowY).toBe('hidden')
    })

    it('stops a defined variable\'s long value at 6 lines and scrolls only vertically', () => {
      const box = definedValueBox({ displayValue: lines(9) })
      expect(box.style.height).toBe('118px')
      expect(box.style.overflowY).toBe('auto')
      expect(box.style.overflowX).toBe('hidden')
    })

    it('shows no box for a secret whose vault is locked', () => {
      inspect({ environment: 'env:e1' }, [usage({ status: 'locked_secret', secret: true, displayValue: '' })])
      render(<VariableCard tabId={TAB} name="creatorAdsUrl" close={vi.fn()} />)
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Unlock the vault to use this secret' })).toBeInTheDocument()
    })
  })
})
