import { useState } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { clashLabel } from '@/features/import-export/lib/labels'
import { pendingClashes } from '@/features/import-export/lib/plan'
import { useImportStore } from '@/features/import-export/store'
import styles from './ImportDialog.module.css'

export function ClashPrompt() {
  const [applyToAll, setApplyToAll] = useState(false)
  const clash = useImportStore((state) => pendingClashes(state.clashes, state.resolutions)[0])
  const remaining = useImportStore((state) => pendingClashes(state.clashes, state.resolutions).length)
  const resolve = useImportStore((state) => state.resolve)
  const back = useImportStore((state) => state.back)
  if (!clash) {
    return null
  }

  return (
    <>
      <Dialog.Body>
        <p className={styles.source} translate="no">
          {clashLabel(clash.kind, clash.existingName)}
        </p>
        <p>Replace it, keep both (the imported one gets “(imported)”), or skip it.</p>
        <label className={styles.checkbox}>
          <input type="checkbox" name="apply-to-all" checked={applyToAll} onChange={(event) => setApplyToAll(event.target.checked)} />
          Apply to all {remaining > 1 ? `${remaining} clashes` : 'clashes'}
        </label>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={back}>
          Cancel
        </Button>
        <Button onClick={() => void resolve('skip', applyToAll)}>Skip</Button>
        <Button onClick={() => void resolve('keep_both', applyToAll)}>Keep Both</Button>
        <Button variant="danger" onClick={() => void resolve('replace', applyToAll)}>
          Replace
        </Button>
      </Dialog.Actions>
    </>
  )
}
