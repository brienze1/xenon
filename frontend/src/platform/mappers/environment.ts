import type {
  EnvironmentList as EnvironmentListDto,
  EnvironmentSummary as EnvironmentSummaryDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const ENVIRONMENT_COLORS = ['none', 'green', 'cyan', 'magenta', 'yellow', 'red'] as const

export type EnvironmentColor = (typeof ENVIRONMENT_COLORS)[number]

export interface EnvironmentSummary {
  id: string
  name: string
  color: EnvironmentColor
  confirmBeforeSend: boolean
  error: string | null
}

export interface EnvironmentList {
  environments: EnvironmentSummary[]
  activeId: string | null
}

function toEnvironmentColor(color: string): EnvironmentColor {
  return (ENVIRONMENT_COLORS as readonly string[]).includes(color) ? (color as EnvironmentColor) : 'none'
}

export function toEnvironmentSummary(dto: EnvironmentSummaryDto): EnvironmentSummary {
  return {
    id: dto.id,
    name: dto.name,
    color: toEnvironmentColor(dto.color),
    confirmBeforeSend: dto.confirm_before_send,
    error: dto.error ? dto.error : null,
  }
}

export function toEnvironmentList(dto: EnvironmentListDto | null): EnvironmentList {
  return {
    environments: (dto?.environments ?? []).map(toEnvironmentSummary),
    activeId: dto?.active_id ? dto.active_id : null,
  }
}
