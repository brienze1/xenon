package property

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"sort"
	"testing"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/pathaction"
)

const (
	editPairs = 100
	seed      = 20261011
)

type edit int

const (
	untouched edit = iota
	modified
	deleted
)

type side struct {
	files  map[string]string
	fields map[string]string
	rows   map[string]string
	order  []string
	name   string
}

type workspace struct {
	files  map[string]string
	fields map[string]string
	rows   map[string]string
	order  []string
}

func TestMergeEngineLosesNoChangeAndReportsEveryOverlap(t *testing.T) {
	random := rand.New(rand.NewPCG(seed, seed))
	for pair := 0; pair < editPairs; pair++ {
		base := generateWorkspace(random)
		mine := editWorkspace(random, base, "mine")
		theirs := editWorkspace(random, base, "theirs")
		if err := checkFiles(base, mine, theirs); err != nil {
			t.Fatalf("pair %d files: %v", pair, err)
		}
		if err := checkCollection(base, mine, theirs); err != nil {
			t.Fatalf("pair %d collection.yaml: %v", pair, err)
		}
		if err := checkEnvironment(base, mine, theirs); err != nil {
			t.Fatalf("pair %d environment: %v", pair, err)
		}
	}
}

func generateWorkspace(random *rand.Rand) workspace {
	ws := workspace{files: map[string]string{}, fields: map[string]string{"name": "Billing", "description": "base"}, rows: map[string]string{}}
	for index := 0; index < 3+random.IntN(8); index++ {
		file := fmt.Sprintf("r%d.yaml", index)
		ws.files["collections/billing/"+file] = "base-" + file
		ws.order = append(ws.order, file)
	}
	for index := 0; index < 2+random.IntN(5); index++ {
		ws.rows[fmt.Sprintf("v%d", index)] = "base"
	}
	return ws
}

func editWorkspace(random *rand.Rand, base workspace, name string) side {
	edited := side{name: name, files: map[string]string{}, fields: map[string]string{}, rows: map[string]string{}}
	for path, content := range base.files {
		switch edit(random.IntN(3)) {
		case modified:
			edited.files[path] = content + "-" + name + fmt.Sprint(random.IntN(2))
		case untouched:
			edited.files[path] = content
		}
	}
	if random.IntN(2) == 0 {
		edited.files[fmt.Sprintf("collections/billing/shared-%d.yaml", random.IntN(2))] = "added-" + fmt.Sprint(random.IntN(2))
	}
	for field, value := range base.fields {
		edited.fields[field] = value
		if random.IntN(2) == 0 {
			edited.fields[field] = value + "-" + name + fmt.Sprint(random.IntN(2))
		}
	}
	for _, entry := range base.order {
		if random.IntN(4) != 0 {
			edited.order = append(edited.order, entry)
		}
	}
	if random.IntN(2) == 0 {
		edited.order = append(edited.order, name+"-new.yaml")
	}
	for key, value := range base.rows {
		switch edit(random.IntN(3)) {
		case modified:
			edited.rows[key] = value + "-" + name + fmt.Sprint(random.IntN(2))
		case untouched:
			edited.rows[key] = value
		}
	}
	if random.IntN(2) == 0 {
		edited.rows["added-"+name] = "new"
	}
	return edited
}

func checkFiles(base workspace, mine, theirs side) error {
	plan := usecase.PlanThreeWay(base.files, mine.files, theirs.files)
	for _, path := range plan.Paths {
		b, m, th := base.files[path.Path], mine.files[path.Path], theirs.files[path.Path]
		switch {
		case m != b && th != b && m != th:
			if !path.IsConflict() {
				return fmt.Errorf("%s changed differently on both sides but planned %s", path.Path, path.Action)
			}
		case m != b:
			if path.IsConflict() || path.Result() != m {
				return fmt.Errorf("%s: my change lost (%s)", path.Path, path.Action)
			}
		case th != b:
			if path.IsConflict() || path.Result() != th {
				return fmt.Errorf("%s: their change lost (%s)", path.Path, path.Action)
			}
		default:
			if path.Action != pathaction.Keep || path.Result() != b {
				return fmt.Errorf("%s untouched but planned %s", path.Path, path.Action)
			}
		}
	}
	return nil
}

func checkCollection(base workspace, mine, theirs side) error {
	merged := usecase.MergeStructuralFile(itemkind.Collection, collectionDoc(base.fields, base.order, base.rows), collectionDoc(mine.fields, mine.order, mine.rows), collectionDoc(theirs.fields, theirs.order, theirs.rows), nil)
	expected := map[string]bool{}
	for field, value := range base.fields {
		if want, conflict := threeWay(value, true, mine.fields[field], true, theirs.fields[field], true); conflict {
			expected[field] = true
		} else if got := merged.Result.Field(field); got == nil || got.Value != want {
			return fmt.Errorf("field %s = %v, want %q", field, got, want)
		}
	}
	rowConflicts, err := checkRows(base.rows, mine, theirs, merged.Result.Field("variables"), "variables")
	if err != nil {
		return err
	}
	for path := range rowConflicts {
		expected[path] = true
	}
	reported := map[string]bool{}
	for _, conflict := range merged.FieldConflicts {
		reported[conflict.Path] = true
	}
	if !reflect.DeepEqual(reported, expected) {
		return fmt.Errorf("conflicts %v, want %v", keys(reported), keys(expected))
	}
	return checkOrder(base.order, mine.order, theirs.order, merged.Result.Field("order"))
}

func checkEnvironment(base workspace, mine, theirs side) error {
	merged := usecase.MergeStructuralFile(itemkind.Environment, environmentDoc(base.rows), environmentDoc(mine.rows), environmentDoc(theirs.rows), nil)
	conflicts, err := checkRows(base.rows, mine, theirs, merged.Result.Field("variables"), "variables")
	if len(conflicts) > 0 {
		if !merged.ItemConflict || len(merged.FieldConflicts) != 0 {
			return fmt.Errorf("overlapping rows %v must make the environment an item conflict", keys(conflicts))
		}
		return nil
	}
	if merged.ItemConflict {
		return fmt.Errorf("unexpected item conflict")
	}
	return err
}

func checkRows(base map[string]string, mine, theirs side, merged *entity.DocumentNode, field string) (map[string]bool, error) {
	conflicts := map[string]bool{}
	_, rows, _ := merged.KeyedRows()
	all := map[string]bool{}
	for _, source := range []map[string]string{base, mine.rows, theirs.rows} {
		for key := range source {
			all[key] = true
		}
	}
	for key := range all {
		b, inBase := base[key]
		m, inMine := mine.rows[key]
		th, inTheirs := theirs.rows[key]
		want, conflict := threeWay(b, inBase, m, inMine, th, inTheirs)
		path := field + "[" + key + "]"
		if conflict {
			if inMine && inTheirs {
				path += ".value"
			}
			conflicts[path] = true
			continue
		}
		row, present := rows[key]
		if want == absent {
			if present {
				return nil, fmt.Errorf("row %s was deleted but is still present", key)
			}
			continue
		}
		if !present || row.Field("value").Value != want {
			return nil, fmt.Errorf("row %s = %v, want %q", key, row, want)
		}
	}
	return conflicts, nil
}

const absent = "\x00absent"

func threeWay(base string, inBase bool, mine string, inMine bool, theirs string, inTheirs bool) (string, bool) {
	b, m, t := present(base, inBase), present(mine, inMine), present(theirs, inTheirs)
	switch {
	case m == t:
		return m, false
	case m == b:
		return t, false
	case t == b:
		return m, false
	default:
		return "", true
	}
}

func present(value string, found bool) string {
	if !found {
		return absent
	}
	return value
}

func checkOrder(base, mine, theirs []string, merged *entity.DocumentNode) error {
	values, ok := merged.ScalarItems()
	if !ok {
		return fmt.Errorf("order is not a list")
	}
	inBase, inMine, inTheirs, inResult := set(base), set(mine), set(theirs), set(values)
	if len(inResult) != len(values) {
		return fmt.Errorf("order has duplicates: %v", values)
	}
	for entry := range union(inBase, inMine, inTheirs) {
		deleted := inBase[entry] && (!inMine[entry] || !inTheirs[entry])
		added := !inBase[entry] && (inMine[entry] || inTheirs[entry])
		kept := inBase[entry] && inMine[entry] && inTheirs[entry]
		if inResult[entry] != (added || kept) || (deleted && inResult[entry]) {
			return fmt.Errorf("order entry %s: in result %v (added %v, kept %v, deleted %v) — %v", entry, inResult[entry], added, kept, deleted, values)
		}
	}
	return nil
}

func collectionDoc(fields map[string]string, order []string, rows map[string]string) *entity.DocumentNode {
	node := entity.NewMappingNode()
	for _, field := range []string{"name", "description"} {
		node.Set(field, entity.NewScalarNode("!!str", fields[field]))
	}
	items := make([]*entity.DocumentNode, 0, len(order))
	for _, entry := range order {
		items = append(items, entity.NewScalarNode("!!str", entry))
	}
	node.Set("order", entity.NewSequenceNode(items))
	node.Set("variables", rowsDoc(rows))
	return node
}

func environmentDoc(rows map[string]string) *entity.DocumentNode {
	return entity.NewMappingNode().Set("name", entity.NewScalarNode("!!str", "hml")).Set("variables", rowsDoc(rows))
}

func rowsDoc(rows map[string]string) *entity.DocumentNode {
	items := make([]*entity.DocumentNode, 0, len(rows))
	for _, key := range keys(toBool(rows)) {
		items = append(items, entity.NewMappingNode().Set("key", entity.NewScalarNode("!!str", key)).Set("value", entity.NewScalarNode("!!str", rows[key])))
	}
	return entity.NewSequenceNode(items)
}

func set(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func union(sets ...map[string]bool) map[string]bool {
	result := map[string]bool{}
	for _, values := range sets {
		for value := range values {
			result[value] = true
		}
	}
	return result
}

func toBool(values map[string]string) map[string]bool {
	result := map[string]bool{}
	for key := range values {
		result[key] = true
	}
	return result
}

func keys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
