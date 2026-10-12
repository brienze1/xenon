import { useEffect, useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import { pickFolder } from '@/platform/services/dialogs'
import { useGitAccountsStore } from '@/features/settings/store'
import { useSyncStore } from '@/features/sync/store'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { errorText } from '@/features/sync/lib/error-text'
import { AccountSelect } from './AccountSelect'
import styles from './SyncPanel.module.css'

export function CloneWorkspaceDialog() {
  const open = useSyncStore((state) => state.cloneOpen)
  const close = useSyncStore((state) => state.closeClone)
  const cloneWorkspace = useSyncStore((state) => state.cloneWorkspace)
  const accounts = useGitAccountsStore((state) => state.accounts)
  const loadAccounts = useGitAccountsStore((state) => state.load)
  const [repository, setRepository] = useState('')
  const [branch, setBranch] = useState('')
  const [subfolder, setSubfolder] = useState('')
  const [accountId, setAccountId] = useState('')
  const [clonePath, setClonePath] = useState('')
  const [name, setName] = useState('')
  const [failure, setFailure] = useState<AppError | null>(null)
  const [cloning, startClone] = useTransition()

  useEffect(() => {
    if (open) {
      void loadAccounts()
    }
  }, [open, loadAccounts])

  async function chooseLocation() {
    const folder = await pickFolder('Choose where Xenon keeps the clone')
    if (folder) {
      setClonePath(folder)
    }
  }

  function dismiss() {
    setFailure(null)
    close()
  }

  function submit() {
    startClone(async () => {
      setFailure(await cloneWorkspace({ repository: repository.trim(), branch: branch.trim(), subfolder: subfolder.trim(), accountId, clonePath, name: name.trim() }))
    })
  }

  const ready = !cloning && repository.trim() !== '' && accountId !== ''

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : dismiss())} className={styles.dialog}>
      <Dialog.Title>New workspace from Git repository</Dialog.Title>
      <Dialog.Body>
        <div className={styles.content}>
          <label className={styles.field}>
            Repository URL
            <input
              className={styles.input}
              name="clone-repository"
              autoComplete="off"
              spellCheck={false}
              placeholder="https://github.com/acme/backend.git"
              value={repository}
              onChange={(event) => setRepository(event.target.value)}
            />
          </label>
          <div className={styles.row}>
            <label className={`${styles.field} ${styles.grow}`}>
              Branch (optional)
              <input
                className={styles.input}
                name="clone-branch"
                autoComplete="off"
                spellCheck={false}
                placeholder="Remote default branch"
                value={branch}
                onChange={(event) => setBranch(event.target.value)}
              />
            </label>
            <label className={`${styles.field} ${styles.grow}`}>
              Subfolder (optional)
              <input
                className={styles.input}
                name="clone-subfolder"
                autoComplete="off"
                spellCheck={false}
                placeholder="api/xenon"
                value={subfolder}
                onChange={(event) => setSubfolder(event.target.value)}
              />
            </label>
          </div>
          <AccountSelect accounts={accounts} value={accountId} disabled={cloning} onChange={setAccountId} />
          <label className={styles.field}>
            Workspace name (optional)
            <input
              className={styles.input}
              name="clone-name"
              autoComplete="off"
              placeholder="Used when the subfolder has no workspace yet"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
          </label>
          <div className={styles.row}>
            <label className={`${styles.field} ${styles.grow}`}>
              Clone location
              <input className={styles.input} name="clone-path" readOnly placeholder="Xenon's app data" value={clonePath} />
            </label>
            <Button size="sm" variant="secondary" onClick={() => void chooseLocation()}>
              Choose…
            </Button>
          </div>
          <p className={styles.hint}>
            Xenon clones the repository with only the subfolder checked out and opens the workspace there. Secret values never travel in Git - you enter them on this machine.
          </p>
          {failure ? (
            <Banner.Root tone="error">
              <Banner.Message>{errorText(failure)}</Banner.Message>
            </Banner.Root>
          ) : null}
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={dismiss}>
          Cancel
        </Button>
        <Button variant="primary" disabled={!ready} onClick={submit}>
          {cloning ? 'Cloning…' : 'Clone and Open'}
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
