import { useState, useTransition, type KeyboardEvent, type ReactNode } from 'react'
import { AppError } from '@/platform/errors'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { DUPLICATE_ENVIRONMENT_NAME } from '@/platform/services/environments'
import { cn } from '@/shared/lib/cn'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useImportStore } from '@/features/import-export/store'
import { useExportStore } from '@/features/export/store'
import type { OpenMode, VariablesItem } from '@/features/sidebar/model'
import { RenameInput } from './RenameInput'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './EnvironmentsView.module.css'

export const NEW_ENVIRONMENT_NAME = 'New Environment'

const ACTION_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

const BROKEN_NEXT_STEP = 'Fix or remove the file in the workspace folder.'

const ADD_GLYPH = <span aria-hidden="true">+</span>

const IMPORT_GLYPH = <span aria-hidden="true">⤓</span>

const ACTIVE_MARK = <span className={styles.activeMark}>Active</span>

const NO_ENVIRONMENTS = <p className={styles.note}>No environments yet. Create one to group variables such as a base URL per stage.</p>

const GLOBALS: VariablesItem = { kind: 'globals' }

function environmentItem(id: string): VariablesItem {
  return { kind: 'environment', environmentId: id }
}

function matches(name: string, query: string): boolean {
  return name.toLowerCase().includes(query.trim().toLowerCase())
}

function isSelected(selected: VariablesItem | null, item: VariablesItem): boolean {
  if (selected === null || selected.kind !== item.kind) {
    return false
  }
  return selected.kind === 'globals' || (item.kind === 'environment' && selected.environmentId === item.environmentId)
}

/** A duplicate name reads better as a sentence than as the backend's error. */
function friendlyError(error: AppError, name: string): AppError {
  if (error.code === DUPLICATE_ENVIRONMENT_NAME) {
    return new AppError(error.code, `An environment named “${name.trim()}” already exists. Choose another name.`)
  }
  return error
}

interface EnvironmentsViewProps {
  /** The entry whose tab is active, highlighted in the list. */
  selected: VariablesItem | null
  onOpen: (item: VariablesItem, mode: OpenMode) => void
}

/** The sidebar's Environments view: the globals and every environment, opened as main-area tabs. */
export function EnvironmentsView({ selected, onOpen }: EnvironmentsViewProps) {
  const environments = useEnvironmentsStore((state) => state.environments)
  const status = useEnvironmentsStore((state) => state.status)
  const createEnvironment = useEnvironmentsStore((state) => state.createEnvironment)
  const openImport = useImportStore((state) => state.open)
  const [query, setQuery] = useState('')
  const [creating, setCreating] = useState(false)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<EnvironmentSummary | null>(null)
  const [error, setError] = useState<AppError | null>(null)

  async function create(name: string): Promise<AppError | null> {
    if (name.trim() === '') {
      setCreating(false)
      return null
    }
    const created = await createEnvironment(name)
    if (!created.ok) {
      return friendlyError(created.error, name)
    }
    setCreating(false)
    onOpen(environmentItem(created.value.id), 'pinned')
    return null
  }

  const searching = query.trim() !== ''
  const shown = searching ? environments.filter((environment) => matches(environment.name, query)) : environments
  const showGlobals = !searching || matches('Globals', query)

  return (
    <div className={styles.view}>
      <div className={styles.toolbar}>
        <IconButton label="New Environment" className={styles.add} onClick={() => setCreating(true)}>
          {ADD_GLYPH}
        </IconButton>
        <IconButton label="Import" className={styles.add} onClick={openImport}>
          {IMPORT_GLYPH}
        </IconButton>
        <input
          type="search"
          name="environment-search"
          aria-label="Search environments"
          placeholder="Search environments…"
          autoComplete="off"
          className={styles.search}
          spellCheck={false}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
      </div>
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>
            <SidebarErrorText error={error} nextStep={ACTION_NEXT_STEP} />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={() => setError(null)}>
              Dismiss Error
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <div className={styles.scroll}>
        <ul aria-label="Environments" className={styles.list}>
          {showGlobals ? (
            <li>
            <ItemButton item={GLOBALS} label="Globals" selected={isSelected(selected, GLOBALS)} onOpen={onOpen}>
              <span className={styles.globalsMark} aria-hidden="true">
                ◎
              </span>
              <span className={styles.name}>Globals</span>
            </ItemButton>
            </li>
          ) : null}
          {creating ? (
            <li className={styles.editing}>
              <RenameInput name={NEW_ENVIRONMENT_NAME} onCommit={create} onCancel={() => setCreating(false)} />
            </li>
          ) : null}
          {shown.map((environment) => (
            <li key={environment.id}>
              {environment.error !== null ? (
                <BrokenItem environment={environment} />
              ) : renamingId === environment.id ? (
                <RenameItem environment={environment} onDone={() => setRenamingId(null)} />
              ) : (
                <EnvironmentItem
                  environment={environment}
                  selected={isSelected(selected, environmentItem(environment.id))}
                  onOpen={onOpen}
                  onRename={() => setRenamingId(environment.id)}
                  onDelete={() => setDeleting(environment)}
                  onError={setError}
                />
              )}
            </li>
          ))}
        </ul>
        {status === 'ready' && environments.length === 0 && !creating ? NO_ENVIRONMENTS : null}
        {searching && environments.length > 0 && shown.length === 0 && !showGlobals ? (
          <p className={styles.note}>No environments match “{query.trim()}”.</p>
        ) : null}
      </div>
      <DeleteEnvironmentDialog environment={deleting} onClose={() => setDeleting(null)} />
    </div>
  )
}

interface ItemButtonProps {
  item: VariablesItem
  /** The accessible name; the visible text can carry marks such as "Active". */
  label: string
  selected: boolean
  onOpen: (item: VariablesItem, mode: OpenMode) => void
  children: ReactNode
}

function ItemButton({ item, label, selected, onOpen, children }: ItemButtonProps) {
  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
      event.preventDefault()
      onOpen(item, 'pinned')
    }
  }
  return (
    <button
      type="button"
      aria-label={label}
      aria-current={selected ? 'true' : undefined}
      className={cn(styles.item, selected && styles.selected)}
      onClick={() => onOpen(item, 'preview')}
      onDoubleClick={() => onOpen(item, 'pinned')}
      onKeyDown={onKeyDown}
    >
      {children}
    </button>
  )
}

interface EnvironmentItemProps {
  environment: EnvironmentSummary
  selected: boolean
  onOpen: (item: VariablesItem, mode: OpenMode) => void
  onRename: () => void
  onDelete: () => void
  onError: (error: AppError | null) => void
}

function EnvironmentItem({ environment, selected, onOpen, onRename, onDelete, onError }: EnvironmentItemProps) {
  const active = useEnvironmentsStore((state) => state.activeId === environment.id)
  const setActive = useEnvironmentsStore((state) => state.setActive)
  const duplicateEnvironment = useEnvironmentsStore((state) => state.duplicateEnvironment)
  const item = environmentItem(environment.id)

  async function duplicate() {
    const duplicated = await duplicateEnvironment(environment.id)
    if (!duplicated.ok) {
      onError(duplicated.error)
      return
    }
    onError(null)
    onOpen(environmentItem(duplicated.value.id), 'preview')
  }

  return (
    <ContextMenu.Root>
      <ContextMenu.Trigger>
        <ItemButton item={item} label={active ? `${environment.name}, active` : environment.name} selected={selected} onOpen={onOpen}>
          <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
          <span className={styles.name} translate="no" title={environment.name}>
            {environment.name}
          </span>
          {active ? ACTIVE_MARK : null}
        </ItemButton>
      </ContextMenu.Trigger>
      <ContextMenu.Menu label={`${environment.name} actions`}>
        <ContextMenu.Item onSelect={() => onOpen(item, 'pinned')}>Open</ContextMenu.Item>
        <ContextMenu.Item disabled={active} onSelect={() => void setActive(environment.id)}>
          Set as Active
        </ContextMenu.Item>
        <ContextMenu.Separator />
        <ContextMenu.Item onSelect={onRename}>Rename</ContextMenu.Item>
        <ContextMenu.Item onSelect={() => void duplicate()}>Duplicate</ContextMenu.Item>
        <ContextMenu.Item onSelect={() => void useExportStore.getState().openItem('environment', environment.id, environment.name)}>Export…</ContextMenu.Item>
        <ContextMenu.Separator />
        <ContextMenu.Item tone="danger" onSelect={onDelete}>
          Delete
        </ContextMenu.Item>
      </ContextMenu.Menu>
    </ContextMenu.Root>
  )
}

interface RenameItemProps {
  environment: EnvironmentSummary
  onDone: () => void
}

function RenameItem({ environment, onDone }: RenameItemProps) {
  const renameEnvironment = useEnvironmentsStore((state) => state.renameEnvironment)

  async function commit(name: string): Promise<AppError | null> {
    if (name.trim() === '' || name === environment.name) {
      onDone()
      return null
    }
    const failure = await renameEnvironment(environment.id, name)
    if (failure !== null) {
      return friendlyError(failure, name)
    }
    onDone()
    return null
  }

  return (
    <div className={styles.editing}>
      <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
      <RenameInput name={environment.name} onCommit={commit} onCancel={onDone} />
    </div>
  )
}

interface BrokenItemProps {
  environment: EnvironmentSummary
}

function BrokenItem({ environment }: BrokenItemProps) {
  const message = withNextStep(environment.error ?? '', BROKEN_NEXT_STEP)
  return (
    <span className={cn(styles.item, styles.broken)} title={message}>
      <span translate="no">{environment.name}</span>
      <span className={styles.brokenMark}>
        <NoTranslateText text={message} terms={[environment.error ?? '']} />
      </span>
    </span>
  )
}

interface DeleteEnvironmentDialogProps {
  environment: EnvironmentSummary | null
  onClose: () => void
}

function DeleteEnvironmentDialog({ environment, onClose }: DeleteEnvironmentDialogProps) {
  return (
    <Dialog.Root open={environment !== null} onOpenChange={(open) => (open ? undefined : onClose())}>
      {environment ? <DeleteEnvironmentConfirm key={environment.id} environment={environment} onClose={onClose} /> : null}
    </Dialog.Root>
  )
}

interface DeleteEnvironmentConfirmProps {
  environment: EnvironmentSummary
  onClose: () => void
}

function DeleteEnvironmentConfirm({ environment, onClose }: DeleteEnvironmentConfirmProps) {
  const deleteEnvironment = useEnvironmentsStore((state) => state.deleteEnvironment)
  const [deleting, startDelete] = useTransition()
  const [error, setError] = useState<AppError | null>(null)

  function confirm() {
    startDelete(async () => {
      const failure = await deleteEnvironment(environment.id)
      setError(failure)
      if (failure === null) {
        onClose()
      }
    })
  }

  return (
    <>
      <Dialog.Title>
        Delete “<span translate="no">{environment.name}</span>”?
      </Dialog.Title>
      <Dialog.Body>
        Its file, local current values and secret values are removed.
        {error ? (
          <Banner.Root tone="error" className={styles.dialogError}>
            <Banner.Message>
              <SidebarErrorText error={error} nextStep={ACTION_NEXT_STEP} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="danger" disabled={deleting} onClick={confirm}>
          {deleting ? 'Deleting…' : 'Delete Environment'}
        </Button>
      </Dialog.Actions>
    </>
  )
}
