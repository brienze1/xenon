import { lazy, Suspense, useState } from 'react'
import type { ApiDocument, ApiOperation, ApiParameter, ApiSecurityScheme } from '@/platform/mappers/documentation'
import { languageFor } from '@/shared/lib/body-format'
import { CodeEditor } from '@/shared/ui/code-editor/CodeEditor'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { groupOperations } from '@/features/documentation/lib/operation-groups'
import styles from './ApiDocView.module.css'

const MarkdownView = lazy(() => import('@/shared/ui/markdown/MarkdownView'))

const UNTAGGED_TITLE = 'Requests'

const LOADING = <span className={styles.muted}>Loading…</span>

interface MarkdownProps {
  text: string
}

function Markdown({ text }: MarkdownProps) {
  return text.trim() === '' ? null : (
    <Suspense fallback={LOADING}>
      <MarkdownView text={text} />
    </Suspense>
  )
}

function describeScheme(scheme: ApiSecurityScheme): string {
  if (scheme.type === 'http') {
    return `HTTP ${scheme.scheme}`
  }
  if (scheme.type === 'apiKey') {
    return scheme.description || `API key in the ${scheme.in} “${scheme.paramName}”`
  }
  if (scheme.type === 'oauth2') {
    return `OAuth 2.0 (${scheme.flow})${scheme.scopes.length > 0 ? ` — scopes ${scheme.scopes.join(', ')}` : ''}`
  }
  return scheme.type
}

interface ParametersProps {
  parameters: ApiParameter[]
}

function Parameters({ parameters }: ParametersProps) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th scope="col">Name</th>
          <th scope="col">In</th>
          <th scope="col">Description</th>
          <th scope="col">Example</th>
        </tr>
      </thead>
      <tbody>
        {parameters.map((parameter) => (
          <tr key={`${parameter.in}:${parameter.name}`}>
            <td className={styles.code} translate="no">
              {parameter.name}
              {parameter.required ? <span className={styles.required}> required</span> : null}
            </td>
            <td>{parameter.variable ? 'path (variable)' : parameter.in}</td>
            <td>{parameter.description}</td>
            <td className={styles.code} translate="no">
              {parameter.example}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

interface ExampleProps {
  label: string
  mediaType: string
  value: string
}

function Example({ label, mediaType, value }: ExampleProps) {
  return (
    <div className={styles.example}>
      <span className={styles.muted} translate="no">
        {mediaType}
      </span>
      {value === '' ? <span className={styles.muted}>Empty</span> : <CodeEditor value={value} language={languageFor(mediaType)} label={label} className={styles.editor} />}
    </div>
  )
}

interface OperationDetailsProps {
  operation: ApiOperation
}

function OperationDetails({ operation }: OperationDetailsProps) {
  const title = operation.summary || `${operation.method} ${operation.path}`
  return (
    <div className={styles.details}>
      {operation.requestId === '' ? <p className={styles.note}>New request — it is created when you save.</p> : null}
      <Markdown text={operation.description} />
      {operation.server ? (
        <p className={styles.muted}>
          Server <code translate="no">{operation.server}</code>
        </p>
      ) : null}
      {operation.parameters.length > 0 ? (
        <section aria-label={`${title} parameters`}>
          <h5 className={styles.heading}>Parameters</h5>
          <Parameters parameters={operation.parameters} />
        </section>
      ) : null}
      {operation.body ? (
        <section aria-label={`${title} body`}>
          <h5 className={styles.heading}>Body</h5>
          {operation.body.fields.length > 0 ? (
            <Parameters
              parameters={operation.body.fields.map((field) => ({
                name: field.name,
                in: 'query',
                description: field.description,
                example: field.binary ? 'file' : field.example,
                required: false,
                variable: false,
              }))}
            />
          ) : (
            <Example label={`${title} body example`} mediaType={operation.body.mediaType} value={operation.body.example} />
          )}
        </section>
      ) : null}
      {operation.responses.length > 0 ? (
        <section aria-label={`${title} responses`}>
          <h5 className={styles.heading}>Responses</h5>
          {operation.responses.map((response) => (
            <div key={response.code} className={styles.response}>
              <span className={styles.status} data-class={String(response.code)[0]}>
                {response.code} {response.description}
              </span>
              {response.examples.map((example) => (
                <div key={example.id || example.name} className={styles.responseExample}>
                  <span className={styles.exampleName}>{example.name}</span>
                  {example.binary ? (
                    <span className={styles.muted}>Binary body</span>
                  ) : (
                    <Example label={`${title} ${example.name} example`} mediaType={example.mediaType} value={example.value} />
                  )}
                </div>
              ))}
            </div>
          ))}
        </section>
      ) : null}
      <p className={styles.muted}>Authentication: {operation.security.length > 0 ? operation.security.join(', ') : 'none'}</p>
    </div>
  )
}

interface OperationViewProps {
  operation: ApiOperation
}

function OperationView({ operation }: OperationViewProps) {
  const [open, setOpen] = useState(false)
  return (
    <details className={styles.operation} open={open} onToggle={(event) => setOpen(event.currentTarget.open)}>
      <summary className={styles.summary}>
        <MethodBadge method={operation.method} />
        <code className={styles.path} translate="no">
          {operation.path}
        </code>
        <span className={styles.name}>{operation.summary}</span>
        {operation.variant ? <span className={styles.tag}>Same method and path</span> : null}
        {operation.requestId === '' ? <span className={styles.tag}>New</span> : null}
      </summary>
      {open ? <OperationDetails operation={operation} /> : null}
    </details>
  )
}

interface ApiDocViewProps {
  document: ApiDocument
}

/** The OpenAPI document rendered offline with the app's colours: info, authentication, then operations by folder. */
export function ApiDocView({ document }: ApiDocViewProps) {
  const groups = groupOperations(document.operations)
  return (
    <article className={styles.doc} aria-label={`${document.title} API documentation`}>
      <header className={styles.header}>
        <h3 className={styles.title}>
          {document.title} <span className={styles.version}>{document.version}</span>
        </h3>
        {document.servers.length > 0 ? (
          <p className={styles.muted}>
            Servers{' '}
            {document.servers.map((server) => (
              <code key={server} className={styles.server} translate="no">
                {server}
              </code>
            ))}
          </p>
        ) : null}
        <Markdown text={document.description} />
      </header>
      {document.securitySchemes.length > 0 ? (
        <section aria-label="Authentication" className={styles.group}>
          <h4 className={styles.groupTitle}>Authentication</h4>
          <ul className={styles.schemes}>
            {document.securitySchemes.map((scheme) => (
              <li key={scheme.name}>
                <code translate="no">{scheme.name}</code> {describeScheme(scheme)}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
      {groups.length === 0 ? <p className={styles.muted}>This collection has no requests yet. Add operations in the YAML to create them.</p> : null}
      {groups.map((group) => (
        <section key={group.tag} aria-label={group.tag || UNTAGGED_TITLE} className={styles.group}>
          <h4 className={styles.groupTitle} translate="no">
            {group.tag || UNTAGGED_TITLE}
          </h4>
          {group.operations.map((operation, index) => (
            <OperationView key={operation.requestId || `${operation.method} ${operation.path} ${index}`} operation={operation} />
          ))}
        </section>
      ))}
    </article>
  )
}
