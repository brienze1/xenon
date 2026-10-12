import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useTransition,
} from "react";
import type { AppError } from "@/platform/errors";
import { GIT_HISTORY_WARNING } from "@/platform/mappers/variable";
import { patchRow, removeRow } from "@/shared/lib/rows";
import { Banner } from "@/shared/ui/banner/Banner";
import { Button } from "@/shared/ui/button/Button";
import { NoTranslateText } from "@/shared/ui/no-translate/NoTranslateText";
import { withNextStep } from "@/shared/lib/next-step";
import type { Result } from "@/shared/types/result";
import { useEnvironmentsStore } from "@/features/environments/store";
import {
  keyCounts,
  keyProblems,
  newDraftRow,
  type DraftRow,
} from "@/features/environments/lib/variable-draft";
import type { TableState } from "@/features/environments/model";
import { TypeSwitchConfirm } from "./TypeSwitchConfirm";
import { VariableTableRow, type RowActions } from "./VariableTableRow";
import styles from "./VariableTable.module.css";

const SAVE_NEXT_STEP =
  "Fix the highlighted rows, or check that the workspace folder is writable, then try again.";
const LOAD_NEXT_STEP =
  "Check that the workspace folder exists, then try again.";

const TABLE_HEAD = (
  <thead>
    <tr>
      <th scope="col">Name</th>
      <th scope="col">Initial Value</th>
      <th scope="col">Current Value</th>
      <th scope="col" className={styles.typeColumn}>
        Type
      </th>
      <th scope="col" className={styles.enabledColumn}>
        Enabled
      </th>
      <th scope="col">Description</th>
      <th scope="col" className={styles.actionColumn}>
        <span className="visually-hidden">Actions</span>
      </th>
    </tr>
  </thead>
);

const GIT_HISTORY_MESSAGE = (
  <Banner.Message>
    Values that were in the workspace files may still be in git history. Rotate
    them if the folder was ever pushed.
  </Banner.Message>
);

const STALE_MESSAGE = (
  <Banner.Message>
    This table changed outside this editor. Save to keep your edits, or discard
    them to see the new values.
  </Banner.Message>
);

type BulkAction = "reset" | "persist";

const BULK_MESSAGE: Record<BulkAction, string> = {
  reset:
    "Reset every current value to its initial value? The local overrides are removed.",
  persist:
    "Persist every current value as its initial value? The values are written to the workspace files.",
};

const BULK_CONFIRM: Record<BulkAction, string> = {
  reset: "Reset All Values",
  persist: "Persist All Values",
};

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
);

/**
 * How long unsaved edits live: `view` drops them when the table unmounts (dialogs guard before
 * closing); `workspace` keeps them in the store until saved or discarded (environment tabs).
 */
export type DraftLifetime = "view" | "workspace";

interface VariableTableProps {
  scope: string;
  label: string;
  draftLifetime?: DraftLifetime;
  /** In a main-area tab the tab saves (Mod+S, close prompt) and shows the unsaved dot, so the table has no save buttons. */
  savedByTab?: boolean;
}

export function VariableTable({
  scope,
  label,
  draftLifetime = "view",
  savedByTab = false,
}: VariableTableProps) {
  const table = useEnvironmentsStore((state) => state.tables[scope]);
  const hasDraft = useEnvironmentsStore(
    (state) => state.drafts[scope] !== undefined,
  );
  const loadTable = useEnvironmentsStore((state) => state.loadTable);
  const discardDraft = useEnvironmentsStore((state) => state.discardDraft);

  useEffect(() => {
    void loadTable(scope);
  }, [scope, loadTable]);

  useEffect(
    () => (draftLifetime === "view" ? () => discardDraft(scope) : undefined),
    [scope, draftLifetime, discardDraft],
  );

  if (!table || (table.status !== "ready" && table.rows.length === 0)) {
    return LOADING;
  }
  if (table.error && table.rows.length === 0) {
    return (
      <Banner.Root tone="error">
        <Banner.Message>
          <ErrorText error={table.error} nextStep={LOAD_NEXT_STEP} />
        </Banner.Message>
        <Banner.Actions>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => void loadTable(scope)}
          >
            Try Again
          </Button>
        </Banner.Actions>
      </Banner.Root>
    );
  }
  return hasDraft ? (
    <VariableTableEditor
      key={scope}
      scope={scope}
      label={label}
      table={table}
      savedByTab={savedByTab}
    />
  ) : (
    LOADING
  );
}

interface ErrorTextProps {
  error: AppError;
  nextStep: string;
}

function ErrorText({ error, nextStep }: ErrorTextProps) {
  return (
    <NoTranslateText
      text={withNextStep(error.description || error.code, nextStep)}
      terms={[error.code]}
    />
  );
}

interface VariableTableEditorProps {
  scope: string;
  label: string;
  table: TableState;
  savedByTab: boolean;
}

function VariableTableEditor({
  scope,
  label,
  table,
  savedByTab,
}: VariableTableEditorProps) {
  const draft = useEnvironmentsStore((state) => state.drafts[scope]);
  const editDraft = useEnvironmentsStore((state) => state.editDraft);
  const discardDraft = useEnvironmentsStore((state) => state.discardDraft);
  const saveDraft = useEnvironmentsStore((state) => state.saveDraft);
  const confirmDraftSave = useEnvironmentsStore(
    (state) => state.confirmDraftSave,
  );
  const cancelDraftSave = useEnvironmentsStore(
    (state) => state.cancelDraftSave,
  );
  const resetValues = useEnvironmentsStore((state) => state.resetValues);
  const persistValues = useEnvironmentsStore((state) => state.persistValues);
  const [confirmingBulk, setConfirmingBulk] = useState<BulkAction | null>(null);
  const [actionFailure, setActionFailure] = useState<AppError | null>(null);
  const [dismissedWarnings, setDismissedWarnings] = useState<string[] | null>(
    null,
  );
  const [pending, startAction] = useTransition();
  const tableRef = useRef<HTMLTableElement>(null);
  const pendingFocus = useRef<NewRowField | null>(null);

  useLayoutEffect(() => {
    const field = pendingFocus.current;
    if (!field) {
      return;
    }
    pendingFocus.current = null;
    const cells = tableRef.current?.querySelectorAll<HTMLInputElement>(
      `tbody input[name="${field}"]:not([data-new-row])`,
    );
    cells?.[cells.length - 1]?.focus();
  });

  if (!draft) {
    return LOADING;
  }

  const busy = pending || draft.saving;
  const confirming = draft.confirming;
  const failure = draft.failure ?? actionFailure;
  const stale = draft.dirty && table.rows !== draft.base;
  const hasOverrides = table.rows.some(
    (row) => !row.secret && row.hasCurrentOverride,
  );
  const showGitHistory =
    table.warnings.includes(GIT_HISTORY_WARNING) &&
    dismissedWarnings !== table.warnings;
  const counts = keyCounts(draft.rows);

  function edit(change: (rows: DraftRow[]) => DraftRow[]) {
    editDraft(scope, change);
  }

  // Typing into the empty last row turns it into a variable; the caret follows into the same cell of the new row.
  function addRow(field: NewRowField, value: string) {
    edit((rows) => [
      ...rows,
      { ...newDraftRow(), [field === "key" ? "key" : "initialValue"]: value },
    ]);
    pendingFocus.current = field;
  }

  function focusFirstInvalidName() {
    tableRef.current
      ?.querySelector<HTMLInputElement>(
        'input[name="key"][aria-invalid="true"]',
      )
      ?.focus();
  }

  const actions: RowActions = {
    onChange: (id, patch) => edit((rows) => patchRow(rows, id, patch)),
    onRemove: (id) => edit((rows) => removeRow(rows, id)),
    onReset: (key) => run(() => resetValues(scope, [key])),
    onPersist: (key) => run(() => persistValues(scope, [key])),
  };

  function run(operation: () => Promise<AppError | null>) {
    startAction(async () => {
      setActionFailure(await operation());
    });
  }

  function runBulk(action: BulkAction) {
    setConfirmingBulk(null);
    run(() =>
      action === "reset" ? resetValues(scope, []) : persistValues(scope, []),
    );
  }

  function discard() {
    setActionFailure(null);
    discardDraft(scope);
  }

  function afterSave(saved: Result<boolean, AppError>) {
    if (!saved.ok) {
      focusFirstInvalidName();
    }
  }

  function requestSave() {
    if (!draft?.dirty) {
      return;
    }
    setActionFailure(null);
    startAction(async () => afterSave(await saveDraft(scope)));
  }

  function confirmSave() {
    startAction(async () => afterSave(await confirmDraftSave(scope)));
  }

  return (
    <div className={styles.root}>
      <div className={styles.toolbar}>
        <span className={styles.spacer} />
        <Button
          variant="ghost"
          size="sm"
          disabled={busy || draft.dirty || !hasOverrides}
          onClick={() => setConfirmingBulk("reset")}
        >
          Reset All
        </Button>
        <Button
          variant="ghost"
          size="sm"
          disabled={busy || draft.dirty || !hasOverrides}
          onClick={() => setConfirmingBulk("persist")}
        >
          Persist All
        </Button>
      </div>
      {confirmingBulk ? (
        <Banner.Root tone="warning">
          <Banner.Message>{BULK_MESSAGE[confirmingBulk]}</Banner.Message>
          <Banner.Actions>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setConfirmingBulk(null)}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              size="sm"
              disabled={busy || draft.dirty}
              onClick={() => runBulk(confirmingBulk)}
            >
              {BULK_CONFIRM[confirmingBulk]}
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {showGitHistory ? (
        <Banner.Root tone="warning">
          {GIT_HISTORY_MESSAGE}
          <Banner.Actions>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setDismissedWarnings(table.warnings)}
            >
              Dismiss
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {stale ? <Banner.Root tone="info">{STALE_MESSAGE}</Banner.Root> : null}
      <div className={styles.scroller}>
        <table ref={tableRef} aria-label={label} className={styles.table}>
          {TABLE_HEAD}
          <tbody>
            {draft.rows.map((row, index) => (
              <VariableTableRow
                key={row.id}
                scope={scope}
                row={row}
                index={index}
                problems={keyProblems(row, counts)}
                savedActionsDisabled={busy || draft.dirty}
                actions={actions}
              />
            ))}
            <NewVariableRow onAdd={addRow} />
          </tbody>
        </table>
      </div>
      {confirming ? (
        <TypeSwitchConfirm
          switches={confirming}
          busy={busy}
          onCancel={() => cancelDraftSave(scope)}
          onConfirm={confirmSave}
        />
      ) : null}
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <ErrorText error={failure} nextStep={SAVE_NEXT_STEP} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      {savedByTab ? null : (
        <div className={styles.footer}>
          <Button
            variant="secondary"
            size="sm"
            disabled={!draft.dirty || busy}
            onClick={discard}
          >
            Discard Changes
          </Button>
          <Button
            variant="primary"
            size="sm"
            disabled={busy || confirming !== null}
            onClick={requestSave}
          >
            {busy ? "Saving…" : "Save Variables"}
          </Button>
        </div>
      )}
    </div>
  );
}

type NewRowField = "key" | "initial-value";

interface NewVariableRowProps {
  onAdd: (field: NewRowField, value: string) => void;
}

/** The always-empty last row, as in Postman: nothing is added until something is typed into it. */
function NewVariableRow({ onAdd }: NewVariableRowProps) {
  return (
    <tr className={styles.newRow}>
      <td>
        <input
          name="key"
          data-new-row=""
          autoComplete="off"
          aria-label="Add variable"
          className={styles.newCell}
          spellCheck={false}
          translate="no"
          placeholder="Add variable"
          title="Add variable"
          value=""
          onChange={(event) => onAdd("key", event.target.value)}
        />
      </td>
      <td>
        <input
          name="initial-value"
          data-new-row=""
          autoComplete="off"
          aria-label="Initial value of a new variable"
          className={styles.newCell}
          spellCheck={false}
          value=""
          onChange={(event) => onAdd("initial-value", event.target.value)}
        />
      </td>
      <td colSpan={5} />
    </tr>
  );
}
