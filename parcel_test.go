package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
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
	db, err := sql.Open("sqlite", "./tracker.db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	// настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("store.Add: %v", err)
	}
	if id <= 0 {
		t.Fatalf("store.Add: invalid id")
	}
	parcel.Number = id
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("store.Get: %v", err)
	}
	if got != parcel {
		t.Fatalf("store.Get: got %v, want %v", got, parcel)
	}

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("store.Delete: %v", err)
	}
	_, err = store.Get(id)
	if err == nil {
		t.Fatalf("store.Delete: invalid id")
	}

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("store.Add: %v", err)
	}
	if id <= 0 {
		t.Fatalf("store.Add: invalid id")
	}

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	if err := store.SetAddress(id, newAddress); err != nil {
		t.Fatalf("store.SetAddress: %v", err)
	}

	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("store.Get: %v", err)
	}
	if got.Address != newAddress {
		t.Fatalf("store.Get: got %v, want %v", got, newAddress)
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("store.Add: %v", err)
	}
	if id <= 0 {
		t.Fatalf("store.Add: invalid id")
	}
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	if err := store.SetStatus(id, newStatus); err != nil {
		t.Fatalf("store.SetStatus: %v", err)
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("store.Get: %v", err)
	}
	if got.Status != newStatus {
		t.Fatalf("store.Get: got %v, want %v", got, newStatus)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close() // настройте подключение к БД

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
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("store.Add: %v", err)
		}
		if id <= 0 {
			t.Fatalf("store.Add: invalid id")
		}
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("store.GetByClient: %v", err)
	}
	if len(storedParcels) != len(parcels) {
		t.Fatalf("store.GetByClient: got %v, want %v", len(storedParcels), len(parcels))
	}
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		if !ok {
			t.Fatalf("store.GetByClient: invalid parcel %v", parcel)
		}
		if parcel.Client != expected.Client {
			t.Fatalf("store.GetByClient: got %v, want %v", parcel.Client, expected.Client)
		}
		if parcel.Status != expected.Status {
			t.Fatalf("store.GetByClient: got %v, want %v", parcel.Status, expected.Status)
		}
		if parcel.CreatedAt != expected.CreatedAt {
			t.Fatalf("store.GetByClient: got %v, want %v", parcel.CreatedAt, expected.CreatedAt)
		}
		if parcel.Address != expected.Address {
			t.Fatalf("store.GetByClient: got %v, want %v", parcel.Address, expected.Address)
		}
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
