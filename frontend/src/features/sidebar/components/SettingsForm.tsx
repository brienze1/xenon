import { useEffect, useState } from 'react'
import type { ItemSettings } from '@/platform/mappers/tree'
import { formatBulk, parseBulk } from '@/platform/services/requests'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '../lib/rows'
import { useSidebarStore } from '../store'
import { AuthFields } from './AuthFields'
import styles from './ItemSettingsDialog.module.css'

interface SettingsFormProps {
  itemId: string
  onClose: () => void
}

export function SettingsForm({ itemId, onClose }: SettingsFormProps) {
  const loadSettings = useSidebarStore((state) => state.loadSettings)
  const saveSettings = useSidebarStore((state) => state.saveSettings)
  const [settings, setSettings] = useState<ItemSettings | null>(null)
  const [bulkText, setBulkText] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let current = true
    void loadSettings(itemId).then((loaded) => {
      if (current) {
        setSettings(loaded)
      }
    })
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

  const update = (patch: Partial<ItemSettings>) => setSettings({ ...settings, ...patch })

  async function toggleBulk(bulk: boolean) {
    if (!settings) {
      return
    }
    if (bulk) {
      setBulkText(await formatBulk(settings.headers))
      return
    }
    update({ headers: await parseBulk(bulkText ?? '') })
    setBulkText(null)
  }

  async function save() {
    if (!settings) {
      return
    }
    setSaving(true)
    const headers = bulkText === null ? settings.headers : await parseBulk(bulkText)
    const saved = await saveSettings({ ...settings, headers })
    setSaving(false)
    if (saved) {
      onClose()
    }
  }

  return (
    <>
      <Dialog.Title>{settings.name} settings</Dialog.Title>
      <Dialog.Body>
        <div className={styles.form}>
          <label className={styles.field}>
            Description
            <textarea
              className={styles.textarea}
              value={settings.description}
              onChange={(event) => update({ description: event.target.value })}
            />
          </label>
          <h3 className={styles.heading}>Authorization</h3>
          <AuthFields kind={settings.kind} auth={settings.auth} onChange={(auth) => update({ auth })} />
          <h3 className={styles.heading}>Headers</h3>
          <p className={styles.note}>Sent with every request below unless a nearer level sets the same header.</p>
          <KeyValueTable
            label={`${settings.name} headers`}
            rows={settings.headers}
            mode={bulkText === null ? 'table' : 'bulk'}
            bulkText={bulkText ?? ''}
            onModeChange={(mode) => void toggleBulk(mode === 'bulk')}
            onBulkTextChange={setBulkText}
            onRowChange={(id, patch) => update({ headers: patchRow(settings.headers, id, patch) })}
            onRowAdd={(row) => update({ headers: addRow(settings.headers, row) })}
            onRowRemove={(id) => update({ headers: removeRow(settings.headers, id) })}
          />
        </div>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="primary" disabled={saving} onClick={() => void save()}>
          Save
        </Button>
      </Dialog.Actions>
    </>
  )
}
