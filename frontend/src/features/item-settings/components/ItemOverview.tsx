import { lazy, Suspense, useId, useState, type KeyboardEvent } from 'react'
import type { AppError } from '@/platform/errors'
import type { ItemSettings } from '@/platform/mappers/tree'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useSidebarStore } from '@/features/sidebar/store'
import styles from './ItemTab.module.css'

const MarkdownView = lazy(() => import('@/shared/ui/markdown/MarkdownView'))

const RENAME_NEXT_STEP = 'Choose another name, then press Enter.'

const LOADING_DESCRIPTION = (
  <span role="status" className={styles.note}>
    Loading…
  </span>
)

interface ItemOverviewProps {
  settings: ItemSettings
  /** "Collection" or "Folder". */
  label: string
}

/** The Overview of a collection or folder: its name (renames it, like the sidebar) and its markdown description. */
export function ItemOverview({ settings, label }: ItemOverviewProps) {
  const itemId = settings.id
  const rename = useSidebarStore((state) => state.rename)
  const adoptName = useItemSettingsStore((state) => state.adoptName)
  const edit = useItemSettingsStore((state) => state.edit)
  const [editing, setEditing] = useState(settings.description.trim() === '')
  const [nameError, setNameError] = useState<AppError | null>(null)
  const nameErrorId = useId()
  const descriptionId = useId()

  async function commitName(value: string) {
    const trimmed = value.trim()
    if (trimmed === '' || trimmed === settings.name) {
      return
    }
    const renamed = await rename(itemId, trimmed)
    if (!renamed.ok) {
      setNameError(renamed.error)
      return
    }
    setNameError(null)
    if (renamed.value) {
      adoptName(itemId, trimmed)
    }
  }

  function onNameKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.currentTarget.blur()
    }
  }

  return (
    <div className={styles.overview}>
      <div className={styles.field}>
        <label className={styles.field}>
          Name
          <input
            name="item-name"
            autoComplete="off"
            key={settings.name}
            className={styles.input}
            translate="no"
            defaultValue={settings.name}
            spellCheck={false}
            aria-invalid={nameError !== null}
            aria-describedby={nameError ? nameErrorId : undefined}
            onChange={() => setNameError(null)}
            onBlur={(event) => void commitName(event.target.value)}
            onKeyDown={onNameKeyDown}
          />
        </label>
        {nameError ? (
          <p id={nameErrorId} role="alert" className={styles.fieldError}>
            <NoTranslateText text={withNextStep(nameError.description || nameError.code, RENAME_NEXT_STEP)} terms={[nameError.code]} />
          </p>
        ) : null}
      </div>
      <div className={styles.field}>
        <div className={styles.descriptionHeader}>
          {editing ? (
            <label htmlFor={descriptionId} className={styles.label}>
              Description
            </label>
          ) : (
            <span className={styles.label}>Description</span>
          )}
          <Button variant="ghost" size="sm" onClick={() => setEditing((current) => !current)}>
            {editing ? 'Preview' : 'Edit Description'}
          </Button>
        </div>
        {editing ? (
          <textarea
            id={descriptionId}
            name="description"
            autoComplete="off"
            className={styles.textarea}
            placeholder={`Describe this ${label.toLowerCase()} in Markdown: what it is for, how to authenticate…`}
            value={settings.description}
            onChange={(event) => edit(itemId, { description: event.target.value })}
          />
        ) : settings.description.trim() === '' ? (
          <span className={styles.note}>No description</span>
        ) : (
          <Suspense fallback={LOADING_DESCRIPTION}>
            <MarkdownView text={settings.description} />
          </Suspense>
        )}
      </div>
    </div>
  )
}
