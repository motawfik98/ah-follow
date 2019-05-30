package models

import "github.com/jinzhu/gorm"

// this struct is to store the `sent to` in each task
type Person struct {
	gorm.Model
	Name          string `json:"name"`
	ActionTaken   string `json:"action_taken"`
	FinalResponse bool   `json:"final_response" gorm:"default:0"`
	Hash          string
	TaskID        uint
}

// this function generates the hash then update the Person created
func (person *Person) AfterCreate(scope *gorm.Scope) error {
	ID := int(person.ID)
	hash := generateHash(ID)
	scope.DB().Model(person).Updates(Person{Hash: hash})
	return nil
}

// this function takes the parameters required to create a new Person, then adds him to the database
func CreatePerson(db *gorm.DB, name string, actionTaken string, id uint, finalResponse bool) int {
	person := Person{
		Name:          name,
		ActionTaken:   actionTaken,
		TaskID:        id,
		FinalResponse: finalResponse,
	}
	db.Create(&person)
	return int(person.ID)
}
