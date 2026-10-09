import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Banner } from './Banner'

describe('Banner', () => {
  it('announces errors as alerts and other tones as status', () => {
    const { rerender } = render(
      <Banner.Root tone="error">
        <Banner.Message>Folder missing</Banner.Message>
        <Banner.Actions>
          <button type="button">Locate folder…</button>
        </Banner.Actions>
      </Banner.Root>,
    )
    expect(screen.getByRole('alert')).toHaveTextContent('Folder missing')
    expect(screen.getByRole('alert').className).toContain('error')
    rerender(
      <Banner.Root tone="warning">
        <Banner.Message>Changed on disk</Banner.Message>
      </Banner.Root>,
    )
    expect(screen.getByRole('status')).toHaveTextContent('Changed on disk')
  })
})
