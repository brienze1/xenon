package entity

import "sort"

type WorktreeUpdate struct {
	Writes    map[string]string
	Removals  []string
	Conflicts []string
}

func NewWorktreeUpdate(current, target map[string]string, worktree SyncedFiles) WorktreeUpdate {
	update := WorktreeUpdate{Writes: map[string]string{}, Removals: []string{}, Conflicts: []string{}}
	for _, path := range unionOfPaths(current, target) {
		incoming := target[path]
		switch {
		case incoming == current[path], worktree[path] == incoming:
		case worktree.Differs(current, path):
			update.Conflicts = append(update.Conflicts, path)
		case incoming == "":
			update.Removals = append(update.Removals, path)
		default:
			update.Writes[path] = incoming
		}
	}
	return update
}

func (w WorktreeUpdate) HasConflicts() bool {
	return len(w.Conflicts) > 0
}

func (w WorktreeUpdate) IsEmpty() bool {
	return len(w.Writes) == 0 && len(w.Removals) == 0
}

func (w WorktreeUpdate) ConflictPlans(current, target map[string]string, worktree SyncedFiles) []PathPlan {
	plans := make([]PathPlan, 0, len(w.Conflicts))
	for _, path := range w.Conflicts {
		plans = append(plans, PathPlan{Path: path, Base: current[path], Mine: worktree[path], Theirs: target[path]})
	}
	return plans
}

func unionOfPaths(first, second map[string]string) []string {
	seen := map[string]bool{}
	paths := make([]string, 0, len(first)+len(second))
	for _, files := range []map[string]string{first, second} {
		for path := range files {
			if !seen[path] {
				seen[path] = true
				paths = append(paths, path)
			}
		}
	}
	sort.Strings(paths)
	return paths
}

// Resolved replaces whatever the update planned for path with the content a conflict resolution chose.
func (w WorktreeUpdate) Resolved(path, desired string, worktree SyncedFiles) WorktreeUpdate {
	resolved := WorktreeUpdate{Writes: map[string]string{}, Removals: withoutPath(w.Removals, path), Conflicts: withoutPath(w.Conflicts, path)}
	for planned, hash := range w.Writes {
		if planned != path {
			resolved.Writes[planned] = hash
		}
	}
	switch {
	case worktree[path] == desired:
	case desired == "":
		resolved.Removals = append(resolved.Removals, path)
	default:
		resolved.Writes[path] = desired
	}
	return resolved
}

func withoutPath(paths []string, path string) []string {
	kept := make([]string, 0, len(paths))
	for _, candidate := range paths {
		if candidate != path {
			kept = append(kept, candidate)
		}
	}
	return kept
}
