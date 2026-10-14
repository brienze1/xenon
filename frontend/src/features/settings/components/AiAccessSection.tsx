import { useId } from 'react'
import { MAX_AI_PORT, MIN_AI_PORT } from '@/platform/mappers/settings'
import type { AIAccessStatus } from '@/platform/mappers/ai-access'
import { CopyButton } from '@/shared/ui/copy-button/CopyButton'
import { useSettingsStore } from '@/features/settings/store'
import { NumberField } from './NumberField'
import styles from './SettingsSection.module.css'

export const AI_PORT_FIELD = 'ai-port'

const SKIP_WARNING = 'The AI then sends to environments marked “Confirm before send” without asking you.'

const RESIDUAL_RISKS = (
  <ul className={styles.note} aria-label="Residual risks">
    <li>
      <span translate="no">Xenon</span> masks every secret value it knows, but a script that transforms a secret (for example reverses or hashes it) makes a
      value it can’t recognise, which may reach the AI.
    </li>
    <li>The AI can send a request, with your secrets filled in, to a destination it controls.</li>
  </ul>
)

function statusText(status: AIAccessStatus | null): string {
  if (status === null) {
    return 'Checking…'
  }
  if (status.state === 'running') {
    const clients = status.clients === 1 ? ' · 1 client connected' : status.clients > 1 ? ` · ${status.clients} clients connected` : ''
    return `Running at ${status.address ?? ''}${clients}`
  }
  if (status.state === 'error') {
    return status.reason ?? 'The server could not start.'
  }
  return 'Off'
}

interface CodeBlockProps {
  label: string
  copyLabel: string
  text: string
  onCopy: (text: string) => Promise<boolean>
}

function CodeBlock({ label, copyLabel, text, onCopy }: CodeBlockProps) {
  const labelId = useId()
  return (
    <div className={styles.field}>
      <span id={labelId} className={styles.codeLabel}>
        {label}
      </span>
      <div className={styles.code}>
        <pre className={styles.pre} aria-labelledby={labelId} translate="no">
          {text}
        </pre>
        <CopyButton label={copyLabel} disabled={text === ''} onCopy={() => onCopy(text)} />
      </div>
    </div>
  )
}

export function AiAccessSection() {
  const enabled = useSettingsStore((state) => state.draft.aiEnabled)
  const port = useSettingsStore((state) => state.draft.aiPort)
  const skipApprovals = useSettingsStore((state) => state.draft.skipApprovals)
  const portError = useSettingsStore((state) => state.fieldErrors.aiPort)
  const status = useSettingsStore((state) => state.aiStatus)
  const edit = useSettingsStore((state) => state.edit)
  const copy = useSettingsStore((state) => state.copy)
  const warningId = useId()

  return (
    <fieldset className={styles.section}>
      <legend className={styles.legend}>AI access (MCP)</legend>
      <p role="status" aria-label="AI access status" className={styles.status}>
        <span className={styles.statusDot} data-state={status?.state ?? 'off'} aria-hidden="true" />
        <span translate="no">{statusText(status)}</span>
      </p>
      <label className={styles.toggle}>
        <input type="checkbox" role="switch" className={styles.switch} checked={enabled} onChange={(event) => edit({ aiEnabled: event.target.checked })} />
        AI Access
      </label>
      <NumberField
        label="Port"
        name={AI_PORT_FIELD}
        value={port}
        min={MIN_AI_PORT}
        max={MAX_AI_PORT}
        hint="AI clients connect to this port on 127.0.0.1 only. Changes apply when you save."
        error={portError}
        onChange={(aiPort) => edit({ aiPort })}
      />
      <div className={styles.field}>
        <label className={styles.toggle}>
          <input
            type="checkbox"
            className={styles.checkbox}
            aria-describedby={warningId}
            checked={skipApprovals}
            onChange={(event) => edit({ skipApprovals: event.target.checked })}
          />
          Dangerously skip approvals
        </label>
        <p id={warningId} className={styles.warning}>
          {SKIP_WARNING} Secrets stay write-only and masked either way.
        </p>
      </div>
      <CodeBlock label="Stdio command" copyLabel="Copy stdio command" text={status?.stdioCommand ?? ''} onCopy={copy} />
      <CodeBlock label="HTTP client config" copyLabel="Copy HTTP config" text={status?.httpConfig ?? ''} onCopy={copy} />
      <CodeBlock label="Stdio client config" copyLabel="Copy stdio config" text={status?.stdioConfig ?? ''} onCopy={copy} />
      {RESIDUAL_RISKS}
    </fieldset>
  )
}
