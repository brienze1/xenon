import { render } from '@testing-library/react'
import { useRef, type CSSProperties } from 'react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { useFitHeight } from './use-fit-height'

// jsdom has no layout: each line of the value is one 18 px line, plus 4 px of padding top and bottom.
const LINE_PX = 18
const PADDING_Y_PX = 8
let overrideScrollHeight: number | null = null

function lines(count: number): string {
  return Array.from({ length: count }, (_, index) => `line ${index + 1}`).join('\n')
}

const BOX: CSSProperties = { lineHeight: `${LINE_PX}px`, padding: '4px 8px', border: '1px solid' }

function Box({ value, maxLines, style = BOX }: { value: string; maxLines?: number; style?: CSSProperties }) {
  const ref = useRef<HTMLTextAreaElement>(null)
  useFitHeight(ref, value, maxLines)
  return <textarea ref={ref} aria-label="Value" rows={1} style={style} value={value} readOnly />
}

function Detached({ value }: { value: string }) {
  const ref = useRef<HTMLTextAreaElement>(null)
  useFitHeight(ref, value, 6)
  return <p>{value}</p>
}

function box(container: HTMLElement): HTMLTextAreaElement {
  return container.querySelector('textarea') as HTMLTextAreaElement
}

describe('useFitHeight', () => {
  beforeEach(() => {
    Object.defineProperty(HTMLTextAreaElement.prototype, 'scrollHeight', {
      configurable: true,
      get(this: HTMLTextAreaElement) {
        return overrideScrollHeight ?? PADDING_Y_PX + LINE_PX * (this.value === '' ? 1 : this.value.split('\n').length)
      },
    })
  })

  afterEach(() => {
    overrideScrollHeight = null
    delete (HTMLTextAreaElement.prototype as { scrollHeight?: number }).scrollHeight
  })

  it('fits the height to the text, border included, and never scrolls when there is no cap', () => {
    const { container } = render(<Box value="https://auth.dev" />)
    expect(box(container).style.height).toBe('28px')
    expect(box(container).style.overflowY).toBe('hidden')
    expect(box(container).style.overflowX).toBe('hidden')
  })

  it('grows up to the cap without scrolling', () => {
    const { container } = render(<Box value={lines(6)} maxLines={6} />)
    expect(box(container).style.height).toBe('118px')
    expect(box(container).style.overflowY).toBe('hidden')
  })

  it('does not scroll when the measured height is one pixel over the cap (rounding at a fractional zoom)', () => {
    overrideScrollHeight = 117
    const { container } = render(<Box value={lines(6)} maxLines={6} />)
    expect(box(container).style.height).toBe('119px')
    expect(box(container).style.overflowY).toBe('hidden')
  })

  it('stops at the cap and scrolls only vertically past it', () => {
    const { container } = render(<Box value={lines(9)} maxLines={6} />)
    expect(box(container).style.height).toBe('118px')
    expect(box(container).style.overflowY).toBe('auto')
    expect(box(container).style.overflowX).toBe('hidden')
  })

  it('grows and shrinks again as the text changes', () => {
    const { container, rerender } = render(<Box value="one" maxLines={6} />)
    expect(box(container).style.height).toBe('28px')
    rerender(<Box value={lines(3)} maxLines={6} />)
    expect(box(container).style.height).toBe('64px')
    rerender(<Box value="one" maxLines={6} />)
    expect(box(container).style.height).toBe('28px')
  })

  it('fits without a cap when the line height is not a length', () => {
    const { container } = render(<Box value={lines(9)} maxLines={6} style={{ ...BOX, lineHeight: 'normal' }} />)
    expect(box(container).style.height).toBe('172px')
    expect(box(container).style.overflowY).toBe('hidden')
  })

  it('does nothing while no textarea is attached', () => {
    expect(() => render(<Detached value="one" />)).not.toThrow()
  })
})
