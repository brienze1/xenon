import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { TerminalArt } from '@/shared/ui/empty-state/illustrations/TerminalArt'

const CONSOLE_PLACEHOLDER = (
  <EmptyState.Root label="Console is empty" size="compact">
    <EmptyState.Art>
      <TerminalArt />
    </EmptyState.Art>
    <EmptyState.Title>Console Arrives with Scripts</EmptyState.Title>
    <EmptyState.Description>Logs from pre-request and test scripts will show up here.</EmptyState.Description>
  </EmptyState.Root>
)

export function ConsoleDrawer() {
  return CONSOLE_PLACEHOLDER
}
