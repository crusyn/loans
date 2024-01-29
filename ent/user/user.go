// Code generated by ent, DO NOT EDIT.

package user

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldSocial holds the string denoting the social field in the database.
	FieldSocial = "social"
	// FieldAddress holds the string denoting the address field in the database.
	FieldAddress = "address"
	// EdgeLoans holds the string denoting the loans edge name in mutations.
	EdgeLoans = "loans"
	// Table holds the table name of the user in the database.
	Table = "users"
	// LoansTable is the table that holds the loans relation/edge.
	LoansTable = "loans"
	// LoansInverseTable is the table name for the Loan entity.
	// It exists in this package in order to avoid circular dependency with the "loan" package.
	LoansInverseTable = "loans"
	// LoansColumn is the table column denoting the loans relation/edge.
	LoansColumn = "borrower_id"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldSocial,
	FieldAddress,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// OrderOption defines the ordering options for the User queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// BySocial orders the results by the social field.
func BySocial(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSocial, opts...).ToFunc()
}

// ByAddress orders the results by the address field.
func ByAddress(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAddress, opts...).ToFunc()
}

// ByLoansCount orders the results by loans count.
func ByLoansCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newLoansStep(), opts...)
	}
}

// ByLoans orders the results by loans terms.
func ByLoans(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newLoansStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newLoansStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(LoansInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, LoansTable, LoansColumn),
	)
}
