import type { AppError } from '@/platform/errors'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { WORKSPACE_MESSAGE_TERMS, workspaceErrorMessage } from '@/features/workspaces/lib/error-message'

interface WorkspaceErrorTextProps {
  error: AppError
}

export function WorkspaceErrorText({ error }: WorkspaceErrorTextProps) {
  return <NoTranslateText text={workspaceErrorMessage(error)} terms={[...WORKSPACE_MESSAGE_TERMS, error.code]} />
}
