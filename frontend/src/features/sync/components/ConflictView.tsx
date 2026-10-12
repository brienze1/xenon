import { useShallow } from 'zustand/react/shallow'
import { useSyncStore } from '@/features/sync/store'
import { errorText } from '@/features/sync/lib/error-text'
import { resolvedCount } from '@/features/sync/lib/conflict-labels'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { ConflictDetail } from './ConflictDetail'
import { ConflictList } from './ConflictList'
import styles from './ConflictView.module.css'

export function ConflictView() {
  const { open, session, selectedId, busy, error } = useSyncStore(
    useShallow((state) => ({ open: state.conflictsOpen, session: state.session, selectedId: state.selectedConflictId, busy: state.busy, error: state.error })),
  )
  const { closeConflicts, selectConflict, resolve, finishPull, cancelPull } = useSyncStore(
    useShallow((state) => ({
      closeConflicts: state.closeConflicts,
      selectConflict: state.selectConflict,
      resolve: state.resolve,
      finishPull: state.finishPull,
      cancelPull: state.cancelPull,
    })),
  )
  if (session === null) {
    return null
  }
  const conflicts = session.conflicts
  const selected = conflicts.find((conflict) => conflict.itemId === selectedId) ?? conflicts[0]
  const resolved = resolvedCount(conflicts)
  const working = busy !== null

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : closeConflicts())} className={styles.dialog}>
      <Dialog.Title>Resolve conflicts</Dialog.Title>
      <Dialog.Body>
        <p className={styles.summary} aria-live="polite">
          {resolved} of {conflicts.length} resolved · nothing changes until you finish; Cancel leaves your workspace as it was.
        </p>
        {error ? (
          <Banner.Root tone="error">
            <Banner.Message>{errorText(error)}</Banner.Message>
          </Banner.Root>
        ) : null}
        <div className={styles.layout}>
          <ConflictList conflicts={conflicts} selectedId={selected?.itemId ?? null} onSelect={selectConflict} />
          {selected ? (
            <ConflictDetail
              key={selected.itemId}
              conflict={selected}
              disabled={working}
              onResolve={(resolution, fieldPath) => void resolve(selected.itemId, resolution, fieldPath)}
            />
          ) : null}
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" disabled={working} onClick={() => void cancelPull()}>
          Cancel pull
        </Button>
        <Button variant="primary" disabled={working || resolved < conflicts.length} onClick={() => void finishPull()}>
          Finish
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
