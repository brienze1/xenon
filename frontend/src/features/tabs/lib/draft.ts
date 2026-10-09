import { emptyDraft, toRequestDto, type RequestDraft } from '@/platform/mappers/request'

const PRISTINE = JSON.stringify(toRequestDto(emptyDraft()))

export function isPristine(draft: RequestDraft): boolean {
  return JSON.stringify(toRequestDto(draft)) === PRISTINE
}
