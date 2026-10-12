import type { ItemSettings } from '@/platform/mappers/tree'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { useItemSettingsStore } from '@/features/item-settings/store'
import styles from './ItemTab.module.css'

const HEADER_EXAMPLES: KeyValueExamples = {
  key: 'X-Team',
  value: 'payments',
  description: 'Sent with every request',
}

const HEADERS_NOTE = (
  <p className={styles.note}>
    <span translate="no">Xenon</span> sends these with every request below unless a nearer level sets the same header.
  </p>
)

interface ItemHeadersProps {
  settings: ItemSettings
}

/** The headers a collection or folder adds to every request below it. */
export function ItemHeaders({ settings }: ItemHeadersProps) {
  const itemId = settings.id
  const bulkText = useItemSettingsStore((state) => state.items[itemId]?.bulkText ?? null)
  const edit = useItemSettingsStore((state) => state.edit)
  const setBulk = useItemSettingsStore((state) => state.setBulk)
  const editBulkText = useItemSettingsStore((state) => state.editBulkText)
  const headers = settings.headers

  return (
    <>
      {HEADERS_NOTE}
      <KeyValueTable
        examples={HEADER_EXAMPLES}
        label={`${settings.name} headers`}
        rows={headers}
        mode={bulkText === null ? 'table' : 'bulk'}
        bulkText={bulkText ?? ''}
        onModeChange={(mode) => void setBulk(itemId, mode === 'bulk')}
        onBulkTextChange={(text) => editBulkText(itemId, text)}
        onRowChange={(id, patch) => edit(itemId, { headers: patchRow(headers, id, patch) })}
        onRowAdd={(row) => edit(itemId, { headers: addRow(headers, row) })}
        onRowRemove={(id) => edit(itemId, { headers: removeRow(headers, id) })}
      />
    </>
  )
}
