import { useId, useMemo, useRef, useState, type CSSProperties, type FormEvent } from 'react'
import { flushSync } from 'react-dom'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSidebarStore } from '@/features/sidebar/store'
import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { containerOptions } from '@/features/tabs/lib/containers'
import type { SaveDialogState } from '@/features/tabs/model'
import { useTabsStore } from '@/features/tabs/store'
import styles from './SaveRequestDialog.module.css'

const NO_COLLECTIONS = <p className={styles.empty}>Create a collection in the sidebar first.</p>

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
const SAVE_NEXT_STEP = 'Choose another name or place, then try again.'

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
  const [saveError, setSaveError] = useState<AppError | null>(null)
  const saveErrorRef = useRef<HTMLDivElement>(null)
  const nameRef = useRef<HTMLInputElement>(null)
  const placesRef = useRef<HTMLFieldSetElement>(null)
  const firstPlaceRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()
  const placeErrorId = useId()

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaveError(null)
    const trimmed = name.trim()
    setNameError(trimmed === '' ? NAME_REQUIRED : null)
    setPlaceError(parentId === null ? PLACE_REQUIRED : null)
    if (trimmed === '') {
      nameRef.current?.focus()
      return
    }
    if (parentId === null) {
      const firstPlace = firstPlaceRef.current ?? placesRef.current
      firstPlace?.focus()
      return
    }
    const failure = await confirmSave(parentId, trimmed)
    if (failure) {
      flushSync(() => setSaveError(failure))
      saveErrorRef.current?.focus()
    }
  }

  return (
    <form aria-label="Save request" className={styles.form} noValidate onSubmit={(event) => void onSubmit(event)}>
      <Dialog.Title>{dialog.mode === 'saveAs' ? 'Save Request As' : 'Save Request'}</Dialog.Title>
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
          <p id={nameErrorId} role="alert" className={styles.fieldError}>
            {nameError}
          </p>
        ) : null}
        <fieldset
          ref={placesRef}
          tabIndex={-1}
          className={styles.places}
          aria-describedby={placeError ? placeErrorId : undefined}
        >
          <legend className={styles.label}>Save to</legend>
          {options.length === 0 ? NO_COLLECTIONS : null}
          {options.map((option, index) => (
            <label key={option.id} className={styles.place} style={{ '--depth': option.depth } as CSSProperties}>
              <input
                ref={index === 0 ? firstPlaceRef : undefined}
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
            <p id={placeErrorId} role="alert" className={styles.fieldError}>
              {placeError}
            </p>
          ) : null}
        </fieldset>
        {saveError ? (
          <div ref={saveErrorRef} tabIndex={-1} className={styles.saveError}>
            <Banner.Root tone="error">
              <Banner.Message>
                <NoTranslateText text={withNextStep(saveError.description || saveError.code, SAVE_NEXT_STEP)} terms={[saveError.code]} />
              </Banner.Message>
            </Banner.Root>
          </div>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={cancelSave}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={saving}>
          {saving ? 'Saving…' : 'Save Request'}
        </Button>
      </Dialog.Actions>
    </form>
  )
}
