package gol

type BookListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Category string
}

// TODO: Implement the interface methods for each struct