package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	assert.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)

	assert.NoError(t, err, "Ошибка добавления в БД")
	assert.NotZero(t, id, "ID должен быть > 0")

	// get
	p, err := store.Get(id)

	assert.NoError(t, err, "Ошибка получения из БД")
	assert.Equal(t, parcel.Address, p.Address, "Address не совпадают")
	assert.Equal(t, parcel.Client, p.Client, "Client не совпадают")
	assert.Equal(t, parcel.Status, p.Status, "Status не совпадают")
	assert.Equal(t, parcel.CreatedAt, p.CreatedAt, "CreatedAt не совпадают")

	// delete
	err = store.Delete(id)

	assert.NoError(t, err, "Ошибка удаления из БД")

	p, err = store.Get(id)

	assert.Error(t, err, "Нет ошибки при получении ранее удаленной записи")

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	assert.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)

	assert.NoError(t, err, "Ошибка добавления в БД")

	// set address
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)

	assert.NoError(t, err, "Ошибка при обновлении адреса для записи")

	// check
	p, err := store.Get(id)

	assert.NoError(t, err, "Ошибка получения из БД")

	assert.Equal(t, newAddress, p.Address, "Адрес записи в БД не совпадает с обновленным ранее")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	assert.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	id, err := store.Add(parcel)

	assert.NoError(t, err, "Ошибка добавления в БД")
	assert.NotZero(t, id, "ID должен быть > 0")

	// set status
	err = store.SetStatus(id, ParcelStatusSent)

	assert.NoError(t, err, "Ошибка обновления статуса записи")

	// check
	p, err := store.Get(id)

	assert.NoError(t, err, "Ошибка получения из БД")
	assert.Equal(t, ParcelStatusSent, p.Status, "Ошибка обновления статуса для записи в БД")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	assert.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		assert.NoError(t, err, "Ошибка добавления в БД")

		parcels[i].Number = id

		parcelMap[i] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client

	assert.NoError(t, err, "Ошибка получения списка посылок")
	assert.Equal(t, 3, len(storedParcels), "Количество полученных записей меньше 3-х (добавленных)")
	// check
	for id, parcel := range storedParcels {
		p, ok := parcelMap[id]
		assert.True(t, ok, "Индекса нет в мапе")
		assert.Equal(t, p.Address, parcel.Address, "Address fields are not equal")
		assert.Equal(t, p.Client, parcel.Client, "Clients fields are not equal")
		assert.Equal(t, p.Status, parcel.Status, "Status fields are not equal")
		assert.Equal(t, p.CreatedAt, parcel.CreatedAt, "CreatedAt fields are not equal")
	}
}
