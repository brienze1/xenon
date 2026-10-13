import type { RequestDraft } from '@/platform/mappers/request'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { VersionRows } from './VersionRows'
import styles from './ConflictView.module.css'

interface RequestVersionViewProps {
  request: RequestDraft
}

function bodyText(request: RequestDraft): string {
  switch (request.body.mode) {
    case 'raw':
      return request.body.raw.content
    case 'graphql':
      return request.body.graphql.query
    case 'binary':
      return request.body.binary.src
    default:
      return ''
  }
}

export function RequestVersionView({ request }: RequestVersionViewProps) {
  const body = bodyText(request)
  const formRows = request.body.mode === 'formdata' ? request.body.formdata : request.body.mode === 'urlencoded' ? request.body.urlencoded : []
  return (
    <div className={styles.version}>
      <p className={styles.requestLine}>
        <MethodBadge method={request.method} />
        <span className={styles.url} translate="no">
          {request.url || '(no URL)'}
        </span>
      </p>
      <dl className={styles.facts}>
        <dt>Auth</dt>
        <dd>{request.auth.type}</dd>
        <dt>Body</dt>
        <dd>{request.body.mode}</dd>
      </dl>
      <VersionRows label="Params" rows={request.params} />
      <VersionRows label="Headers" rows={request.headers} />
      <VersionRows label="Body fields" rows={formRows} />
      {body !== '' ? (
        <pre className={styles.text} aria-label="Body" translate="no">
          {body}
        </pre>
      ) : null}
      {request.scripts.preRequest !== '' ? (
        <pre className={styles.text} aria-label="Pre-request script" translate="no">
          {request.scripts.preRequest}
        </pre>
      ) : null}
      {request.scripts.postResponse !== '' ? (
        <pre className={styles.text} aria-label="Post-response script" translate="no">
          {request.scripts.postResponse}
        </pre>
      ) : null}
    </div>
  )
}
