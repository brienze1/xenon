import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PanelSection } from './PanelSection'

describe('PanelSection', () => {
  it('renders a labelled section with a level-3 heading and a note', () => {
    render(
      <PanelSection.Root label="Scripts">
        <PanelSection.Heading>Query Params</PanelSection.Heading>
        <PanelSection.Note>Nothing to show.</PanelSection.Note>
      </PanelSection.Root>,
    )
    const region = screen.getByRole('region', { name: 'Scripts' })
    expect(region).toContainElement(screen.getByRole('heading', { level: 3, name: 'Query Params' }))
    expect(region).toHaveTextContent('Nothing to show.')
  })
})
