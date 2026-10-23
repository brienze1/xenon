import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Swatch } from '@/shared/ui/swatch/Swatch'
import { useEnvironmentsStore } from '@/features/environments/store'
import { unavailableAliasMessage, useUnavailableAlias } from '@/features/request-builder/hooks/use-unavailable-alias'

interface UnavailableAliasBannerProps {
  tabId: string
  className?: string
}

/** Explains why Send is off while the alias on screen is not for the active environment, with one switch per environment it is for. */
export function UnavailableAliasBanner({ tabId, className }: UnavailableAliasBannerProps) {
  const unavailable = useUnavailableAlias(tabId)
  const setActive = useEnvironmentsStore((state) => state.setActive)
  if (unavailable === null) {
    return null
  }
  const switches = unavailable.labels.filter((label) => !label.deleted)
  return (
    <Banner.Root tone="warning" className={className}>
      <Banner.Message>
        <NoTranslateText text={unavailableAliasMessage(unavailable)} terms={[unavailable.alias.name]} />
      </Banner.Message>
      {switches.length > 0 ? (
        <Banner.Actions>
          {switches.map((label) => (
            <Button key={label.id} variant="secondary" size="sm" onClick={() => void setActive(label.id)}>
              <Swatch color={label.color} />
              Switch to <span translate="no">{label.name}</span>
            </Button>
          ))}
        </Banner.Actions>
      ) : null}
    </Banner.Root>
  )
}
