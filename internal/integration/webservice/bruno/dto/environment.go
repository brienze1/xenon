package dto

import (
	"math"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

var neonColors = map[enums.EnvColor][3]float64{
	envcolor.Green:   {0x3D, 0xFF, 0x4F},
	envcolor.Cyan:    {0x3E, 0xC8, 0xFF},
	envcolor.Magenta: {0xFF, 0x3D, 0xFF},
	envcolor.Yellow:  {0xFF, 0xE5, 0x3D},
	envcolor.Red:     {0xFF, 0x3D, 0x5A},
}

type Environment struct {
	Name      string
	Color     string
	Variables []EnvironmentVariable
}

func NewEnvironment(name, color string) Environment {
	return Environment{Name: name, Color: color, Variables: []EnvironmentVariable{}}
}

func (e *Environment) ToEntity(newId func() string) entity.ImportVariables {
	environment := entity.NewEnvironment(newId(), e.Name)
	if color, found := nearestNeon(e.Color); found {
		environment.Color = color
		environment.Bruno = entity.SourceFields{"color": e.Color}
	}
	for _, variable := range e.Variables {
		row := entity.NewVariableRow(variable.Key, variable.Value)
		row.Enabled = variable.Enabled
		row.Secret = variable.Secret
		environment.Variables = append(environment.Variables, row.WithoutSecretValue())
	}
	return entity.NewImportVariables(environment, map[string]string{})
}

func nearestNeon(color string) (enums.EnvColor, bool) {
	rgb, valid := hexColor(color)
	if !valid {
		return envcolor.None, false
	}
	nearest, distance := envcolor.None, math.MaxFloat64
	for _, candidate := range []enums.EnvColor{envcolor.Green, envcolor.Cyan, envcolor.Magenta, envcolor.Yellow, envcolor.Red} {
		neon := neonColors[candidate]
		if current := math.Pow(rgb[0]-neon[0], 2) + math.Pow(rgb[1]-neon[1], 2) + math.Pow(rgb[2]-neon[2], 2); current < distance {
			nearest, distance = candidate, current
		}
	}
	return nearest, true
}

func hexColor(color string) ([3]float64, bool) {
	hex := strings.TrimPrefix(strings.TrimSpace(color), "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	value, err := strconv.ParseUint(hex, 16, 32)
	if len(hex) != 6 || err != nil {
		return [3]float64{}, false
	}
	return [3]float64{float64(value >> 16 & 0xFF), float64(value >> 8 & 0xFF), float64(value & 0xFF)}, true
}
