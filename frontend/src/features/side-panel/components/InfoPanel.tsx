import { lazy, Suspense, useId, useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { selectLocation, useSidebarStore } from '@/features/sidebar/store'
import { selectNameErrorText, useTabsStore } from '@/features/tabs/store'
import { Button } from '@/shared/ui/button/Button'
import { EditableName } from '@/shared/ui/editable-name/EditableName'
import styles from './InfoPanel.module.css'

const MarkdownView = lazy(() => import('@/shared/ui/markdown/MarkdownView'))

const INFO_HEADING = <h2 className={styles.title}>Info</h2>

const NO_DESCRIPTION = <span className={styles.muted}>No description</span>

const LOADING_DESCRIPTION = (
  <span role="status" className={styles.muted}>
    Loading…
  </span>
)

const NAME_LABEL = <span className={styles.label}>Name</span>

const LOCATION_LABEL = <span className={styles.label}>Location</span>

const DESCRIPTION_LABEL = <span className={styles.label}>Description</span>

interface InfoPanelProps {
  tabId: string
}

export function InfoPanel({ tabId }: InfoPanelProps) {
  const requestId = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.id ?? '')
  const name = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name ?? '')
  const description = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.description ?? '')
  const setDescription = useRequestBuilderStore((state) => state.setDescription)
  const location = useSidebarStore(useShallow((state) => (requestId ? selectLocation(state.tree, requestId).slice(0, -1) : [])))
  const nameError = useTabsStore((state) => selectNameErrorText(state, tabId))
  const editTabName = useTabsStore((state) => state.editTabName)
  const cancelTabName = useTabsStore((state) => state.cancelTabName)
  const [editing, setEditing] = useState(false)
  const descriptionId = useId()

  return (
    <section aria-label="Info" className={styles.info}>
      {INFO_HEADING}
      <div className={styles.field}>
        {NAME_LABEL}
        <EditableName
          value={name}
          label="Rename request"
          error={nameError}
          onChange={(next) => editTabName(tabId, next)}
          onCancel={() => cancelTabName(tabId)}
        />
      </div>
      <div className={styles.field}>
        {LOCATION_LABEL}
        {location.length > 0 ? (
          <nav aria-label="Location" className={styles.breadcrumb}>
            {location.join(' / ')}
          </nav>
        ) : (
          <span className={styles.muted}>{requestId ? 'Unknown' : 'Not saved yet'}</span>
        )}
      </div>
      <div className={styles.field}>
        <div className={styles.descriptionHeader}>
          {editing ? (
            <label htmlFor={descriptionId} className={styles.label}>
              Description
            </label>
          ) : (
            DESCRIPTION_LABEL
          )}
          <Button variant="ghost" size="sm" onClick={() => setEditing((current) => !current)}>
            {editing ? 'Done Editing' : 'Edit Description'}
          </Button>
        </div>
        {editing ? (
          <textarea
            id={descriptionId}
            name="description"
            autoComplete="off"
            className={styles.textarea}
            value={description}
            onChange={(event) => setDescription(tabId, event.target.value)}
          />
        ) : description.trim() === '' ? (
          NO_DESCRIPTION
        ) : (
          <Suspense fallback={LOADING_DESCRIPTION}>
            <MarkdownView text={description} />
          </Suspense>
        )}
      </div>
    </section>
  )
}
