import { Clipboard } from '@wailsio/runtime'
import * as AIAccessService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aiaccessservice'
import { toAppError } from '../errors'
import { toAIAccessStatus, toPendingApproval, type AIAccessStatus, type PendingApproval } from '../mappers/ai-access'

export const APPROVAL_ALREADY_RESOLVED = 'MCP-05409'
export const APPROVAL_UNKNOWN = 'MCP-07400'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getAIAccessStatus(): Promise<AIAccessStatus> {
  return call(async () => toAIAccessStatus(await AIAccessService.Status()))
}

/** The AI sends waiting for an answer, oldest first. */
export function listPendingApprovals(): Promise<PendingApproval[]> {
  return call(async () => {
    const pending = await AIAccessService.PendingApprovals()
    return (pending ?? []).map(toPendingApproval)
  })
}

export function answerApproval(id: string, approved: boolean): Promise<void> {
  return call(async () => {
    await AIAccessService.AnswerApproval({ id, approved })
  })
}

/** The client configs and the stdio command hold no secret, so the webview may copy them. */
export function copyText(text: string): Promise<boolean> {
  return call(() => Clipboard.SetText(text)).then(() => true)
}
