import { Banner } from '@/shared/ui/banner/Banner'
import { useImportStore } from '@/features/import-export/store'
import styles from './ImportDialog.module.css'

export function ImportErrorBanner() {
  const error = useImportStore((state) => state.error)
  if (!error) {
    return null
  }
  return (
    <div className={styles.banner}>
      <Banner.Root tone="error">
        <Banner.Message>
          <span translate="no">{error.description}</span>
        </Banner.Message>
      </Banner.Root>
    </div>
  )
}
