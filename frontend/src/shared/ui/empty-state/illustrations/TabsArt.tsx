import { Illustration } from './Illustration'

const TABS_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="cyan" d="M16 34h26l4-10h26l4 10h28v44H16z" />
    <Illustration.Stroke tone="magenta" d="M46 52h28M60 38v28" />
    <Illustration.Drip tone="cyan" d="M30 78v8M88 78v5" />
  </Illustration.Canvas>
)

export function TabsArt() {
  return TABS_ART
}
