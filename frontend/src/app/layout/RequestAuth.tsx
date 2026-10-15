import { AuthEditor } from '@/features/auth/components/AuthEditor'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { aliasAuthOwner } from '@/platform/mappers/request'

interface RequestAuthProps {
  tabId: string
}

/**
 * The request's Authorization tab: the auth editor on the request builder's draft; an unsaved request owns its token
 * by tab, and an alias on screen owns its own token.
 */
export function RequestAuth({ tabId }: RequestAuthProps) {
  const auth = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.auth)
  const requestId = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.id ?? '')
  const aliasId = useRequestBuilderStore((state) => state.tabs[tabId]?.aliasId ?? null)
  const updateAuth = useRequestBuilderStore((state) => state.updateAuth)
  if (!auth) {
    return null
  }
  const ownerId = requestId && aliasId ? aliasAuthOwner(requestId, aliasId) : requestId || tabId
  return <AuthEditor auth={auth} ownerId={ownerId} kind="request" onChange={(next) => updateAuth(tabId, next)} />
}
