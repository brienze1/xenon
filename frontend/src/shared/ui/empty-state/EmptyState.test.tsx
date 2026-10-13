import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { ComponentType } from 'react'
import { EmptyState } from './EmptyState'
import { BoltRingArt } from './illustrations/BoltRingArt'
import { ClockArt } from './illustrations/ClockArt'
import { FolderQuestionArt } from './illustrations/FolderQuestionArt'
import { FolderStarArt } from './illustrations/FolderStarArt'
import { PanelPlusArt } from './illustrations/PanelPlusArt'
import { StackArt } from './illustrations/StackArt'
import { TerminalArt } from './illustrations/TerminalArt'

const ILLUSTRATIONS: Array<[string, ComponentType]> = [
  ['folder star', FolderStarArt],
  ['stack', StackArt],
  ['panel plus', PanelPlusArt],
  ['bolt ring', BoltRingArt],
  ['clock', ClockArt],
  ['terminal', TerminalArt],
  ['folder question', FolderQuestionArt],
]

describe('EmptyState', () => {
  it('renders a labelled section with the marker title, description and actions', () => {
    render(
      <EmptyState.Root label="No response">
        <EmptyState.Art>
          <BoltRingArt />
        </EmptyState.Art>
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

  it.each(ILLUSTRATIONS)('adds the %s illustration without anything for assistive tech to read', (_name, Illustration) => {
    render(
      <EmptyState.Root label="Empty">
        <EmptyState.Art>
          <Illustration />
        </EmptyState.Art>
        <EmptyState.Title>Title</EmptyState.Title>
      </EmptyState.Root>,
    )
    const region = screen.getByRole('region', { name: 'Empty' })
    expect(within(region).getByRole('heading', { name: 'Title' })).toBeInTheDocument()
    expect(region).toHaveTextContent(/^Title$/)
  })
})
