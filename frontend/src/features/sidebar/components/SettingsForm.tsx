import { useEffect, useState } from 'react'
import type { ItemSettings } from '@/platform/mappers/tree'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { useSidebarStore } from '@/features/sidebar/store'
import { AuthFields } from './AuthFields'
import styles from './SettingsForm.module.css'

interface SettingsFormProps {
  itemId: string
  onClose: () => void
}

export function SettingsForm({ itemId, onClose }: SettingsFormProps) {
  const loadSettings = useSidebarStore((state) => state.loadSettings)
  const saveSettings = useSidebarStore((state) => state.saveSettings)
  const formatBulk = useSidebarStore((state) => state.formatBulk)
  const parseBulk = useSidebarStore((state) => state.parseBulk)
  const [settings, setSettings] = useState<ItemSettings | null>(null)
  const [bulkText, setBulkText] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let current = true
    async function load() {
      const loaded = await loadSettings(itemId)
      if (current) {
        setSettings(loaded)
      }
    }
    void load()
    return () => {
      current = false
    }
  }, [itemId, loadSettings])

  if (!settings) {
    return (
      <>
        <Dialog.Title>Settings</Dialog.Title>
        <Dialog.Body>
          <p className={styles.note}>Loading…</p>
        </Dialog.Body>
      </>
    )
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
      if (text !== null) {
        setBulkText(text)
      }
      return
    }
    const headers = await parseBulk(bulkText ?? '')
    if (headers !== null) {
      update({ headers })
      setBulkText(null)
    }
  }

  async function save() {
    if (!settings) {
      return
    }
    setSaving(true)
    const headers = bulkText === null ? settings.headers : await parseBulk(bulkText)
    if (headers === null) {
      setSaving(false)
      return
    }
    const saved = await saveSettings({ ...settings, headers })
    setSaving(false)
    if (saved) {
      onClose()
    }
  }

  return (
    <>
      <Dialog.Title>{settings.name} Settings</Dialog.Title>
      <Dialog.Body>
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
          <h3 className={styles.heading}>Authorization</h3>
          <AuthFields kind={settings.kind} auth={settings.auth} onChange={(auth) => update({ auth })} />
          <h3 className={styles.heading}>Headers</h3>
          <p className={styles.note}>
            <span translate="no">Xenon</span> sends these with every request below unless a nearer level sets the same header.
          </p>
          <KeyValueTable
            examples={{
              key: 'X-Team',
              value: 'payments',
              description: 'Sent with every request',
            }}
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
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="primary" disabled={saving} onClick={() => void save()}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </Dialog.Actions>
    </>
  )
}
