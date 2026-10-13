import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  getDocumentation: vi.fn(),
  previewDocumentation: vi.fn(),
  planDocumentation: vi.fn(),
  applyDocumentation: vi.fn(),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { isDocumentationDirty, PREVIEW_DELAY_MS, useDocumentationStore } = await import('./store')
const { editedYaml, emptyPlan, generated, generatedDocument, plan } = await import('./test-fixtures')

function store() {
  return useDocumentationStore.getState()
}

const problem = new AppError('DOC-01400', 'line 4, column 4: mapping values are not allowed in this context', [
  { attribute: 'line', messages: ['4'] },
  { attribute: 'column', messages: ['4'] },
])

describe('documentation store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    vi.mocked(tree.getDocumentation).mockResolvedValue(generated)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('generates the documentation once per collection and starts on the preview', async () => {
    await store().open('c1')
    await store().open('c1')
    expect(tree.getDocumentation).toHaveBeenCalledTimes(1)
    expect(tree.getDocumentation).toHaveBeenCalledWith('ws', 'c1')
    expect(store().entries.c1).toMatchObject({ status: 'ready', draft: generated.yaml, dirty: false, view: 'preview', preview: generatedDocument })
  })

  it('marks edits dirty and reads the edited YAML shortly after typing stops', async () => {
    vi.useFakeTimers()
    const edited = { ...generatedDocument, title: 'Billing v2' }
    vi.mocked(tree.previewDocumentation).mockResolvedValue({ yaml: editedYaml, document: edited })
    await store().open('c1')
    store().edit('c1', editedYaml)
    expect(isDocumentationDirty(store(), 'c1')).toBe(true)
    expect(tree.previewDocumentation).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(PREVIEW_DELAY_MS)
    expect(tree.previewDocumentation).toHaveBeenCalledWith('ws', 'c1', editedYaml)
    expect(store().entries.c1?.preview?.title).toBe('Billing v2')
    store().edit('c1', generated.yaml)
    expect(isDocumentationDirty(store(), 'c1')).toBe(false)
  })

  it('keeps the problem Go found in the edited YAML', async () => {
    vi.useFakeTimers()
    vi.mocked(tree.previewDocumentation).mockRejectedValue(problem)
    await store().open('c1')
    store().edit('c1', 'broken: [')
    await vi.advanceTimersByTimeAsync(PREVIEW_DELAY_MS)
    expect(store().entries.c1?.problem).toBe(problem)
  })

  it('plans a save and writes it only once confirmed', async () => {
    const applied = { ...plan, documentation: { yaml: editedYaml, document: generatedDocument } }
    vi.mocked(tree.planDocumentation).mockResolvedValue(plan)
    vi.mocked(tree.applyDocumentation).mockResolvedValue(applied)
    await store().open('c1')
    store().edit('c1', editedYaml)
    await expect(store().save('c1')).resolves.toEqual({ ok: true, value: false })
    expect(tree.planDocumentation).toHaveBeenCalledWith('ws', 'c1', editedYaml)
    expect(tree.applyDocumentation).not.toHaveBeenCalled()
    expect(store().entries.c1?.plan).toEqual(plan)
    await expect(store().confirm('c1', true)).resolves.toEqual({ ok: true, value: true })
    expect(tree.applyDocumentation).toHaveBeenCalledWith('ws', 'c1', editedYaml, true)
    expect(store().entries.c1).toMatchObject({ plan: null, dirty: false, draft: editedYaml })
    expect(store().saved).toMatchObject({ localId: 'ws', itemId: 'c1', collectionChanged: true, deleted: plan.deleted })
  })

  it('takes a save that changes nothing as saved without asking', async () => {
    vi.mocked(tree.planDocumentation).mockResolvedValue(emptyPlan)
    await store().open('c1')
    store().edit('c1', `${generated.yaml}# a comment\n`)
    await expect(store().save('c1')).resolves.toEqual({ ok: true, value: true })
    expect(store().entries.c1).toMatchObject({ plan: null, dirty: false, draft: generated.yaml })
    expect(tree.applyDocumentation).not.toHaveBeenCalled()
  })

  it('shows why a plan or a save failed and keeps the edits', async () => {
    vi.mocked(tree.planDocumentation).mockRejectedValue(problem)
    await store().open('c1')
    store().edit('c1', 'broken: [')
    const saved = await store().save('c1')
    expect(saved.ok).toBe(false)
    expect(store().entries.c1).toMatchObject({ dirty: true, failure: problem, problem, plan: null })
    vi.mocked(tree.planDocumentation).mockResolvedValue(plan)
    vi.mocked(tree.applyDocumentation).mockRejectedValue(new AppError('DOC-04400', 'confirm the deletions first'))
    await store().save('c1')
    const confirmed = await store().confirm('c1', false)
    expect(confirmed.ok).toBe(false)
    expect(store().entries.c1).toMatchObject({ dirty: true, plan, failure: { code: 'DOC-04400' } })
    store().cancelPlan('c1')
    expect(store().entries.c1?.plan).toBeNull()
  })

  it('regenerates when not edited and marks unsaved YAML stale when the collection changed', async () => {
    await store().open('c1')
    vi.mocked(tree.getDocumentation).mockResolvedValue({ ...generated, yaml: `${generated.yaml}  description: new\n` })
    await store().refreshAll()
    expect(store().entries.c1).toMatchObject({ dirty: false, stale: false, draft: `${generated.yaml}  description: new\n` })
    store().edit('c1', editedYaml)
    vi.mocked(tree.getDocumentation).mockResolvedValue(generated)
    await store().load('c1')
    expect(store().entries.c1).toMatchObject({ dirty: true, stale: true, draft: editedYaml })
    store().keepDraft('c1')
    expect(store().entries.c1?.stale).toBe(false)
    store().discard('c1')
    expect(store().entries.c1).toMatchObject({ dirty: false, draft: generated.yaml })
    store().drop('c1')
    expect(store().entries.c1).toBeUndefined()
  })
})
