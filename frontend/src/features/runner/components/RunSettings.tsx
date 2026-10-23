import type { FunctionalSettings, RunType } from '@/platform/mappers/runner'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { selectedCount } from '@/features/runner/lib/run-order'
import { useRunnerStore } from '@/features/runner/store'
import { DataFileField } from './DataFileField'
import { EnvironmentSelect } from './EnvironmentSelect'
import { NumberSetting } from './NumberSetting'
import { PerformanceSettings } from './PerformanceSettings'
import fields from './SettingFields.module.css'
import styles from './RunSettings.module.css'

const MAX_DELAY_MS = 3_600_000

type FunctionalToggle = 'persistResponses' | 'disableLogs' | 'stopOnError' | 'keepVariableValues' | 'skipStoredCookies' | 'saveCookies'

/** Postman's run toggles, in Postman's order and with Postman's labels. */
const TOGGLES: ReadonlyArray<{ key: FunctionalToggle; label: string }> = [
  { key: 'persistResponses', label: 'Persist responses for a session' },
  { key: 'disableLogs', label: 'Turn off logs during run' },
  { key: 'stopOnError', label: 'Stop run if an error occurs' },
  { key: 'keepVariableValues', label: 'Keep variable values' },
  { key: 'skipStoredCookies', label: 'Run collection without using stored cookies' },
  { key: 'saveCookies', label: 'Save cookies after collection run' },
]

const RUN_TYPES: ReadonlyArray<{ type: RunType; label: string }> = [
  { type: 'functional', label: 'Functional' },
  { type: 'performance', label: 'Performance' },
]

const DATA_FILE_MISSING = (
  <Banner.Message>The remembered iteration data file no longer exists, so it was cleared. Select another file if this run needs one.</Banner.Message>
)

interface RunSettingsProps {
  itemId: string
}

/** The configuration column of the Setup sub-tab: run type, its settings, the environment and Start run. */
export function RunSettings({ itemId }: RunSettingsProps) {
  const config = useRunnerStore((state) => state.targets[itemId]?.config ?? null)
  const fieldErrors = useRunnerStore((state) => state.targets[itemId]?.fieldErrors)
  const dataFileMissing = useRunnerStore((state) => state.targets[itemId]?.dataFileMissing === true)
  const selected = useRunnerStore((state) => selectedCount(state.targets[itemId]?.rows ?? []))
  const running = useRunnerStore((state) => state.active[itemId] !== undefined)
  const starting = useRunnerStore((state) => state.targets[itemId]?.starting === true)
  const updateConfig = useRunnerStore((state) => state.updateConfig)
  const dismissDataFileMissing = useRunnerStore((state) => state.dismissDataFileMissing)
  const start = useRunnerStore((state) => state.start)
  if (config === null) {
    return null
  }
  const functional = config.functional

  function editFunctional(patch: Partial<FunctionalSettings>) {
    updateConfig(itemId, (current) => ({ ...current, functional: { ...current.functional, ...patch } }))
  }

  return (
    <section className={styles.settings} aria-label="Run configuration">
      <fieldset className={styles.types}>
        <legend className="visually-hidden">Run type</legend>
        {RUN_TYPES.map((option) => (
          <label key={option.type} className={styles.type}>
            <input
              type="radio"
              name={`${itemId}-run-type`}
              className={styles.typeInput}
              checked={config.type === option.type}
              onChange={() => updateConfig(itemId, (current) => ({ ...current, type: option.type }))}
            />
            {option.label}
          </label>
        ))}
      </fieldset>
      {dataFileMissing ? (
        <Banner.Root tone="warning">
          {DATA_FILE_MISSING}
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={() => dismissDataFileMissing(itemId)}>
              Dismiss
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {config.type === 'performance' ? (
        <PerformanceSettings itemId={itemId} />
      ) : (
        <div className={fields.group}>
          <div className={fields.row}>
            <NumberSetting
              label="Iterations"
              name="iterations"
              value={functional.iterations}
              min={1}
              max={1_000_000}
              error={fieldErrors?.iterations}
              onChange={(iterations) => editFunctional({ iterations })}
            />
            <NumberSetting
              label="Delay"
              name="delay_ms"
              unit="ms"
              value={functional.delayMs}
              min={0}
              max={MAX_DELAY_MS}
              error={fieldErrors?.delay_ms}
              onChange={(delayMs) => editFunctional({ delayMs })}
            />
          </div>
          <DataFileField itemId={itemId} path={functional.dataFile} />
          <fieldset className={fields.checks}>
            <legend className={fields.legend}>Advanced settings</legend>
            {TOGGLES.map((toggle) => (
              <label key={toggle.key} className={fields.check}>
                <input type="checkbox" checked={functional[toggle.key]} onChange={(event) => editFunctional({ [toggle.key]: event.target.checked })} />
                {toggle.label}
              </label>
            ))}
          </fieldset>
        </div>
      )}
      <EnvironmentSelect itemId={itemId} />
      <div className={styles.actions}>
        <Button variant="primary" disabled={selected === 0 || running || starting} onClick={() => void start(itemId)}>
          {starting ? 'Starting…' : 'Start run'}
        </Button>
        {running ? <span className={styles.note}>A run is already in progress here.</span> : null}
      </div>
    </section>
  )
}
