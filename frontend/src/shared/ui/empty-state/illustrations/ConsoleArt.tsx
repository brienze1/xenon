import { Illustration } from './Illustration'

const CONSOLE_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="magenta" d="M14 18h92v58H14z" />
    <Illustration.Stroke tone="green" d="M28 38l12 10-12 10M50 60h22" />
    <Illustration.Drip tone="magenta" d="M36 76v9M90 76v6" />
  </Illustration.Canvas>
)

export function ConsoleArt() {
  return CONSOLE_ART
}
