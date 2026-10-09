import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { NoTranslateText } from './NoTranslateText'

describe('NoTranslateText', () => {
  it('keeps protected terms out of translation', () => {
    render(
      <p>
        <NoTranslateText text="Update Xenon to open it." terms={['Xenon']} />
      </p>,
    )
    expect(screen.getByText('Xenon')).toHaveAttribute('translate', 'no')
    expect(screen.getByText(/Update/)).toHaveTextContent('Update Xenon to open it.')
  })
})
