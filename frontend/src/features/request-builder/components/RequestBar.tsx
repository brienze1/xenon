import { useState, type KeyboardEvent } from 'react'
import { STANDARD_METHODS } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useRequestBuilderStore } from '../store'
import styles from './RequestBar.module.css'

interface RequestBarProps {
  tabId: string
}

const CUSTOM_METHOD = /^[A-Za-z][A-Za-z0-9_-]{0,31}$/

export function RequestBar({ tabId }: RequestBarProps) {
  const method = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.method ?? 'GET')
  const url = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.url ?? '')
  const sending = useRequestBuilderStore((state) => state.tabs[tabId]?.pending != null)
  const saved = useRequestBuilderStore((state) => (state.tabs[tabId]?.draft.id ?? '') !== '')
  const saving = useRequestBuilderStore((state) => state.tabs[tabId]?.saving === true)
  const save = useRequestBuilderStore((state) => state.save)
  const setMethod = useRequestBuilderStore((state) => state.setMethod)
  const setUrl = useRequestBuilderStore((state) => state.setUrl)
  const send = useRequestBuilderStore((state) => state.send)
  const cancel = useRequestBuilderStore((state) => state.cancel)
  const [customMethod, setCustomMethod] = useState('')

  function onUrlKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter' && !event.metaKey && !event.ctrlKey) {
      event.preventDefault()
      void send(tabId)
    }
  }

  function onCustomKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    event.stopPropagation()
    if (event.key === 'Enter' && CUSTOM_METHOD.test(customMethod.trim())) {
      setMethod(tabId, customMethod)
      setCustomMethod('')
    }
  }

  return (
    <div className={styles.bar}>
      <Dropdown.Root className={styles.method}>
        <Dropdown.Trigger label={`Method ${method}`} className={styles.methodTrigger}>
          <MethodBadge method={method} />
        </Dropdown.Trigger>
        <Dropdown.Menu label="HTTP method">
          {STANDARD_METHODS.map((option) => (
            <Dropdown.Item key={option} checked={option === method} onSelect={() => setMethod(tabId, option)}>
              <MethodBadge method={option} />
            </Dropdown.Item>
          ))}
          <input
            aria-label="Custom method"
            placeholder="Custom…"
            className={styles.customMethod}
            value={customMethod}
            spellCheck={false}
            onChange={(event) => setCustomMethod(event.target.value)}
            onKeyDown={onCustomKeyDown}
          />
        </Dropdown.Menu>
      </Dropdown.Root>
      <input
        aria-label="Request URL"
        className={styles.url}
        placeholder="Enter URL or paste text"
        spellCheck={false}
        value={url}
        onChange={(event) => setUrl(tabId, event.target.value)}
        onKeyDown={onUrlKeyDown}
      />
      {sending ? (
        <Button variant="secondary" className={styles.send} onClick={() => cancel(tabId)}>
          Cancel
        </Button>
      ) : (
        <Button variant="accent" className={styles.send} onClick={() => void send(tabId)}>
          Send
        </Button>
      )}
      <Button
        variant="secondary"
        disabled={!saved || saving}
        title={saved ? 'Save' : 'Open a request from the sidebar to save it'}
        onClick={() => void save(tabId)}
      >
        Save
      </Button>
    </div>
  )
}
