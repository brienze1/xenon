import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'

const MESSAGES: Record<string, string> = {
  'WKS-00400': 'Give the workspace a name (up to 200 characters) and choose a folder.',
  'WKS-01409': 'That folder isn’t empty. Choose an empty folder or a new one.',
  'WKS-02409': 'Another workspace already uses that folder, or the folder sits inside or around one. Choose a folder outside your other workspaces.',
  'WKS-06404': 'The workspace folder is missing. Someone moved or deleted it outside Xenon. Locate the folder again, or remove the workspace.',
  'WKS-07404': 'Xenon has no record of that workspace. Open its folder again with Open Workspace….',
  'WKS-03404': 'That folder isn’t a Xenon workspace (it has no workspace.yaml). Choose the folder that contains workspace.yaml.',
  'WKS-05500': 'Xenon couldn’t move the folder to the trash. You can remove the workspace from Xenon and keep its files instead.',
  'WKS-22400': 'Give the workspace a name (up to 200 characters).',
  'WKS-04422': 'A newer version of Xenon wrote this workspace. Update Xenon to open it.',
}

const NEXT_STEP = 'Try again, or check that the folder exists and you have permission to write to it.'

export const WORKSPACE_MESSAGE_TERMS: readonly string[] = ['Xenon', 'workspace.yaml']

export function workspaceErrorMessage(error: AppError): string {
  return MESSAGES[error.code] ?? withNextStep(error.description || 'Something went wrong.', NEXT_STEP)
}
