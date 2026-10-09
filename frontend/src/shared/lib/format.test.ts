import { describe, expect, it } from 'vitest'
import { formatBytes } from './format-bytes'
import { formatDuration } from './format-duration'

describe('formatBytes', () => {
  it.each([
    [0, '0 B'],
    [1023, '1023 B'],
    [1024, '1.00 KB'],
    [20 * 1024, '20.0 KB'],
    [8 * 1024 * 1024, '8.00 MB'],
    [3 * 1024 * 1024 * 1024, '3.00 GB'],
  ])('%d bytes is %s', (bytes, text) => {
    expect(formatBytes(bytes)).toBe(text)
  })
})

describe('formatDuration', () => {
  it.each([
    [0.4, '0 ms'],
    [12.6, '13 ms'],
    [999, '999 ms'],
    [1500, '1.50 s'],
  ])('%d ms is %s', (milliseconds, text) => {
    expect(formatDuration(milliseconds)).toBe(text)
  })
})
