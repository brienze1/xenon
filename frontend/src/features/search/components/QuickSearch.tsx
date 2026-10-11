import { useId, type KeyboardEvent } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { Banner } from '@/shared/ui/banner/Banner'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useSearchStore } from '@/features/search/store'
import styles from './QuickSearch.module.css'

/** Cmd/Ctrl+K: searches the workspace's requests by name, URL and folder path and opens the chosen one. */
export function QuickSearch() {
  const open = useSearchStore((state) => state.open)
  const close = useSearchStore((state) => state.close)
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      {open ? <QuickSearchPanel /> : null}
    </Dialog.Root>
  )
}

function QuickSearchPanel() {
  const { query, results, status, error, activeIndex } = useSearchStore(
    useShallow((state) => ({ query: state.query, results: state.results, status: state.status, error: state.error, activeIndex: state.activeIndex })),
  )
  const setQuery = useSearchStore((state) => state.setQuery)
  const move = useSearchStore((state) => state.move)
  const choose = useSearchStore((state) => state.choose)
  const listId = useId()
  const optionId = (index: number) => `${listId}-option-${index}`

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      move(event.key === 'ArrowDown' ? 1 : -1)
    } else if (event.key === 'Enter') {
      event.preventDefault()
      void choose()
    }
  }

  return (
    <>
      <Dialog.Title>Search Requests</Dialog.Title>
      <input
        type="search"
        role="combobox"
        name="quick-search"
        autoComplete="off"
        spellCheck={false}
        aria-label="Search requests"
        aria-expanded={results.length > 0}
        aria-controls={listId}
        aria-autocomplete="list"
        aria-activedescendant={results.length > 0 ? optionId(activeIndex) : undefined}
        placeholder="Name, URL or folder…"
        className={styles.input}
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        onKeyDown={onKeyDown}
      />
      <ul id={listId} role="listbox" aria-label="Matching requests" className={styles.results}>
        {results.map((hit, index) => (
          <li
            key={hit.id}
            id={optionId(index)}
            role="option"
            aria-selected={index === activeIndex}
            className={styles.result}
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => void choose(index)}
          >
            <MethodBadge method={hit.method} className={styles.method} />
            <span className={styles.name} translate="no">
              {hit.name}
            </span>
            <span className={styles.path} translate="no">
              {hit.path}
            </span>
            <span className={styles.url} translate="no">
              {hit.url}
            </span>
          </li>
        ))}
      </ul>
      <p role="status" aria-live="polite" className={styles.status}>
        {status === 'idle' ? 'Type to search the requests of this workspace.' : null}
        {status === 'ready' && results.length === 0 ? `No requests match “${query}”.` : null}
        {status === 'ready' && results.length > 0 ? `${results.length} ${results.length === 1 ? 'request' : 'requests'} found.` : null}
      </p>
      {error ? (
        <Banner.Root tone="error">
          <Banner.Message>{error.description || error.code}</Banner.Message>
        </Banner.Root>
      ) : null}
    </>
  )
}
