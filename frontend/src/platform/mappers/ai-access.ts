import type {
  AIAccessStatus as AIAccessStatusDto,
  AIApproval as AIApprovalRowDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import type {
  AIAccessStatusChanged as AIAccessStatusChangedDto,
  AIApprovalRequested as AIApprovalDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/publisher/dto/models'

export type AIAccessState = 'off' | 'running' | 'error'

/** What the MCP server is doing now; `address` when running, `reason` when it failed. */
export interface AIAccessRuntime {
  state: AIAccessState
  address: string | null
  reason: string | null
  code: string | null
  clients: number
}

/** The runtime status plus the ready-to-paste client configs Settings shows. */
export interface AIAccessStatus extends AIAccessRuntime {
  stdioCommand: string
  httpConfig: string
  stdioConfig: string
}

/** An AI send waiting for the human to approve it (confirm-before-send environment). */
export interface PendingApproval {
  id: string
  client: string
  method: string
  /** The URL as written, unresolved, so no secret appears. */
  url: string
  environment: { id: string; name: string; color: string }
  createdAt: string
}

const STATES: readonly AIAccessState[] = ['off', 'running', 'error']

function toState(state: string): AIAccessState {
  return STATES.find((known) => known === state) ?? 'off'
}

function orNull(value: string | undefined): string | null {
  return value ? value : null
}

export function toAIAccessRuntime(dto: AIAccessStatusChangedDto): AIAccessRuntime {
  return { state: toState(dto.state), address: orNull(dto.address), reason: orNull(dto.reason), code: orNull(dto.code), clients: dto.clients }
}

export function toAIAccessStatus(dto: AIAccessStatusDto | null): AIAccessStatus {
  if (!dto) {
    return { state: 'off', address: null, reason: null, code: null, clients: 0, stdioCommand: '', httpConfig: '', stdioConfig: '' }
  }
  return { ...toAIAccessRuntime(dto), stdioCommand: dto.stdio_command, httpConfig: dto.http_config, stdioConfig: dto.stdio_config }
}

/** Maps a `PendingApprovals` row or an `ai:approval-requested` payload (the same shape). */
export function toPendingApproval(dto: AIApprovalRowDto | AIApprovalDto): PendingApproval {
  return {
    id: dto.id,
    client: dto.client,
    method: dto.method,
    url: dto.url,
    environment: { id: dto.environment.id, name: dto.environment.name, color: dto.environment.color },
    createdAt: dto.created_at,
  }
}
