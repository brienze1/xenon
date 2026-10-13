import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useExportStore } from '@/features/export/store'
import styles from './ExportDialog.module.css'

const EXPORT_NEXT_STEP = 'Check the save location and the options, then try again.'
const CANCELLED = 'The export was cancelled; no file was written.'

export function ExportErrorBanner() {
  const error = useExportStore((state) => state.error)
  const cancelled = useExportStore((state) => state.cancelled)
  if (error) {
    return (
      <div className={styles.banner}>
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={withNextStep(error.description || error.code, EXPORT_NEXT_STEP)} terms={[error.code]} />
          </Banner.Message>
        </Banner.Root>
      </div>
    )
  }
  return cancelled ? (
    <div className={styles.banner}>
      <Banner.Root tone="info">
        <Banner.Message>{CANCELLED}</Banner.Message>
      </Banner.Root>
    </div>
  ) : null
}
