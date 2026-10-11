import { AuthEditor } from '@/features/auth/components/AuthEditor'
import { useRequestBuilderStore } from '@/features/request-builder/store'

interface RequestAuthProps {
  tabId: string
}

/** The request's Authorization tab: the auth editor on the request builder's draft; an unsaved request owns its token by tab. */
export function RequestAuth({ tabId }: RequestAuthProps) {
  const auth = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.auth)
  const requestId = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.id ?? '')
  const updateAuth = useRequestBuilderStore((state) => state.updateAuth)
  if (!auth) {
    return null
  }
  return <AuthEditor auth={auth} ownerId={requestId || tabId} kind="request" onChange={(next) => updateAuth(tabId, next)} />
}
