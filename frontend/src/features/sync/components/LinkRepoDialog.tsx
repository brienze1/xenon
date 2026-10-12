import { useEffect, useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import type { LinkMode } from '@/platform/mappers/sync'
import { SUBFOLDER_HAS_WORKSPACE } from '@/platform/services/sync'
import { useGitAccountsStore } from '@/features/settings/store'
import { useSyncStore } from '@/features/sync/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { errorText } from '@/features/sync/lib/error-text'
import { AccountSelect } from './AccountSelect'
import styles from './SyncPanel.module.css'

const DEFAULT_BRANCH = 'main'

export function LinkRepoDialog() {
  const open = useSyncStore((state) => state.linkOpen)
  const close = useSyncStore((state) => state.closeLink)
  const link = useSyncStore((state) => state.link)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)
  const accounts = useGitAccountsStore((state) => state.accounts)
  const loadAccounts = useGitAccountsStore((state) => state.load)
  const [repository, setRepository] = useState('')
  const [branch, setBranch] = useState(DEFAULT_BRANCH)
  const [subfolder, setSubfolder] = useState('')
  const [accountId, setAccountId] = useState('')
  const [clonePath, setClonePath] = useState('')
  const [mode, setMode] = useState<LinkMode | null>(null)
  const [needsChoice, setNeedsChoice] = useState(false)
  const [failure, setFailure] = useState<AppError | null>(null)
  const [linking, startLink] = useTransition()

  useEffect(() => {
    if (open) {
      void loadAccounts()
    }
  }, [open, loadAccounts])

  async function chooseLocation() {
    const folder = await chooseFolder('Choose where Xenon keeps its clone')
    if (folder) {
      setClonePath(folder)
    }
  }

  function submit() {
    startLink(async () => {
      const error = await link({ repository: repository.trim(), branch: branch.trim(), subfolder: subfolder.trim(), accountId, clonePath, mode })
      if (error?.code === SUBFOLDER_HAS_WORKSPACE) {
        setNeedsChoice(true)
      }
      setFailure(error)
    })
  }

  const ready = !linking && repository.trim() !== '' && branch.trim() !== '' && accountId !== '' && (!needsChoice || mode !== null)

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      <Dialog.Title>Link a Git repository</Dialog.Title>
      <Dialog.Body>
        <div className={styles.content}>
          <label className={styles.field}>
            Repository URL
            <input
              className={styles.input}
              name="sync-repository"
              autoComplete="off"
              spellCheck={false}
              placeholder="https://github.com/acme/backend.git"
              value={repository}
              onChange={(event) => setRepository(event.target.value)}
            />
          </label>
          <div className={styles.row}>
            <label className={`${styles.field} ${styles.grow}`}>
              Branch
              <input className={styles.input} name="sync-branch" autoComplete="off" spellCheck={false} value={branch} onChange={(event) => setBranch(event.target.value)} />
            </label>
            <label className={`${styles.field} ${styles.grow}`}>
              Subfolder (optional)
              <input
                className={styles.input}
                name="sync-subfolder"
                autoComplete="off"
                spellCheck={false}
                placeholder="api/xenon"
                value={subfolder}
                onChange={(event) => setSubfolder(event.target.value)}
              />
            </label>
          </div>
          <AccountSelect accounts={accounts} value={accountId} disabled={linking} onChange={setAccountId} />
          <div className={styles.row}>
            <label className={`${styles.field} ${styles.grow}`}>
              Clone location
              <input className={styles.input} name="sync-clone-path" readOnly placeholder="Xenon's app data" value={clonePath} />
            </label>
            <Button size="sm" variant="secondary" onClick={() => void chooseLocation()}>
              Choose…
            </Button>
          </div>
          <p className={styles.hint}>Xenon keeps its own clone with only the subfolder checked out; your current folder stays as it is.</p>
          {needsChoice ? (
            <fieldset className={styles.choices}>
              <legend className={styles.hint}>The subfolder already holds a Xenon workspace.</legend>
              <label>
                <input type="radio" name="sync-mode" checked={mode === 'adopt_repo'} onChange={() => setMode('adopt_repo')} /> Use the repo&apos;s workspace
              </label>
              <label>
                <input type="radio" name="sync-mode" checked={mode === 'replace_with_mine'} onChange={() => setMode('replace_with_mine')} /> Replace it with mine
              </label>
            </fieldset>
          ) : null}
          {failure && failure.code !== SUBFOLDER_HAS_WORKSPACE ? (
            <Banner.Root tone="error">
              <Banner.Message>{errorText(failure)}</Banner.Message>
            </Banner.Root>
          ) : null}
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={close}>
          Cancel
        </Button>
        <Button variant="primary" disabled={!ready} onClick={submit}>
          {linking ? 'Linking…' : 'Link'}
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
