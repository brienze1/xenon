package dto

func (t TreeNode) MethodName() string {
	if t.Method == nil {
		return ""
	}
	return *t.Method
}
