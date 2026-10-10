import { ExampleTab } from '@/features/examples/components/ExampleTab'
import { RequestBuilder } from '@/features/request-builder/components/RequestBuilder'

interface ExampleWorkspaceProps {
  tabId: string
  onSave: () => void
}

/** An example tab: the request editors in example mode above the example's response editor. */
export function ExampleWorkspace({ tabId, onSave }: ExampleWorkspaceProps) {
  return <ExampleTab tabId={tabId} requestEditor={<RequestBuilder tabId={tabId} mode="example" onSave={onSave} />} />
}
