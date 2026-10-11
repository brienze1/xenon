import { lazy, Suspense, useEffect, useId, useState, type ReactNode } from 'react'
import { useShallow } from 'zustand/react/shallow'
import type { SnippetOption } from '@/platform/mappers/snippet'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { currentSnippetLanguage, editorLanguageOf, optionValue, rememberedOptions } from '@/features/side-panel/lib/snippet-options'
import { useSidePanelStore } from '@/features/side-panel/store'
import { useTabsStore } from '@/features/tabs/store'
import styles from './CodePanel.module.css'

const CodeEditor = lazy(() => import('@/shared/ui/code-editor/CodeEditor').then((module) => ({ default: module.CodeEditor })))

export const SNIPPET_DELAY_MS = 150

type CopyState = 'idle' | 'copied' | 'failed'

const COPY_ANNOUNCEMENTS: Record<CopyState, ReactNode> = {
  idle: null,
  copied: 'The snippet is on the clipboard.',
  failed: 'The snippet could not be copied.',
}

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
  const [copyState, setCopyState] = useState<CopyState>('idle')
  const language = currentSnippetLanguage(saved, languages)
  const options = rememberedOptions(saved, language)
  const spec = languages.find((candidate) => candidate.id === language)
  const languageId = useId()

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

  async function copyText() {
    const failure = await copy(tabId, { language, options })
    setCopyState(failure ? 'failed' : 'copied')
  }

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
        <Button size="sm" variant="ghost" disabled={snippet === null} onClick={() => void copyText()}>
          {copyState === 'copied' ? 'Copied' : 'Copy'}
        </Button>
        <span role="status" aria-live="polite" className="visually-hidden">
          {COPY_ANNOUNCEMENTS[copyState]}
        </span>
      </div>
      <label className={styles.check}>
        <input type="checkbox" name="include-secrets" checked={includeSecrets} onChange={(event) => setIncludeSecrets(event.target.checked)} />
        Include Secrets
      </label>
      {spec && spec.options.length > 0 ? (
        <details className={styles.options}>
          <summary className={styles.summary}>Options</summary>
          {spec.options.map((option) => (
            <SnippetOptionField key={option.id} option={option} value={optionValue(options, option.id, option.defaultValue)} onChange={(value) => setOption(option, value)} />
          ))}
        </details>
      ) : null}
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
    </div>
  )
}

interface SnippetOptionFieldProps {
  option: SnippetOption
  value: string
  onChange: (value: string) => void
}

function SnippetOptionField({ option, value, onChange }: SnippetOptionFieldProps) {
  if (option.type === 'boolean') {
    return (
      <label className={styles.check}>
        <input type="checkbox" name={option.id} checked={value === 'true'} onChange={(event) => onChange(String(event.target.checked))} />
        {option.label}
      </label>
    )
  }
  if (option.type === 'enum') {
    return (
      <label className={styles.field}>
        {option.label}
        <select name={option.id} className={styles.select} value={value} onChange={(event) => onChange(event.target.value)}>
          {option.values.map((choice) => (
            <option key={choice} value={choice}>
              {choice}
            </option>
          ))}
        </select>
      </label>
    )
  }
  return (
    <label className={styles.field}>
      {option.label}
      <input
        type="number"
        name={option.id}
        min={0}
        step={1}
        inputMode="numeric"
        autoComplete="off"
        className={styles.number}
        value={value}
        onChange={(event) => onChange(event.target.value === '' ? '0' : event.target.value)}
      />
    </label>
  )
}
