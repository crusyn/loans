package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SharedLoan holds the schema definition for the SharedLoan entity.
type SharedLoan struct {
	ent.Schema
}

// Fields of the SharedLoan.
func (SharedLoan) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("loan_id"),
	}
}

// Edges of the SharedLoan.
func (SharedLoan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("shared_loan").
			Field("user_id").
			Required().
			Unique(),
		edge.From("loan", Loan.Type).
			Ref("shared_loan").
			Field("loan_id").
			Required().
			Unique(),
	}
}
