import type { SnippetLanguage } from "@/platform/mappers/snippet";
import type { SnippetOptionsState } from "@/platform/mappers/ui-state";
import type { CodeLanguage } from "@/shared/ui/code-editor/languages";

export const DEFAULT_SNIPPET_LANGUAGE = "curl";

/** The language last used: the last remembered entry, or cURL. */
export function currentSnippetLanguage(
  saved: readonly SnippetOptionsState[],
  languages: readonly SnippetLanguage[],
): string {
  const known = new Set(languages.map((language) => language.id));
  const last = [...saved]
    .reverse()
    .find((entry) => known.size === 0 || known.has(entry.language));
  return last?.language ?? languages[0]?.id ?? DEFAULT_SNIPPET_LANGUAGE;
}

const NO_OPTIONS: Readonly<Record<string, string>> = {};

/** The options remembered for a language (the generator applies its defaults to the rest). */
export function rememberedOptions(
  saved: readonly SnippetOptionsState[],
  language: string,
): Readonly<Record<string, string>> {
  return (
    saved.find((entry) => entry.language === language)?.options ?? NO_OPTIONS
  );
}

/** The value an option control shows: the remembered one or the generator's default. */
export function optionValue(
  options: Readonly<Record<string, string>>,
  id: string,
  defaultValue: string,
): string {
  return options[id] ?? defaultValue;
}

const EDITOR_LANGUAGES: Readonly<Record<string, CodeLanguage>> = {
  curl: "shell",
  js_fetch: "javascript",
  go_native: "go",
  python_requests: "python",
};

export function editorLanguageOf(language: string): CodeLanguage {
  return EDITOR_LANGUAGES[language] ?? "text";
}
