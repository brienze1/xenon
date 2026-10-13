import { useResponseViewerStore } from '@/features/response-viewer/store'

const MISSING_FILE_PREFIX = /^File not found: /

export function useMissingFile(tabId: string): string | null {
  return useResponseViewerStore((state) => {
    const error = state.entries[tabId]?.result?.error
    return error?.kind === 'missing_file' ? error.message.replace(MISSING_FILE_PREFIX, '') : null
  })
}
