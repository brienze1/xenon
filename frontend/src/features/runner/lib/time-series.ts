import type { SeriesPoint } from '@/platform/mappers/runner'

export type SeriesKey = keyof Omit<SeriesPoint, 'second'>

/** An SVG path through the points, scaled to the panel; a stepped line holds each value until the next second. */
export function seriesPath(series: readonly SeriesPoint[], key: SeriesKey, stepped: boolean, width: number, height: number): string {
  if (series.length === 0) {
    return ''
  }
  const max = Math.max(1, ...series.map((point) => point[key]))
  const last = Math.max(1, series.length - 1)
  const x = (index: number) => Math.round((width * index * 10) / last) / 10
  const y = (value: number) => Math.round((height - (height * value) / max) * 10) / 10
  return series
    .map((point, index) => {
      if (index === 0) {
        return `M${x(0)},${y(point[key])}`
      }
      return stepped ? `H${x(index)}V${y(point[key])}` : `L${x(index)},${y(point[key])}`
    })
    .join('')
}
