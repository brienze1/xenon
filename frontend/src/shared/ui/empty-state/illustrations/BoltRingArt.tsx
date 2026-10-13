import { Illustration } from './Illustration'

const BOLT_RING_ART = (
  <Illustration.Canvas>
    <Illustration.Ring tone="cyan" cx={60} cy={46} r={30} />
    <Illustration.Stroke tone="green" d="M64 22L48 50h14l-6 24 18-30H60z" />
    <Illustration.Drip tone="cyan" d="M44 74v10M78 72v6" />
  </Illustration.Canvas>
)

export function BoltRingArt() {
  return BOLT_RING_ART
}
