import styles from './Illustration.module.css'

const DOTS = [
  [8, 14, 1.6],
  [14, 8, 1],
  [20, 18, 1.2],
  [104, 74, 1.6],
  [112, 66, 1],
  [98, 84, 1.2],
  [110, 86, 0.8],
] as const

export function Spray() {
  return (
    <g className={styles.spray}>
      {DOTS.map(([cx, cy, r]) => (
        <circle key={`${cx}-${cy}`} cx={cx} cy={cy} r={r} />
      ))}
    </g>
  )
}
