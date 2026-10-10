import { useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope } from '@/platform/mappers/variable'
import { DUPLICATE_ENVIRONMENT_NAME } from '@/platform/services/environments'
import { cn } from '@/shared/lib/cn'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { VariableTable } from './VariableTable'
import styles from './EnvironmentsManager.module.css'

const ENVIRONMENT_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

function errorText(error: AppError, name: string): string {
  if (error.code === DUPLICATE_ENVIRONMENT_NAME) {
    return `An environment named “${name.trim()}” already exists. Choose another name.`
  }
  return withNextStep(error.description || error.code, ENVIRONMENT_NEXT_STEP)
}

interface ErrorBannerProps {
  error: AppError
  name: string
}

function ErrorBanner({ error, name }: ErrorBannerProps) {
  return (
    <Banner.Root tone="error">
      <Banner.Message>
        <NoTranslateText text={errorText(error, name)} terms={[error.code]} />
      </Banner.Message>
    </Banner.Root>
  )
}

const NO_ENVIRONMENTS = <p className={styles.note}>No environments yet. Create one to group variables such as a base URL per stage.</p>

export function EnvironmentsManager() {
  const open = useEnvironmentsStore((state) => state.editor === 'manager')
  const closeEditor = useEnvironmentsStore((state) => state.closeEditor)
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : closeEditor())} className={styles.dialog}>
      {open ? <ManagerContent onClose={closeEditor} /> : null}
    </Dialog.Root>
  )
}

interface ManagerContentProps {
  onClose: () => void
}

function ManagerContent({ onClose }: ManagerContentProps) {
  const environments = useEnvironmentsStore((state) => state.environments)
  const createEnvironment = useEnvironmentsStore((state) => state.createEnvironment)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [createError, setCreateError] = useState<AppError | null>(null)
  const [creating, startCreate] = useTransition()
  const usable = environments.filter((environment) => environment.error === null)
  const selected = usable.find((environment) => environment.id === selectedId) ?? usable[0] ?? null

  function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = newName
    startCreate(async () => {
      const created = await createEnvironment(name)
      if (!created.ok) {
        setCreateError(created.error)
        return
      }
      setCreateError(null)
      setNewName('')
      setSelectedId(created.value.id)
    })
  }

  return (
    <>
      <Dialog.Title>Environments</Dialog.Title>
      <Dialog.Body>
        <div className={styles.layout}>
          <div className={styles.sidebar}>
            <ul aria-label="Environments" className={styles.list}>
              {environments.map((environment) => (
                <li key={environment.id}>
                  <EnvironmentItem
                    environment={environment}
                    selected={environment.id === selected?.id}
                    onSelect={() => setSelectedId(environment.id)}
                  />
                </li>
              ))}
            </ul>
            {environments.length === 0 ? NO_ENVIRONMENTS : null}
            <form className={styles.create} onSubmit={create}>
              <input
                name="environment-name"
                autoComplete="off"
                aria-label="New environment name"
                placeholder="New environment…"
                className={styles.input}
                value={newName}
                onChange={(event) => setNewName(event.target.value)}
              />
              <Button type="submit" variant="secondary" size="sm" disabled={creating || newName.trim() === ''}>
                Create
              </Button>
            </form>
            {createError ? <ErrorBanner error={createError} name={newName} /> : null}
          </div>
          <div className={styles.detail}>
            {selected ? <EnvironmentDetail key={selected.id} environment={selected} onDuplicated={setSelectedId} /> : null}
          </div>
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Close
        </Button>
      </Dialog.Actions>
    </>
  )
}

interface EnvironmentItemProps {
  environment: EnvironmentSummary
  selected: boolean
  onSelect: () => void
}

function EnvironmentItem({ environment, selected, onSelect }: EnvironmentItemProps) {
  if (environment.error !== null) {
    return (
      <span className={cn(styles.item, styles.broken)} title={environment.error}>
        <span translate="no">{environment.name}</span>
        <span className={styles.brokenMark}>{environment.error}</span>
      </span>
    )
  }
  return (
    <button type="button" aria-current={selected ? 'true' : undefined} className={styles.item} onClick={onSelect}>
      <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
      <span className={styles.itemName} translate="no">
        {environment.name}
      </span>
    </button>
  )
}

interface EnvironmentDetailProps {
  environment: EnvironmentSummary
  onDuplicated: (id: string) => void
}

function EnvironmentDetail({ environment, onDuplicated }: EnvironmentDetailProps) {
  const renameEnvironment = useEnvironmentsStore((state) => state.renameEnvironment)
  const duplicateEnvironment = useEnvironmentsStore((state) => state.duplicateEnvironment)
  const deleteEnvironment = useEnvironmentsStore((state) => state.deleteEnvironment)
  const [name, setName] = useState(environment.name)
  const [error, setError] = useState<AppError | null>(null)
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [busy, startAction] = useTransition()

  function rename(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    startAction(async () => setError(await renameEnvironment(environment.id, name)))
  }

  function duplicate() {
    startAction(async () => {
      const duplicated = await duplicateEnvironment(environment.id)
      if (duplicated.ok) {
        onDuplicated(duplicated.value.id)
      } else {
        setError(duplicated.error)
      }
    })
  }

  function remove() {
    startAction(async () => {
      const failure = await deleteEnvironment(environment.id)
      setError(failure)
      setConfirmingDelete(false)
    })
  }

  return (
    <div className={styles.detailBody}>
      <div className={styles.detailBar}>
        <form className={styles.rename} onSubmit={rename}>
          <input
            name="name"
            autoComplete="off"
            aria-label="Environment name"
            className={styles.input}
            translate="no"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
          <Button type="submit" variant="secondary" size="sm" disabled={busy || name.trim() === '' || name === environment.name}>
            Rename
          </Button>
        </form>
        <Button variant="ghost" size="sm" disabled={busy} onClick={duplicate}>
          Duplicate
        </Button>
        <Button variant="ghost" size="sm" disabled={busy} onClick={() => setConfirmingDelete(true)}>
          Delete
        </Button>
      </div>
      {confirmingDelete ? (
        <Banner.Root tone="warning">
          <Banner.Message>
            Delete “<span translate="no">{environment.name}</span>”? Its file, local current values and secret values are removed.
          </Banner.Message>
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => setConfirmingDelete(false)}>
              Cancel
            </Button>
            <Button variant="danger" size="sm" disabled={busy} onClick={remove}>
              Delete Environment
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {error ? <ErrorBanner error={error} name={name} /> : null}
      <VariableTable scope={environmentScope(environment.id)} label={`${environment.name} variables`} />
    </div>
  )
}
