package entity

type BrunoCollection struct {
	Root         ImportNode
	Environments []ImportVariables
	Scripts      []BrunoScripts
	Warnings     []ImportWarning
}
