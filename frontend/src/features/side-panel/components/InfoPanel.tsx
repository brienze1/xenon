import { lazy, Suspense, useState, type KeyboardEvent } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { selectLocation, useSidebarStore } from '@/features/sidebar/store'
import { Button } from '@/shared/ui/button/Button'
import styles from './InfoPanel.module.css'

const MarkdownView = lazy(() => import('./MarkdownView'))

const INFO_HEADING = <h2 className={styles.title}>Info</h2>

const NO_DESCRIPTION = <span className={styles.muted}>No description</span>

const LOADING_DESCRIPTION = <span className={styles.muted}>Loading…</span>

interface InfoPanelProps {
  tabId: string
}

export function InfoPanel({ tabId }: InfoPanelProps) {
  const requestId = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.id ?? '')
  const name = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name ?? '')
  const description = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.description ?? '')
  const setName = useRequestBuilderStore((state) => state.setName)
  const setDescription = useRequestBuilderStore((state) => state.setDescription)
  const adoptRename = useRequestBuilderStore((state) => state.adoptRename)
  const location = useSidebarStore(useShallow((state) => (requestId ? selectLocation(state.tree, requestId).slice(0, -1) : [])))
  const rename = useSidebarStore((state) => state.rename)
  const [editing, setEditing] = useState(false)

  async function commitName(value: string) {
    const trimmed = value.trim()
    if (trimmed === '' || trimmed === name) {
      return
    }
    if (requestId === '') {
      setName(tabId, trimmed)
      return
    }
    if (await rename(requestId, trimmed)) {
      await adoptRename(requestId)
    }
  }

  function onNameKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.currentTarget.blur()
    }
  }

  return (
    <section aria-label="Info" className={styles.info}>
      {INFO_HEADING}
      <label className={styles.field}>
        Name
        <input
          name="request-name"
          autoComplete="off"
          key={`${requestId}:${name}`}
          className={styles.name}
          defaultValue={name}
          spellCheck={false}
          onBlur={(event) => void commitName(event.target.value)}
          onKeyDown={onNameKeyDown}
        />
      </label>
      <div className={styles.field}>
        <span className={styles.label}>Location</span>
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
          <span className={styles.label}>Description</span>
          <Button variant="ghost" size="sm" onClick={() => setEditing((current) => !current)}>
            {editing ? 'Done Editing' : 'Edit Description'}
          </Button>
        </div>
        {editing ? (
          <textarea
            name="description"
            autoComplete="off"
            aria-label="Description"
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
