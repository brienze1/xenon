import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
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
const { useDocumentationStore } = await import('@/features/documentation/store')
const { editedYaml, generated, generatedDocument, plan } = await import('@/features/documentation/test-fixtures')
const { DocumentationPanel } = await import('./DocumentationPanel')

function store() {
  return useDocumentationStore.getState()
}

describe('DocumentationPanel', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    vi.mocked(tree.getDocumentation).mockResolvedValue(generated)
  })

  it('renders the documentation by folder and opens an operation to show its parameters and examples', async () => {
    const user = userEvent.setup()
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    const doc = await screen.findByRole('article', { name: 'Billing API documentation' })
    expect(within(doc).getByRole('heading', { name: /Billing/, level: 3 })).toBeInTheDocument()
    expect(within(doc).getByText('{{host}}')).toBeInTheDocument()
    const invoices = within(doc).getByRole('region', { name: 'Invoices' })
    expect(within(doc).getByRole('region', { name: 'Requests' })).toHaveTextContent('/ping')
    expect(within(doc).getByRole('region', { name: 'Authentication' })).toHaveTextContent('bearerAuth HTTP bearer')
    await user.click(within(invoices).getByText('/invoices'))
    expect(await within(invoices).findByRole('region', { name: 'List invoices parameters' })).toHaveTextContent('statusqueryInvoice statusopen')
    expect(within(invoices).getByRole('region', { name: 'List invoices responses' })).toHaveTextContent('200 OK')
    expect(within(invoices).getByText('A page')).toBeInTheDocument()
  })

  it('switches between the preview and the YAML editor', async () => {
    const user = userEvent.setup()
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    await screen.findByRole('article')
    await user.click(screen.getByRole('button', { name: 'YAML' }))
    expect(screen.getByRole('button', { name: 'YAML' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('textbox', { name: 'OpenAPI document' })).toHaveTextContent('title: Billing')
    expect(screen.queryByRole('article')).not.toBeInTheDocument()
    expect(store().entries.c1?.view).toBe('yaml')
    await user.click(screen.getByRole('button', { name: 'Preview' }))
    expect(screen.getByRole('article')).toBeInTheDocument()
  })

  it('shows where the edited YAML cannot be read and marks it in the editor', async () => {
    const user = userEvent.setup()
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    await screen.findByRole('article')
    store().edit('c1', 'broken')
    useDocumentationStore.setState((state) => ({
      entries: {
        ...state.entries,
        c1: {
          ...state.entries.c1!,
          problem: new AppError('DOC-01400', 'line 4, column 4: mapping values are not allowed in this context', [
            { attribute: 'line', messages: ['4'] },
            { attribute: 'column', messages: ['4'] },
          ]),
        },
      },
    }))
    expect(await screen.findByRole('alert')).toHaveTextContent('The YAML cannot be read at line 4, column 4: mapping values are not allowed in this context')
    await user.click(screen.getByRole('button', { name: 'Show YAML' }))
    expect(screen.getByRole('list', { name: 'OpenAPI document problems' })).toHaveTextContent('Line 4, column 4: mapping values are not allowed in this context')
    await user.click(screen.getByRole('button', { name: 'Discard Changes' }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(store().entries.c1?.dirty).toBe(false)
  })

  it('lists what saving changes and needs the deletions ticked before writing them', async () => {
    const user = userEvent.setup()
    vi.mocked(tree.planDocumentation).mockResolvedValue(plan)
    vi.mocked(tree.applyDocumentation).mockResolvedValue({ ...plan, documentation: { yaml: editedYaml, document: generatedDocument } })
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    await screen.findByRole('article')
    store().edit('c1', editedYaml)
    await store().save('c1')
    const dialog = await screen.findByRole('dialog', { name: 'Save Documentation Changes?' })
    expect(dialog).toHaveTextContent('1 updated, 1 created, 1 deleted.')
    expect(within(dialog).getByRole('region', { name: 'Updated' })).toHaveTextContent('List invoices')
    expect(within(dialog).getByRole('region', { name: 'Created' })).toHaveTextContent('Archive invoicein Invoices / Archive')
    expect(within(dialog).getByRole('region', { name: 'Deleted' })).toHaveTextContent('Ping')
    const save = within(dialog).getByRole('button', { name: 'Save Changes' })
    expect(save).toBeDisabled()
    await user.click(within(dialog).getByRole('checkbox', { name: 'Delete this request from disk' }))
    await user.click(save)
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(tree.applyDocumentation).toHaveBeenCalledWith('ws', 'c1', editedYaml, true)
    expect(store().entries.c1?.dirty).toBe(false)
  })

  it('cancels a planned save without writing anything', async () => {
    const user = userEvent.setup()
    vi.mocked(tree.planDocumentation).mockResolvedValue({ ...plan, deleted: [] })
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    await screen.findByRole('article')
    store().edit('c1', editedYaml)
    await store().save('c1')
    const dialog = await screen.findByRole('dialog')
    await user.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(tree.applyDocumentation).not.toHaveBeenCalled()
    expect(store().entries.c1?.dirty).toBe(true)
  })

  it('offers to try again when the documentation cannot be generated', async () => {
    const user = userEvent.setup()
    vi.mocked(tree.getDocumentation).mockRejectedValueOnce(new AppError('DOC-05400', 'only collections have documentation'))
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    expect(await screen.findByRole('alert')).toHaveTextContent('only collections have documentation')
    await user.click(screen.getByRole('button', { name: 'Try Again' }))
    expect(await screen.findByRole('article')).toBeInTheDocument()
  })

  it('shows x-tagGroups as sections holding their folders, each folder with its description', async () => {
    vi.mocked(tree.getDocumentation).mockResolvedValue({
      ...generated,
      document: {
        ...generatedDocument,
        tags: [{ name: 'Invoices', description: 'Everything about **invoices**.' }],
        tagGroups: [{ name: 'Money', tags: ['Invoices'] }],
      },
    })
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    const doc = await screen.findByRole('article', { name: 'Billing API documentation' })
    const money = within(doc).getByRole('region', { name: 'Money' })
    expect(within(money).getByRole('heading', { name: 'Money', level: 4 })).toBeInTheDocument()
    const invoices = within(money).getByRole('region', { name: 'Invoices' })
    expect(within(invoices).getByRole('heading', { name: 'Invoices', level: 5 })).toBeInTheDocument()
    expect(within(invoices).getByText('/invoices')).toBeInTheDocument()
    expect(await within(invoices).findByText('invoices')).toBeInTheDocument()
    const requests = within(doc).getByRole('region', { name: 'Requests' })
    expect(requests).toHaveTextContent('/ping')
    expect(within(requests).queryByText('/invoices')).not.toBeInTheDocument()
  })

  it('puts a top-level folder\'s own requests under its section heading once, and names sub-folders within it', async () => {
    const [invoice, ping] = generatedDocument.operations
    vi.mocked(tree.getDocumentation).mockResolvedValue({
      ...generated,
      document: {
        ...generatedDocument,
        tags: [
          { name: 'Ads', description: 'Everything about **ads**.' },
          { name: 'Ads / CRUD', description: '' },
        ],
        tagGroups: [{ name: 'Ads', tags: ['Ads', 'Ads / CRUD'] }],
        operations: [
          { ...invoice!, tag: 'Ads', path: '/ads' },
          { ...ping!, requestId: 'r3', tag: 'Ads / CRUD', path: '/ads/1' },
        ],
      },
    })
    render(<DocumentationPanel itemId="c1" kind="collection" />)
    const doc = await screen.findByRole('article', { name: 'Billing API documentation' })
    expect(within(doc).getAllByRole('heading', { name: 'Ads' })).toHaveLength(1)
    const ads = within(doc).getByRole('region', { name: 'Ads' })
    expect(ads).toHaveTextContent('/ads')
    expect(await within(ads).findByText('ads')).toBeInTheDocument()
    const crud = within(ads).getByRole('region', { name: 'Ads / CRUD' })
    expect(within(crud).getByRole('heading', { name: 'CRUD', level: 5 })).toBeInTheDocument()
    expect(crud).toHaveTextContent('/ads/1')
  })

  it('speaks of the folder in a folder\'s documentation and its save plan', async () => {
    vi.mocked(tree.getDocumentation).mockResolvedValue({ ...generated, document: { ...generatedDocument, operations: [], tags: [], tagGroups: [] } })
    vi.mocked(tree.planDocumentation).mockResolvedValue({ ...plan, created: [{ ...plan.created[0]!, folder: '' }], deleted: [] })
    render(<DocumentationPanel itemId="f1" kind="folder" />)
    expect(await screen.findByText(/This folder has no requests yet/)).toBeInTheDocument()
    expect(screen.getByText(/Built from the requests of this folder/)).toBeInTheDocument()
    store().edit('f1', editedYaml)
    await store().save('f1')
    const dialog = await screen.findByRole('dialog', { name: 'Save Documentation Changes?' })
    expect(dialog).toHaveTextContent('Archive invoicein the folder')
    expect(dialog).toHaveTextContent('The folder’s name or description changes too.')
  })
})
