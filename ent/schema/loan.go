package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Loan holds the schema definition for the Loan entity.
type Loan struct {
	ent.Schema
}

// Fields of the Loan.
func (Loan) Fields() []ent.Field {
	return []ent.Field{
		field.Int("amount"), // like other currency fields we should amount in cents to avoid floating point math
		field.Float("rate"),
		field.Int("term"), // In months
		field.Int("borrower_id"),
	}
}

// Edges of the Loan.
func (Loan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("borrower", User.Type).
			Ref("loans").
			Field("borrower_id").
			Required().
			Unique(),
		edge.To("shared_loan", SharedLoan.Type),
	}
}
