import { PanelSection } from '@/shared/ui/panel-section/PanelSection'

const SCRIPTS_PLACEHOLDER = (
  <PanelSection.Root label="Scripts">
    <PanelSection.Heading>Pre-Request and Post-Response Scripts</PanelSection.Heading>
    <PanelSection.Note>Scripts arrive in a later stage of Xenon.</PanelSection.Note>
  </PanelSection.Root>
)

export function ScriptsPanel() {
  return SCRIPTS_PLACEHOLDER
}
