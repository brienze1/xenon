import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { workspaceErrorMessage } from '../lib/error-message'
import { useWorkspacesStore } from '../store'

interface MissingFolderNoticeProps {
  localId: string
}

export function MissingFolderNotice({ localId }: MissingFolderNoticeProps) {
  const workspace = useWorkspacesStore((state) => state.workspaces.find((candidate) => candidate.localId === localId))
  const error = useWorkspacesStore((state) => state.error)
  const locate = useWorkspacesStore((state) => state.locate)
  const remove = useWorkspacesStore((state) => state.remove)

  return (
    <EmptyState.Root label="Workspace folder missing">
      <EmptyState.Title>Folder missing</EmptyState.Title>
      <EmptyState.Description>
        The folder of “{workspace?.name ?? 'this workspace'}” was moved or deleted outside Xenon
        {workspace ? ` (${workspace.path})` : ''}. Locate it again, or remove the workspace from Xenon.
      </EmptyState.Description>
      <EmptyState.Actions>
        <Button variant="primary" onClick={() => void locate(localId)}>
          Locate folder…
        </Button>
        <Button variant="secondary" onClick={() => void remove(localId, false)}>
          Remove from Xenon
        </Button>
      </EmptyState.Actions>
      {error && error.code !== 'WKS-06404' ? (
        <Banner.Root tone="error">
          <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
        </Banner.Root>
      ) : null}
    </EmptyState.Root>
  )
}
