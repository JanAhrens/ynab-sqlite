package main

// https://api.youneedabudget.com/v1/budgets/:id/categories
type CategoryGroup struct {
	Id string
	Name string
	Hidden bool
	Deleted bool
	Categories []Category
}

// https://api.youneedabudget.com/v1/budgets/:id/categories
type Category struct {
	Id string
	Name string
	CategoryGroup CategoryGroup
	Hidden bool
	Note string
	Budgeted int
	GoalType string
	GoalCreationMonth string
	GoalTarget int
	GoalTargetMonth string
}

