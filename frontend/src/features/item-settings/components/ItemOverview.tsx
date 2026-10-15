import { lazy, Suspense, useId, useState } from 'react'
import type { ItemSettings } from '@/platform/mappers/tree'
import { Button } from '@/shared/ui/button/Button'
import { useItemSettingsStore } from '@/features/item-settings/store'
import styles from './ItemTab.module.css'

const MarkdownView = lazy(() => import('@/shared/ui/markdown/MarkdownView'))

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

/** The Overview of a collection or folder: its markdown description (the name is edited in the tab's heading). */
export function ItemOverview({ settings, label }: ItemOverviewProps) {
  const itemId = settings.id
  const edit = useItemSettingsStore((state) => state.edit)
  const [editing, setEditing] = useState(settings.description.trim() === '')
  const descriptionId = useId()

  return (
    <div className={styles.overview}>
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
