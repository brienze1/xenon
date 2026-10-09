import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { EmptyState } from './EmptyState'

describe('EmptyState', () => {
  it('renders a labelled section with the marker title, description and actions', () => {
    render(
      <EmptyState.Root label="No response">
        <EmptyState.Art variant="response" />
        <EmptyState.Title>Nothing here yet</EmptyState.Title>
        <EmptyState.Description>Send a request to see the response.</EmptyState.Description>
        <EmptyState.Actions>
          <button type="button">Send</button>
        </EmptyState.Actions>
      </EmptyState.Root>,
    )
    const region = screen.getByRole('region', { name: 'No response' })
    expect(region).toContainElement(screen.getByRole('heading', { level: 2, name: 'Nothing here yet' }))
    expect(screen.getByText('Send a request to see the response.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Send' })).toBeInTheDocument()
  })

  it.each(['workspace', 'collections', 'tabs', 'response', 'history', 'console', 'missing'] as const)('adds the %s illustration without anything for assistive tech to read', (variant) => {
    render(
      <EmptyState.Root label="Empty">
        <EmptyState.Art variant={variant} />
        <EmptyState.Title>Title</EmptyState.Title>
      </EmptyState.Root>,
    )
    const region = screen.getByRole('region', { name: 'Empty' })
    expect(within(region).getByRole('heading', { name: 'Title' })).toBeInTheDocument()
    expect(region).toHaveTextContent(/^Title$/)
  })
})
