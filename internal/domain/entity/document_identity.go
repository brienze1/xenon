package entity

const (
	idField    = "id"
	nameField  = "name"
	orderField = "order"
	stringTag  = "!!str"
)

// Renamed copies a mapping document with a new id and name; other fields are shared, not cloned.
func (d *DocumentNode) Renamed(id, name string) *DocumentNode {
	renamed := NewMappingNode()
	for _, key := range d.Keys {
		renamed.Set(key, d.Fields[key])
	}
	return renamed.Set(idField, NewScalarNode(stringTag, id)).Set(nameField, NewScalarNode(stringTag, name))
}

// WithOrderEntryAfter inserts added right after existing in the order list; false when the list does not name existing.
func (d *DocumentNode) WithOrderEntryAfter(existing, added string) (*DocumentNode, bool) {
	entries, scalars := d.Field(orderField).ScalarItems()
	if !scalars {
		return d, false
	}
	items := make([]*DocumentNode, 0, len(entries)+1)
	found := false
	for index, entry := range entries {
		items = append(items, d.Field(orderField).Items[index])
		if entry == existing && !found {
			found = true
			items = append(items, NewScalarNode(stringTag, added))
		}
	}
	if !found {
		return d, false
	}
	updated := NewMappingNode()
	for _, key := range d.Keys {
		updated.Set(key, d.Fields[key])
	}
	return updated.Set(orderField, NewSequenceNode(items)), true
}
