import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'

/** What a tab's name names, for the "Type a name for the …" message. */
export type NameSubject = 'request' | 'example' | 'collection' | 'folder' | 'environment'

/** Why a tab's pending name was not saved; shown under the name until it is edited again. */
export type NameError = { kind: 'required'; subject: NameSubject } | { kind: 'clash'; name: string } | { kind: 'failed'; error: AppError }

const FAILED_NEXT_STEP = 'Choose another name, then save.'

/** Whether a tab's name draft is an unsaved change: never for an item with no saved name (a request not saved yet). */
export function isNamePending(draft: string | undefined, saved: string | null): boolean {
  return saved !== null && draft !== undefined && draft.trim() !== saved
}

export function nameErrorMessage(error: NameError): string {
  if (error.kind === 'required') {
    return `Type a name for the ${error.subject}.`
  }
  if (error.kind === 'clash') {
    return `An environment named “${error.name.trim()}” already exists. Choose another name.`
  }
  return withNextStep(error.error.description || error.error.code, FAILED_NEXT_STEP)
}
