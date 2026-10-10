import { useExamplesStore } from '@/features/examples/store'
import { useTabsStore } from '@/features/tabs/store'
import { Button } from '@/shared/ui/button/Button'

interface SaveAsExampleButtonProps {
  tabId: string
}

/** Offers "Save as Example" for a saved request's successful response. */
export function SaveAsExampleButton({ tabId }: SaveAsExampleButtonProps) {
  const requestId = useTabsStore((state) => {
    const tab = state.tabs.find((candidate) => candidate.id === tabId)
    return tab && tab.kind !== 'example' ? tab.requestId : null
  })
  const openSaveDialog = useExamplesStore((state) => state.openSaveDialog)

  return (
    <Button
      size="sm"
      variant="ghost"
      disabled={requestId === null}
      title={requestId === null ? 'Save the request first to keep its examples' : 'Save this response as an example of the request'}
      onClick={() => (requestId === null ? undefined : openSaveDialog(tabId, requestId))}
    >
      Save as Example
    </Button>
  )
}
