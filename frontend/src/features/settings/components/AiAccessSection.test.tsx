import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AIAccessStatus } from '@/platform/mappers/ai-access'
import { defaultAppSettings } from '@/platform/mappers/settings'

vi.mock('@/platform/services/settings', () => ({
  SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400',
  HISTORY_USER_CAP_OUT_OF_RANGE: 'SET-02400',
  HISTORY_AI_CAP_OUT_OF_RANGE: 'SET-03400',
  AI_PORT_OUT_OF_RANGE: 'SET-04400',
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
}))
vi.mock('@/platform/services/ai-access', () => ({ getAIAccessStatus: vi.fn(), copyText: vi.fn() }))

const aiAccess = await import('@/platform/services/ai-access')
const { useSettingsStore } = await import('@/features/settings/store')
const { AiAccessSection } = await import('./AiAccessSection')

const HTTP_CONFIG = '{\n  "mcpServers": { "xenon": { "type": "http", "url": "http://127.0.0.1:47821/mcp" } }\n}'
const STDIO_CONFIG = '{\n  "mcpServers": { "xenon": { "command": "/Applications/Xenon.app/Contents/MacOS/xenon", "args": ["mcp"] } }\n}'

const running: AIAccessStatus = {
  state: 'running',
  address: 'http://127.0.0.1:47821/mcp',
  reason: null,
  code: null,
  clients: 0,
  stdioCommand: '/Applications/Xenon.app/Contents/MacOS/xenon mcp',
  httpConfig: HTTP_CONFIG,
  stdioConfig: STDIO_CONFIG,
}

function settings() {
  return useSettingsStore.getState()
}

function statusLine() {
  return screen.getByRole('status', { name: 'AI access status' })
}

describe('AiAccessSection', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useSettingsStore.setState({ settings: defaultAppSettings(), fieldErrors: {}, aiStatus: running })
    settings().discard()
  })

  afterEach(() => {
    act(() => useSettingsStore.setState({ aiStatus: null, fieldErrors: {} }))
  })

  it('shows the status line and follows the ai:access-status-changed updates', () => {
    render(<AiAccessSection />)
    expect(statusLine()).toHaveTextContent('Running at http://127.0.0.1:47821/mcp')
    act(() => settings().applyStatus({ state: 'running', address: 'http://127.0.0.1:47821/mcp', reason: null, code: null, clients: 2 }))
    expect(statusLine()).toHaveTextContent('Running at http://127.0.0.1:47821/mcp · 2 clients connected')
    act(() => settings().applyStatus({ state: 'error', address: null, reason: 'Port 47821 is in use', code: 'MCP-01500', clients: 0 }))
    expect(statusLine()).toHaveTextContent('Port 47821 is in use')
    act(() => settings().applyStatus({ state: 'off', address: null, reason: null, code: null, clients: 0 }))
    expect(statusLine()).toHaveTextContent(/^Off$/)
  })

  it('edits the switch and the port in the form, applied on Save', async () => {
    render(<AiAccessSection />)
    const toggle = screen.getByRole('switch', { name: 'AI Access' })
    expect(toggle).toBeChecked()
    await userEvent.click(toggle)
    expect(toggle).not.toBeChecked()
    expect(settings().draft.aiEnabled).toBe(false)
    const port = screen.getByRole('spinbutton', { name: /Port/ })
    expect(port).toHaveValue(47821)
    await userEvent.clear(port)
    await userEvent.type(port, '50000')
    expect(settings().draft.aiPort).toBe('50000')
    act(() => useSettingsStore.setState({ fieldErrors: { aiPort: 'Enter a port from 1024 to 65535.' } }))
    expect(screen.getByRole('alert')).toHaveTextContent('Enter a port from 1024 to 65535.')
    expect(port).toHaveAttribute('aria-invalid', 'true')
  })

  it('offers "Dangerously skip approvals", off by default, with its warning', async () => {
    render(<AiAccessSection />)
    const skip = screen.getByRole('checkbox', { name: 'Dangerously skip approvals' })
    expect(skip).not.toBeChecked()
    expect(skip).toHaveAccessibleDescription(/sends to environments marked “Confirm before send” without asking you/)
    await userEvent.click(skip)
    expect(settings().draft.skipApprovals).toBe(true)
  })

  it('shows the stdio command and both client configs, each with a copy button', async () => {
    vi.mocked(aiAccess.copyText).mockResolvedValue(true)
    render(<AiAccessSection />)
    expect(screen.getByText('/Applications/Xenon.app/Contents/MacOS/xenon mcp')).toBeInTheDocument()
    expect(screen.getByLabelText('HTTP client config')).toHaveTextContent('"url": "http://127.0.0.1:47821/mcp"')
    expect(screen.getByLabelText('Stdio client config')).toHaveTextContent('"args": ["mcp"]')
    await userEvent.click(screen.getByRole('button', { name: 'Copy HTTP config' }))
    expect(aiAccess.copyText).toHaveBeenLastCalledWith(HTTP_CONFIG)
    await userEvent.click(screen.getByRole('button', { name: 'Copy stdio config' }))
    expect(aiAccess.copyText).toHaveBeenLastCalledWith(STDIO_CONFIG)
    await userEvent.click(screen.getByRole('button', { name: 'Copy stdio command' }))
    expect(aiAccess.copyText).toHaveBeenLastCalledWith('/Applications/Xenon.app/Contents/MacOS/xenon mcp')
  })

  it('notes the two accepted residual risks', () => {
    render(<AiAccessSection />)
    const risks = screen.getByRole('list', { name: 'Residual risks' })
    expect(risks).toHaveTextContent('transforms a secret')
    expect(risks).toHaveTextContent('to a destination it controls')
  })

  it('says it is still checking before the first status arrives', () => {
    useSettingsStore.setState({ aiStatus: null })
    render(<AiAccessSection />)
    expect(statusLine()).toHaveTextContent('Checking…')
    expect(screen.getByRole('button', { name: 'Copy HTTP config' })).toBeDisabled()
  })
})
