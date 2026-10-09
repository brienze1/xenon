import type { AutoHeader, KeyValueRow } from '@/platform/mappers/request'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './HeadersPanel.module.css'

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
      <PanelSection.Root>
        <PanelSection.Heading>Headers</PanelSection.Heading>
        <KeyValueTable
          examples={{
            key: 'Authorization',
            value: 'Bearer abc123',
            description: 'Why this header is sent',
          }}
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
      </PanelSection.Root>
      {autoHeaders.length > 0 ? (
        <PanelSection.Root>
          <PanelSection.Heading>Auto-Generated Headers</PanelSection.Heading>
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
                      name="auto-header-enabled"
                      type="checkbox"
                      aria-label={`Send ${header.key}`}
                      checked={!header.disabled}
                      onChange={(event) => setAutoHeaderEnabled(tabId, header.key, event.target.checked)}
                    />
                  </td>
                  <td translate="no">{header.key}</td>
                  <td translate="no">{header.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </PanelSection.Root>
      ) : null}
    </div>
  )
}
