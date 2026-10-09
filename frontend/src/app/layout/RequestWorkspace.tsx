import { RequestBuilder } from '@/features/request-builder/components/RequestBuilder'
import { ResponseViewer } from '@/features/response-viewer/components/ResponseViewer'
import { LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { SplitPane } from '@/shared/ui/split-pane/SplitPane'
import { useShellLayout } from './shell-layout'

interface RequestWorkspaceProps {
  tabId: string
  onSave: () => void
}

export function RequestWorkspace({ tabId, onSave }: RequestWorkspaceProps) {
  const { layout, onLayoutChange } = useShellLayout()
  return (
    <SplitPane.Root
      direction={layout.split === 'side-by-side' ? 'row' : 'column'}
      ratio={layout.splitRatio}
      min={LAYOUT_LIMITS.split.min}
      max={LAYOUT_LIMITS.split.max}
      onRatioChange={(splitRatio) => onLayoutChange({ splitRatio })}
    >
      <SplitPane.Pane>
        <RequestBuilder tabId={tabId} onSave={onSave} />
      </SplitPane.Pane>
      <SplitPane.Divider label="Resize request and response" />
      <SplitPane.Pane>
        <ResponseViewer tabId={tabId} />
      </SplitPane.Pane>
    </SplitPane.Root>
  )
}
