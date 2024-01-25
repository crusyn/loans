package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		// Having a unique social security number ensures we don't insert the same user twice.
		// We should remember to hash the social and only store the last 4 digits
		// before we take this to production.
		field.String("social").Unique(),
		field.String("address").
			Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
