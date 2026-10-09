import { useId, useMemo, useRef, useState, type CSSProperties, type FormEvent } from 'react'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { containerOptions } from '@/features/tabs/lib/containers'
import { useTabsStore, type SaveDialogState } from '@/features/tabs/store'
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

const NAME_REQUIRED = 'Enter a name for the request.'
const PLACE_REQUIRED = 'Choose where to save the request.'

function initialName(dialog: SaveDialogState): string {
  const currentName = useRequestBuilderStore.getState().tabs[dialog.tabId]?.draft.name ?? ''
  return dialog.mode === 'saveAs' ? `${currentName} Copy` : currentName
}

function SaveRequestForm({ dialog }: SaveRequestFormProps) {
  const saving = useRequestBuilderStore((state) => state.tabs[dialog.tabId]?.saving === true)
  const tree = useSidebarStore((state) => state.tree)
  const confirmSave = useTabsStore((state) => state.confirmSave)
  const cancelSave = useTabsStore((state) => state.cancelSave)
  const options = useMemo(() => containerOptions(tree), [tree])
  const [name, setName] = useState(() => initialName(dialog))
  const [parentId, setParentId] = useState<string | null>(() => options[0]?.id ?? null)
  const [nameError, setNameError] = useState<string | null>(null)
  const [placeError, setPlaceError] = useState<string | null>(null)
  const nameRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()
  const placeErrorId = useId()

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    setNameError(trimmed === '' ? NAME_REQUIRED : null)
    setPlaceError(parentId === null ? PLACE_REQUIRED : null)
    if (trimmed === '') {
      nameRef.current?.focus()
      return
    }
    if (parentId !== null) {
      void confirmSave(parentId, trimmed)
    }
  }

  return (
    <form aria-label="Save request" className={styles.form} noValidate onSubmit={onSubmit}>
      <Dialog.Title>{dialog.mode === 'saveAs' ? 'Save request as' : 'Save request'}</Dialog.Title>
      <Dialog.Body>
        <label className={styles.field}>
          <span className={styles.label}>Request name</span>
          <input
            ref={nameRef}
            className={styles.input}
            name="request-name"
            value={name}
            maxLength={200}
            autoComplete="off"
            aria-invalid={nameError !== null}
            aria-describedby={nameError ? nameErrorId : undefined}
            onChange={(event) => {
              setName(event.target.value)
              setNameError(null)
            }}
          />
        </label>
        {nameError ? (
          <p id={nameErrorId} className={styles.fieldError}>
            {nameError}
          </p>
        ) : null}
        <fieldset className={styles.places} aria-describedby={placeError ? placeErrorId : undefined}>
          <legend className={styles.label}>Save to</legend>
          {options.length === 0 ? <p className={styles.empty}>Create a collection in the sidebar first.</p> : null}
          {options.map((option) => (
            <label key={option.id} className={styles.place} style={{ '--depth': option.depth } as CSSProperties}>
              <input
                type="radio"
                name="parent"
                checked={parentId === option.id}
                onChange={() => {
                  setParentId(option.id)
                  setPlaceError(null)
                }}
              />
              {option.name}
            </label>
          ))}
          {placeError ? (
            <p id={placeErrorId} className={styles.fieldError}>
              {placeError}
            </p>
          ) : null}
        </fieldset>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={cancelSave}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={saving}>
          Save
        </Button>
      </Dialog.Actions>
    </form>
  )
}
