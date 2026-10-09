import { useResponseViewerStore } from '@/features/response-viewer/store'
import { BODY_MODES, RAW_LANGUAGES, type BodyMode, type FormDataRow, type RawLanguage } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { CodeEditor } from '@/shared/ui/code-editor/CodeEditor'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { KeyValueTable } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '../lib/rows'
import { useRequestBuilderStore } from '../store'
import styles from './Panel.module.css'

interface BodyPanelProps {
  tabId: string
}

const MODE_LABELS: Record<BodyMode, string> = {
  none: 'none',
  formdata: 'form-data',
  urlencoded: 'x-www-form-urlencoded',
  raw: 'raw',
  binary: 'binary',
  graphql: 'GraphQL',
}

const LANGUAGE_LABELS: Record<RawLanguage, string> = {
  json: 'JSON',
  xml: 'XML',
  text: 'Text',
  html: 'HTML',
  javascript: 'JavaScript',
}

const EMPTY_FORM_ROW: Omit<FormDataRow, 'id'> = { key: '', value: '', description: '', enabled: true, type: 'text', src: '' }

function useMissingFile(tabId: string): string | null {
  return useResponseViewerStore((state) => {
    const error = state.entries[tabId]?.result?.error
    return error?.kind === 'missing_file' ? error.message.replace(/^File not found: /, '') : null
  })
}

export function BodyPanel({ tabId }: BodyPanelProps) {
  const body = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.body)
  const bulkText = useRequestBuilderStore((state) => state.tabs[tabId]?.bulk.urlencoded)
  const updateBody = useRequestBuilderStore((state) => state.updateBody)
  const setBulkMode = useRequestBuilderStore((state) => state.setBulkMode)
  const setBulkText = useRequestBuilderStore((state) => state.setBulkText)
  const chooseBodyFile = useRequestBuilderStore((state) => state.chooseBodyFile)
  const missingFile = useMissingFile(tabId)
  if (!body) {
    return null
  }

  const formdata = body.formdata
  const urlencoded = body.urlencoded

  return (
    <section className={styles.section}>
      <div role="radiogroup" aria-label="Body type" className={styles.modes}>
        {BODY_MODES.map((mode) => (
          <label key={mode} className={styles.toggle}>
            <input type="radio" name={`${tabId}-body-mode`} checked={body.mode === mode} onChange={() => updateBody(tabId, { mode })} />
            {MODE_LABELS[mode]}
          </label>
        ))}
        {body.mode === 'raw' ? (
          <select
            aria-label="Raw language"
            className={styles.select}
            value={body.raw.language}
            onChange={(event) => updateBody(tabId, { raw: { ...body.raw, language: event.target.value as RawLanguage } })}
          >
            {RAW_LANGUAGES.map((language) => (
              <option key={language} value={language}>
                {LANGUAGE_LABELS[language]}
              </option>
            ))}
          </select>
        ) : null}
      </div>

      {body.mode === 'none' ? <p className={styles.note}>This request does not have a body.</p> : null}

      {body.mode === 'raw' ? (
        <CodeEditor
          label="Raw body"
          className={styles.editor}
          language={body.raw.language}
          value={body.raw.content}
          onChange={(content) => updateBody(tabId, { raw: { ...body.raw, content } })}
        />
      ) : null}

      {body.mode === 'urlencoded' ? (
        <KeyValueTable
          label="Url-encoded body"
          rows={urlencoded}
          mode={bulkText === undefined ? 'table' : 'bulk'}
          bulkText={bulkText ?? ''}
          onModeChange={(mode) => void setBulkMode(tabId, 'urlencoded', mode === 'bulk')}
          onBulkTextChange={(text) => setBulkText(tabId, 'urlencoded', text)}
          onRowChange={(id, patch) => updateBody(tabId, { urlencoded: patchRow(urlencoded, id, patch) })}
          onRowAdd={(row) => updateBody(tabId, { urlencoded: addRow(urlencoded, row) })}
          onRowRemove={(id) => updateBody(tabId, { urlencoded: removeRow(urlencoded, id) })}
        />
      ) : null}

      {body.mode === 'formdata' ? (
        <table aria-label="Form-data body" className={styles.table}>
          <thead>
            <tr>
              <th className={styles.checkColumn}>
                <span className="visually-hidden">Enabled</span>
              </th>
              <th>Key</th>
              <th>Type</th>
              <th>Value</th>
              <th className={styles.checkColumn}>
                <span className="visually-hidden">Actions</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {formdata.map((row) => (
              <tr key={row.id}>
                <td>
                  <input
                    type="checkbox"
                    aria-label={`Enable ${row.key || 'row'}`}
                    checked={row.enabled}
                    onChange={(event) => updateBody(tabId, { formdata: patchRow(formdata, row.id, { enabled: event.target.checked }) })}
                  />
                </td>
                <td>
                  <input
                    aria-label="Key"
                    className={styles.cell}
                    value={row.key}
                    onChange={(event) => updateBody(tabId, { formdata: patchRow(formdata, row.id, { key: event.target.value }) })}
                  />
                </td>
                <td>
                  <select
                    aria-label={`Type of ${row.key || 'row'}`}
                    className={styles.select}
                    value={row.type}
                    onChange={(event) =>
                      updateBody(tabId, { formdata: patchRow(formdata, row.id, { type: event.target.value as FormDataRow['type'] }) })
                    }
                  >
                    <option value="text">Text</option>
                    <option value="file">File</option>
                  </select>
                </td>
                <td>
                  {row.type === 'file' ? (
                    <div className={styles.file}>
                      <Button size="sm" onClick={() => void chooseBodyFile(tabId, row.id)}>
                        Select file
                      </Button>
                      <span className={styles.filePath}>{row.src}</span>
                      {missingFile !== null && row.src !== '' && missingFile.endsWith(row.src.replace(/^.*[\\/]/, '')) ? (
                        <span className={styles.missing}>missing</span>
                      ) : null}
                    </div>
                  ) : (
                    <input
                      aria-label="Value"
                      className={styles.cell}
                      value={row.value}
                      onChange={(event) => updateBody(tabId, { formdata: patchRow(formdata, row.id, { value: event.target.value }) })}
                    />
                  )}
                </td>
                <td>
                  <IconButton label={`Remove ${row.key || 'row'}`} onClick={() => updateBody(tabId, { formdata: removeRow(formdata, row.id) })}>
                    ×
                  </IconButton>
                </td>
              </tr>
            ))}
            <tr>
              <td colSpan={5}>
                <Button size="sm" variant="ghost" onClick={() => updateBody(tabId, { formdata: addRow(formdata, EMPTY_FORM_ROW) })}>
                  Add field
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      ) : null}

      {body.mode === 'binary' ? (
        <div className={styles.file}>
          <Button onClick={() => void chooseBodyFile(tabId)}>Select file</Button>
          <span className={styles.filePath}>{body.binary.src || 'No file selected'}</span>
          {missingFile !== null && body.binary.src !== '' ? (
            <span role="status" className={styles.missing}>
              File missing — choose it again
            </span>
          ) : null}
        </div>
      ) : null}

      {body.mode === 'graphql' ? (
        <>
          <h3 className={styles.heading}>Query</h3>
          <CodeEditor
            label="GraphQL query"
            className={styles.editor}
            language="text"
            value={body.graphql.query}
            onChange={(query) => updateBody(tabId, { graphql: { ...body.graphql, query } })}
          />
          <h3 className={styles.heading}>GraphQL Variables</h3>
          <CodeEditor
            label="GraphQL variables"
            className={styles.editor}
            language="json"
            value={body.graphql.variables}
            onChange={(variables) => updateBody(tabId, { graphql: { ...body.graphql, variables } })}
          />
        </>
      ) : null}
    </section>
  )
}
