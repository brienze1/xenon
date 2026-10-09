import type { KeyValueRow, PathVariableRow } from '@/platform/mappers/request'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './Panel.module.css'

interface ParamsPanelProps {
  tabId: string
}

const NO_PARAMS: KeyValueRow[] = []
const NO_PATH_VARIABLES: PathVariableRow[] = []

export function ParamsPanel({ tabId }: ParamsPanelProps) {
  const params = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.params ?? NO_PARAMS)
  const pathVariables = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.pathVariables ?? NO_PATH_VARIABLES)
  const bulkText = useRequestBuilderStore((state) => state.tabs[tabId]?.bulk.params)
  const updateParams = useRequestBuilderStore((state) => state.updateParams)
  const updatePathVariables = useRequestBuilderStore((state) => state.updatePathVariables)
  const setBulkMode = useRequestBuilderStore((state) => state.setBulkMode)
  const setBulkText = useRequestBuilderStore((state) => state.setBulkText)

  return (
    <div>
      <section className={styles.section}>
        <h3 className={styles.heading}>Query Params</h3>
        <KeyValueTable
          label="Query params"
          rows={params}
          mode={bulkText === undefined ? 'table' : 'bulk'}
          bulkText={bulkText ?? ''}
          onModeChange={(mode) => void setBulkMode(tabId, 'params', mode === 'bulk')}
          onBulkTextChange={(text) => setBulkText(tabId, 'params', text)}
          onRowChange={(id, patch) => updateParams(tabId, patchRow(params, id, patch))}
          onRowAdd={(row) => updateParams(tabId, addRow(params, row))}
          onRowRemove={(id) => updateParams(tabId, removeRow(params, id))}
        />
      </section>
      {pathVariables.length > 0 ? (
        <section className={styles.section}>
          <h3 className={styles.heading}>Path Variables</h3>
          <table aria-label="Path variables" className={styles.table}>
            <thead>
              <tr>
                <th>Key</th>
                <th>Value</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              {pathVariables.map((variable) => (
                <tr key={variable.id}>
                  <td>{variable.key}</td>
                  <td>
                    <input
                      name="path-variable-value"
                      autoComplete="off"
                      aria-label={`Value of ${variable.key}`}
                      className={styles.cell}
                      spellCheck={false}
                      value={variable.value}
                      onChange={(event) => updatePathVariables(tabId, patchRow(pathVariables, variable.id, { value: event.target.value }))}
                    />
                  </td>
                  <td>
                    <input
                      name="path-variable-description"
                      autoComplete="off"
                      aria-label={`Description of ${variable.key}`}
                      className={styles.cell}
                      value={variable.description}
                      onChange={(event) =>
                        updatePathVariables(tabId, patchRow(pathVariables, variable.id, { description: event.target.value }))
                      }
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ) : null}
    </div>
  )
}
