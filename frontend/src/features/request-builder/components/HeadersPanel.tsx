import type { AutoHeader, KeyValueRow } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'
import { useTableFields } from '@/features/request-builder/hooks/use-table-fields'
import { useVariableEditing } from '@/features/request-builder/hooks/use-variable-editing'
import { useAutoHeadersHidden } from '@/features/request-builder/hooks/use-auto-headers-hidden'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { formatNumber } from '@/shared/lib/number-format'
import styles from './HeadersPanel.module.css'

interface HeadersPanelProps {
  tabId: string
}

const NO_HEADERS: KeyValueRow[] = []
const NO_AUTO_HEADERS: AutoHeader[] = []

const HEADER_EXAMPLES: KeyValueExamples = {
  key: 'Authorization',
  value: 'Bearer abc123',
  description: 'Why this header is sent',
}

const HEADERS_HEADING = <PanelSection.Heading>Headers</PanelSection.Heading>

const EYE = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M2.06 12.35a1 1 0 0 1 0-.7 10.75 10.75 0 0 1 19.88 0 1 1 0 0 1 0 .7 10.75 10.75 0 0 1-19.88 0" />
    <circle cx="12" cy="12" r="3" />
  </svg>
)

const EYE_OFF = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c4.7 0 8.6 3 9.94 6.65a1 1 0 0 1 0 .7 10.75 10.75 0 0 1-1.44 2.49" />
    <path d="M14.08 14.16a3 3 0 0 1-4.24-4.24" />
    <path d="M17.48 17.5a10.75 10.75 0 0 1-15.42-5.15 1 1 0 0 1 0-.7 10.75 10.75 0 0 1 4.45-5.14" />
    <path d="m2 2 20 20" />
  </svg>
)

const AUTO_HEADERS_HEADING = <PanelSection.Heading>Auto-Generated Headers</PanelSection.Heading>

const AUTO_HEADERS_HEAD = (
  <thead>
    <tr>
      <th className={styles.checkColumn}>
        <span className="visually-hidden">Sent</span>
      </th>
      <th>Key</th>
      <th>Value</th>
    </tr>
  </thead>
)

export function HeadersPanel({ tabId }: HeadersPanelProps) {
  const headers = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.headers ?? NO_HEADERS)
  const autoHeaders = useRequestBuilderStore((state) => state.tabs[tabId]?.autoHeaders ?? NO_AUTO_HEADERS)
  const bulkText = useRequestBuilderStore((state) => state.tabs[tabId]?.bulk.headers)
  const updateHeaders = useRequestBuilderStore((state) => state.updateHeaders)
  const setAutoHeaderEnabled = useRequestBuilderStore((state) => state.setAutoHeaderEnabled)
  const setBulkMode = useRequestBuilderStore((state) => state.setBulkMode)
  const setBulkText = useRequestBuilderStore((state) => state.setBulkText)
  const editing = useVariableEditing(tabId)
  const { cellDecorations, onCellSelection } = useTableFields('headers', headers, editing)
  const [autoHidden, toggleAutoHidden] = useAutoHeadersHidden()
  const toggleLabel = autoHidden ? 'Show auto-generated headers' : 'Hide auto-generated headers'

  return (
    <div>
      <PanelSection.Root>
        <div className={styles.headingRow}>
          {HEADERS_HEADING}
          {autoHeaders.length > 0 ? (
            <Button
              variant="ghost"
              size="sm"
              aria-label={toggleLabel}
              aria-pressed={!autoHidden}
              title={toggleLabel}
              onClick={toggleAutoHidden}
            >
              {autoHidden ? EYE_OFF : EYE}
              {autoHidden ? <span>{formatNumber(autoHeaders.length, 0)} hidden</span> : null}
            </Button>
          ) : null}
        </div>
        <KeyValueTable
          examples={HEADER_EXAMPLES}
          label="Headers"
          rows={headers}
          mode={bulkText === undefined ? 'table' : 'bulk'}
          bulkText={bulkText ?? ''}
          onModeChange={(mode) => void setBulkMode(tabId, 'headers', mode === 'bulk')}
          onBulkTextChange={(text) => setBulkText(tabId, 'headers', text)}
          onRowChange={(id, patch) => updateHeaders(tabId, patchRow(headers, id, patch))}
          onRowAdd={(row) => updateHeaders(tabId, addRow(headers, row))}
          onRowRemove={(id) => updateHeaders(tabId, removeRow(headers, id))}
          cellDecorations={cellDecorations}
          completionSource={editing.completionSource}
          onCellSelection={onCellSelection}
        />
      </PanelSection.Root>
      {autoHeaders.length > 0 && !autoHidden ? (
        <PanelSection.Root>
          {AUTO_HEADERS_HEADING}
          <table aria-label="Auto-generated headers" className={styles.table}>
            {AUTO_HEADERS_HEAD}
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
