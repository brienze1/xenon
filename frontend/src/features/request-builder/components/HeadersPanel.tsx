import type { AutoHeader, KeyValueRow } from '@/platform/mappers/request'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '../lib/rows'
import { useRequestBuilderStore } from '../store'
import styles from './Panel.module.css'

interface HeadersPanelProps {
  tabId: string
}

const NO_HEADERS: KeyValueRow[] = []
const NO_AUTO_HEADERS: AutoHeader[] = []

export function HeadersPanel({ tabId }: HeadersPanelProps) {
  const headers = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.headers ?? NO_HEADERS)
  const autoHeaders = useRequestBuilderStore((state) => state.tabs[tabId]?.autoHeaders ?? NO_AUTO_HEADERS)
  const bulkText = useRequestBuilderStore((state) => state.tabs[tabId]?.bulk.headers)
  const updateHeaders = useRequestBuilderStore((state) => state.updateHeaders)
  const setAutoHeaderEnabled = useRequestBuilderStore((state) => state.setAutoHeaderEnabled)
  const setBulkMode = useRequestBuilderStore((state) => state.setBulkMode)
  const setBulkText = useRequestBuilderStore((state) => state.setBulkText)

  return (
    <div>
      <section className={styles.section}>
        <h3 className={styles.heading}>Headers</h3>
        <KeyValueTable
          label="Headers"
          rows={headers}
          mode={bulkText === undefined ? 'table' : 'bulk'}
          bulkText={bulkText ?? ''}
          onModeChange={(mode) => void setBulkMode(tabId, 'headers', mode === 'bulk')}
          onBulkTextChange={(text) => setBulkText(tabId, 'headers', text)}
          onRowChange={(id, patch) => updateHeaders(tabId, patchRow(headers, id, patch))}
          onRowAdd={(row) => updateHeaders(tabId, addRow(headers, row))}
          onRowRemove={(id) => updateHeaders(tabId, removeRow(headers, id))}
        />
      </section>
      {autoHeaders.length > 0 ? (
        <section className={styles.section}>
          <h3 className={styles.heading}>Auto-generated headers</h3>
          <table aria-label="Auto-generated headers" className={styles.table}>
            <thead>
              <tr>
                <th className={styles.checkColumn}>
                  <span className="visually-hidden">Sent</span>
                </th>
                <th>Key</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {autoHeaders.map((header) => (
                <tr key={header.key} className={styles.auto}>
                  <td>
                    <input
                      type="checkbox"
                      aria-label={`Send ${header.key}`}
                      checked={!header.disabled}
                      onChange={(event) => setAutoHeaderEnabled(tabId, header.key, event.target.checked)}
                    />
                  </td>
                  <td>{header.key}</td>
                  <td>{header.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ) : null}
    </div>
  )
}
