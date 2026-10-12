import type { ChosenResolution, ConflictResolution } from '@/platform/mappers/sync'
import type { ConflictChoice } from '@/features/sync/lib/conflict-labels'
import { Button } from '@/shared/ui/button/Button'
import styles from './ConflictView.module.css'

interface ChoiceButtonsProps {
  label: string
  choices: readonly ConflictChoice[]
  selected: ConflictResolution
  disabled: boolean
  onChoose: (resolution: ChosenResolution) => void
}

export function ChoiceButtons({ label, choices, selected, disabled, onChoose }: ChoiceButtonsProps) {
  return (
    <div role="group" aria-label={label} className={styles.choices}>
      {choices.map((choice) => (
        <Button
          key={choice.resolution}
          size="sm"
          variant={selected === choice.resolution ? 'accent' : 'secondary'}
          aria-pressed={selected === choice.resolution}
          disabled={disabled}
          onClick={() => onChoose(choice.resolution)}
        >
          {choice.label}
        </Button>
      ))}
    </div>
  )
}
