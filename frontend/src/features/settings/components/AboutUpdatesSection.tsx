import { Button } from '@/shared/ui/button/Button'
import { useSettingsStore } from '@/features/settings/store'
import { updateRow } from '@/features/settings/lib/update-row'
import { useUpdatesStore } from '@/features/updates/store'
import styles from './SettingsSection.module.css'

/** The first Settings section: the running version, the update status and Update automatically. */
export function AboutUpdatesSection() {
  const status = useUpdatesStore((state) => state.status)
  const check = useUpdatesStore((state) => state.check)
  const update = useUpdatesStore((state) => state.update)
  const restart = useUpdatesStore((state) => state.restart)
  const openReleasePage = useUpdatesStore((state) => state.openReleasePage)
  const autoUpdate = useSettingsStore((state) => state.draft.autoUpdate)
  const edit = useSettingsStore((state) => state.edit)
  const row = updateRow(status)

  return (
    <fieldset className={styles.section}>
      <legend className={styles.legend}>About &amp; Updates</legend>
      <p className={styles.version} translate="no">
        Version {status.currentVersion}
      </p>
      <div role="status" aria-label="Update status" className={styles.updateStatus}>
        {row.message !== null ? <p className={styles.updateMessage}>{row.message}</p> : null}
        {row.progress !== null ? (
          <progress className={styles.progress} aria-label="Download progress" max={100} value={row.progress} />
        ) : null}
        {row.showLatest ? (
          <p className={styles.hint}>
            Latest version <span translate="no">{status.latestVersion}</span>
          </p>
        ) : null}
        {row.showWhatsNew ? (
          <button type="button" className={styles.link} onClick={() => void openReleasePage()}>
            What's new
          </button>
        ) : null}
        {row.showReleasePageLink ? (
          <button type="button" className={styles.link} onClick={() => void openReleasePage()}>
            Open the release page
          </button>
        ) : null}
      </div>
      {row.actions.length > 0 ? (
        <div className={styles.updateActions}>
          {row.actions.includes('update') ? (
            <Button variant="primary" size="sm" onClick={() => void update()}>
              Update
            </Button>
          ) : null}
          {row.actions.includes('restart') ? (
            <Button variant="primary" size="sm" onClick={() => void restart()}>
              Restart now
            </Button>
          ) : null}
          {row.actions.includes('check') ? (
            <Button size="sm" disabled={row.checkDisabled} onClick={() => void check()}>
              Check for updates
            </Button>
          ) : null}
        </div>
      ) : null}
      <label className={styles.toggle}>
        <input type="checkbox" role="switch" className={styles.switch} checked={autoUpdate} onChange={(event) => edit({ autoUpdate: event.target.checked })} />
        Update automatically
      </label>
    </fieldset>
  )
}
