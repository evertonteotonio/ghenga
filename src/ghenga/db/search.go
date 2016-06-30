package db

// FuzzyFindPersons searches the database for persons related to the query
// string.
func (db *Database) FuzzyFindPersons(query string) ([]*Person, error) {
	var result []*Person

	err := db.dbmap.Select(&result, "SELECT * FROM people WHERE name ILIKE $1", "%"+query+"%")
	if err != nil {
		return nil, err
	}

	return result, nil
}
