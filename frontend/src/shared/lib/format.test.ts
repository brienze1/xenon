import { describe, expect, it } from 'vitest'
import { formatBytes } from './format-bytes'
import { formatDuration } from './format-duration'

const NBSP = ' '

describe('formatBytes', () => {
  it.each([
    [0, `0${NBSP}B`],
    [1023, `1,023${NBSP}B`],
    [1024, `1.00${NBSP}KB`],
    [20 * 1024, `20.0${NBSP}KB`],
    [8 * 1024 * 1024, `8.00${NBSP}MB`],
    [3 * 1024 * 1024 * 1024, `3.00${NBSP}GB`],
  ])('%d bytes is %s', (bytes, text) => {
    expect(formatBytes(bytes, 'en-US')).toBe(text)
  })

  it('formats the number for the locale', () => {
    expect(formatBytes(1536, 'de-DE')).toBe(`1,50${NBSP}KB`)
  })
})

describe('formatDuration', () => {
  it.each([
    [0.4, `0${NBSP}ms`],
    [12.6, `13${NBSP}ms`],
    [999, `999${NBSP}ms`],
    [1500, `1.50${NBSP}s`],
  ])('%d ms is %s', (milliseconds, text) => {
    expect(formatDuration(milliseconds, 'en-US')).toBe(text)
  })

  it('formats the number for the locale', () => {
    expect(formatDuration(12500, 'de-DE')).toBe(`12,50${NBSP}s`)
  })
})
