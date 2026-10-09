import { useMemo, useState, type FormEvent } from 'react'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { containerOptions } from '../lib/containers'
import { useTabsStore, type SaveDialogState } from '../store'
import styles from './SaveRequestDialog.module.css'

export function SaveRequestDialog() {
  const dialog = useTabsStore((state) => state.saveDialog)
  const cancelSave = useTabsStore((state) => state.cancelSave)
  return (
    <Dialog.Root open={dialog !== null} onOpenChange={(open) => (open ? undefined : cancelSave())}>
      {dialog ? <SaveRequestForm key={`${dialog.tabId}:${dialog.mode}`} dialog={dialog} /> : null}
    </Dialog.Root>
  )
}

interface SaveRequestFormProps {
  dialog: SaveDialogState
}

function SaveRequestForm({ dialog }: SaveRequestFormProps) {
  const currentName = useRequestBuilderStore((state) => state.tabs[dialog.tabId]?.draft.name ?? '')
  const saving = useRequestBuilderStore((state) => state.tabs[dialog.tabId]?.saving === true)
  const tree = useSidebarStore((state) => state.tree)
  const confirmSave = useTabsStore((state) => state.confirmSave)
  const cancelSave = useTabsStore((state) => state.cancelSave)
  const options = useMemo(() => containerOptions(tree), [tree])
  const [name, setName] = useState(dialog.mode === 'saveAs' ? `${currentName} Copy` : currentName)
  const [parentId, setParentId] = useState<string | null>(options[0]?.id ?? null)

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (parentId !== null && name.trim() !== '') {
      void confirmSave(parentId, name.trim())
    }
  }

  return (
    <form aria-label="Save request" className={styles.form} onSubmit={onSubmit}>
      <Dialog.Title>{dialog.mode === 'saveAs' ? 'Save request as' : 'Save request'}</Dialog.Title>
      <Dialog.Body>
        <label className={styles.field}>
          <span className={styles.label}>Request name</span>
          <input className={styles.input} value={name} maxLength={200} autoComplete="off" onChange={(event) => setName(event.target.value)} />
        </label>
        <fieldset className={styles.places}>
          <legend className={styles.label}>Save to</legend>
          {options.length === 0 ? <p className={styles.empty}>Create a collection in the sidebar first.</p> : null}
          {options.map((option) => (
            <label key={option.id} className={styles.place} style={{ paddingInlineStart: `calc(${option.depth} * var(--space-4))` }}>
              <input type="radio" name="parent" checked={parentId === option.id} onChange={() => setParentId(option.id)} />
              {option.name}
            </label>
          ))}
        </fieldset>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={cancelSave}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={saving || parentId === null || name.trim() === ''}>
          Save
        </Button>
      </Dialog.Actions>
    </form>
  )
}
