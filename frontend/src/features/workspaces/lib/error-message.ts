import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'

const MESSAGES: Record<string, string> = {
  'WKS-00400': 'Give the workspace a name (up to 200 characters) and choose a folder.',
  'WKS-01409': 'That folder is not empty. Choose an empty folder or a new one.',
  'WKS-02409': 'That folder is already used by another workspace, or is inside or around one. Choose a folder outside your other workspaces.',
  'WKS-06404': 'The workspace folder is missing. It was moved or deleted outside Xenon. Locate the folder again, or remove the workspace.',
  'WKS-07404': 'That workspace is no longer registered. Open its folder again with Open Workspace….',
  'WKS-03404': 'That folder is not a Xenon workspace (it has no workspace.yaml). Choose the folder that contains workspace.yaml.',
  'WKS-05500': 'The folder could not be moved to the trash. You can remove the workspace from Xenon and keep its files instead.',
  'WKS-22400': 'Give the workspace a name (up to 200 characters).',
  'WKS-04422': 'This workspace was written by a newer version of Xenon. Update Xenon to open it.',
}

const NEXT_STEP = 'Try again, or check that the folder exists and you have permission to write to it.'

export function workspaceErrorMessage(error: AppError): string {
  return MESSAGES[error.code] ?? withNextStep(error.description || 'Something went wrong.', NEXT_STEP)
}
