import { useCallback, useMemo } from 'react'
import type { EditorSelection } from '@/shared/ui/code-editor/editor-extensions'
import type { CodeField } from '@/shared/ui/key-value-table/KeyValueTable'
import type { VariableEditing } from '@/features/request-builder/hooks/use-variable-editing'

interface RowLike {
  id: string
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
