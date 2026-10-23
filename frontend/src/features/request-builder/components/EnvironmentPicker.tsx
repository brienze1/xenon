import { useId } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Swatch } from '@/shared/ui/swatch/Swatch'
import { useEnvironmentsStore } from '@/features/environments/store'
import { aliasEnvironmentLabels } from '@/features/request-builder/lib/alias-availability'
import { pickableEnvironments, type EnvironmentChoice } from '@/features/request-builder/lib/environment-choice'
import styles from './EnvironmentPicker.module.css'

export const PICK_ONE_HINT = 'Pick at least one environment.'

interface EnvironmentPickerProps {
  choice: EnvironmentChoice
  /** The alias's stored ids, so ones whose environment was deleted can be shown as removed on save. */
  storedIds?: string[]
  complete: boolean
  onChange: (choice: EnvironmentChoice) => void
}

/** All environments, or only the ticked ones; an alias limited to environments can only be sent while one of them is active. */
export function EnvironmentPicker({ choice, storedIds = [], complete, onChange }: EnvironmentPickerProps) {
  const environments = useEnvironmentsStore(useShallow((state) => pickableEnvironments(state.environments)))
  const deleted = aliasEnvironmentLabels({ environmentIds: storedIds }, environments).filter((label) => label.deleted)
  const name = useId()
  const hintId = useId()

  function toggle(id: string, checked: boolean) {
    onChange({ mode: 'only', ids: checked ? [...choice.ids, id] : choice.ids.filter((candidate) => candidate !== id) })
  }

  return (
    <fieldset className={styles.picker} aria-describedby={complete ? undefined : hintId}>
      <legend className={styles.legend}>Environments</legend>
      <label className={styles.option}>
        <input type="radio" name={name} checked={choice.mode === 'all'} onChange={() => onChange({ mode: 'all', ids: choice.ids })} />
        All environments
      </label>
      <label className={styles.option}>
        <input type="radio" name={name} checked={choice.mode === 'only'} onChange={() => onChange({ mode: 'only', ids: choice.ids })} />
        Only these environments
      </label>
      {choice.mode === 'only' ? (
        <ul className={styles.list}>
          {environments.map((environment) => (
            <li key={environment.id}>
              <label className={styles.environment}>
                <input type="checkbox" checked={choice.ids.includes(environment.id)} onChange={(event) => toggle(environment.id, event.target.checked)} />
                <Swatch color={environment.color} />
                <NoTranslateText text={environment.name} terms={[]} />
              </label>
            </li>
          ))}
          {deleted.map((label) => (
            <li key={label.id} className={styles.deleted}>
              <Swatch color="none" />
              {label.name} · Removed when you save
            </li>
          ))}
        </ul>
      ) : null}
      {complete ? null : (
        <p id={hintId} className={styles.hint}>
          {PICK_ONE_HINT}
        </p>
      )}
    </fieldset>
  )
}
