package models

import "github.com/jinzhu/gorm"

type Person struct {
	gorm.Model
	Name        string
	ActionTaken string
	Hash        string
	TaskID      int
}

func (person *Person) AfterCreate(scope *gorm.Scope) error {
	ID := int(person.ID)
	hash := generateHash(ID)
	scope.DB().Model(person).Updates(Person{Hash: hash})
	return nil
}
