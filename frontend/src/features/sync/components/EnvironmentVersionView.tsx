import type { SyncEnvironmentVersion } from '@/platform/mappers/sync'
import { VersionRows } from './VersionRows'
import styles from './ConflictView.module.css'

interface EnvironmentVersionViewProps {
  environment: SyncEnvironmentVersion
}

export function EnvironmentVersionView({ environment }: EnvironmentVersionViewProps) {
  return (
    <div className={styles.version}>
      <p className={styles.environmentName} translate="no">
        {environment.name}
      </p>
      {environment.variables.length > 0 ? <VersionRows label="Variables" rows={environment.variables} /> : <p className={styles.muted}>No variables</p>}
    </div>
  )
}
