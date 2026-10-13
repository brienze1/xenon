package entity

type BrunoCollection struct {
	Environments []ImportVariables
	Scripts      []BrunoScripts
	Warnings     []ImportWarning
	Root         ImportNode
}
