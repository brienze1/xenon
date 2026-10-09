import { Illustration } from './Illustration'

const MISSING_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="yellow" d="M18 30h28l8 9h46v41H18z" />
    <Illustration.Stroke tone="red" d="M50 52a10 10 0 1 1 14 9c-3 1.5-4 3-4 6M60 74v1" />
    <Illustration.Drip tone="yellow" d="M30 80v8M88 80v5" />
  </Illustration.Canvas>
)

export function MissingArt() {
  return MISSING_ART
}
