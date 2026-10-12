import { lazy, Suspense, useEffect, useId, useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import type { SnippetOption } from '@/platform/mappers/snippet'
import { Banner } from '@/shared/ui/banner/Banner'
import { CopyButton } from '@/shared/ui/copy-button/CopyButton'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { currentSnippetLanguage, editorLanguageOf, rememberedOptions } from '@/features/side-panel/lib/snippet-options'
import { useSidePanelStore } from '@/features/side-panel/store'
import { useTabsStore } from '@/features/tabs/store'
import { SnippetSettingsDialog } from './SnippetSettingsDialog'
import styles from './CodePanel.module.css'

const CodeEditor = lazy(() => import('@/shared/ui/code-editor/CodeEditor').then((module) => ({ default: module.CodeEditor })))

export const SNIPPET_DELAY_MS = 150

const GEAR = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
    <circle cx="12" cy="12" r="3" />
  </svg>
)

const LOADING_EDITOR = (
  <p role="status" className={styles.note}>
    Loading editor…
  </p>
)

interface CodePanelProps {
  tabId: string
}

/** The request as cURL, JavaScript fetch, Go or Python, generated in Go with secrets masked unless included. */
export function CodePanel({ tabId }: CodePanelProps) {
  const languages = useSidePanelStore((state) => state.languages)
  const languagesError = useSidePanelStore((state) => state.languagesError)
  const loadLanguages = useSidePanelStore((state) => state.loadLanguages)
  const includeSecrets = useSidePanelStore((state) => state.includeSecrets)
  const setIncludeSecrets = useSidePanelStore((state) => state.setIncludeSecrets)
  const generate = useSidePanelStore((state) => state.generate)
  const copy = useSidePanelStore((state) => state.copy)
  const { snippet, status, error } = useSidePanelStore(
    useShallow((state) => ({ snippet: state.snippets[tabId]?.snippet ?? null, status: state.snippets[tabId]?.status ?? 'idle', error: state.snippets[tabId]?.error ?? null })),
  )
  const saved = useTabsStore((state) => state.snippetOptions)
  const setSnippetOptions = useTabsStore((state) => state.setSnippetOptions)
  const draft = useRequestBuilderStore((state) => state.tabs[tabId]?.draft)
  const language = currentSnippetLanguage(saved, languages)
  const options = rememberedOptions(saved, language)
  const spec = languages.find((candidate) => candidate.id === language)
  const languageId = useId()
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    void loadLanguages()
  }, [loadLanguages])

  useEffect(() => {
    if (!draft) {
      return
    }
    const timer = setTimeout(() => void generate(tabId, { language, options }), SNIPPET_DELAY_MS)
    return () => clearTimeout(timer)
  }, [tabId, draft, language, options, includeSecrets, generate])

  function setOption(option: SnippetOption, value: string) {
    setSnippetOptions(language, { ...options, [option.id]: value })
  }

  return (
    <div className={styles.panel}>
      <div className={styles.toolbar}>
        <label htmlFor={languageId} className="visually-hidden">
          Language
        </label>
        <select
          id={languageId}
          name="snippet-language"
          className={styles.select}
          value={language}
          onChange={(event) => setSnippetOptions(event.target.value, rememberedOptions(saved, event.target.value))}
        >
          {languages.map((candidate) => (
            <option key={candidate.id} value={candidate.id}>
              {candidate.label}
            </option>
          ))}
        </select>
        <div className={styles.actions}>
        {spec && spec.options.length > 0 ? (
          <IconButton label="Snippet settings" onClick={() => setSettingsOpen(true)}>
            {GEAR}
          </IconButton>
        ) : null}
        <CopyButton
          label="Copy snippet"
          disabled={snippet === null}
          copiedMessage="The snippet is on the clipboard."
          failedMessage="The snippet could not be copied."
          onCopy={async () => (await copy(tabId, { language, options })) === null}
        />
        </div>
      </div>
      <label className={styles.check}>
        <input type="checkbox" name="include-secrets" checked={includeSecrets} onChange={(event) => setIncludeSecrets(event.target.checked)} />
        Include Secrets
      </label>
      {snippet && snippet.notices.length > 0 ? (
        <Banner.Root tone="warning">
          {snippet.notices.map((notice) => (
            <Banner.Message key={notice}>{notice}</Banner.Message>
          ))}
        </Banner.Root>
      ) : null}
      {error || languagesError ? (
        <Banner.Root tone="error">
          <Banner.Message>{(error ?? languagesError)?.description || (error ?? languagesError)?.code}</Banner.Message>
        </Banner.Root>
      ) : null}
      {snippet ? (
        <Suspense fallback={LOADING_EDITOR}>
          <CodeEditor label={`${spec?.label ?? language} snippet`} className={styles.editor} language={editorLanguageOf(language)} value={snippet.text} />
        </Suspense>
      ) : (
        <p role="status" className={styles.note}>
          {status === 'loading' ? 'Generating…' : 'Open a request to see it as code.'}
        </p>
      )}
      {settingsOpen && spec ? <SnippetSettingsDialog language={spec} options={options} onChange={setOption} onClose={() => setSettingsOpen(false)} /> : null}
    </div>
  )
}
