import { useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import { ENVIRONMENT_COLORS, type EnvironmentColor, type EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope } from '@/platform/mappers/variable'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { EditableName } from '@/shared/ui/editable-name/EditableName'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { selectEnvironmentTabId, selectNameErrorText, useTabsStore } from '@/features/tabs/store'
import { VariableTable } from './VariableTable'
import styles from './EnvironmentTab.module.css'

const ENVIRONMENT_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

const BROKEN_NEXT_STEP = 'Fix or remove the file in the workspace folder.'

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

function errorText(error: AppError): string {
  return withNextStep(error.description || error.code, ENVIRONMENT_NEXT_STEP)
}

interface EnvironmentTabProps {
  environmentId: string
}

/** The main-area tab of one environment: its name (edited in place, saved with the tab), colour, "Confirm before send" and variables. */
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
  const updateEnvironmentLook = useEnvironmentsStore((state) => state.updateEnvironmentLook)
  const name = useEnvironmentsStore((state) => state.nameDrafts[environment.id] ?? environment.name)
  const tabId = useTabsStore((state) => selectEnvironmentTabId(state, environment.id))
  const nameError = useTabsStore((state) => selectNameErrorText(state, tabId))
  const editTabName = useTabsStore((state) => state.editTabName)
  const cancelTabName = useTabsStore((state) => state.cancelTabName)
  const [error, setError] = useState<AppError | null>(null)
  const [busy, startAction] = useTransition()

  function changeLook(color: EnvironmentColor, confirmBeforeSend: boolean) {
    startAction(async () => setError(await updateEnvironmentLook(environment.id, { color, confirmBeforeSend })))
  }

  return (
    <section className={styles.tab} aria-label={`${environment.name} environment`}>
      <header className={styles.header}>
        <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
        <h2 className={styles.title}>
          <EditableName
            value={name}
            label="Rename environment"
            error={nameError}
            onChange={(next) => tabId !== null && editTabName(tabId, next)}
            onCancel={() => tabId !== null && cancelTabName(tabId)}
          />
        </h2>
        {active ? ACTIVE_MARK : null}
      </header>
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
            <NoTranslateText text={errorText(error)} terms={[error.code]} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <VariableTable scope={environmentScope(environment.id)} label={`${environment.name} variables`} />
    </section>
  )
}
