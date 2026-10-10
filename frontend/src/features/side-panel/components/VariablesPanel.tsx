import { useCallback } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { VariableScope } from '@/platform/mappers/inspection'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { MoveToSecret } from './MoveToSecret'
import { VariableItem } from './VariableItem'
import styles from './VariablesPanel.module.css'

interface VariablesPanelProps {
  tabId: string
}

const VARIABLES_HEADING = <h2 className={styles.title}>Variables</h2>

const NO_VARIABLES = (
  <EmptyState.Root label="No variables" size="compact">
    <EmptyState.Description>This request uses no {'{{variables}}'}. Type {'{{'} in a field to pick one.</EmptyState.Description>
  </EmptyState.Root>
)

function sourceLabel(scope: VariableScope, environmentName: string | null): string {
  switch (scope) {
    case 'environment':
      return environmentName ? `Environment: ${environmentName}` : 'Environment'
    case 'collection':
      return 'Collection'
    case 'globals':
      return 'Globals'
    case 'dynamic':
      return 'Dynamic variable'
    default:
      return 'Not defined in any scope'
  }
}

/** Every variable the request uses (inherited auth and headers included) with value, source and status. */
export function VariablesPanel({ tabId }: VariablesPanelProps) {
  const inspection = useRequestBuilderStore((state) => state.tabs[tabId]?.inspection ?? null)
  const inspect = useRequestBuilderStore((state) => state.inspect)
  const environmentName = useEnvironmentsStore(
    (state) => state.environments.find((environment) => environment.id === state.activeId)?.name ?? null,
  )
  const refresh = useCallback(() => void inspect(tabId), [inspect, tabId])

  return (
    <section className={styles.panel} aria-label="Variables">
      {VARIABLES_HEADING}
      {inspection ? <MoveToSecret tabId={tabId} scopes={inspection.scopes} /> : null}
      {inspection === null || inspection.variables.length === 0 ? (
        NO_VARIABLES
      ) : (
        <ul className={styles.list}>
          {inspection.variables.map((variable) => (
            <VariableItem
              key={variable.name}
              variable={variable}
              scopes={inspection.scopes}
              scopeLabel={sourceLabel(variable.scope, environmentName)}
              onChanged={refresh}
            />
          ))}
        </ul>
      )}
    </section>
  )
}
