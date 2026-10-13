import { AuthEditor } from '@/features/auth/components/AuthEditor'
import type { ItemAuthSlot } from '@/features/item-settings/model'

interface ItemAuthProps {
  slot: ItemAuthSlot
}

/** The auth of a collection or folder in its tab; tokens obtained here are shared by the requests that inherit it. */
export function ItemAuth({ slot }: ItemAuthProps) {
  return <AuthEditor auth={slot.auth} ownerId={slot.itemId} kind={slot.kind} onChange={slot.onChange} />
}
