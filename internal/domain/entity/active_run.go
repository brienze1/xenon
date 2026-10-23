package entity

type ActiveRun struct {
	Control *RunControl
	Run     Run
}

func ActiveRunOf(active []ActiveRun, itemId string) (Run, bool) {
	for _, run := range active {
		if run.Run.Target.ItemID == itemId {
			return run.Run, true
		}
	}
	return Run{}, false
}
