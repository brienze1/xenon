import { Illustration } from './Illustration'

const FOLDER_STAR_ART = (
  <Illustration.Canvas>
    <Illustration.Stroke tone="cyan" d="M18 30h28l8 9h46v41H18z" />
    <Illustration.Drip tone="cyan" d="M34 80v8M70 80v5" />
    <Illustration.Stroke tone="magenta" d="M60 47l4 10 11 1-8 7 3 11-10-6-10 6 3-11-8-7 11-1z" />
  </Illustration.Canvas>
)

export function FolderStarArt() {
  return FOLDER_STAR_ART
}
