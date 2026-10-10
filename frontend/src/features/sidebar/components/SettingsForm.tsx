import { useEffect, useEffectEvent, useState, useTransition, type ReactNode } from 'react'
import type { AppError } from '@/platform/errors'
import type { ItemSettings } from '@/platform/mappers/tree'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { useSidebarStore } from '@/features/sidebar/store'
import { AuthFields } from './AuthFields'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './SettingsForm.module.css'

const HEADER_EXAMPLES: KeyValueExamples = {
  key: 'X-Team',
  value: 'payments',
  description: 'Sent with every request',
}

const SETTINGS_TITLE = <Dialog.Title>Settings</Dialog.Title>

const AUTH_HEADING = <h3 className={styles.heading}>Authorization</h3>

const HEADERS_INTRO = (
  <>
    <h3 className={styles.heading}>Headers</h3>
    <p className={styles.note}>
      <span translate="no">Xenon</span> sends these with every request below unless a nearer level sets the same header.
    </p>
  </>
)

const LOAD_NEXT_STEP = 'Check that the workspace folder exists, then try again.'
const SAVE_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'
const FORMAT_NEXT_STEP = 'Try again, or keep editing the headers in the table.'
const BULK_NEXT_STEP = 'Check that each line reads key:value, then try again.'

interface SettingsFailure {
  error: AppError
  nextStep: string
}

const SETTINGS_LOADING = (
  <>
    {SETTINGS_TITLE}
    <Dialog.Body>
      <p role="status" className={styles.note}>
        Loading…
      </p>
    </Dialog.Body>
  </>
)

type SettingsTab = 'general' | 'variables'

const SETTINGS_TABS = (
  <Tabs.List label="Collection settings">
    <Tabs.Tab value="general">General</Tabs.Tab>
    <Tabs.Tab value="variables">Variables</Tabs.Tab>
  </Tabs.List>
)

interface SettingsFormProps {
  itemId: string
  onClose: () => void
  /** Renders the collection's variable editor; composed by the app so the sidebar stays unaware of it. */
  collectionVariables?: (collectionId: string) => ReactNode
}

export function SettingsForm({ itemId, onClose, collectionVariables }: SettingsFormProps) {
  const loadSettings = useSidebarStore((state) => state.loadSettings)
  const saveSettings = useSidebarStore((state) => state.saveSettings)
  const formatBulk = useSidebarStore((state) => state.formatBulk)
  const parseBulk = useSidebarStore((state) => state.parseBulk)
  const [settings, setSettings] = useState<ItemSettings | null>(null)
  const [bulkText, setBulkText] = useState<string | null>(null)
  const [saving, startSave] = useTransition()
  const [loadError, setLoadError] = useState<AppError | null>(null)
  const [failure, setFailure] = useState<SettingsFailure | null>(null)
  const [loadAttempt, setLoadAttempt] = useState(0)
  const [tab, setTab] = useState<SettingsTab>('general')
  const closeUnavailable = useEffectEvent(onClose)

  useEffect(() => {
    let current = true
    async function load() {
      const loaded = await loadSettings(itemId)
      if (!current) {
        return
      }
      if (!loaded.ok) {
        setLoadError(loaded.error)
      } else if (loaded.value) {
        setSettings(loaded.value)
      } else {
        closeUnavailable()
      }
    }
    void load()
    return () => {
      current = false
    }
  }, [itemId, loadAttempt, loadSettings])

  function retryLoad() {
    setLoadError(null)
    setLoadAttempt((attempt) => attempt + 1)
  }

  if (loadError) {
    return (
      <>
        {SETTINGS_TITLE}
        <Dialog.Body>
          <Banner.Root tone="error">
            <Banner.Message>
              <SidebarErrorText error={loadError} nextStep={LOAD_NEXT_STEP} />
            </Banner.Message>
          </Banner.Root>
        </Dialog.Body>
        <Dialog.Actions>
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
          <Button variant="primary" onClick={retryLoad}>
            Try Again
          </Button>
        </Dialog.Actions>
      </>
    )
  }

  if (!settings) {
    return SETTINGS_LOADING
  }

  const update = (patch: Partial<ItemSettings>) => setSettings((current) => (current ? { ...current, ...patch } : current))

  const updateHeaders = (change: (headers: ItemSettings['headers']) => ItemSettings['headers']) =>
    setSettings((current) => (current ? { ...current, headers: change(current.headers) } : current))

  async function toggleBulk(bulk: boolean) {
    if (!settings) {
      return
    }
    if (bulk) {
      const text = await formatBulk(settings.headers)
      if (!text.ok) {
        setFailure({ error: text.error, nextStep: FORMAT_NEXT_STEP })
        return
      }
      setFailure(null)
      setBulkText(text.value)
      return
    }
    const headers = await parseBulk(bulkText ?? '')
    if (!headers.ok) {
      setFailure({ error: headers.error, nextStep: BULK_NEXT_STEP })
      return
    }
    setFailure(null)
    update({ headers: headers.value })
    setBulkText(null)
  }

  function save(current: ItemSettings) {
    startSave(async () => {
      const headers = bulkText === null ? { ok: true as const, value: current.headers } : await parseBulk(bulkText)
      if (!headers.ok) {
        setFailure({ error: headers.error, nextStep: BULK_NEXT_STEP })
        return
      }
      const saveError = await saveSettings({ ...current, headers: headers.value })
      if (saveError) {
        setFailure({ error: saveError, nextStep: SAVE_NEXT_STEP })
        return
      }
      onClose()
    })
  }

  const general = (
    <div className={styles.form}>
      <label className={styles.field}>
        Description
        <textarea
          name="description"
          autoComplete="off"
          className={styles.textarea}
          value={settings.description}
          onChange={(event) => update({ description: event.target.value })}
        />
      </label>
      {AUTH_HEADING}
      <AuthFields kind={settings.kind} auth={settings.auth} onChange={(auth) => update({ auth })} />
      {HEADERS_INTRO}
      <KeyValueTable
        examples={HEADER_EXAMPLES}
        label={`${settings.name} headers`}
        rows={settings.headers}
        mode={bulkText === null ? 'table' : 'bulk'}
        bulkText={bulkText ?? ''}
        onModeChange={(mode) => void toggleBulk(mode === 'bulk')}
        onBulkTextChange={setBulkText}
        onRowChange={(id, patch) => updateHeaders((headers) => patchRow(headers, id, patch))}
        onRowAdd={(row) => updateHeaders((headers) => addRow(headers, row))}
        onRowRemove={(id) => updateHeaders((headers) => removeRow(headers, id))}
      />
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <SidebarErrorText error={failure.error} nextStep={failure.nextStep} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
    </div>
  )

  return (
    <>
      <Dialog.Title>{settings.name} Settings</Dialog.Title>
      <Dialog.Body>
        {settings.kind === 'collection' && collectionVariables ? (
          <Tabs.Root value={tab} onValueChange={(next) => setTab(next as SettingsTab)}>
            {SETTINGS_TABS}
            <Tabs.Panel value="general" className={styles.panel}>
              {general}
            </Tabs.Panel>
            <Tabs.Panel value="variables" className={styles.panel}>
              {collectionVariables(settings.id)}
            </Tabs.Panel>
          </Tabs.Root>
        ) : (
          general
        )}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="primary" disabled={saving} onClick={() => save(settings)}>
          {saving ? 'Saving…' : 'Save Settings'}
        </Button>
      </Dialog.Actions>
    </>
  )
}
