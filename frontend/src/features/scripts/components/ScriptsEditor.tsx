import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react'
import type { ScriptPhase, ScriptSet } from '@/platform/mappers/request'
import { memberCompletions, variableCompletions, variableNameQuery, type VariableOption } from '@/features/scripts/lib/completion'
import { useSyntaxErrors } from '@/features/scripts/hooks/use-syntax-errors'
import { useScriptsStore } from '@/features/scripts/store'
import type { CodeEditorHandle } from '@/shared/ui/code-editor/CodeEditor'
import type { CompletionSource } from '@/shared/ui/code-editor/editor-extensions'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import styles from './ScriptsEditor.module.css'

const CodeEditor = lazy(() => import('@/shared/ui/code-editor/CodeEditor').then((module) => ({ default: module.CodeEditor })))

const PHASES: ReadonlyArray<{ phase: ScriptPhase; label: string }> = [
  { phase: 'preRequest', label: 'Pre-request' },
  { phase: 'postResponse', label: 'Post-response' },
]

const EDITOR_HINT = (
  <p className={styles.hint}>
    Type <code>pm.</code> for the API. A test written as <code>async () =&gt;</code> reports a failure after an <code>await</code> as a script
    error; values a script transforms (substrings, base64) are not masked in the console.
  </p>
)

function textOf(scripts: ScriptSet, phase: ScriptPhase): string {
  return phase === 'preRequest' ? scripts.preRequest : scripts.postResponse
}

function withText(scripts: ScriptSet, phase: ScriptPhase, text: string): ScriptSet {
  return phase === 'preRequest' ? { ...scripts, preRequest: text } : { ...scripts, postResponse: text }
}

interface PhaseEditorProps {
  label: string
  phase: ScriptPhase
  scripts: ScriptSet
  onChange: (scripts: ScriptSet) => void
  completionSource: CompletionSource
}

function PhaseEditor({ label, phase, scripts, onChange, completionSource }: PhaseEditorProps) {
  const snippets = useScriptsStore((state) => state.snippets)
  const editorRef = useRef<CodeEditorHandle>(null)
  const text = textOf(scripts, phase)
  const errors = useSyntaxErrors(text)
  const disabled = scripts.disabled.includes(phase)
  const available = snippets.filter((snippet) => snippet.phase === 'any' || snippet.phase === phase)
  return (
    <div className={styles.phase}>
      <div className={styles.toolbar}>
        <Dropdown.Root>
          <Dropdown.Trigger label={`Snippets for the ${label} script`}>Snippets</Dropdown.Trigger>
          <Dropdown.Menu label="Snippets">
            {available.map((snippet) => (
              <Dropdown.Item key={snippet.id} onSelect={() => editorRef.current?.insert(snippet.text)}>
                {snippet.title}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>
        </Dropdown.Root>
        {disabled ? (
          <p className={styles.disabled}>
            This script is disabled and does not run.
            <Button size="sm" variant="ghost" onClick={() => onChange({ ...scripts, disabled: scripts.disabled.filter((item) => item !== phase) })}>
              Enable
            </Button>
          </p>
        ) : null}
      </div>
      <div className={styles.editor}>
        <Suspense fallback={null}>
          <CodeEditor
            ref={editorRef}
            value={text}
            language="javascript"
            label={`${label} script`}
            onChange={(next) => onChange(withText(scripts, phase, next))}
            diagnostics={errors}
            completionSource={completionSource}
          />
        </Suspense>
      </div>
    </div>
  )
}

interface ScriptsEditorProps {
  scripts: ScriptSet
  onChange: (scripts: ScriptSet) => void
  /** The "Inherit parent scripts" toggle; `null` hides it (collections have no parent). */
  inheritScripts: boolean | null
  onInheritChange?: (inherit: boolean) => void
  /** Variable names offered inside pm.<scope>.get/set/unset/has("…"). */
  suggestVariables?: (prefix: string) => Promise<VariableOption[]>
}

export function ScriptsEditor({ scripts, onChange, inheritScripts, onInheritChange, suggestVariables }: ScriptsEditorProps) {
  const load = useScriptsStore((state) => state.load)
  const [phase, setPhase] = useState<string>('preRequest')
  const suggestRef = useRef(suggestVariables)

  useEffect(() => {
    suggestRef.current = suggestVariables
  }, [suggestVariables])

  useEffect(() => {
    void load()
  }, [load])

  const completionSource = useCallback<CompletionSource>(async (text, position) => {
    const variable = variableNameQuery(text, position)
    if (variable) {
      const suggest = suggestRef.current
      return suggest ? variableCompletions(variable, await suggest(variable.prefix)) : null
    }
    return memberCompletions(useScriptsStore.getState().reference, text, position)
  }, [])

  return (
    <div className={styles.scripts}>
      <Tabs.Root value={phase} onValueChange={setPhase} className={styles.tabs}>
        <div className={styles.header}>
          <Tabs.List label="Script phase">
            {PHASES.map(({ phase: value, label }) => (
              <Tabs.Tab key={value} value={value}>
                {label}
                {textOf(scripts, value).trim() !== '' ? <span className={styles.dot} aria-label="has a script" /> : null}
              </Tabs.Tab>
            ))}
          </Tabs.List>
          {inheritScripts === null ? null : (
            <label className={styles.inherit}>
              <input type="checkbox" checked={inheritScripts} onChange={(event) => onInheritChange?.(event.target.checked)} />
              Inherit parent scripts
            </label>
          )}
        </div>
        {PHASES.map(({ phase: value, label }) => (
          <Tabs.Panel key={value} value={value} className={styles.panel}>
            <PhaseEditor label={label} phase={value} scripts={scripts} onChange={onChange} completionSource={completionSource} />
          </Tabs.Panel>
        ))}
      </Tabs.Root>
      {EDITOR_HINT}
    </div>
  )
}
