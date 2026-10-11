import styles from './ObjectTree.module.css'

type TaggedValue = { $type: string; value?: unknown }

function isTagged(value: unknown): value is TaggedValue {
  return typeof value === 'object' && value !== null && !Array.isArray(value) && typeof (value as TaggedValue).$type === 'string'
}

/** A logged value as text: strings quoted, special values (undefined, functions, dates…) as JavaScript shows them. */
export function previewValue(value: unknown): string {
  if (typeof value === 'string') {
    return JSON.stringify(value)
  }
  if (isTagged(value)) {
    return value.value === undefined ? value.$type : `${value.$type} ${String(previewValue(value.value))}`
  }
  if (Array.isArray(value)) {
    return `Array(${value.length})`
  }
  if (typeof value === 'object' && value !== null) {
    return `{…} ${Object.keys(value).length} keys`
  }
  return String(value)
}

function children(value: unknown): Array<[string, unknown]> | null {
  if (isTagged(value)) {
    return Array.isArray(value.value) ? value.value.map((item, index) => [String(index), item]) : null
  }
  if (Array.isArray(value)) {
    return value.map((item, index) => [String(index), item])
  }
  if (typeof value === 'object' && value !== null) {
    return Object.entries(value)
  }
  return null
}

interface ObjectTreeProps {
  label: string
  value: unknown
}

/** An expandable view of a logged object; nested levels open on demand. */
export function ObjectTree({ label, value }: ObjectTreeProps) {
  const nested = children(value)
  if (nested === null) {
    return (
      <div className={styles.leaf}>
        <span className={styles.key}>{label}</span>
        <span className={styles.value}>{previewValue(value)}</span>
      </div>
    )
  }
  return (
    <details className={styles.node}>
      <summary className={styles.summary}>
        <span className={styles.key}>{label}</span>
        <span className={styles.value}>{previewValue(value)}</span>
      </summary>
      <div className={styles.children}>
        {nested.map(([key, item]) => (
          <ObjectTree key={key} label={key} value={item} />
        ))}
      </div>
    </details>
  )
}
