package webservice

import (
	"bufio"
	"context"
	"encoding/json"
	"math/big"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

const dynamicValuesGolden = "testdata/dynamic_values.golden.json"

func seededDynamicValues(t *testing.T) *dynamicValuesWebservice {
	t.Helper()
	frozen := time.Date(2026, time.October, 9, 12, 34, 56, 789000000, time.UTC)
	timeutils.SetCurrentTimeProvider(func() time.Time { return frozen })
	t.Cleanup(func() { timeutils.SetCurrentTimeProvider(time.Now) })
	return NewDynamicValuesWebservice(rand.NewPCG(1, 2)).(*dynamicValuesWebservice)
}

func vendoredDynamicNames(t *testing.T) []string {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "..", "test", "integration", "resources", "postman-dynamic-variables.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	names := make([]string, 0, 118)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" && !strings.HasPrefix(line, "#") {
			names = append(names, line)
		}
	}
	return names
}

func TestDynamicNamesMatchTheVendoredPostmanList(t *testing.T) {
	names := seededDynamicValues(t).Names(context.Background())
	vendored := vendoredDynamicNames(t)
	if len(vendored) != 118 || !slices.Equal(names, vendored) {
		t.Fatalf("names differ from the vendored list:\n got %v\nwant %v", names, vendored)
	}
	names[0] = "changed"
	if seededDynamicValues(t).Names(context.Background())[0] != "$guid" {
		t.Error("Names must return a copy")
	}
}

func TestUnknownDynamicNamesAreNotGenerated(t *testing.T) {
	webservice := seededDynamicValues(t)
	for _, name := range []string{"$nope", "$GUID", "guid", ""} {
		if value, found := webservice.Generate(context.Background(), name); found || value != "" {
			t.Errorf("Generate(%q) = %q, %v", name, value, found)
		}
	}
}

func TestSeededDynamicValuesMatchTheGoldenFile(t *testing.T) {
	webservice := seededDynamicValues(t)
	generated := map[string][]string{}
	for _, name := range webservice.Names(context.Background()) {
		for range 3 {
			value, found := webservice.Generate(context.Background(), name)
			if !found {
				t.Fatalf("%s is not generated", name)
			}
			generated[name] = append(generated[name], value)
		}
	}
	if os.Getenv("UPDATE_GOLDEN") != "" {
		encoded, _ := json.MarshalIndent(generated, "", "  ")
		if err := os.WriteFile(dynamicValuesGolden, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(dynamicValuesGolden)
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string][]string
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	for name, values := range golden {
		if !slices.Equal(generated[name], values) {
			t.Errorf("%s: got %q, want %q", name, generated[name], values)
		}
	}
	if len(golden) != len(generated) {
		t.Errorf("golden has %d names, generated %d", len(golden), len(generated))
	}
}

func TestIbansCarryValidCheckDigits(t *testing.T) {
	webservice := seededDynamicValues(t)
	for range 500 {
		iban, _ := webservice.Generate(context.Background(), "$randomBankAccountIban")
		rearranged := iban[4:] + iban[:4]
		var digits strings.Builder
		for _, character := range rearranged {
			if character >= 'A' && character <= 'Z' {
				digits.WriteString(big.NewInt(int64(character - 'A' + 10)).String())
				continue
			}
			digits.WriteRune(character)
		}
		number, _ := new(big.Int).SetString(digits.String(), 10)
		if new(big.Int).Mod(number, big.NewInt(97)).Int64() != 1 {
			t.Fatalf("%s fails the mod 97 check", iban)
		}
	}
}
