import { Illustration } from './Illustration'

const HISTORY_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="cyan" d="M32 46a28 28 0 1 0 9-20.5" />
    <Illustration.Stroke tone="cyan" d="M40 14v12h12" />
    <Illustration.Stroke tone="magenta" d="M60 30v18l12 8" />
    <Illustration.Drip tone="cyan" d="M48 72v12M82 70v6" />
  </Illustration.Canvas>
)

export function HistoryArt() {
  return HISTORY_ART
}
