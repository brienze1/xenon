import type { OAuth2Param, OAuth2ParamTarget } from '@/platform/mappers/request'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import styles from './AuthForm.module.css'

interface OAuth2ParamsProps {
  label: string
  name: string
  params: OAuth2Param[]
  targets: OAuth2ParamTarget[]
  onChange: (params: OAuth2Param[]) => void
}

const TARGET_LABELS: Record<OAuth2ParamTarget, string> = { request_url: 'URL', request_body: 'Request body', request_header: 'Header' }

export function OAuth2Params({ label, name, params, targets, onChange }: OAuth2ParamsProps) {
  const update = (index: number, change: Partial<OAuth2Param>) => onChange(params.map((param, at) => (at === index ? { ...param, ...change } : param)))
  return (
    <fieldset className={styles.params}>
      <legend>{label}</legend>
      {params.map((param, index) => (
        <div key={index} className={styles.paramRow}>
          <input
            type="checkbox"
            aria-label={`Enable ${label} row ${index + 1}`}
            checked={param.enabled}
            onChange={(event) => update(index, { enabled: event.target.checked })}
          />
          <input
            name={`${name}-key-${index}`}
            aria-label={`${label} key ${index + 1}`}
            autoComplete="off"
            spellCheck={false}
            className={styles.input}
            value={param.key}
            onChange={(event) => update(index, { key: event.target.value })}
          />
          <input
            name={`${name}-value-${index}`}
            aria-label={`${label} value ${index + 1}`}
            autoComplete="off"
            spellCheck={false}
            className={styles.input}
            value={param.value}
            onChange={(event) => update(index, { value: event.target.value })}
          />
          <select
            aria-label={`${label} send as ${index + 1}`}
            className={styles.select}
            value={param.sendAs}
            onChange={(event) => update(index, { sendAs: event.target.value as OAuth2ParamTarget })}
          >
            {targets.map((target) => (
              <option key={target} value={target}>
                {TARGET_LABELS[target]}
              </option>
            ))}
          </select>
          <IconButton label={`Remove ${label} row ${index + 1}`} onClick={() => onChange(params.filter((_, at) => at !== index))}>
            ×
          </IconButton>
        </div>
      ))}
      <Button size="sm" variant="ghost" onClick={() => onChange([...params, { key: '', value: '', enabled: true, sendAs: targets[0] ?? 'request_body' }])}>
        Add Parameter
      </Button>
    </fieldset>
  )
}
