import { lazy, Suspense } from 'react'
import { BODY_MODES, RAW_LANGUAGES, type BodyMode, type FormDataRow, type KeyValueRow, type RawLanguage } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { KeyValueTable, type KeyValueExamples } from '@/shared/ui/key-value-table/KeyValueTable'
import { addRow, patchRow, removeRow } from '@/shared/lib/rows'
import { SingleLineEditor } from '@/shared/ui/single-line-editor/SingleLineEditor'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'
import { useMissingFile } from '@/features/request-builder/hooks/use-missing-file'
import { useTableFields } from '@/features/request-builder/hooks/use-table-fields'
import { useVariableEditing } from '@/features/request-builder/hooks/use-variable-editing'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { beautify, canBeautify } from '@/shared/lib/beautify'
import styles from './BodyPanel.module.css'

const CodeEditor = lazy(() => import('@/shared/ui/code-editor/CodeEditor').then((module) => ({ default: module.CodeEditor })))

interface BodyPanelProps {
  tabId: string
}

const MODE_LABELS: Record<BodyMode, string> = {
  none: 'None',
  formdata: 'form-data',
  urlencoded: 'x-www-form-urlencoded',
  raw: 'Raw',
  binary: 'Binary',
  graphql: 'GraphQL',
}

const TECHNICAL_MODES: ReadonlySet<BodyMode> = new Set(['formdata', 'urlencoded', 'graphql'])

const LANGUAGE_LABELS: Record<RawLanguage, string> = {
  json: 'JSON',
  xml: 'XML',
  text: 'Text',
  html: 'HTML',
  javascript: 'JavaScript',
}

const TECHNICAL_LANGUAGES: ReadonlySet<RawLanguage> = new Set(['json', 'xml', 'html', 'javascript'])

const MISSING_FILE = (
  <span role="status" className={styles.missing}>
    <span translate="no">Xenon</span> can’t find this file — select it again
  </span>
)

const REPICK_FILE = (
  <span role="status" className={styles.missing}>
    Imported from another machine — select the file again before sending
  </span>
)

const URLENCODED_EXAMPLES: KeyValueExamples = {
  key: 'grant_type',
  value: 'client_credentials',
  description: 'What this field carries',
}

const REMOVE_GLYPH = <span aria-hidden="true">×</span>

const LOADING_EDITOR = (
  <p role="status" className={styles.loading}>
    Loading editor…
  </p>
)

const NO_BODY_NOTE = <PanelSection.Note>This request does not have a body.</PanelSection.Note>

const FORM_DATA_HEAD = (
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
)

const FORM_DATA_TYPE_OPTIONS = (
  <>
    <option value="text">Text</option>
    <option value="file">File</option>
  </>
)

const GRAPHQL_QUERY_HEADING = <PanelSection.Heading>Query</PanelSection.Heading>

const GRAPHQL_VARIABLES_HEADING = (
  <PanelSection.Heading>
    <span translate="no">GraphQL</span> Variables
  </PanelSection.Heading>
)

const BODY_TYPE_LEGEND = <legend className="visually-hidden">Body type</legend>

const DIRECTORY_PREFIX = /^.*[\\/]/

const NO_ROWS: KeyValueRow[] = []

const EMPTY_FORM_ROW: Omit<FormDataRow, 'id'> = {
  key: '',
  value: '',
  description: '',
  enabled: true,
  type: 'text',
  src: '',
  srcState: 'ok',
}

export function BodyPanel({ tabId }: BodyPanelProps) {
  const body = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.body)
  const bulkText = useRequestBuilderStore((state) => state.tabs[tabId]?.bulk.urlencoded)
  const updateBody = useRequestBuilderStore((state) => state.updateBody)
  const setBulkMode = useRequestBuilderStore((state) => state.setBulkMode)
  const setBulkText = useRequestBuilderStore((state) => state.setBulkText)
  const chooseBodyFile = useRequestBuilderStore((state) => state.chooseBodyFile)
  const missingFile = useMissingFile(tabId)
  const editing = useVariableEditing(tabId)
  const { cellDecorations, onCellSelection } = useTableFields('body.urlencoded', body?.urlencoded ?? NO_ROWS, editing)
  if (!body) {
    return null
  }

  const formdata = body.formdata
  const urlencoded = body.urlencoded

  return (
    <PanelSection.Root>
      <div className={styles.modes}>
        <fieldset className={styles.modeGroup}>
          {BODY_TYPE_LEGEND}
          {BODY_MODES.map((mode) => (
            <label key={mode} className={styles.toggle}>
              <input type="radio" name={`${tabId}-body-mode`} checked={body.mode === mode} onChange={() => updateBody(tabId, { mode })} />
              <span translate={TECHNICAL_MODES.has(mode) ? 'no' : undefined}>{MODE_LABELS[mode]}</span>
            </label>
          ))}
        </fieldset>
        {body.mode === 'raw' ? (
          <select
            aria-label="Raw language"
            name="raw-language"
            className={styles.select}
            value={body.raw.language}
            onChange={(event) =>
              updateBody(tabId, {
                raw: {
                  ...body.raw,
                  language: event.target.value as RawLanguage,
                },
              })
            }
          >
            {RAW_LANGUAGES.map((language) => (
              <option key={language} value={language} translate={TECHNICAL_LANGUAGES.has(language) ? 'no' : undefined}>
                {LANGUAGE_LABELS[language]}
              </option>
            ))}
          </select>
        ) : null}
        {body.mode === 'raw' && canBeautify(body.raw.language) ? (
          <Button
            size="sm"
            variant="ghost"
            className={styles.beautify}
            disabled={body.raw.language === 'json' && beautify(body.raw.content, 'json') === null}
            onClick={() => {
              const content = beautify(body.raw.content, body.raw.language)
              if (content !== null && content !== body.raw.content) {
                updateBody(tabId, { raw: { ...body.raw, content } })
              }
            }}
          >
            Beautify
          </Button>
        ) : null}
      </div>

      {body.mode === 'none' ? NO_BODY_NOTE : null}

      {body.mode === 'raw' ? (
        <Suspense fallback={LOADING_EDITOR}>
          <CodeEditor
            label="Raw body"
            className={styles.editor}
            checkSyntax
            language={body.raw.language}
            value={body.raw.content}
            decorations={editing.decorationsOf('body.raw')}
            completionSource={editing.completionSource}
            onChange={(content) => updateBody(tabId, { raw: { ...body.raw, content } })}
          />
        </Suspense>
      ) : null}

      {body.mode === 'urlencoded' ? (
        <KeyValueTable
          examples={URLENCODED_EXAMPLES}
          label="Url-encoded body"
          rows={urlencoded}
          mode={bulkText === undefined ? 'table' : 'bulk'}
          bulkText={bulkText ?? ''}
          onModeChange={(mode) => void setBulkMode(tabId, 'urlencoded', mode === 'bulk')}
          onBulkTextChange={(text) => setBulkText(tabId, 'urlencoded', text)}
          onRowChange={(id, patch) => updateBody(tabId, { urlencoded: patchRow(urlencoded, id, patch) })}
          onRowAdd={(row) => updateBody(tabId, { urlencoded: addRow(urlencoded, row) })}
          onRowRemove={(id) => updateBody(tabId, { urlencoded: removeRow(urlencoded, id) })}
          cellDecorations={cellDecorations}
          completionSource={editing.completionSource}
          onCellSelection={onCellSelection}
        />
      ) : null}

      {body.mode === 'formdata' ? (
        <table aria-label="Form-data body" className={styles.table}>
          {FORM_DATA_HEAD}
          <tbody>
            {formdata.map((row, index) => (
              <tr key={row.id}>
                <td>
                  <input
                    type="checkbox"
                    aria-label={`Enable ${row.key || 'row'}`}
                    name="formdata-enabled"
                    checked={row.enabled}
                    onChange={(event) =>
                      updateBody(tabId, {
                        formdata: patchRow(formdata, row.id, {
                          enabled: event.target.checked,
                        }),
                      })
                    }
                  />
                </td>
                <td>
                  <SingleLineEditor
                    label="Key"
                    className={styles.codeCell}
                    value={row.key}
                    decorations={editing.decorationsOf(`body.formdata.${index}.key`)}
                    completionSource={editing.completionSource}
                    onSelectionChange={(selection) => editing.selectIn(`body.formdata.${index}.key`, selection)}
                    onChange={(key) => updateBody(tabId, { formdata: patchRow(formdata, row.id, { key }) })}
                  />
                </td>
                <td>
                  <select
                    aria-label={`Type of ${row.key || 'row'}`}
                    name="formdata-type"
                    className={styles.select}
                    value={row.type}
                    onChange={(event) =>
                      updateBody(tabId, {
                        formdata: patchRow(formdata, row.id, {
                          type: event.target.value as FormDataRow['type'],
                        }),
                      })
                    }
                  >
                    {FORM_DATA_TYPE_OPTIONS}
                  </select>
                </td>
                <td>
                  {row.type === 'file' ? (
                    <div className={styles.file}>
                      <Button size="sm" onClick={() => void chooseBodyFile(tabId, row.id)}>
                        Select File
                      </Button>
                      <span className={styles.filePath} translate="no">
                        {row.src}
                      </span>
                      {row.srcState === 'repick' ? REPICK_FILE : null}
                      {missingFile !== null && row.src !== '' && missingFile.endsWith(row.src.replace(DIRECTORY_PREFIX, '')) ? MISSING_FILE : null}
                    </div>
                  ) : (
                    <SingleLineEditor
                      label="Value"
                      className={styles.codeCell}
                      value={row.value}
                      decorations={editing.decorationsOf(`body.formdata.${index}.value`)}
                      completionSource={editing.completionSource}
                      onSelectionChange={(selection) => editing.selectIn(`body.formdata.${index}.value`, selection)}
                      onChange={(value) => updateBody(tabId, { formdata: patchRow(formdata, row.id, { value }) })}
                    />
                  )}
                </td>
                <td>
                  <IconButton
                    label={`Remove ${row.key || 'Row'}`}
                    onClick={() =>
                      updateBody(tabId, {
                        formdata: removeRow(formdata, row.id),
                      })
                    }
                  >
                    {REMOVE_GLYPH}
                  </IconButton>
                </td>
              </tr>
            ))}
            <tr>
              <td colSpan={5}>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    updateBody(tabId, {
                      formdata: addRow(formdata, EMPTY_FORM_ROW),
                    })
                  }
                >
                  Add Field
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      ) : null}

      {body.mode === 'binary' ? (
        <div className={styles.file}>
          <Button onClick={() => void chooseBodyFile(tabId)}>Select File</Button>
          <span className={styles.filePath} translate={body.binary.src ? 'no' : undefined}>
            {body.binary.src || 'No file selected'}
          </span>
          {body.binary.srcState === 'repick' ? REPICK_FILE : null}
          {missingFile !== null && body.binary.src !== '' ? MISSING_FILE : null}
        </div>
      ) : null}

      {body.mode === 'graphql' ? (
        <Suspense fallback={LOADING_EDITOR}>
          {GRAPHQL_QUERY_HEADING}
          <CodeEditor
            label="GraphQL query"
            className={styles.editor}
            language="text"
            value={body.graphql.query}
            decorations={editing.decorationsOf('body.graphql.query')}
            completionSource={editing.completionSource}
            onChange={(query) => updateBody(tabId, { graphql: { ...body.graphql, query } })}
          />
          {GRAPHQL_VARIABLES_HEADING}
          <CodeEditor
            label="GraphQL variables"
            className={styles.editor}
            language="json"
            checkSyntax
            value={body.graphql.variables}
            decorations={editing.decorationsOf('body.graphql.variables')}
            completionSource={editing.completionSource}
            onChange={(variables) => updateBody(tabId, { graphql: { ...body.graphql, variables } })}
          />
        </Suspense>
      ) : null}
    </PanelSection.Root>
  )
}
