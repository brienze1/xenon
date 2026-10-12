import { useEffect } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { useGitAccountsStore } from '@/features/settings/store'
import { useSyncStore } from '@/features/sync/store'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { errorText, lastErrorText } from '@/features/sync/lib/error-text'
import { statusLabel } from '@/features/sync/lib/status-label'
import { AccountSelect } from './AccountSelect'
import { ChangeList } from './ChangeList'
import styles from './SyncPanel.module.css'

const PHASE_LABELS: Record<string, string> = { fetch: 'Fetching…', merge: 'Merging…', write: 'Writing files…', push: 'Pushing…', clone: 'Cloning…' }

export function SyncPanel() {
  const { open, status, changes, excluded, message, busy, progress, error } = useSyncStore(
    useShallow((state) => ({
      open: state.panelOpen,
      status: state.status,
      changes: state.changes,
      excluded: state.excluded,
      message: state.message,
      busy: state.busy,
      progress: state.progress,
      error: state.error,
    })),
  )
  const { closePanel, openLink, openConflicts, toggle, setMessage, commit, push, pull, saveSettings } = useSyncStore(
    useShallow((state) => ({
      closePanel: state.closePanel,
      openLink: state.openLink,
      openConflicts: state.openConflicts,
      toggle: state.toggle,
      setMessage: state.setMessage,
      commit: state.commit,
      push: state.push,
      pull: state.pull,
      saveSettings: state.saveSettings,
    })),
  )
  const accounts = useGitAccountsStore((state) => state.accounts)
  const loadAccounts = useGitAccountsStore((state) => state.load)

  useEffect(() => {
    if (open) {
      void loadAccounts()
    }
  }, [open, loadAccounts])

  const working = busy !== null
  const linked = status !== null && status.state !== 'not_linked'
  const selected = changes.filter((change) => !excluded.includes(change.itemId)).length
  const canCommit = !working && selected > 0 && message.trim() !== '' && status?.settings.accountId !== ''

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : closePanel())} className={styles.dialog}>
      <Dialog.Title>Sync</Dialog.Title>
      <Dialog.Body>
        {linked ? (
          <div className={styles.content}>
            <p className={styles.summary} translate="no">
              {status.settings.repository} · {status.settings.branch}
              {status.settings.subfolder ? ` · ${status.settings.subfolder}` : ''}
            </p>
            <p className={styles.state} aria-live="polite">
              {working ? (PHASE_LABELS[progress?.phase ?? ''] ?? 'Syncing…') : statusLabel(status)}
            </p>
            <AccountSelect accounts={accounts} value={status.settings.accountId} disabled={working} onChange={(accountId) => void saveSettings(accountId)} />
            {status.notice ? (
              <Banner.Root tone="warning">
                <Banner.Message>{status.notice}</Banner.Message>
              </Banner.Root>
            ) : null}
            {error ? (
              <Banner.Root tone="error">
                <Banner.Message>{errorText(error)}</Banner.Message>
              </Banner.Root>
            ) : status.lastError ? (
              <Banner.Root tone="error">
                <Banner.Message>{lastErrorText(status.lastError)}</Banner.Message>
              </Banner.Root>
            ) : null}
            {status.state === 'conflict' ? (
              <Banner.Root tone="warning">
                <Banner.Message>The last pull found items changed on both sides. Nothing changes until you resolve them - your work is as it was.</Banner.Message>
                <Banner.Actions>
                  <Button size="sm" variant="accent" disabled={working} onClick={() => void openConflicts()}>
                    Resolve conflicts…
                  </Button>
                </Banner.Actions>
              </Banner.Root>
            ) : null}
            <ChangeList changes={changes} excluded={excluded} disabled={working} onToggle={toggle} />
            <label className={styles.field}>
              Commit message
              <textarea className={styles.message} name="commit-message" rows={3} value={message} disabled={working} onChange={(event) => setMessage(event.target.value)} />
            </label>
            <p className={styles.hint}>Xenon does not run Git hooks or sign commits.</p>
          </div>
        ) : (
          <div className={styles.content}>
            <p className={styles.hint}>This workspace is not linked to a Git repository. Everything works offline as before; link it to commit, push and pull.</p>
          </div>
        )}
      </Dialog.Body>
      <Dialog.Actions>
        {linked ? (
          <>
            <Button variant="secondary" disabled={working} onClick={() => void pull()}>
              Pull
            </Button>
            <Button variant="secondary" disabled={working || status.ahead === 0} onClick={() => void push()}>
              Push
            </Button>
            <Button variant="secondary" disabled={!canCommit} onClick={() => void commit(false)}>
              Commit
            </Button>
            <Button variant="primary" disabled={!canCommit} onClick={() => void commit(true)}>
              Commit and Push
            </Button>
          </>
        ) : (
          <Button variant="primary" disabled={status === null} onClick={openLink}>
            Link repository…
          </Button>
        )}
      </Dialog.Actions>
    </Dialog.Root>
  )
}
