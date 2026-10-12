import { ExampleTab } from '@/features/examples/components/ExampleTab'
import { RequestBuilder } from '@/features/request-builder/components/RequestBuilder'
import { RequestAuth } from './RequestAuth'

interface ExampleWorkspaceProps {
  tabId: string
}

/** An example tab: the request editors in example mode above the example's response editor. */
export function ExampleWorkspace({ tabId }: ExampleWorkspaceProps) {
  return <ExampleTab tabId={tabId} requestEditor={<RequestBuilder tabId={tabId} mode="example" auth={<RequestAuth tabId={tabId} />} />} />
}
