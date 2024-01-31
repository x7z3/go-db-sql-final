package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

	require.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)

	require.NoError(t, err, "Ошибка добавления в БД")
	require.NotZero(t, id, "ID должен быть > 0")

	parcel.Number = id

	// get
	p, err := store.Get(id)

	require.NoError(t, err, "Ошибка получения из БД")
	require.Equal(t, parcel, p, "Добавленная посылка не равна полученной из БД")

	// delete
	err = store.Delete(id)

	require.NoError(t, err, "Ошибка удаления из БД")

	p, err = store.Get(id)

	require.Error(t, err, "Нет ошибки при получении ранее удаленной записи")

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	require.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)

	require.NoError(t, err, "Ошибка добавления в БД")
	require.NotZero(t, id, "Id не должен равняться нулю")

	// set address
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)

	require.NoError(t, err, "Ошибка при обновлении адреса для записи")

	// check
	p, err := store.Get(id)

	require.NoError(t, err, "Ошибка получения из БД")

	require.Equal(t, newAddress, p.Address, "Адрес записи в БД не совпадает с обновленным ранее")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	require.NoError(t, err, "SQL DB Connection error")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	id, err := store.Add(parcel)

	require.NoError(t, err, "Ошибка добавления в БД")
	require.NotZero(t, id, "ID должен быть > 0")

	// set status
	err = store.SetStatus(id, ParcelStatusSent)

	require.NoError(t, err, "Ошибка обновления статуса записи")

	// check
	p, err := store.Get(id)

	require.NoError(t, err, "Ошибка получения из БД")
	require.Equal(t, ParcelStatusSent, p.Status, "Ошибка обновления статуса для записи в БД")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД

	require.NoError(t, err, "SQL DB Connection error")

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
		require.NoError(t, err, "Ошибка добавления в БД")

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client

	require.NoError(t, err, "Ошибка получения списка посылок")
	require.Equal(t, len(parcels), len(storedParcels), "Количество полученных записей меньше 3-х (добавленных)")
	// check
	for _, parcel := range storedParcels {
		p, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Индекса нет в мапе")
		require.Equal(t, p, parcel, "Посылка полученная с БД не равна добавленной")
	}
}
