import { Illustration } from './Illustration'

const COLLECTIONS_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="magenta" d="M30 20h56v14H30z" />
    <Illustration.Stroke tone="cyan" d="M24 40h68v14H24z" />
    <Illustration.Stroke tone="green" d="M18 60h80v14H18z" />
    <Illustration.Drip tone="green" d="M40 74v10M76 74v6" />
  </Illustration.Canvas>
)

export function CollectionsArt() {
  return COLLECTIONS_ART
}
