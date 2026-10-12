const INDENT = '  '

/**
 * Re-indents a JSON body without parsing it, so `{{variables}}` outside strings,
 * big numbers, duplicate keys and key order survive exactly as typed.
 * Returns null when the brackets or strings do not balance.
 */
export function beautifyJson(text: string): string | null {
  let out = ''
  let depth = 0
  let inString = false
  let escaped = false
  const stack: string[] = []

  const newline = () => `\n${INDENT.repeat(depth)}`

  for (let index = 0; index < text.length; index++) {
    const char = text[index] as string

    if (inString) {
      out += char
      if (escaped) {
        escaped = false
      } else if (char === '\\') {
        escaped = true
      } else if (char === '"') {
        inString = false
      }
      continue
    }

    if (/\s/.test(char)) {
      continue
    }

    if (char === '{' && text[index + 1] === '{') {
      const end = text.indexOf('}}', index + 2)
      if (end === -1) {
        return null
      }
      out += text.slice(index, end + 2)
      index = end + 1
      continue
    }

    switch (char) {
      case '"':
        inString = true
        out += char
        break
      case '{':
      case '[': {
        const expected = char === '{' ? '}' : ']'
        const close = nextSignificant(text, index + 1)
        if (close === expected) {
          out += char + close
          index = text.indexOf(close, index + 1)
          break
        }
        stack.push(expected)
        depth++
        out += char + newline()
        break
      }
      case '}':
      case ']':
        if (stack.pop() !== char) {
          return null
        }
        depth--
        out += newline() + char
        break
      case ',':
        out += char + newline()
        break
      case ':':
        out += ': '
        break
      default:
        out += char
    }
  }

  return inString || stack.length > 0 ? null : out
}

function nextSignificant(text: string, from: number): string | undefined {
  for (let index = from; index < text.length; index++) {
    const char = text[index] as string
    if (!/\s/.test(char)) {
      return char
    }
  }
  return undefined
}
