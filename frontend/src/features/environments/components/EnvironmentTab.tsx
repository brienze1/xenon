import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { ENVIRONMENT_COLORS, type EnvironmentColor, type EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope } from '@/platform/mappers/variable'
import { DUPLICATE_ENVIRONMENT_NAME } from '@/platform/services/environments'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { VariableTable } from './VariableTable'
import styles from './EnvironmentTab.module.css'

const ENVIRONMENT_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

const BROKEN_NEXT_STEP = 'Fix or remove the file in the workspace folder.'

const NAME_REQUIRED = 'Type a name for the environment.'

const COLOR_NAMES: Record<EnvironmentColor, string> = {
  none: 'No colour',
  green: 'Neon green',
  cyan: 'Neon cyan',
  magenta: 'Neon magenta',
  yellow: 'Neon yellow',
  red: 'Neon red',
}

const COLOR_LEGEND = <legend className={styles.legend}>Colour</legend>

const ACTIVE_MARK = <span className={styles.activeMark}>Active</span>

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
)

const MISSING = (
  <EmptyState.Root label="Environment not found" size="compact">
    <EmptyState.Title>Environment Not Found</EmptyState.Title>
    <EmptyState.Description>This environment no longer exists. Close the tab, or pick another environment in the sidebar.</EmptyState.Description>
  </EmptyState.Root>
)

function errorText(error: AppError, name: string): string {
  if (error.code === DUPLICATE_ENVIRONMENT_NAME) {
    return `An environment named “${name.trim()}” already exists. Choose another name.`
  }
  return withNextStep(error.description || error.code, ENVIRONMENT_NEXT_STEP)
}

interface EnvironmentTabProps {
  environmentId: string
}

/** The main-area tab of one environment: its name, colour, "Confirm before send" and variables. */
export function EnvironmentTab({ environmentId }: EnvironmentTabProps) {
  const environment = useEnvironmentsStore((state) => state.environments.find((candidate) => candidate.id === environmentId) ?? null)
  const loaded = useEnvironmentsStore((state) => state.status === 'ready')
  const active = useEnvironmentsStore((state) => state.activeId === environmentId)

  if (environment === null) {
    return loaded ? MISSING : LOADING
  }
  if (environment.error !== null) {
    return (
      <div className={styles.tab}>
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={withNextStep(environment.error, BROKEN_NEXT_STEP)} terms={[environment.error]} />
          </Banner.Message>
        </Banner.Root>
      </div>
    )
  }
  return <EnvironmentEditor key={environment.id} environment={environment} active={active} />
}

interface EnvironmentEditorProps {
  environment: EnvironmentSummary
  active: boolean
}

function EnvironmentEditor({ environment, active }: EnvironmentEditorProps) {
  const renameEnvironment = useEnvironmentsStore((state) => state.renameEnvironment)
  const updateEnvironmentLook = useEnvironmentsStore((state) => state.updateEnvironmentLook)
  const [name, setName] = useState(environment.name)
  const [error, setError] = useState<AppError | null>(null)
  const [nameRequired, setNameRequired] = useState(false)
  const [busy, startAction] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()
  const [syncedName, setSyncedName] = useState(environment.name)

  if (environment.name !== syncedName) {
    setSyncedName(environment.name)
    setName(environment.name)
  }

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

  return (
    <section className={styles.tab} aria-label={`${environment.name} environment`}>
      <header className={styles.header}>
        <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
        <h2 className={styles.title} translate="no" title={environment.name}>
          {environment.name}
        </h2>
        {active ? ACTIVE_MARK : null}
      </header>
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
      {error ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={errorText(error, name)} terms={[error.code]} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <VariableTable scope={environmentScope(environment.id)} label={`${environment.name} variables`} draftLifetime="workspace" savedByTab />
    </section>
  )
}
