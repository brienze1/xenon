import { GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { VariableTable } from './VariableTable'
import styles from './EnvironmentTab.module.css'

const GLOBALS_HEADER = (
  <header className={styles.header}>
    <h2 className={styles.title}>Globals</h2>
  </header>
)

const GLOBALS_INTRO = (
  <p className={styles.note}>Globals apply to every request in this workspace, after the active environment and the collection.</p>
)

/** The main-area tab of the workspace globals. */
export function GlobalsTab() {
  return (
    <section className={styles.tab} aria-label="Globals">
      {GLOBALS_HEADER}
      {GLOBALS_INTRO}
      <VariableTable scope={GLOBALS_SCOPE} label="Global variables" draftLifetime="workspace" savedByTab />
    </section>
  )
}
