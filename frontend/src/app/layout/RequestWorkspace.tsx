import { useState } from 'react'
import { RequestBuilder } from '@/features/request-builder/components/RequestBuilder'
import { ResponseViewer } from '@/features/response-viewer/components/ResponseViewer'
import { SplitPane } from '@/shared/ui/split-pane/SplitPane'

const DEFAULT_SPLIT = 0.5

interface RequestWorkspaceProps {
  tabId: string
}

export function RequestWorkspace({ tabId }: RequestWorkspaceProps) {
  const [ratio, setRatio] = useState(DEFAULT_SPLIT)
  return (
    <SplitPane.Root direction="column" ratio={ratio} onRatioChange={setRatio}>
      <SplitPane.Pane>
        <RequestBuilder tabId={tabId} />
      </SplitPane.Pane>
      <SplitPane.Divider label="Resize request and response" />
      <SplitPane.Pane>
        <ResponseViewer tabId={tabId} />
      </SplitPane.Pane>
    </SplitPane.Root>
  )
}
