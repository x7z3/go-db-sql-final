package main

import (
	"database/sql"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	r, err := s.db.Exec(`
	INSERT INTO parcel (client, status, address, created_at) 
	VALUES (?, ?, ?, ?)
	`, p.Client, p.Status, p.Address, p.CreatedAt)

	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	r := s.db.QueryRow(`
	SELECT number, client, status, address, created_at FROM parcel
	WHERE number = ? 
	LIMIT 1
	`, number)

	p := Parcel{}

	err := r.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel

	rows, err := s.db.Query(`
	SELECT number, client, status, address, created_at FROM parcel
	WHERE client = ?
	`, client)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Parcel{}

		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(`
	UPDATE parcel
	SET status=?
	WHERE number=?
	`, status, number)

	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec(`
	UPDATE parcel
	SET address=?
	WHERE number=? AND status=?
	`, address, number, ParcelStatusRegistered)

	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	_, err := s.db.Exec(`
	DELETE FROM parcel
	WHERE number=? AND status=?
	`, number, ParcelStatusRegistered)

	if err != nil {
		return err
	}

	return nil
}
