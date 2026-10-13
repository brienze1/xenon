import type { SyncItemVersion } from '@/platform/mappers/sync'
import { EnvironmentVersionView } from './EnvironmentVersionView'
import { RequestVersionView } from './RequestVersionView'
import styles from './ConflictView.module.css'

interface ItemVersionViewProps {
  title: string
  version: SyncItemVersion | null
}

function VersionBody({ version }: { version: SyncItemVersion | null }) {
  if (version === null) {
    return <p className={styles.muted}>Deleted on this side</p>
  }
  if (version.request !== null) {
    return <RequestVersionView request={version.request} />
  }
  if (version.environment !== null) {
    return <EnvironmentVersionView environment={version.environment} />
  }
  return (
    <pre className={styles.text} translate="no">
      {version.text}
    </pre>
  )
}

export function ItemVersionView({ title, version }: ItemVersionViewProps) {
  return (
    <section className={styles.side} aria-label={title}>
      <h4 className={styles.sideTitle}>{title}</h4>
      <VersionBody version={version} />
    </section>
  )
}
