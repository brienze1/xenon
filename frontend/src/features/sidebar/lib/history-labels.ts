import type { HistoryActivity, HistorySourceFilter } from '@/platform/mappers/history'

export type OutcomeTone = 'success' | 'redirect' | 'error' | 'info'

/** The tag on every AI row: "AI · Claude Code". */
export function aiTag(client: string | null): string {
  return client ? `AI · ${client}` : 'AI'
}

/** "tree.create", or just the tool when the call has no action. */
export function activityName(activity: HistoryActivity | null): string {
  if (!activity) {
    return 'unknown'
  }
  return activity.action ? `${activity.tool}.${activity.action}` : activity.tool
}

/** ok, the error code, denied or timed out. */
export function activityOutcome(activity: HistoryActivity | null): { label: string; tone: OutcomeTone } {
  switch (activity?.outcome) {
    case 'ok':
      return { label: 'ok', tone: 'success' }
    case 'denied':
      return { label: 'denied', tone: 'redirect' }
    case 'timed_out':
      return { label: 'timed out', tone: 'redirect' }
    default:
      return { label: activity?.code ?? 'error', tone: 'error' }
  }
}

/** What Clear History removes under each source filter. */
export const CLEAR_SCOPES: Record<HistorySourceFilter, { title: string; action: string }> = {
  all: { title: 'Clear History?', action: 'Clear History' },
  user: { title: 'Clear User History?', action: 'Clear User History' },
  ai: { title: 'Clear AI History?', action: 'Clear AI History' },
}
