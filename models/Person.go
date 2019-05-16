package models

import "github.com/jinzhu/gorm"

type Person struct {
	gorm.Model
	Name string `json:"name"`
	Hash string
}

func (person *Person) AfterCreate(scope *gorm.Scope) error {
	ID := int(person.ID)
	hash := generateHash(ID)
	scope.DB().Model(person).Updates(Person{Hash: hash})
	return nil
}

func CreatePerson(db *gorm.DB, name string, actionTaken string, id int) {
	person := Person{
		Name: name,
	}
	db.Create(&person)
}
