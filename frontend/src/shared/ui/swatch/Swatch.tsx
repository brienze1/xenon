import { cn } from '@/shared/lib/cn'
import styles from './Swatch.module.css'

interface SwatchProps {
  /** One of the neon colour names (`green`, `cyan`, `magenta`, `yellow`, `red`); anything else is a neutral dot. */
  color: string
  className?: string
}

/** A small colour dot, such as an environment's colour; decorative, so the name next to it carries the meaning. */
export function Swatch({ color, className }: SwatchProps) {
  return <span className={cn(styles.swatch, className)} data-color={color} aria-hidden="true" />
}
