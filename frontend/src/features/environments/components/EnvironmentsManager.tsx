import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { ENVIRONMENT_COLORS, type EnvironmentColor, type EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope } from '@/platform/mappers/variable'
import { DUPLICATE_ENVIRONMENT_NAME } from '@/platform/services/environments'
import { cn } from '@/shared/lib/cn'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useDiscardGuard } from '@/features/environments/hooks/use-discard-guard'
import { useEnvironmentsStore } from '@/features/environments/store'
import { DiscardChangesConfirm } from './DiscardChangesConfirm'
import { VariableTable } from './VariableTable'
import styles from './EnvironmentsManager.module.css'

const ENVIRONMENT_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

const NAME_REQUIRED = 'Type a name for the environment.'

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

const COLOR_NAMES: Record<EnvironmentColor, string> = {
  none: 'No colour',
  green: 'Neon green',
  cyan: 'Neon cyan',
  magenta: 'Neon magenta',
  yellow: 'Neon yellow',
  red: 'Neon red',
}

const COLOR_LEGEND = <legend className={styles.legend}>Colour</legend>

const NO_ENVIRONMENTS = <p className={styles.note}>No environments yet. Create one to group variables such as a base URL per stage.</p>

export function EnvironmentsManager() {
  const open = useEnvironmentsStore((state) => state.editor === 'manager')
  const closeEditor = useEnvironmentsStore((state) => state.closeEditor)
  return open ? <ManagerContent onClose={closeEditor} /> : null
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
  const [nameRequired, setNameRequired] = useState(false)
  const [creating, startCreate] = useTransition()
  const newNameRef = useRef<HTMLInputElement>(null)
  const newNameErrorId = useId()
  const usable = environments.filter((environment) => environment.error === null)
  const selected = usable.find((environment) => environment.id === selectedId) ?? usable[0] ?? null
  const { confirming, guard, discard, keepEditing } = useDiscardGuard(selected ? [environmentScope(selected.id)] : [])
  const close = () => guard(onClose)

  function select(id: string) {
    if (id !== selected?.id) {
      guard(() => setSelectedId(id))
    }
  }

  function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = newName
    if (name.trim() === '') {
      setNameRequired(true)
      newNameRef.current?.focus()
      return
    }
    startCreate(async () => {
      const created = await createEnvironment(name)
      if (!created.ok) {
        setCreateError(created.error)
        return
      }
      setCreateError(null)
      setNewName('')
      select(created.value.id)
    })
  }

  return (
    <Dialog.Root open onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
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
                    onSelect={() => select(environment.id)}
                  />
                </li>
              ))}
            </ul>
            {environments.length === 0 ? NO_ENVIRONMENTS : null}
            <form className={styles.create} noValidate onSubmit={create}>
              <input
                ref={newNameRef}
                name="environment-name"
                autoComplete="off"
                aria-label="New environment name"
                aria-invalid={nameRequired ? true : undefined}
                aria-describedby={nameRequired ? newNameErrorId : undefined}
                placeholder="New environment…"
                className={styles.input}
                value={newName}
                onChange={(event) => {
                  setNewName(event.target.value)
                  setNameRequired(false)
                }}
              />
              <Button type="submit" variant="secondary" size="sm" disabled={creating}>
                Create
              </Button>
            </form>
            {nameRequired ? (
              <p id={newNameErrorId} role="alert" className={styles.fieldError}>
                {NAME_REQUIRED}
              </p>
            ) : null}
            {createError ? <ErrorBanner error={createError} name={newName} /> : null}
          </div>
          <div className={styles.detail}>
            {selected ? <EnvironmentDetail key={selected.id} environment={selected} onDuplicated={select} /> : null}
          </div>
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        {confirming ? (
          <DiscardChangesConfirm onKeepEditing={keepEditing} onDiscard={discard} />
        ) : (
          <Button variant="secondary" onClick={close}>
            Close
          </Button>
        )}
      </Dialog.Actions>
    </Dialog.Root>
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
  const updateEnvironmentLook = useEnvironmentsStore((state) => state.updateEnvironmentLook)
  const duplicateEnvironment = useEnvironmentsStore((state) => state.duplicateEnvironment)
  const deleteEnvironment = useEnvironmentsStore((state) => state.deleteEnvironment)
  const [name, setName] = useState(environment.name)
  const [error, setError] = useState<AppError | null>(null)
  const [nameRequired, setNameRequired] = useState(false)
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [busy, startAction] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()

  function rename(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (name.trim() === '') {
      setNameRequired(true)
      nameRef.current?.focus()
      return
    }
    if (name === environment.name) {
      return
    }
    startAction(async () => setError(await renameEnvironment(environment.id, name)))
  }

  function changeLook(color: EnvironmentColor, confirmBeforeSend: boolean) {
    startAction(async () => setError(await updateEnvironmentLook(environment.id, { color, confirmBeforeSend })))
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
        <form className={styles.rename} noValidate onSubmit={rename}>
          <input
            ref={nameRef}
            name="name"
            autoComplete="off"
            aria-label="Environment name"
            aria-invalid={nameRequired ? true : undefined}
            aria-describedby={nameRequired ? nameErrorId : undefined}
            className={styles.input}
            translate="no"
            value={name}
            onChange={(event) => {
              setName(event.target.value)
              setNameRequired(false)
            }}
          />
          <Button type="submit" variant="secondary" size="sm" disabled={busy}>
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
      {nameRequired ? (
        <p id={nameErrorId} role="alert" className={styles.fieldError}>
          {NAME_REQUIRED}
        </p>
      ) : null}
      <div className={styles.look}>
        <fieldset className={styles.colors}>
          {COLOR_LEGEND}
          {ENVIRONMENT_COLORS.map((color) => (
            <label key={color} className={styles.colorOption} title={COLOR_NAMES[color]}>
              <input
                type="radio"
                name={`${environment.id}-color`}
                className="visually-hidden"
                checked={environment.color === color}
                disabled={busy}
                aria-label={COLOR_NAMES[color]}
                onChange={() => changeLook(color, environment.confirmBeforeSend)}
              />
              <span className={styles.colorSwatch} data-color={color} aria-hidden="true" />
            </label>
          ))}
        </fieldset>
        <label className={styles.confirm}>
          <input
            type="checkbox"
            name="confirm-before-send"
            checked={environment.confirmBeforeSend}
            disabled={busy}
            onChange={(event) => changeLook(environment.color, event.target.checked)}
          />
          Confirm before send
        </label>
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
