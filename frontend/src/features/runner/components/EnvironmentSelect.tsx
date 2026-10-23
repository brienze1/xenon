import { useId } from 'react'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { ACTIVE_ENVIRONMENT, NO_ENVIRONMENT } from '@/platform/mappers/runner'
import { useRunnerStore } from '@/features/runner/store'
import styles from './EnvironmentSelect.module.css'

const NO_ENVIRONMENTS: EnvironmentSummary[] = []

interface EnvironmentSelectProps {
  itemId: string
}

/** The run's environment: the workspace's active one by default, another one, or none; the active environment stays as is. */
export function EnvironmentSelect({ itemId }: EnvironmentSelectProps) {
  const id = useId()
  const environments = useRunnerStore((state) => state.targets[itemId]?.environments ?? NO_ENVIRONMENTS)
  const activeId = useRunnerStore((state) => state.targets[itemId]?.activeEnvironmentId ?? null)
  const chosen = useRunnerStore((state) => state.targets[itemId]?.config?.environmentId ?? ACTIVE_ENVIRONMENT)
  const updateConfig = useRunnerStore((state) => state.updateConfig)
  const value = chosen === ACTIVE_ENVIRONMENT ? (activeId ?? NO_ENVIRONMENT) : chosen
  const shown = environments.find((environment) => environment.id === value) ?? null

  return (
    <div className={styles.field}>
      <label htmlFor={id} className={styles.label}>
        Environment
      </label>
      <div className={styles.control}>
        {shown?.confirmBeforeSend ? <span className={styles.chip} data-color={shown.color} aria-hidden="true" /> : null}
        <select id={id} className={styles.select} value={value} onChange={(event) => updateConfig(itemId, (config) => ({ ...config, environmentId: event.target.value }))}>
          <option value={NO_ENVIRONMENT}>No environment</option>
          {environments.map((environment) => (
            <option key={environment.id} value={environment.id} disabled={environment.error !== null}>
              {environment.name}
              {environment.id === activeId ? ' (active)' : ''}
            </option>
          ))}
        </select>
      </div>
      {shown?.confirmBeforeSend ? <p className={styles.hint}>This environment asks for a confirmation before the run starts.</p> : null}
    </div>
  )
}
