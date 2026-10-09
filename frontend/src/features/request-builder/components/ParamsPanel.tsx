import type { KeyValueRow, PathVariableRow } from '@/platform/mappers/request'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './ParamsPanel.module.css'

interface ParamsPanelProps {
  tabId: string
}

const NO_PARAMS: KeyValueRow[] = []
const NO_PATH_VARIABLES: PathVariableRow[] = []

const PARAM_EXAMPLES: KeyValueExamples = { key: 'page', value: '1', description: 'Page to fetch' }

const QUERY_PARAMS_HEADING = <PanelSection.Heading>Query Params</PanelSection.Heading>

const PATH_VARIABLES_HEADING = <PanelSection.Heading>Path Variables</PanelSection.Heading>

const PATH_VARIABLES_HEAD = (
  <thead>
    <tr>
      <th>Key</th>
      <th>Value</th>
      <th>Description</th>
    </tr>
  </thead>
)

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
      <PanelSection.Root>
        {QUERY_PARAMS_HEADING}
        <KeyValueTable
          examples={PARAM_EXAMPLES}
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
      </PanelSection.Root>
      {pathVariables.length > 0 ? (
        <PanelSection.Root>
          {PATH_VARIABLES_HEADING}
          <table aria-label="Path variables" className={styles.table}>
            {PATH_VARIABLES_HEAD}
            <tbody>
              {pathVariables.map((variable) => (
                <tr key={variable.id}>
                  <td translate="no">{variable.key}</td>
                  <td>
                    <input
                      name="path-variable-value"
                      autoComplete="off"
                      aria-label={`Value of ${variable.key}`}
                      className={styles.cell}
                      spellCheck={false}
                      value={variable.value}
                      onChange={(event) =>
                        updatePathVariables(
                          tabId,
                          patchRow(pathVariables, variable.id, {
                            value: event.target.value,
                          }),
                        )
                      }
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
                        updatePathVariables(
                          tabId,
                          patchRow(pathVariables, variable.id, {
                            description: event.target.value,
                          }),
                        )
                      }
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </PanelSection.Root>
      ) : null}
    </div>
  )
}
