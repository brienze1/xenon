package entity

type TreeMove struct {
	ID             string
	FileName       string
	SourceParentID string
	TargetParentID string
	SourceOrder    []string
	TargetOrder    []string
}

func NewTreeMove(item, source, target TreeNode, fileName string, index int) TreeMove {
	move := TreeMove{ID: item.ID, FileName: fileName, SourceParentID: source.ID, TargetParentID: target.ID}
	if move.IsReorder() {
		move.TargetOrder = target.OrderMoving(item.FileName, index)
		return move
	}
	move.SourceOrder = source.OrderRemoving(item.FileName)
	move.TargetOrder = target.OrderInserting(fileName, index)
	return move
}

func (t TreeMove) IsReorder() bool {
	return t.SourceParentID == t.TargetParentID
}
