import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { CompletionSource, EditorDecoration, EditorSelection } from '@/shared/ui/code-editor/editor-extensions'
import type { CodeField } from '@/shared/ui/key-value-table/KeyValueTable'
import { decorationsByField } from '@/features/request-builder/lib/variable-decorations'
import { toCompletions, variableCompletionQuery } from '@/features/request-builder/lib/variable-completion'

interface VariableEditing {
  /** Decorations of a request field (`url`, `params.0.value`, `body.raw`, …). */
  decorationsOf: (field: string) => readonly EditorDecoration[] | undefined
  completionSource: CompletionSource
  selectIn: (field: string, selection: EditorSelection) => void
}

interface RowLike {
  id: string
}

const NO_DECORATIONS: readonly EditorDecoration[] = []

export function useActiveEnvironmentName(): string | null {
  return useEnvironmentsStore((state) => state.environments.find((environment) => environment.id === state.activeId)?.name ?? null)
}

/** Feeds the request editors from VariableService: decorations from Inspect, `{{` completion from Suggest, selections for "Move to secret variable". */
export function useVariableEditing(tabId: string): VariableEditing {
  const inspection = useRequestBuilderStore((state) => state.tabs[tabId]?.inspection ?? null)
  const suggest = useRequestBuilderStore((state) => state.suggest)
  const setSelection = useRequestBuilderStore((state) => state.setSelection)
  const environmentName = useActiveEnvironmentName()
  const environmentNameRef = useRef(environmentName)

  useEffect(() => {
    environmentNameRef.current = environmentName
  }, [environmentName])

  const byField = useMemo(() => decorationsByField(inspection, environmentName), [inspection, environmentName])

  const decorationsOf = useCallback((field: string) => byField.get(field) ?? NO_DECORATIONS, [byField])

  const completionSource = useCallback<CompletionSource>(
    async (text, position) => {
      const query = variableCompletionQuery(text, position)
      if (query === null) {
        return null
      }
      return toCompletions(query, await suggest(tabId, query.prefix), environmentNameRef.current)
    },
    [suggest, tabId],
  )

  const selectIn = useCallback(
    (field: string, selection: EditorSelection) => setSelection(tabId, selection.from === selection.to ? null : { field, ...selection }),
    [setSelection, tabId],
  )

  return { decorationsOf, completionSource, selectIn }
}

/** Maps a key/value table's row ids to the Inspect field names of that table (`<prefix>.<index>.<key|value>`). */
export function useTableFields(prefix: string, rows: readonly RowLike[], editing: VariableEditing) {
  const { decorationsOf, selectIn } = editing
  const indexOf = useMemo(() => new Map(rows.map((row, index) => [row.id, index])), [rows])
  const cellDecorations = useCallback(
    (id: string, field: CodeField) => decorationsOf(`${prefix}.${indexOf.get(id) ?? -1}.${field}`),
    [decorationsOf, indexOf, prefix],
  )
  const onCellSelection = useCallback(
    (id: string, field: CodeField, selection: EditorSelection) => selectIn(`${prefix}.${indexOf.get(id) ?? -1}.${field}`, selection),
    [indexOf, prefix, selectIn],
  )
  return { cellDecorations, onCellSelection }
}
