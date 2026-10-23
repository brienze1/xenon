import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ScriptSet } from '@/platform/mappers/request'
import { injectStylesheets } from '@/test/stylesheets'

vi.mock('@/platform/services/scripts', () => ({ checkScript: vi.fn(), scriptReference: vi.fn(), scriptSnippets: vi.fn() }))

const scriptsService = await import('@/platform/services/scripts')
const { useScriptsStore } = await import('@/features/scripts/store')
const { ScriptsEditor } = await import('./ScriptsEditor')

const snippets = [
  { id: 'get-environment-variable', title: 'Get an environment variable', phase: 'any' as const, text: 'pm.environment.get("variable_key");' },
  { id: 'status-code-200', title: 'Status code: Code is 200', phase: 'postResponse' as const, text: 'pm.test("Status code is 200", function () {});' },
]

function Harness({ initial, inherit, onChange }: { initial: ScriptSet; inherit: boolean | null; onChange?: (scripts: ScriptSet) => void }) {
  const [scripts, setScripts] = useState(initial)
  const [inheritScripts, setInherit] = useState(inherit)
  return (
    <ScriptsEditor
      scripts={scripts}
      onChange={(next) => {
        setScripts(next)
        onChange?.(next)
      }}
      inheritScripts={inheritScripts}
      onInheritChange={setInherit}
      suggestVariables={async () => [{ name: 'base_url', detail: 'dev' }]}
    />
  )
}

describe('ScriptsEditor', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useScriptsStore.setState({ reference: [], snippets: [], status: 'idle' })
    vi.mocked(scriptsService.scriptReference).mockResolvedValue([])
    vi.mocked(scriptsService.scriptSnippets).mockResolvedValue(snippets)
    vi.mocked(scriptsService.checkScript).mockResolvedValue([])
  })

  it('edits the pre-request and post-response scripts in JavaScript editors', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ preRequest: 'console.log("pre")', postResponse: '', disabled: [] }} inherit={true} />)
    expect(await screen.findByLabelText('Pre-request script')).toHaveTextContent('console.log("pre")')
    expect(screen.getByRole('tab', { name: /Pre-request/ })).toContainElement(screen.getByLabelText('has a script'))
    await user.click(screen.getByRole('tab', { name: 'Post-response' }))
    expect(await screen.findByLabelText('Post-response script')).toHaveAttribute('contenteditable', 'true')
    expect(scriptsService.scriptReference).toHaveBeenCalledOnce()
  })

  it('inserts Postman snippets for the phase at the cursor', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<Harness initial={{ preRequest: '', postResponse: '', disabled: [] }} inherit={null} onChange={onChange} />)
    await screen.findByLabelText('Pre-request script')
    await user.click(screen.getByRole('button', { name: 'Snippets for the Pre-request script' }))
    expect(screen.queryByRole('menuitem', { name: 'Status code: Code is 200' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('menuitem', { name: 'Get an environment variable' }))
    expect(onChange).toHaveBeenLastCalledWith({ preRequest: 'pm.environment.get("variable_key");', postResponse: '', disabled: [] })
    await user.click(screen.getByRole('tab', { name: /Post-response/ }))
    await screen.findByLabelText('Post-response script')
    await user.click(screen.getByRole('button', { name: 'Snippets for the Post-response script' }))
    await user.click(screen.getByRole('menuitem', { name: 'Status code: Code is 200' }))
    expect(onChange).toHaveBeenLastCalledWith(expect.objectContaining({ postResponse: 'pm.test("Status code is 200", function () {});' }))
  })

  it('toggles inheritance and has no toggle for collections', async () => {
    const user = userEvent.setup()
    const { unmount } = render(<Harness initial={{ preRequest: '', postResponse: '', disabled: [] }} inherit={true} />)
    const toggle = screen.getByRole('checkbox', { name: 'Inherit parent scripts' })
    expect(toggle).toBeChecked()
    await user.click(toggle)
    expect(toggle).not.toBeChecked()
    unmount()
    render(<Harness initial={{ preRequest: '', postResponse: '', disabled: [] }} inherit={null} />)
    expect(screen.queryByRole('checkbox', { name: 'Inherit parent scripts' })).not.toBeInTheDocument()
  })

  it('lists syntax errors from Go under the editor', async () => {
    vi.mocked(scriptsService.checkScript).mockResolvedValue([{ message: "SyntaxError: Unexpected token ';'", line: 1, column: 11 }])
    render(<Harness initial={{ preRequest: 'const a = ;', postResponse: '', disabled: [] }} inherit={true} />)
    const problems = await screen.findByRole('list', { name: 'Pre-request script problems' })
    expect(problems).toHaveTextContent("Line 1, column 11: SyntaxError: Unexpected token ';'")
    expect(scriptsService.checkScript).toHaveBeenCalledWith('const a = ;')
    expect(await screen.findByTitle("SyntaxError: Unexpected token ';'")).toHaveTextContent(';')
  })

  it('says a disabled script does not run and enables it', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<Harness initial={{ preRequest: 'x()', postResponse: '', disabled: ['preRequest'] }} inherit={true} onChange={onChange} />)
    const notice = await screen.findByText('This script is disabled and does not run.')
    await user.click(within(notice).getByRole('button', { name: 'Enable' }))
    expect(onChange).toHaveBeenLastCalledWith({ preRequest: 'x()', postResponse: '', disabled: [] })
    await waitFor(() => expect(screen.queryByText('This script is disabled and does not run.')).not.toBeInTheDocument())
  })
  it('keeps its rows from overlapping in a short pane', async () => {
    injectStylesheets('src/shared/ui/tabs/Tabs.module.css', 'src/features/scripts/components/ScriptsEditor.module.css')
    const { container } = render(<Harness initial={{ preRequest: '', postResponse: '', disabled: [] }} inherit={true} />)
    await screen.findByLabelText('Pre-request script')
    const header = screen.getByRole('tablist', { name: 'Script phase' }).parentElement!
    const panel = screen.getByRole('tabpanel')
    const toolbar = screen.getByRole('button', { name: 'Snippets for the Pre-request script' }).closest('.toolbar')!
    expect(header).toContainElement(screen.getByRole('checkbox', { name: 'Inherit parent scripts' }))
    expect(getComputedStyle(header).flexShrink).toBe('0')
    expect(getComputedStyle(screen.getByText(/for the API/).closest('p')!).flexShrink).toBe('0')
    expect(getComputedStyle(toolbar).flexShrink).toBe('0')
    expect(getComputedStyle(toolbar).flexWrap).toBe('wrap')
    for (const element of [container.querySelector('.tabs')!, panel, panel.querySelector('.phase')!]) {
      expect(getComputedStyle(element).minHeight).toBe('auto')
    }
    expect(getComputedStyle(panel).overflow).toBe('visible')
    expect(getComputedStyle(panel.querySelector('.editor')!).minHeight).toBe('160px')
  })
})

describe('scripts store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useScriptsStore.setState({ reference: [], snippets: [], status: 'idle' })
  })

  it('loads the reference and snippets once and reports check failures as no errors', async () => {
    vi.mocked(scriptsService.scriptReference).mockResolvedValue([{ path: 'pm', signature: 'pm', description: 'd', kind: 'object' }])
    vi.mocked(scriptsService.scriptSnippets).mockResolvedValue(snippets)
    await useScriptsStore.getState().load()
    await useScriptsStore.getState().load()
    expect(scriptsService.scriptReference).toHaveBeenCalledOnce()
    expect(useScriptsStore.getState()).toMatchObject({ status: 'ready', snippets })
    vi.mocked(scriptsService.checkScript).mockRejectedValue(new Error('down'))
    expect(await useScriptsStore.getState().check('x')).toEqual([])
  })

  it('can load again after a failure', async () => {
    vi.mocked(scriptsService.scriptReference).mockRejectedValue(new Error('down'))
    vi.mocked(scriptsService.scriptSnippets).mockResolvedValue([])
    await useScriptsStore.getState().load()
    expect(useScriptsStore.getState().status).toBe('error')
    vi.mocked(scriptsService.scriptReference).mockResolvedValue([])
    await useScriptsStore.getState().load()
    expect(useScriptsStore.getState().status).toBe('ready')
  })
})
