import { RangeSetBuilder, type Extension } from '@codemirror/state'
import { Decoration, EditorView, ViewPlugin, type DecorationSet, type ViewUpdate } from '@codemirror/view'

// Matches CodeMirror's default left padding of a line.
const LINE_PADDING_PX = 6
const LEADING_SPACE = /^[ \t]*/

// A wrapped line continues under its own indentation (hanging indent), as in Postman.
function hangingIndents(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  for (const { from, to } of view.visibleRanges) {
    for (let position = from; position <= to; ) {
      const line = view.state.doc.lineAt(position)
      const width = LEADING_SPACE.exec(line.text)?.[0].replace(/\t/g, ' '.repeat(view.state.tabSize)).length ?? 0
      if (width > 0) {
        builder.add(
          line.from,
          line.from,
          Decoration.line({ attributes: { style: `padding-left: calc(${width}ch + ${LINE_PADDING_PX}px); text-indent: -${width}ch` } }),
        )
      }
      position = line.to + 1
    }
  }
  return builder.finish()
}

export const wrapIndent: Extension = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = hangingIndents(view)
    }

    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = hangingIndents(update.view)
      }
    }
  },
  { decorations: (plugin) => plugin.decorations },
)
