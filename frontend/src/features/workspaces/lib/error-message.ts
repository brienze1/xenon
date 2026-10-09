import type { AppError } from '@/platform/errors'

const MESSAGES: Record<string, string> = {
  'WKS-00400': 'Give the workspace a name (up to 200 characters) and choose a folder.',
  'WKS-01409': 'That folder is not empty. Choose an empty folder or a new one.',
  'WKS-02409': 'That folder is already used by another workspace, or is inside or around one.',
  'WKS-06404': 'The workspace folder is missing. It was moved or deleted outside Xenon.',
  'WKS-07404': 'That workspace is no longer registered.',
  'WKS-04422': 'This workspace was written by a newer version of Xenon.',
}

export function workspaceErrorMessage(error: AppError): string {
  return MESSAGES[error.code] ?? (error.description || 'Something went wrong.')
}
