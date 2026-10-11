import { RequestBuilder } from '@/features/request-builder/components/RequestBuilder'
import { ResponseViewer } from '@/features/response-viewer/components/ResponseViewer'
import { LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { SplitPane } from '@/shared/ui/split-pane/SplitPane'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import { RequestAuth } from './RequestAuth'
import { RequestScripts } from './RequestScripts'

const DIVIDER = <SplitPane.Divider label="Resize request and response" />

interface RequestWorkspaceProps {
  tabId: string
  onSave: () => void
}

export function RequestWorkspace({ tabId, onSave }: RequestWorkspaceProps) {
  const {
    state: { layout },
    actions: { update },
  } = useShellLayout()
  return (
    <SplitPane.Root
      direction={layout.split === 'side-by-side' ? 'row' : 'column'}
      ratio={layout.splitRatio}
      min={LAYOUT_LIMITS.split.min}
      max={LAYOUT_LIMITS.split.max}
      onRatioChange={(splitRatio) => update({ splitRatio })}
    >
      <SplitPane.Pane>
        <RequestBuilder tabId={tabId} onSave={onSave} scripts={<RequestScripts tabId={tabId} />} auth={<RequestAuth tabId={tabId} />} />
      </SplitPane.Pane>
      {DIVIDER}
      <SplitPane.Pane>
        <ResponseViewer tabId={tabId} />
      </SplitPane.Pane>
    </SplitPane.Root>
  )
}
