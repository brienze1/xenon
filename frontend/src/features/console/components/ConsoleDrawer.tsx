import { EmptyState } from '@/shared/ui/empty-state/EmptyState'

const CONSOLE_PLACEHOLDER = (
  <EmptyState.Root label="Console is empty" size="compact">
    <EmptyState.Art variant="console" />
    <EmptyState.Title>Console Arrives with Scripts</EmptyState.Title>
    <EmptyState.Description>Logs from pre-request and test scripts will show up here.</EmptyState.Description>
  </EmptyState.Root>
)

export function ConsoleDrawer() {
  return CONSOLE_PLACEHOLDER
}
