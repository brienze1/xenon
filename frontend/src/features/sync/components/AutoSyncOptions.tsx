import type { SyncSettingsChange } from '@/features/sync/store'
import styles from './SyncPanel.module.css'

interface AutoSyncOptionsProps {
  autoPull: boolean
  autoPush: boolean
  disabled: boolean
  onChange: (change: SyncSettingsChange) => void
}

export function AutoSyncOptions({ autoPull, autoPush, disabled, onChange }: AutoSyncOptionsProps) {
  return (
    <fieldset className={styles.choices}>
      <legend className={styles.hint}>Automatic sync (runs in the background; problems show here)</legend>
      <label>
        <input type="checkbox" name="sync-auto-pull" checked={autoPull} disabled={disabled} onChange={(event) => onChange({ autoPull: event.target.checked })} /> Pull when the
        workspace is opened
      </label>
      <label>
        <input type="checkbox" name="sync-auto-push" checked={autoPush} disabled={disabled} onChange={(event) => onChange({ autoPush: event.target.checked })} /> Push after
        each save (3 seconds after the last one)
      </label>
    </fieldset>
  )
}
