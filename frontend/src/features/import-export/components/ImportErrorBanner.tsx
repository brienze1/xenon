import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useImportStore } from '@/features/import-export/store'
import styles from './ImportDialog.module.css'

const IMPORT_NEXT_STEP = 'Check the source and the target workspace, then try again.'

export function ImportErrorBanner() {
  const error = useImportStore((state) => state.error)
  if (!error) {
    return null
  }
  return (
    <div className={styles.banner}>
      <Banner.Root tone="error">
        <Banner.Message>
          <NoTranslateText text={withNextStep(error.description || error.code, IMPORT_NEXT_STEP)} terms={[error.code]} />
        </Banner.Message>
      </Banner.Root>
    </div>
  )
}
