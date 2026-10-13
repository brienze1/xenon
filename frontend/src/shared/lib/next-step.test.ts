import { describe, expect, it } from 'vitest'
import { withNextStep } from './next-step'

describe('withNextStep', () => {
  it.each([
    ['disk full', 'Try again.', 'disk full. Try again.'],
    ['Disk is full.', 'Try again.', 'Disk is full. Try again.'],
    ['  Is the disk full?  ', 'Try again.', 'Is the disk full? Try again.'],
    ['', 'Try again.', 'Try again.'],
  ])('joins %j and %j', (message, nextStep, text) => {
    expect(withNextStep(message, nextStep)).toBe(text)
  })
})
