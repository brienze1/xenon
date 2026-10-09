import { EmptyState } from '@/shared/ui/empty-state/EmptyState'

export function ConsoleDrawer() {
  return (
    <EmptyState.Root label="Console is empty" size="compact">
      <EmptyState.Art variant="console" />
      <EmptyState.Title>Console arrives with scripts</EmptyState.Title>
      <EmptyState.Description>Logs from pre-request and test scripts will show up here.</EmptyState.Description>
    </EmptyState.Root>
  )
}
