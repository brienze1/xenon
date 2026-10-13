import { createContext, type ReactNode } from 'react'

/** Renders the card of a decorated range (its `card` key); `close` dismisses it, e.g. after an action. */
export type RenderDecorationCard = (card: string, close: () => void) => ReactNode

/**
 * Editors inside this context open a card when a decoration with a `card` key is hovered. Whoever provides it decides
 * what the card holds, so the editors stay domain-free.
 */
export const DecorationCardContext = createContext<RenderDecorationCard | null>(null)
