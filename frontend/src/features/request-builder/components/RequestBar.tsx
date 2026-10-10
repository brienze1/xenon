import { useCallback, useState, type KeyboardEvent } from 'react'
import { STANDARD_METHODS } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { SingleLineEditor } from '@/shared/ui/single-line-editor/SingleLineEditor'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useVariableEditing } from '@/features/request-builder/hooks/use-variable-editing'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './RequestBar.module.css'

interface RequestBarProps {
  tabId: string
  onSave: () => void
  mode?: 'request' | 'example'
}

const CUSTOM_METHOD = /^[A-Za-z][A-Za-z0-9_-]{0,31}$/

export function RequestBar({ tabId, onSave, mode = 'request' }: RequestBarProps) {
  const method = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.method ?? 'GET')
  const url = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.url ?? '')
  const sending = useRequestBuilderStore((state) => state.tabs[tabId]?.pending != null)
  const saved = useRequestBuilderStore((state) => (state.tabs[tabId]?.draft.id ?? '') !== '')
  const saving = useRequestBuilderStore((state) => state.tabs[tabId]?.saving === true)
  const setMethod = useRequestBuilderStore((state) => state.setMethod)
  const setUrl = useRequestBuilderStore((state) => state.setUrl)
  const send = useRequestBuilderStore((state) => state.send)
  const cancel = useRequestBuilderStore((state) => state.cancel)
  const [customMethod, setCustomMethod] = useState('')
  const { decorationsOf, completionSource, selectIn } = useVariableEditing(tabId)
  const onUrlChange = useCallback((value: string) => setUrl(tabId, value), [setUrl, tabId])
  const sendable = mode === 'request'
  const onUrlEnter = useCallback(() => (sendable ? void send(tabId) : undefined), [send, sendable, tabId])
  const onUrlSelection = useCallback((selection: Parameters<typeof selectIn>[1]) => selectIn('url', selection), [selectIn])

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
            <Dropdown.RadioItem key={option} checked={option === method} onSelect={() => setMethod(tabId, option)}>
              <MethodBadge method={option} />
            </Dropdown.RadioItem>
          ))}
          <input
            name="custom-method"
            autoComplete="off"
            aria-label="Custom method"
            placeholder="PURGE…"
            translate="no"
            className={styles.customMethod}
            value={customMethod}
            spellCheck={false}
            onChange={(event) => setCustomMethod(event.target.value)}
            onKeyDown={onCustomKeyDown}
          />
        </Dropdown.Menu>
      </Dropdown.Root>
      <SingleLineEditor
        label="Request URL"
        className={styles.url}
        placeholder="https://api.example.com/users…"
        value={url}
        decorations={decorationsOf('url')}
        completionSource={completionSource}
        onChange={onUrlChange}
        onEnter={onUrlEnter}
        onSelectionChange={onUrlSelection}
      />
      {!sendable ? null : sending ? (
        <Button variant="running" className={styles.send} onClick={() => cancel(tabId)}>
          Cancel Request
        </Button>
      ) : (
        <Button variant="accent" className={styles.send} onClick={() => void send(tabId)}>
          Send
        </Button>
      )}
      <Button variant="secondary" disabled={saving} title={saved || !sendable ? 'Save' : 'Save into a collection'} onClick={onSave}>
        {saving ? 'Saving…' : 'Save'}
      </Button>
    </div>
  )
}
