package gocosmos_test

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/gocosmos"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/*======================================================================*/

const numApps = 4
const numLogicalPartitions = 16
const numCategories = 19

var dataList []gocosmos.DocInfo

func _initDataSubPartitions(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	totalRu := 0.0
	randList := make([]int, numItem)
	for i := 0; i < numItem; i++ {
		randList[i] = i*2 + 1
	}
	rand.Shuffle(numItem, func(i, j int) {
		randList[i], randList[j] = randList[j], randList[i]
	})
	dataList = make([]gocosmos.DocInfo, numItem)
	for i := 0; i < numItem; i++ {
		category := randList[i] % numCategories
		app := "app" + strconv.Itoa(i%numApps)
		username := "user" + strconv.Itoa(i%numLogicalPartitions)
		docInfo := gocosmos.DocInfo{
			"id":       fmt.Sprintf("%05d", i),
			"app":      app,
			"username": username,
			"email":    "user" + strconv.Itoa(i) + "@domain.com",
			"grade":    float64(randList[i]),
			"category": float64(category),
			"active":   i%10 == 0,
			"big":      fmt.Sprintf("%05d", i) + "/" + strings.Repeat("this is a very long string/", 256),
		}
		dataList[i] = docInfo
		if result := client.CreateDocument(gocosmos.DocumentSpec{DbName: db, CollName: container, PartitionKeyValues: []interface{}{app, username}, DocumentData: docInfo}); result.Error() != nil {
			t.Fatalf("%s failed: %s", testName, result.Error())
		} else {
			totalRu += result.RequestCharge
		}
	}
	// fmt.Printf("\t%s - total RU charged: %0.3f\n", testName+"/Insert", totalRu)
}

func _initDataSubPartitionsSmallRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 400})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/app", "/username"}, "kind": "MultiHash", "version": 2},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               400,
	})
	_initDataSubPartitions(t, testName, client, db, container, numItem)
}

func _initDataSubPartitionsLargeRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 20000})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/app", "/username"}, "kind": "MultiHash", "version": 2},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               20000,
	})
	_initDataSubPartitions(t, testName, client, db, container, numItem)
}

func _initData(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	totalRu := 0.0
	randList := make([]int, numItem)
	for i := 0; i < numItem; i++ {
		randList[i] = i*2 + 1
	}
	rand.Shuffle(numItem, func(i, j int) {
		randList[i], randList[j] = randList[j], randList[i]
	})
	dataList = make([]gocosmos.DocInfo, numItem)
	for i := 0; i < numItem; i++ {
		category := randList[i] % numCategories
		username := "user" + strconv.Itoa(i%numLogicalPartitions)
		docInfo := gocosmos.DocInfo{
			"id":       fmt.Sprintf("%05d", i),
			"username": username,
			"email":    "user" + strconv.Itoa(i) + "@domain.com",
			"grade":    float64(randList[i]),
			"category": float64(category),
			"active":   i%10 == 0,
			"big":      fmt.Sprintf("%05d", i) + "/" + strings.Repeat("this is a very long string/", 256),
		}
		dataList[i] = docInfo
		if result := client.CreateDocument(gocosmos.DocumentSpec{DbName: db, CollName: container, PartitionKeyValues: []interface{}{username}, DocumentData: docInfo}); result.Error() != nil {
			t.Fatalf("%s failed: %s", testName, result.Error())
		} else {
			totalRu += result.RequestCharge
		}
	}
	// fmt.Printf("\t%s - total RU charged: %0.3f\n", testName+"/Insert", totalRu)
}

func _initDataSmallRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 400})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/username"}, "kind": "Hash"},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               400,
	})
	_initData(t, testName, client, db, container, numItem)
}

func _initDataLargeRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string, numItem int) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 20000})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/username"}, "kind": "Hash"},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               20000,
	})
	_initData(t, testName, client, db, container, numItem)
}

/*----------------------------------------------------------------------*/

func _initDataFamilies(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	dataMapFamilies = make(map[string]gocosmos.DocInfo)
	dataListFamilies = make([]gocosmos.DocInfo, 0)
	err := json.Unmarshal([]byte(_testDataFamilies), &dataListFamilies)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	for _, doc := range dataListFamilies {
		if result := client.CreateDocument(gocosmos.DocumentSpec{DbName: db, CollName: container, PartitionKeyValues: []interface{}{doc["id"]}, DocumentData: doc}); result.Error() != nil {
			t.Fatalf("%s failed: %s", testName, result.Error())
		}
		dataMapFamilies[doc.Id()] = doc
	}
}

func _initDataFamliesSmallRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 400})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		Ru:               400,
	})
	_initDataFamilies(t, testName, client, db, container)
}

func _initDataFamliesLargeRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 20000})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               20000,
	})
	_initDataFamilies(t, testName, client, db, container)
}

func _initDataVolcano(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	dataListVolcano = make([]gocosmos.DocInfo, 0)
	err := json.Unmarshal([]byte(_testDataVolcano), &dataListVolcano)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	for _, doc := range dataListVolcano {
		if result := client.CreateDocument(gocosmos.DocumentSpec{DbName: db, CollName: container, PartitionKeyValues: []interface{}{doc["id"]}, DocumentData: doc}); result.Error() != nil {
			t.Fatalf("%s failed: %s", testName, result.Error())
		}
	}
}

func _initDataVolcanoSmallRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 400})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		Ru:               400,
	})
	_initDataVolcano(t, testName, client, db, container)
}

func _initDataVolcanoLargeRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 20000})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               20000,
	})
	_initDataVolcano(t, testName, client, db, container)
}

var dataListFamilies, dataListVolcano []gocosmos.DocInfo
var dataMapFamilies map[string]gocosmos.DocInfo

func _toJson(data interface{}) string {
	js, _ := json.Marshal(data)
	return string(js)
}

const _testDataFamilies = `
[
  {
    "id": "AndersenFamily",
    "lastName": "Andersen",
    "parents": [
      {
        "firstName": "Thomas"
      },
      {
        "firstName": "Mary Kay"
      }
    ],
    "children": [
      {
        "firstName": "Henriette Thaulow",
        "gender": "female",
        "grade": 5,
        "pets": [
          {
            "givenName": "Fluffy"
          }
        ]
      }
    ],
    "address": {
      "state": "WA",
      "county": "King",
      "city": "Seattle"
    },
    "creationDate": 1431620472,
    "isRegistered": true
  },
  {
    "id": "WakefieldFamily",
    "parents": [
      {
        "familyName": "Wakefield",
        "givenName": "Robin"
      },
      {
        "familyName": "Miller",
        "givenName": "Ben"
      }
    ],
    "children": [
      {
        "familyName": "Merriam",
        "givenName": "Jesse",
        "gender": "female",
        "grade": 1,
        "pets": [
          {
            "givenName": "Goofy"
          },
          {
            "givenName": "Shadow"
          }
        ]
      },
      {
        "familyName": "Miller",
        "givenName": "Lisa",
        "gender": "female",
        "grade": 8
      }
    ],
    "address": {
      "state": "NY",
      "county": "Manhattan",
      "city": "NY"
    },
    "creationDate": 1431620462,
    "isRegistered": false
  }
]
`

const _testDataVolcano = `
[
  {
    "Volcano Name": "Abu",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        131.6,
        34.5
      ]
    },
    "Elevation": 571,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "4cb67ab0-ba1a-0e8a-8dfc-d48472fd5766"
  },
  {
    "Volcano Name": "Acamarachi",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.62,
        -23.3
      ]
    },
    "Elevation": 6046,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "246927ec-11c6-56da-b97c-00e5ed69fd3f"
  },
  {
    "Volcano Name": "Acatenango",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.876,
        14.501
      ]
    },
    "Elevation": 3976,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a6297b2d-d004-8caa-bc42-a349ff046bc4"
  },
  {
    "Volcano Name": "Acigol-Nevsehir",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        34.52,
        38.57
      ]
    },
    "Elevation": 1689,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cd080a05-b245-b78a-0dbe-1cb32eac3a74"
  },
  {
    "Volcano Name": "Adams",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.49,
        46.206
      ]
    },
    "Elevation": 3742,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "9e3c494e-8367-3f50-1f56-8c6fcb961363"
  },
  {
    "Volcano Name": "Adatara",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.28,
        37.62
      ]
    },
    "Elevation": 1718,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "81ed06ee-8319-4555-dbd4-74923ad4130a"
  },
  {
    "Volcano Name": "Adwa",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.84,
        10.07
      ]
    },
    "Elevation": 1733,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6802c282-225a-2fb9-db1a-3e1b7af39a59"
  },
  {
    "Volcano Name": "Afdera",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.85,
        13.08
      ]
    },
    "Elevation": 1295,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "1edbcfd7-408e-954b-2dd6-2cefd0291fd4"
  },
  {
    "Volcano Name": "Agmagan-Karadag",
    "Country": "Armenia",
    "Region": "Armenia",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.75,
        40.275
      ]
    },
    "Elevation": 3560,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7971f48f-af85-78eb-112e-945bd26ba468"
  },
  {
    "Volcano Name": "Agrigan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.67,
        18.77
      ]
    },
    "Elevation": 965,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e26b4342-2bdf-6970-6a7c-7f4c78de3eb0"
  },
  {
    "Volcano Name": "Agua",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.743,
        14.465
      ]
    },
    "Elevation": 3760,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4bf9255d-f545-4376-b097-df4ca4da9689"
  },
  {
    "Volcano Name": "Agua de Pau",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.47,
        37.77
      ]
    },
    "Elevation": 947,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "d44c94b6-81f8-4b27-4970-f79b149529d3"
  },
  {
    "Volcano Name": "Aguilera",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.83,
        -50.17
      ]
    },
    "Elevation": 0,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "7408d446-fb51-e70e-9955-560a0c966b68"
  },
  {
    "Volcano Name": "Agung",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        115.508,
        -8.342
      ]
    },
    "Elevation": 3142,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "670b15d5-6c84-4f9e-0836-4e1b87f02673"
  },
  {
    "Volcano Name": "Ahyi",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.03,
        20.42
      ]
    },
    "Elevation": -137,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e581b14e-480f-dffc-85b5-ae9f1b6ab4d6"
  },
  {
    "Volcano Name": "Akademia Nauk",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.45,
        53.98
      ]
    },
    "Elevation": 1180,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0054666d-0a62-542a-2e6b-8614260f3124"
  },
  {
    "Volcano Name": "Akademia Nauk",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.45,
        53.983
      ]
    },
    "Elevation": 1180,
    "Type": "Stratovolcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "767afe71-aa85-a080-2ae5-bc2a3792c27d"
  },
  {
    "Volcano Name": "Akagi",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.18,
        36.53
      ]
    },
    "Elevation": 1828,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "a8dc9372-32e9-a067-d040-0f7513e60db8"
  },
  {
    "Volcano Name": "Akan",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.02,
        43.38
      ]
    },
    "Elevation": 1499,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "947ad983-c955-4ace-40bd-fa0f23eaa721"
  },
  {
    "Volcano Name": "Akhtang",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.65,
        55.43
      ]
    },
    "Elevation": 1956,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f38dec4b-4591-ccee-a668-943e4943e758"
  },
  {
    "Volcano Name": "Akita-Komaga-take",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.8,
        39.75
      ]
    },
    "Elevation": 1637,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "993427a7-4be8-af07-90cd-ee92b0c71bb0"
  },
  {
    "Volcano Name": "Akita-Yake-yama",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.77,
        39.97
      ]
    },
    "Elevation": 1366,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9037065f-18c0-5faa-b002-360059d0f051"
  },
  {
    "Volcano Name": "Akuseki-jima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.597,
        29.461
      ]
    },
    "Elevation": 584,
    "Type": "Stratovolcanoes",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "7f0304a0-fa2d-471b-d079-ab562429a127"
  },
  {
    "Volcano Name": "Akuseki-jima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.6,
        29.45
      ]
    },
    "Elevation": 586,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d6241563-02e6-60bb-cac8-71b5e7e94bc1"
  },
  {
    "Volcano Name": "Akutan",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -165.97,
        54.13
      ]
    },
    "Elevation": 1303,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "35971ea9-b611-b44f-f4f5-b6161eafae74"
  },
  {
    "Volcano Name": "Alaid",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.55,
        50.858
      ]
    },
    "Elevation": 2339,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ef9da8b3-173b-399c-201d-4ad045128994"
  },
  {
    "Volcano Name": "Alamagan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.83,
        17.6
      ]
    },
    "Elevation": 744,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "54f43f96-5d10-e87d-0051-0cb4799e58f0"
  },
  {
    "Volcano Name": "Alayta",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.57,
        12.88
      ]
    },
    "Elevation": 1501,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c223a890-f75b-7fc8-b2b3-1b29c36555a0"
  },
  {
    "Volcano Name": "Albano, Monte",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        12.7,
        41.73
      ]
    },
    "Elevation": 949,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "c3bb1fef-fcb1-56e3-0389-f88583c3ce0d"
  },
  {
    "Volcano Name": "Alcedo, Volcan",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.12,
        -0.43
      ]
    },
    "Elevation": 1130,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "877a940e-e580-2479-1b50-e33985624ef2"
  },
  {
    "Volcano Name": "Ale Bagu",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.63,
        13.52
      ]
    },
    "Elevation": 1031,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dc269bca-e610-9cd8-553c-93fb3019ed61"
  },
  {
    "Volcano Name": "Alid",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.92,
        14.88
      ]
    },
    "Elevation": 904,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0be79920-dda4-dcd7-9312-3bc6a717cd39"
  },
  {
    "Volcano Name": "Alligator Lake",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -135.42,
        60.42
      ]
    },
    "Elevation": 2217,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4eca56ec-987b-f1a1-8f5d-6cd89edcb3e3"
  },
  {
    "Volcano Name": "Almolonga",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.48,
        14.82
      ]
    },
    "Elevation": 3197,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "13af25d2-9636-41d2-45f2-0d9052ed3263"
  },
  {
    "Volcano Name": "Alney-Chashakondzha",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.65,
        56.7
      ]
    },
    "Elevation": 2598,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "461c8ff1-cb91-01d8-5f8e-8b518a5fe5ea"
  },
  {
    "Volcano Name": "Alngey",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.4,
        57.7
      ]
    },
    "Elevation": 1853,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a2af9028-874f-f97d-e39f-05bca3f8e25d"
  },
  {
    "Volcano Name": "Alu",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.55,
        13.82
      ]
    },
    "Elevation": 429,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a0b4d81c-5dbc-f91b-ea8c-96f4377280b4"
  },
  {
    "Volcano Name": "Alutu",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.78,
        7.77
      ]
    },
    "Elevation": 2335,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "afe9b90a-ebc6-ca7e-0968-88f5c9e15319"
  },
  {
    "Volcano Name": "Amak",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.15,
        55.42
      ]
    },
    "Elevation": 513,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "bc66b762-937f-5266-661d-25af8e506369"
  },
  {
    "Volcano Name": "Amasing",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.5,
        -0.55
      ]
    },
    "Elevation": 1030,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e527e160-d953-0dea-03aa-87d3f101d6c0"
  },
  {
    "Volcano Name": "Ambalatungan Group",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.1,
        17.32
      ]
    },
    "Elevation": 2329,
    "Type": "Compound volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "545f3bb5-2bf2-54e3-b919-19ebd58d8e12"
  },
  {
    "Volcano Name": "Ambang",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.42,
        0.75
      ]
    },
    "Elevation": 1795,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ac3d7c60-3349-b57d-256e-a57b8320fabc"
  },
  {
    "Volcano Name": "Ambitle",
    "Country": "Papua New Guinea",
    "Region": "New Ireland-SW Pacif",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.65,
        -4.08
      ]
    },
    "Elevation": 450,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ff6223b4-06fc-caf4-c16a-4c6034d1eb98"
  },
  {
    "Volcano Name": "Amboy",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -115.78,
        34.55
      ]
    },
    "Elevation": 288,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eddf3848-cfa2-8f3d-9774-4acbf0f35f45"
  },
  {
    "Volcano Name": "Ambre-Bobaomby",
    "Country": "Madagascar",
    "Region": "Madagascar",
    "Location": {
      "type": "Point",
      "coordinates": [
        49.1,
        -12.48
      ]
    },
    "Elevation": 1475,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c6540ff4-87a5-1490-041a-39a858c6bfd2"
  },
  {
    "Volcano Name": "Ambrym",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.12,
        -16.25
      ]
    },
    "Elevation": 1334,
    "Type": "Pyroclastic shield",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "22b7cf2c-4a9c-9ff3-51ce-9973eba2bd3e"
  },
  {
    "Volcano Name": "Amorong",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.805,
        15.828
      ]
    },
    "Elevation": 376,
    "Type": "Unknown",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "989919f0-6eee-3f9f-7c49-3f648811509d"
  },
  {
    "Volcano Name": "Amsterdam Island",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        77.52,
        -37.83
      ]
    },
    "Elevation": 881,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d0b6abc3-4ac6-ca05-5dff-840a296de750"
  },
  {
    "Volcano Name": "Amukta",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -171.25,
        52.5
      ]
    },
    "Elevation": 1066,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "e1617302-f42f-fbb8-5570-6843eb244cde"
  },
  {
    "Volcano Name": "Anatahan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.67,
        16.35
      ]
    },
    "Elevation": 788,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b9ae0093-791d-a289-abc9-deceebc12ff8"
  },
  {
    "Volcano Name": "Anaun",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.83,
        56.32
      ]
    },
    "Elevation": 1828,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "36145835-b8fd-de58-bed8-8171b316c4c0"
  },
  {
    "Volcano Name": "Andahua Valley",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.33,
        -15.42
      ]
    },
    "Elevation": 4713,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0434f857-91b7-d817-8163-8a9c1437da35"
  },
  {
    "Volcano Name": "Andrus",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -132.33,
        -75.8
      ]
    },
    "Elevation": 2978,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "f79ffb06-5ed0-4eb4-c146-44bfd8e2a743"
  },
  {
    "Volcano Name": "Aneityum",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        169.83,
        -20.2
      ]
    },
    "Elevation": 852,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "b5b5a510-6af1-4681-ce9e-fd17db02277d"
  },
  {
    "Volcano Name": "Aniakchak",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -158.17,
        56.88
      ]
    },
    "Elevation": 1341,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b7d647f3-7d8b-d4d8-8239-68f9d96343ad"
  },
  {
    "Volcano Name": "Ankaizina Field",
    "Country": "Madagascar",
    "Region": "Madagascar",
    "Location": {
      "type": "Point",
      "coordinates": [
        48.67,
        -14.3
      ]
    },
    "Elevation": 2878,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c011a810-6686-e39b-9aa8-ddb99653b92a"
  },
  {
    "Volcano Name": "Ankaratra Field",
    "Country": "Madagascar",
    "Region": "Madagascar",
    "Location": {
      "type": "Point",
      "coordinates": [
        47.2,
        -19.4
      ]
    },
    "Elevation": 2644,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eabf717d-a74a-352c-f6a8-35f0aedf9ec3"
  },
  {
    "Volcano Name": "Antillanca Group",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.153,
        -40.771
      ]
    },
    "Elevation": 1990,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "44db6cc6-db8b-43ba-28b2-8f17d9d071ba"
  },
  {
    "Volcano Name": "Antipodes Island",
    "Country": "New Zealand",
    "Region": "Pacific-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.767,
        -49.683
      ]
    },
    "Elevation": 402,
    "Type": "Pyroclastic cones",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "ca021eec-d2ea-4405-c158-96cf42aba5f7"
  },
  {
    "Volcano Name": "Antisana",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.141,
        -0.481
      ]
    },
    "Elevation": 5753,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "19654a02-76a6-c78a-02bb-0a5fd2b7a8c2"
  },
  {
    "Volcano Name": "Antofagasta de la Sierra",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.5,
        -26.08
      ]
    },
    "Elevation": 4000,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e13eb27a-3d26-f1f3-d4c2-549451990869"
  },
  {
    "Volcano Name": "Antofagasta de la Sierra",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.5,
        -26.083
      ]
    },
    "Elevation": 4000,
    "Type": "Scoria cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e00eb831-c762-fd6e-c35b-b196563e444e"
  },
  {
    "Volcano Name": "Antofalla",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68,
        -25.53
      ]
    },
    "Elevation": 6100,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c9ba5e67-c546-844e-da77-21f25370371d"
  },
  {
    "Volcano Name": "Antuco",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.349,
        -37.406
      ]
    },
    "Elevation": 2979,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "3dadd649-ef04-e7f4-973f-ee77bbc99e6d"
  },
  {
    "Volcano Name": "Aoba",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.83,
        -15.4
      ]
    },
    "Elevation": 1496,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "45cefa28-7958-2c6f-a392-8aa6696f7f90"
  },
  {
    "Volcano Name": "Aoga-shima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.77,
        32.45
      ]
    },
    "Elevation": 423,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "a55f3ccd-cb7c-1eb0-87d9-42aab950b7ab"
  },
  {
    "Volcano Name": "Apaneca Range",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.786,
        13.891
      ]
    },
    "Elevation": 2036,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "009641f2-2f40-87c3-1170-fadcfaf3069c"
  },
  {
    "Volcano Name": "Apastepeque Field",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.77,
        13.72
      ]
    },
    "Elevation": 700,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7ce2129c-e071-a869-4186-8f9601e249e1"
  },
  {
    "Volcano Name": "Apo",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.273,
        6.987
      ]
    },
    "Elevation": 2954,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4734c07b-d7cf-761d-1425-efe1c0dcc86d"
  },
  {
    "Volcano Name": "Apoyeque",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.342,
        12.242
      ]
    },
    "Elevation": 518,
    "Type": "Pyroclastic shield",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "2d45109c-babc-fbb1-abd3-b58ab0fabf5d"
  },
  {
    "Volcano Name": "Aracar",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.77,
        -24.27
      ]
    },
    "Elevation": 6082,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "38b8c21f-5068-20ff-9add-a4cb033ee164"
  },
  {
    "Volcano Name": "Aracar",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.767,
        -24.25
      ]
    },
    "Elevation": 6082,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "ed2cbae9-6496-43ab-fc10-f09cc545f896"
  },
  {
    "Volcano Name": "Aragats",
    "Country": "Armenia",
    "Region": "Armenia",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.12,
        40.55
      ]
    },
    "Elevation": 4090,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7033abb7-f273-e66d-e6b0-dcbd76983ece"
  },
  {
    "Volcano Name": "Aramuaca, Laguna",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.105,
        13.428
      ]
    },
    "Elevation": 180,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4f7adb9e-0a94-334c-e7a6-e002b233cccf"
  },
  {
    "Volcano Name": "Ararat",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.28,
        39.7
      ]
    },
    "Elevation": 5165,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c0db7559-9662-bd60-30dd-df7834dbb6b5"
  },
  {
    "Volcano Name": "Arayat",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.742,
        15.2
      ]
    },
    "Elevation": 1026,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8b0e5458-71e1-3002-9c1c-4dd5175f9f6c"
  },
  {
    "Volcano Name": "Ardoukoba",
    "Country": "Djibouti",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.47,
        11.58
      ]
    },
    "Elevation": 298,
    "Type": "Fissure vent",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "73b16e0e-b976-ba4a-2198-37341e874649"
  },
  {
    "Volcano Name": "Arenal",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -84.703,
        10.463
      ]
    },
    "Elevation": 1657,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5fef8698-88e1-1f36-9fd8-5d23cec0da67"
  },
  {
    "Volcano Name": "Arenales",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.483,
        -47.2
      ]
    },
    "Elevation": 3437,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d313723b-b5e2-9fae-7d40-aeb1215a02bf"
  },
  {
    "Volcano Name": "Arhab, Harra of",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.08,
        15.63
      ]
    },
    "Elevation": 3100,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "2519a08b-f0b6-bef5-0aa8-8117f8218985"
  },
  {
    "Volcano Name": "Arjuno-Welirang",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.58,
        -7.725
      ]
    },
    "Elevation": 3339,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "fc64f9bf-4df4-5662-3565-f63e5dc9ec4b"
  },
  {
    "Volcano Name": "Arshan",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.7,
        47.5
      ]
    },
    "Elevation": null,
    "Type": "Cinder cones",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "0bd87c2e-8ab3-432e-8745-f7ce59b5b4b9"
  },
  {
    "Volcano Name": "Asama",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.53,
        36.4
      ]
    },
    "Elevation": 2560,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "003dd58b-3cf1-9093-bf9d-2d8d08cf5521"
  },
  {
    "Volcano Name": "Asavyo",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.6,
        13.07
      ]
    },
    "Elevation": 1200,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a18af74e-d421-0a8f-59d3-ef86a61811d8"
  },
  {
    "Volcano Name": "Ascension",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -14.37,
        -7.95
      ]
    },
    "Elevation": 858,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "de336a06-df33-62dc-b2b4-7f1450c975dc"
  },
  {
    "Volcano Name": "Askja",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.75,
        65.03
      ]
    },
    "Elevation": 1516,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "72c18795-b674-cdaf-117f-5eba81988d23"
  },
  {
    "Volcano Name": "Aso",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        131.1,
        32.88
      ]
    },
    "Elevation": 1592,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a67ef0f1-6438-7d06-17e7-3f0904f1fcde"
  },
  {
    "Volcano Name": "Assab Volc Field",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.43,
        12.95
      ]
    },
    "Elevation": 987,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a66cbbb6-de30-65fd-5199-96d17327a9b2"
  },
  {
    "Volcano Name": "Asuncion",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.4,
        19.67
      ]
    },
    "Elevation": 857,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "38080c8d-d495-ef83-6f73-371d930f7deb"
  },
  {
    "Volcano Name": "Atacazo",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.617,
        -0.353
      ]
    },
    "Elevation": 4463,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "0c88d416-3c0d-550d-0f00-9692fc6e3812"
  },
  {
    "Volcano Name": "Atakor Volc Field",
    "Country": "Algeria",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        5.833,
        23.333
      ]
    },
    "Elevation": 2918,
    "Type": "Scoria cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b2872c4c-3001-9e0d-68fc-6fab296d18d8"
  },
  {
    "Volcano Name": "Atitlan",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.186,
        14.583
      ]
    },
    "Elevation": 3535,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "95773841-cfbb-241f-00d5-c2ffbd88d7d2"
  },
  {
    "Volcano Name": "Atka",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.15,
        52.38
      ]
    },
    "Elevation": 1533,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2dbc466a-0183-e3dc-8083-31f1c94efae1"
  },
  {
    "Volcano Name": "Atlasova",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.65,
        57.97
      ]
    },
    "Elevation": 1764,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a11212aa-c166-0d86-b14a-5423079d520e"
  },
  {
    "Volcano Name": "Atlixcos, Cerro Los",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -96.526,
        19.809
      ]
    },
    "Elevation": 800,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d7ce555f-3d2a-0abc-5956-3eef2a71b78f"
  },
  {
    "Volcano Name": "Atsonupuri",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.13,
        44.804
      ]
    },
    "Elevation": 1205,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ee50b1c6-644e-6ad8-7b21-c788b45f45ad"
  },
  {
    "Volcano Name": "Atuel, Caldera del",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.05,
        -34.65
      ]
    },
    "Elevation": 5189,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0e39b054-3c72-cbef-511a-a1460787395c"
  },
  {
    "Volcano Name": "Auckland Field",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        174.87,
        -36.9
      ]
    },
    "Elevation": 260,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "82d8401c-5e05-5156-06a5-b3ee6639d55c"
  },
  {
    "Volcano Name": "Augustine",
    "Country": "United States",
    "Region": "Alaska-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -153.42,
        59.37
      ]
    },
    "Elevation": 1252,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "34f5fd05-c023-b845-42fa-d58bb5f3c967"
  },
  {
    "Volcano Name": "Avachinsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.83,
        53.255
      ]
    },
    "Elevation": 2741,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ea683539-477a-50ad-7f58-f86c0697c1dd"
  },
  {
    "Volcano Name": "Awu",
    "Country": "Indonesia",
    "Region": "Sangihe Is-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.5,
        3.67
      ]
    },
    "Elevation": 1320,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "682efb9c-ddec-96ef-7664-3b4b7ab05be1"
  },
  {
    "Volcano Name": "Axial Seamount",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130,
        45.95
      ]
    },
    "Elevation": -1500,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "14f5ae05-fb6f-f989-4239-05c4bfb7a3c7"
  },
  {
    "Volcano Name": "Ayelu",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.702,
        10.082
      ]
    },
    "Elevation": 2145,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2aa37dce-f10e-23eb-c3c3-eeb1d4a28a4e"
  },
  {
    "Volcano Name": "Azufral, Volcan",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.68,
        1.08
      ]
    },
    "Elevation": 4070,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "b90da8c7-2324-9882-3608-23500183af5f"
  },
  {
    "Volcano Name": "Azul, Cerro",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.42,
        -0.9
      ]
    },
    "Elevation": 1690,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9ea3d20d-ae99-9f7e-9c70-69781d4d7c84"
  },
  {
    "Volcano Name": "Azul, Cerro [Quizapu]",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.761,
        -35.653
      ]
    },
    "Elevation": 3788,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1e0d17ef-8829-cb64-3089-ed897ddc6233"
  },
  {
    "Volcano Name": "Azul, Volcan",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -83.87,
        12.53
      ]
    },
    "Elevation": 201,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "417b8267-b75d-2c47-6b95-3f6081afaf64"
  },
  {
    "Volcano Name": "Azuma",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.25,
        37.73
      ]
    },
    "Elevation": 2024,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f48b85f5-bc5a-31e7-8506-c2dd6d07ed68"
  },
  {
    "Volcano Name": "Babuyan Claro",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.94,
        19.523
      ]
    },
    "Elevation": 1180,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e31f6a7f-351f-b113-611d-440c403b9115"
  },
  {
    "Volcano Name": "Bachelor",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.688,
        43.979
      ]
    },
    "Elevation": 2763,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "3d7ab64b-f71a-07cd-86f9-7a47f14d2244"
  },
  {
    "Volcano Name": "Bagana",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.195,
        -6.14
      ]
    },
    "Elevation": 1750,
    "Type": "Lava cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f76f6c9f-23a7-8ea5-2d37-46632f97a567"
  },
  {
    "Volcano Name": "Bakenin",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.07,
        53.905
      ]
    },
    "Elevation": 2278,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fad71886-a263-78b7-1af6-3190b77d94fe"
  },
  {
    "Volcano Name": "Baker",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.82,
        48.786
      ]
    },
    "Elevation": 3285,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "1d1913ad-fff9-c2aa-2ff9-9f67deda3578"
  },
  {
    "Volcano Name": "Bal Haf, Harra of",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        48.33,
        14.05
      ]
    },
    "Elevation": 233,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9a347c65-d2dc-749a-6b7b-a126ec79befa"
  },
  {
    "Volcano Name": "Balatocan",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.92,
        8.8
      ]
    },
    "Elevation": 2450,
    "Type": "Compound volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c1b1e852-fff8-0b1b-3aca-5e77ed56cc5a"
  },
  {
    "Volcano Name": "Balbi",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.98,
        -5.83
      ]
    },
    "Elevation": 2715,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b469a510-e7ac-1251-2556-f01f15039484"
  },
  {
    "Volcano Name": "Bald Knoll",
    "Country": "United States",
    "Region": "US-Utah",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.408,
        37.328
      ]
    },
    "Elevation": 2135,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "b3a51bfb-13ca-a3c5-2ee2-44e59ed4d266"
  },
  {
    "Volcano Name": "Baluan",
    "Country": "United States",
    "Region": "Admiralty Is-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.28,
        -2.57
      ]
    },
    "Elevation": 254,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "63b6909e-c991-18e5-3d5a-ad40cdb23f6d"
  },
  {
    "Volcano Name": "Baluran",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        114.37,
        -7.85
      ]
    },
    "Elevation": 1247,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "8a4dfebf-21de-a8ea-281e-f0ab5b86ebd0"
  },
  {
    "Volcano Name": "Balut",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.375,
        5.4
      ]
    },
    "Elevation": 852,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c5b03a9c-d242-6c72-5290-0c6cf20b681d"
  },
  {
    "Volcano Name": "Bam",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.85,
        -3.6
      ]
    },
    "Elevation": 685,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "1ad58577-ae8e-f0b8-6a06-de9e2d2dcf6e"
  },
  {
    "Volcano Name": "Bamus",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.23,
        -5.2
      ]
    },
    "Elevation": 2248,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b58d00d0-4be8-cb65-6b51-b3d07f82f3b3"
  },
  {
    "Volcano Name": "Banahao",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.48,
        14.07
      ]
    },
    "Elevation": 2158,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3480dffb-00a5-efad-bad5-21547ac94cb9"
  },
  {
    "Volcano Name": "Banda Api",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.871,
        -4.525
      ]
    },
    "Elevation": 640,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1aeb479d-9b7a-4a59-1c28-35f0adc06240"
  },
  {
    "Volcano Name": "Bandai",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.08,
        37.6
      ]
    },
    "Elevation": 1819,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "23ab07cd-6635-047b-dd96-f05833027ff6"
  },
  {
    "Volcano Name": "Banua Wuhu",
    "Country": "Indonesia",
    "Region": "Sangihe Is-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.491,
        3.138
      ]
    },
    "Elevation": -5,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a1d7f038-11cb-a215-7bc4-50f1eb456847"
  },
  {
    "Volcano Name": "Baransky",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.02,
        45.1
      ]
    },
    "Elevation": 1132,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "1673197a-1479-6b1e-363e-440ebd3c8235"
  },
  {
    "Volcano Name": "Barcena",
    "Country": "Mexico",
    "Region": "Mexico-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -110.82,
        19.3
      ]
    },
    "Elevation": 332,
    "Type": "Cinder cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2b2fdadc-a059-29b4-5761-eafb2cb07f35"
  },
  {
    "Volcano Name": "Bardarbunga",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.53,
        64.63
      ]
    },
    "Elevation": 2000,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "af682842-f50b-33cd-4140-f1b721192559"
  },
  {
    "Volcano Name": "Barkhatnaya Sopka",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.27,
        52.823
      ]
    },
    "Elevation": 870,
    "Type": "Lava dome",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "39187c90-ef91-c3a9-1689-0be046c24e62"
  },
  {
    "Volcano Name": "Barren Island",
    "Country": "India",
    "Region": "Andaman Is-Indian O",
    "Location": {
      "type": "Point",
      "coordinates": [
        93.875,
        12.292
      ]
    },
    "Elevation": 354,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9fe273a6-32e6-4e12-3f60-0798f701e2fe"
  },
  {
    "Volcano Name": "Barrier, The",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.57,
        2.32
      ]
    },
    "Elevation": 1032,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ca7f5089-11e8-1ca2-b4ea-fcebfa4058c8"
  },
  {
    "Volcano Name": "Baru",
    "Country": "Panama",
    "Region": "Panama",
    "Location": {
      "type": "Point",
      "coordinates": [
        -82.558,
        8.8
      ]
    },
    "Elevation": 3477,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "1f7825f6-b06e-9d0d-42e6-3c0b72241299"
  },
  {
    "Volcano Name": "Barva",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -84.1,
        10.135
      ]
    },
    "Elevation": 2906,
    "Type": "Complex volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "94f3a5c2-644b-65ee-eacb-73eae34c7785"
  },
  {
    "Volcano Name": "Bas Dong Nai",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.2,
        10.8
      ]
    },
    "Elevation": 392,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "72f1ef6b-9f55-dbe9-9d37-97467e1dfba1"
  },
  {
    "Volcano Name": "Batur",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        115.375,
        -8.242
      ]
    },
    "Elevation": 1717,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "84b5c55c-8005-90e1-41fe-2ef275085c54"
  },
  {
    "Volcano Name": "Bayo, Cerro",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.58,
        -25.42
      ]
    },
    "Elevation": 5401,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ae5a1a64-d1f9-fe70-5b87-e19f8a6a729e"
  },
  {
    "Volcano Name": "Bayonnaise Rocks",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.92,
        31.92
      ]
    },
    "Elevation": 10,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a8bfa577-5d16-ad99-e978-702f44d9073c"
  },
  {
    "Volcano Name": "Bayuda Volc Field",
    "Country": "Sudan",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        32.75,
        18.33
      ]
    },
    "Elevation": 0,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "aabacb57-aee1-bcd1-912d-6c61d731bfb4"
  },
  {
    "Volcano Name": "Bazman",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        60,
        28.07
      ]
    },
    "Elevation": 3490,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3e98a252-34a2-e770-d8f4-b1ce847a87d8"
  },
  {
    "Volcano Name": "Behm Canal-Rudyerd Bay",
    "Country": "United States",
    "Region": "Alaska-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -131.05,
        55.32
      ]
    },
    "Elevation": 500,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "3999c8e4-4a6c-1ed1-c766-153e84617134"
  },
  {
    "Volcano Name": "Belenkaya",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.27,
        51.75
      ]
    },
    "Elevation": 892,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7fd7ac46-adca-e165-e8b6-030ee4c6331c"
  },
  {
    "Volcano Name": "Belirang-Beriti",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.18,
        -2.82
      ]
    },
    "Elevation": 1958,
    "Type": "Compound volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "38ae9e9c-2345-eb82-f59f-7cb839c6a780"
  },
  {
    "Volcano Name": "Belknap",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.841,
        44.285
      ]
    },
    "Elevation": 2095,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "2bb9f2d1-c7a1-cedd-c061-8775cf389cc5"
  },
  {
    "Volcano Name": "Bely",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.53,
        57.88
      ]
    },
    "Elevation": 2080,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "69feb266-b3bd-71ba-b224-c3b0a7d31a00"
  },
  {
    "Volcano Name": "Berlin",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -136,
        -76.05
      ]
    },
    "Elevation": 3478,
    "Type": "Shield volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0546b0a5-087c-a5e7-f26d-25b7c14293e4"
  },
  {
    "Volcano Name": "Beru",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.75,
        8.95
      ]
    },
    "Elevation": 1100,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a2f42b40-02da-d8b2-8c97-f83e1e22bf18"
  },
  {
    "Volcano Name": "Berutarube",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.93,
        44.47
      ]
    },
    "Elevation": 1220,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "44835bb5-87ab-6726-1a96-16c778fe1a6d"
  },
  {
    "Volcano Name": "Besar",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.67,
        -4.43
      ]
    },
    "Elevation": 1899,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "97512e77-dd1c-877e-4894-c77702e8626b"
  },
  {
    "Volcano Name": "Bezymianny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.587,
        55.978
      ]
    },
    "Elevation": 2882,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "495a1e1c-89ba-3fcc-56b9-28bdb0c8bf6f"
  },
  {
    "Volcano Name": "Bibinoi",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.72,
        -0.78
      ]
    },
    "Elevation": 900,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "210ea78f-70fd-d317-2ab8-476efbb91554"
  },
  {
    "Volcano Name": "Big Cave",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.365,
        40.955
      ]
    },
    "Elevation": 1259,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "610bd44f-8ab4-6a5c-4e8d-5dd40679974d"
  },
  {
    "Volcano Name": "Bilate River Field",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.1,
        7.07
      ]
    },
    "Elevation": 1700,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ee0a958e-60c0-3b8f-abce-a74de86fdeb0"
  },
  {
    "Volcano Name": "Biliran",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.534,
        11.523
      ]
    },
    "Elevation": 1187,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "7234cecc-fed6-5ddf-6103-23cc212e62b7"
  },
  {
    "Volcano Name": "Billy Mitchell",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.225,
        -6.092
      ]
    },
    "Elevation": 1544,
    "Type": "Pyroclastic shield",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "1d0d9ec9-c054-62ff-5217-54aa1bbe7162"
  },
  {
    "Volcano Name": "Bir Borhut",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        50.63,
        15.55
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "ffe61023-35a7-1163-43be-8c489a1ec20f"
  },
  {
    "Volcano Name": "Birk, Harrat al",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.633,
        18.367
      ]
    },
    "Elevation": 381,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4e73ca77-2c31-de13-7f2b-d5594a332a12"
  },
  {
    "Volcano Name": "Bishoftu Volc Field",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.98,
        8.78
      ]
    },
    "Elevation": 1850,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "77f6d13f-5e27-1cf9-da80-02febfb32084"
  },
  {
    "Volcano Name": "Biu Plateau",
    "Country": "Nigeria",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        12,
        10.75
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "962003b1-c2f1-4720-720e-51ad5ce5fdd8"
  },
  {
    "Volcano Name": "Black Peak",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -158.8,
        56.53
      ]
    },
    "Elevation": 1032,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "e7035762-c0aa-857e-8768-92be010d8c9c"
  },
  {
    "Volcano Name": "Black Rock Desert",
    "Country": "United States",
    "Region": "US-Utah",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.5,
        38.97
      ]
    },
    "Elevation": 1800,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "2127b71b-90cd-7671-30ee-34cd6240fbcb"
  },
  {
    "Volcano Name": "Blanca, Loma",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.009,
        -36.286
      ]
    },
    "Elevation": 2268,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5f170f4d-f12e-39c9-057a-82ed2eb45832"
  },
  {
    "Volcano Name": "Bliznets",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.78,
        56.97
      ]
    },
    "Elevation": 1244,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "950d2c43-30aa-8bdc-a5dd-b546c913464e"
  },
  {
    "Volcano Name": "Bliznetsy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        161.367,
        57.35
      ]
    },
    "Elevation": 265,
    "Type": "Lava cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "61dbb263-b7fc-be46-672c-c21f1e1e76b4"
  },
  {
    "Volcano Name": "Blue Lake Crater",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.77,
        44.42
      ]
    },
    "Elevation": 1230,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "d86b6779-0335-d257-6aff-1990f19ce4fa"
  },
  {
    "Volcano Name": "Blup Blup",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.62,
        -3.508
      ]
    },
    "Elevation": 402,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cb8e5a8f-e135-67e6-f226-c498f630823a"
  },
  {
    "Volcano Name": "Bobrof",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.43,
        51.9
      ]
    },
    "Elevation": 738,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "34892def-2f5f-097e-2a45-b9c7ce8504cd"
  },
  {
    "Volcano Name": "Bogatyr Ridge",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.37,
        44.83
      ]
    },
    "Elevation": 1634,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ea394244-f66a-8d47-9a4c-71e848e9611a"
  },
  {
    "Volcano Name": "Bogoslof",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -168.03,
        53.93
      ]
    },
    "Elevation": 101,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7bf90b76-6d3a-1326-552d-f5b85010d5d6"
  },
  {
    "Volcano Name": "Boisa",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.963,
        -3.994
      ]
    },
    "Elevation": 240,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7c6f20ac-3edd-0de3-810a-8954cc9cdb8d"
  },
  {
    "Volcano Name": "Bola",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.03,
        -5.15
      ]
    },
    "Elevation": 1155,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cfe77918-a077-5ad1-a5a9-68eb2504b8de"
  },
  {
    "Volcano Name": "Bolshe-Bannaya",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.78,
        52.9
      ]
    },
    "Elevation": 1200,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "87aa08ae-d65b-d2ac-70c2-2d7584c31c13"
  },
  {
    "Volcano Name": "Bolshoi Payalpan",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.78,
        55.88
      ]
    },
    "Elevation": 1906,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c1b3fea6-5c3e-7954-3340-bff6fda815ab"
  },
  {
    "Volcano Name": "Bolshoi Semiachik",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.02,
        54.32
      ]
    },
    "Elevation": 1720,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ced85d8f-e2a9-c698-b6c7-3fa53f631839"
  },
  {
    "Volcano Name": "Bolshoi-Kekuknaysky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.8,
        56.47
      ]
    },
    "Elevation": 1401,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "989ca4e4-13e2-ad3b-ee7a-bce203d3afa6"
  },
  {
    "Volcano Name": "Bombalai",
    "Country": "Malaysia",
    "Region": "Borneo",
    "Location": {
      "type": "Point",
      "coordinates": [
        117.88,
        4.4
      ]
    },
    "Elevation": 531,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "7302faad-33e3-9920-e24c-59b99cb084c9"
  },
  {
    "Volcano Name": "Bona-Churchill",
    "Country": "United States",
    "Region": "Alaska-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -141.75,
        61.38
      ]
    },
    "Elevation": 5005,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "0911c097-4184-873c-e80b-7a1cec234a9c"
  },
  {
    "Volcano Name": "Boomerang Seamount",
    "Country": "Antarctica",
    "Region": "Indian O.-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        77.825,
        -37.721
      ]
    },
    "Elevation": -650,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4fd125ee-dd43-f3e9-733d-b1400565b4a5"
  },
  {
    "Volcano Name": "Bora-Bericcio Complex",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.03,
        8.27
      ]
    },
    "Elevation": 2285,
    "Type": "Pumice cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "437aa29f-3fa8-282f-14b0-4261cecf0fda"
  },
  {
    "Volcano Name": "Borale Ale",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.6,
        13.725
      ]
    },
    "Elevation": 668,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "92e5d228-7894-9b01-31fe-7d507c73c515"
  },
  {
    "Volcano Name": "Borawli",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.98,
        13.3
      ]
    },
    "Elevation": 812,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4d754b95-336c-1e0d-02fa-b1e837716ff5"
  },
  {
    "Volcano Name": "Borawli Complex",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.45,
        11.63
      ]
    },
    "Elevation": 875,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4ca14ac2-0887-48d0-cd67-68ff239abf83"
  },
  {
    "Volcano Name": "Boset-Bericha",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.475,
        8.558
      ]
    },
    "Elevation": 2447,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "859314b7-aa49-8dbd-baff-9a635ae0fa7e"
  },
  {
    "Volcano Name": "Bouvet",
    "Country": "Bouvet I.",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        3.35,
        -54.42
      ]
    },
    "Elevation": 780,
    "Type": "Shield volcano",
    "Status": "Magnetism",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fe6b7d6d-40a9-db5d-58d1-35a3009b1ab7"
  },
  {
    "Volcano Name": "Bratan",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        115.13,
        -8.28
      ]
    },
    "Elevation": 2276,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "35e56ed5-faca-21bd-fca1-3b06be63d1f6"
  },
  {
    "Volcano Name": "Brava",
    "Country": "Cape Verde",
    "Region": "Cape Verde Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -24.72,
        14.85
      ]
    },
    "Elevation": 900,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f066e823-4556-1eb2-c960-69eb8230a2b2"
  },
  {
    "Volcano Name": "Bravo, Cerro",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.3,
        5.092
      ]
    },
    "Elevation": 4000,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "39498c3e-869e-5442-58fe-0388a8086d24"
  },
  {
    "Volcano Name": "Bravo, Cerro",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.3,
        5.092
      ]
    },
    "Elevation": 4000,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "3600aa52-f6ed-1cc2-ac82-36b1a749bc0e"
  },
  {
    "Volcano Name": "Brennisteinsfjoll",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.83,
        63.92
      ]
    },
    "Elevation": 626,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "1242222a-0581-0243-0f8d-051e08c06fe7"
  },
  {
    "Volcano Name": "Bridge River Cones",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -123.4,
        50.8
      ]
    },
    "Elevation": 2500,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "757d8809-7db6-03bb-6d9d-18501baacb58"
  },
  {
    "Volcano Name": "Bridgeman Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -56.75,
        -62.05
      ]
    },
    "Elevation": 240,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "4b4bfd5a-11b8-de97-0b08-9fb747521c0d"
  },
  {
    "Volcano Name": "Brimstone Island",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.92,
        -30.23
      ]
    },
    "Elevation": -2000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "3404ee74-26d4-4c7f-2956-36553378232e"
  },
  {
    "Volcano Name": "Bristol Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -26.58,
        -59.03
      ]
    },
    "Elevation": 1100,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ef41d571-a462-259f-21f8-89b9a05f2be7"
  },
  {
    "Volcano Name": "Brothers",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        179.075,
        -34.875
      ]
    },
    "Elevation": -1350,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c4fb1c9a-61bf-5a51-d273-c7ab1db31fe9"
  },
  {
    "Volcano Name": "Brushy Butte",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.443,
        41.178
      ]
    },
    "Elevation": 1174,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "fb7b16a4-c44f-16c3-06d4-306e0364cbca"
  },
  {
    "Volcano Name": "Buckle Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        163.25,
        -66.8
      ]
    },
    "Elevation": 1239,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "55f7e1c8-f960-3474-4524-f5d457c5e50d"
  },
  {
    "Volcano Name": "Bud Dajo",
    "Country": "Philippines",
    "Region": "Sulu Is-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.07,
        5.95
      ]
    },
    "Elevation": 440,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "437f0f44-be12-6f42-142f-787df9e7986b"
  },
  {
    "Volcano Name": "Bufumbira",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.72,
        -1.23
      ]
    },
    "Elevation": 2440,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e763151e-d980-d843-2e93-f4ab923bf902"
  },
  {
    "Volcano Name": "Buldir",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        175.98,
        52.37
      ]
    },
    "Elevation": 656,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8a176a1b-2526-dd69-cf70-ad7ce960f38a"
  },
  {
    "Volcano Name": "Bulusan",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.05,
        12.77
      ]
    },
    "Elevation": 1565,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "50202cec-2c50-cc22-0e9c-367b18b8a654"
  },
  {
    "Volcano Name": "Bunyaruguru Field",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        30.08,
        -0.2
      ]
    },
    "Elevation": 1554,
    "Type": "Explosion crater",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a7fe6b59-1d94-8f7e-1b4b-eb799226fb39"
  },
  {
    "Volcano Name": "Burney, Monte",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.4,
        -52.33
      ]
    },
    "Elevation": 1758,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2f0d020f-b4f4-2484-474c-96977740f666"
  },
  {
    "Volcano Name": "Bus-Obo",
    "Country": "Mongolia",
    "Region": "Mongolia",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.08,
        47.12
      ]
    },
    "Elevation": 1162,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "9ad51d0b-c59a-ec74-22cf-8f25d196e655"
  },
  {
    "Volcano Name": "Butajiri-Silti Field",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.35,
        8.05
      ]
    },
    "Elevation": 2281,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "067a9c9d-7222-f9c4-68df-73ae43b6e41f"
  },
  {
    "Volcano Name": "Buzzard Creek",
    "Country": "United States",
    "Region": "Alaska-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -148.42,
        64.07
      ]
    },
    "Elevation": 830,
    "Type": "Tuff rings",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "72a763b4-f3d6-8aaf-021d-a469bb87849b"
  },
  {
    "Volcano Name": "Cabalian",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.22,
        10.287
      ]
    },
    "Elevation": 945,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8cf68617-4cc6-e615-c532-0ba44f090099"
  },
  {
    "Volcano Name": "Caburga",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.83,
        -39.2
      ]
    },
    "Elevation": 995,
    "Type": "Cinder cone",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "d3a0e04e-4450-3051-5948-7cc88ff62d61"
  },
  {
    "Volcano Name": "Cagua",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.123,
        18.222
      ]
    },
    "Elevation": 1133,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4cf90b47-3061-b8ec-4d4e-111caadc93c5"
  },
  {
    "Volcano Name": "Calabozos",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.496,
        -35.558
      ]
    },
    "Elevation": 3508,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1b2b8c2d-bf65-1d44-5674-6ba2932b8844"
  },
  {
    "Volcano Name": "Calatrava Volc Field",
    "Country": "Spain",
    "Region": "Spain",
    "Location": {
      "type": "Point",
      "coordinates": [
        -4.017,
        38.867
      ]
    },
    "Elevation": 1117,
    "Type": "Pyroclastic cones",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "894a24f8-426f-af90-303f-0f1176139e50"
  },
  {
    "Volcano Name": "Calayo",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.068,
        7.877
      ]
    },
    "Elevation": 646,
    "Type": "Tuff cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "67e8b444-dc40-8ce2-4d8e-f65a10c7b85e"
  },
  {
    "Volcano Name": "Calbuco",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.614,
        -41.326
      ]
    },
    "Elevation": 2003,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "342cc74f-38fa-8b04-d3ca-7a0c3cfba215"
  },
  {
    "Volcano Name": "Callaqui",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.45,
        -37.92
      ]
    },
    "Elevation": 3164,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4a34dbb8-aa2e-4fac-43af-426f53c88b5c"
  },
  {
    "Volcano Name": "Cameroon, Mt.",
    "Country": "Cameroon",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        9.17,
        4.203
      ]
    },
    "Elevation": 4095,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6690f4c2-cd53-1a7a-db57-1a4f848395fb"
  },
  {
    "Volcano Name": "Camiguin de Babuyanes",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.86,
        18.83
      ]
    },
    "Elevation": 712,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "402ef35c-4565-acee-abe6-a390ca0d6ad4"
  },
  {
    "Volcano Name": "Campi Flegrei",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.139,
        40.827
      ]
    },
    "Elevation": 458,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "ed3dd87e-bdd9-b9ae-5a23-31e50a3e0f7d"
  },
  {
    "Volcano Name": "Campi Flegrei Mar Sicilia",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        12.7,
        37.1
      ]
    },
    "Elevation": -8,
    "Type": "Submarine volcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "7caf5345-9f8b-e017-e74b-773674efd67e"
  },
  {
    "Volcano Name": "Candlemas Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -26.67,
        -57.08
      ]
    },
    "Elevation": 550,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "10b6045c-714c-2f39-226c-6f1fa69d1683"
  },
  {
    "Volcano Name": "Canlaon",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.132,
        10.412
      ]
    },
    "Elevation": 2435,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "569e3fce-c47f-eb40-66c7-a915abf826b7"
  },
  {
    "Volcano Name": "Carlisle",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -170.05,
        52.9
      ]
    },
    "Elevation": 1620,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "67c17209-73d6-0fa8-ce81-1c267023efb8"
  },
  {
    "Volcano Name": "Carran-Los Venados",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.07,
        -40.35
      ]
    },
    "Elevation": 1114,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "481968e5-2fc8-cecf-92ac-7ec7aac8289c"
  },
  {
    "Volcano Name": "Carrizozo",
    "Country": "United States",
    "Region": "US-New Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -105.93,
        33.78
      ]
    },
    "Elevation": 1731,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2f7e115d-3882-c673-e520-6ec16a9e9aaf"
  },
  {
    "Volcano Name": "Casiri, Nevados",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.82,
        -17.47
      ]
    },
    "Elevation": 5650,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "50abfe09-b9c5-bb95-3e8b-f1d1763bcc8c"
  },
  {
    "Volcano Name": "Cayambe",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.986,
        0.029
      ]
    },
    "Elevation": 5790,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "173b5441-a0e1-f584-f864-b900d30d93c8"
  },
  {
    "Volcano Name": "Cayambe",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.986,
        0.029
      ]
    },
    "Elevation": 5790,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "dc12ea68-b2ab-719c-9576-e5da3c83b661"
  },
  {
    "Volcano Name": "Cayute-La Vigueria",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.27,
        -41.25
      ]
    },
    "Elevation": 506,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eaffed24-e6b1-26fe-2ce4-31df22e381dd"
  },
  {
    "Volcano Name": "Cayutu?-La Viguer?a",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.267,
        -41.25
      ]
    },
    "Elevation": 506,
    "Type": "Pyroclastic cones",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f714d051-9473-2e82-5b6e-2191059829f5"
  },
  {
    "Volcano Name": "Ceboruco",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.508,
        21.125
      ]
    },
    "Elevation": 2280,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "557b64e4-fab4-9671-8f32-a6cd9572c1d0"
  },
  {
    "Volcano Name": "Cendres, Ile des",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.014,
        10.158
      ]
    },
    "Elevation": -20,
    "Type": "Submarine volcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "bdf9891e-51cb-9c50-25f4-a1539e0749dc"
  },
  {
    "Volcano Name": "Central Island",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.042,
        3.5
      ]
    },
    "Elevation": 550,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4f46fb01-4707-9663-43c7-2263c1d01837"
  },
  {
    "Volcano Name": "Cereme",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        108.4,
        -6.892
      ]
    },
    "Elevation": 3078,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "f32ea845-504d-8d2f-6fd1-a2c2eb4226b4"
  },
  {
    "Volcano Name": "Ch'uga-ryong",
    "Country": "North Korea",
    "Region": "Korea",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.33,
        38.33
      ]
    },
    "Elevation": 452,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "dd0fe6e8-8905-9805-daf8-787b51ed5f34"
  },
  {
    "Volcano Name": "Chacana",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.25,
        -0.375
      ]
    },
    "Elevation": 4643,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "da0e52c6-dc4a-9ead-b6de-b16cb3061f64"
  },
  {
    "Volcano Name": "Chachani, Nevado",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.53,
        -16.191
      ]
    },
    "Elevation": 6057,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "db9571e6-21f3-8eba-ebaa-25e4c0d923a3"
  },
  {
    "Volcano Name": "Chachani, Nevado",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.53,
        -16.191
      ]
    },
    "Elevation": 6057,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "dc58113b-e5ab-92f8-b817-6b827fe9ab39"
  },
  {
    "Volcano Name": "Chagulak",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -171.13,
        52.57
      ]
    },
    "Elevation": 1142,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "70206682-0382-77bd-0bca-a81dcd2c848c"
  },
  {
    "Volcano Name": "Chaine des Puys",
    "Country": "France",
    "Region": "France",
    "Location": {
      "type": "Point",
      "coordinates": [
        2.967,
        45.775
      ]
    },
    "Elevation": 1464,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "a95511d1-d93c-b20f-b136-7a5ccb08fdf5"
  },
  {
    "Volcano Name": "Chaiten",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.646,
        -42.833
      ]
    },
    "Elevation": 962,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "e763a69d-3288-81da-1e18-8a16378d70c0"
  },
  {
    "Volcano Name": "Changbaishan",
    "Country": "North Korea",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        128.08,
        41.98
      ]
    },
    "Elevation": 2744,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "08bc1dbf-b7bd-7bae-2b86-146211219e12"
  },
  {
    "Volcano Name": "Chapulul, Cerro",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.08,
        -38.37
      ]
    },
    "Elevation": 2143,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "12a6eb40-91b5-2441-24f5-a39dab94b8a4"
  },
  {
    "Volcano Name": "Cherny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.67,
        56.82
      ]
    },
    "Elevation": 1778,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f5ffff0c-7fff-3cc0-8dfb-4a82787c2df5"
  },
  {
    "Volcano Name": "Cherpuk Group",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.467,
        55.55
      ]
    },
    "Elevation": 1868,
    "Type": "Pyroclastic cones",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "562a648f-511a-34dd-1980-1f3ce46667c7"
  },
  {
    "Volcano Name": "Chichinautzin",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -99.13,
        19.08
      ]
    },
    "Elevation": 3930,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "e07b8125-3937-b868-f788-47e5f62cffeb"
  },
  {
    "Volcano Name": "Chichon, El",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -93.228,
        17.36
      ]
    },
    "Elevation": 1150,
    "Type": "Tuff cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "faf487ff-cd10-5c3f-f32a-cab9e73aa688"
  },
  {
    "Volcano Name": "Chiginagak",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -157,
        57.13
      ]
    },
    "Elevation": 2075,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "06b2f7d5-dd13-1e06-b536-ca779a6b8d4c"
  },
  {
    "Volcano Name": "Chikurachki",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.458,
        50.325
      ]
    },
    "Elevation": 1816,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d76d892c-1e9e-9dee-401e-1ef6162c48cd"
  },
  {
    "Volcano Name": "Chiliques",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.7,
        -23.583
      ]
    },
    "Elevation": 5778,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "da4f8d3a-19c4-3ea0-1859-0cfdfc7d9ff1"
  },
  {
    "Volcano Name": "Chiliques",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.7,
        -23.58
      ]
    },
    "Elevation": 5778,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "541df5a3-c230-4fa0-0624-623778995979"
  },
  {
    "Volcano Name": "Chillan, Nevados de",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.377,
        -36.863
      ]
    },
    "Elevation": 3212,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "58516d2b-b633-73e4-d799-e8fcaed7513d"
  },
  {
    "Volcano Name": "Chimborazo",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.815,
        -1.464
      ]
    },
    "Elevation": 6310,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "a7f1a365-514d-8d45-bd8c-2b4fcf62fd72"
  },
  {
    "Volcano Name": "Chinameca",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.32,
        13.475
      ]
    },
    "Elevation": 1228,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a700fed8-b67e-c877-a04d-b47ceb9b2b2f"
  },
  {
    "Volcano Name": "Chingo Volc Field",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.73,
        14.12
      ]
    },
    "Elevation": 1775,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "091ca4cb-c800-fd40-7f16-91e1e08103e9"
  },
  {
    "Volcano Name": "Chiquimula Volc Field",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.55,
        14.83
      ]
    },
    "Elevation": 1192,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0e5e0b51-1fbb-2c34-0237-5d8afcc0d304"
  },
  {
    "Volcano Name": "Chiracha",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.12,
        6.65
      ]
    },
    "Elevation": 1650,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "8eed8fae-72c9-b3ba-02b0-5b55bcc53014"
  },
  {
    "Volcano Name": "Chirinkotan",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.48,
        48.98
      ]
    },
    "Elevation": 724,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ceb72742-a919-bd8a-cc94-fe0123e266c4"
  },
  {
    "Volcano Name": "Chirip",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.92,
        45.38
      ]
    },
    "Elevation": 1589,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "f800dc7a-9cb7-fa79-888b-07e0292aaa28"
  },
  {
    "Volcano Name": "Chirpoi",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.875,
        46.525
      ]
    },
    "Elevation": 742,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "232d8459-a5be-84c3-b482-8d71dcf7b3fa"
  },
  {
    "Volcano Name": "Chokai",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.03,
        39.08
      ]
    },
    "Elevation": 2237,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2c41334c-8731-45d5-e464-28e7bc61b55d"
  },
  {
    "Volcano Name": "Chyulu Hills",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.88,
        -2.68
      ]
    },
    "Elevation": 2188,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "4bea921a-04b0-19f9-afd3-303d3fbb31f8"
  },
  {
    "Volcano Name": "Ciguatepe, Cerro El",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.142,
        12.53
      ]
    },
    "Elevation": 603,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "626c26a2-498b-df78-cb17-4df369de8611"
  },
  {
    "Volcano Name": "Cinnamon Butte",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.108,
        43.241
      ]
    },
    "Elevation": 1956,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "7c819d46-26cd-ede4-f3fe-343b59a3302b"
  },
  {
    "Volcano Name": "Cinotepec, Cerro",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.25,
        14.02
      ]
    },
    "Elevation": 665,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d27d9f7d-cd57-dbeb-18ca-5aaee2950b5d"
  },
  {
    "Volcano Name": "Clark",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        177.839,
        -36.446
      ]
    },
    "Elevation": -860,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "704b8334-ad36-5fce-b1df-39e2b7be1986"
  },
  {
    "Volcano Name": "Clear Lake",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.77,
        38.97
      ]
    },
    "Elevation": 1439,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b822b690-d1f3-73a9-f6e1-fd55abfac4cf"
  },
  {
    "Volcano Name": "Cleft Segment",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.3,
        44.83
      ]
    },
    "Elevation": -2140,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c95ebcae-2da2-26d3-2248-c5cc3f34c377"
  },
  {
    "Volcano Name": "Cleveland",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.95,
        52.82
      ]
    },
    "Elevation": 1730,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7edb0cf1-eadb-b99d-69d1-aed972759b14"
  },
  {
    "Volcano Name": "Coatepeque Caldera",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.55,
        13.87
      ]
    },
    "Elevation": 746,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dc28f7ae-08df-fb8e-a651-7249d448db54"
  },
  {
    "Volcano Name": "Cochons, Ile Aux",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        50.23,
        -46.1
      ]
    },
    "Elevation": 775,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6749a10c-f15a-3c25-25f9-1f8daa7f604f"
  },
  {
    "Volcano Name": "Cofre de Perote",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.15,
        19.492
      ]
    },
    "Elevation": 4282,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "69b2bc37-3731-f0ee-6877-259b8206fdbe"
  },
  {
    "Volcano Name": "Colachi",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.65,
        -23.23
      ]
    },
    "Elevation": 5631,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8252ac9f-f999-9dd0-6a1a-7b9c34361aed"
  },
  {
    "Volcano Name": "Coleman Seamount",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.17,
        -8.83
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b5c2a85d-1002-f685-4df0-548aa617ffb2"
  },
  {
    "Volcano Name": "Colima",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -103.62,
        19.514
      ]
    },
    "Elevation": 3850,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4bdc26cd-2851-fe01-c693-ab733429099a"
  },
  {
    "Volcano Name": "Colluma, Cerro",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.07,
        -18.5
      ]
    },
    "Elevation": 3876,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "0f25b67b-f3f4-ec66-7209-7101586e5417"
  },
  {
    "Volcano Name": "Colo [Una Una]",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.608,
        -0.17
      ]
    },
    "Elevation": 507,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5844e3bf-3f89-08ff-b6d2-a0042d4d65c8"
  },
  {
    "Volcano Name": "Comondu-La Purisima",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -111.92,
        26
      ]
    },
    "Elevation": 780,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "9b08586f-6dc3-a38a-7137-2133bf3d4aef"
  },
  {
    "Volcano Name": "Concepcion",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.622,
        11.538
      ]
    },
    "Elevation": 1700,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5dc96725-c25b-5f32-a527-d9aa16692ebb"
  },
  {
    "Volcano Name": "Conchagua",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.853,
        13.277
      ]
    },
    "Elevation": 1250,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "fd33fa4f-3551-34dc-9ec9-76f81e9f511a"
  },
  {
    "Volcano Name": "Conchaguita",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.765,
        13.22
      ]
    },
    "Elevation": 550,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ee900092-64b1-b4b9-afc3-0203e85aa10e"
  },
  {
    "Volcano Name": "Cook, Isla",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.27,
        -54.95
      ]
    },
    "Elevation": 150,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "aa9b43f8-113b-bf0d-3995-28ee5a6e30b3"
  },
  {
    "Volcano Name": "Copahue",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.17,
        -37.85
      ]
    },
    "Elevation": 2965,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f62290a2-1a45-7368-1e46-03d7df65c1bc"
  },
  {
    "Volcano Name": "Copiapo",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.13,
        -27.3
      ]
    },
    "Elevation": 6052,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "f8f2463d-b186-2003-ddfb-29fd44d36353"
  },
  {
    "Volcano Name": "Corbetti Caldera",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.43,
        7.18
      ]
    },
    "Elevation": 2320,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "16a08449-6923-9315-8e86-d09189403e9c"
  },
  {
    "Volcano Name": "Corcovado",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.8,
        -43.18
      ]
    },
    "Elevation": 2300,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "84eab039-e7d2-3518-a667-8461a1a270e0"
  },
  {
    "Volcano Name": "Cordon Chalviri",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.62,
        -23.85
      ]
    },
    "Elevation": 5623,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a90e1e20-faab-2b41-9d86-d05a3bfcc276"
  },
  {
    "Volcano Name": "Cordon de Puntas Negras",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.53,
        -23.75
      ]
    },
    "Elevation": 5852,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "59806d6d-30a9-dd8f-b74d-0051a05100a1"
  },
  {
    "Volcano Name": "Cordon del Azufre",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.52,
        -25.33
      ]
    },
    "Elevation": 5463,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5d7999bc-861f-8dae-7621-129e5c2ea4bb"
  },
  {
    "Volcano Name": "Coronado",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.513,
        29.08
      ]
    },
    "Elevation": 440,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "68348146-0881-695e-16ce-4cba497357c8"
  },
  {
    "Volcano Name": "Coropuna",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.65,
        -15.52
      ]
    },
    "Elevation": 6377,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3a750ac6-5b83-bb49-1be3-c9d49c94b511"
  },
  {
    "Volcano Name": "Corvo",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -31.07,
        39.67
      ]
    },
    "Elevation": 715,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c572a259-cb9a-1c35-6595-06f53c2c7048"
  },
  {
    "Volcano Name": "Cosiguina",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.57,
        12.98
      ]
    },
    "Elevation": 872,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "1f6efe76-57c1-8204-7e8b-55ad4944c559"
  },
  {
    "Volcano Name": "Coso Volc Field",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -117.82,
        36.03
      ]
    },
    "Elevation": 2400,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "dc897081-5572-0b2b-f4dc-64f394ef9183"
  },
  {
    "Volcano Name": "Cotopaxi",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.436,
        -0.677
      ]
    },
    "Elevation": 5911,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b887ece8-a2a5-02ba-f63d-d854809aed66"
  },
  {
    "Volcano Name": "Crater Basalt Volc Field",
    "Country": "Argentina",
    "Region": "Chile-S/Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.183,
        -42.017
      ]
    },
    "Elevation": 1359,
    "Type": "Cinder cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "99768c38-40f1-bf2d-e376-4c8ef660347d"
  },
  {
    "Volcano Name": "Crater Lake",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.12,
        42.93
      ]
    },
    "Elevation": 2487,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "338db807-b171-b77d-89f4-029cbd0f9aba"
  },
  {
    "Volcano Name": "Crater Mountain",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.08,
        -6.58
      ]
    },
    "Elevation": 3233,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "12990220-2a1f-2644-b496-ac7eb14583e9"
  },
  {
    "Volcano Name": "Craters of the Moon",
    "Country": "United States",
    "Region": "US-Idaho",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.5,
        43.42
      ]
    },
    "Elevation": 2005,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "6065fc46-753e-b1d9-069d-e67578e1416b"
  },
  {
    "Volcano Name": "Crow Lagoon",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.23,
        54.7
      ]
    },
    "Elevation": 335,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8f81b3d6-cf7f-4e5f-74ab-a53441148ecb"
  },
  {
    "Volcano Name": "Cu-Lao Re Group",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.12,
        15.38
      ]
    },
    "Elevation": 181,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b34fc294-efb7-4c73-6e8d-f70b4e2b49bd"
  },
  {
    "Volcano Name": "Cuicocha",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.364,
        0.308
      ]
    },
    "Elevation": 3246,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "a3e28893-7f55-1068-06d3-f0bc6c98b499"
  },
  {
    "Volcano Name": "Cuilapa-Barbarena",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.4,
        14.33
      ]
    },
    "Elevation": 1454,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "df259dfc-0a11-e8bb-376f-3d15a6228608"
  },
  {
    "Volcano Name": "Cuilapa-Barbarena",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.4,
        14.333
      ]
    },
    "Elevation": 1454,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "98395c77-d55f-71d0-c785-e25f27f9f082"
  },
  {
    "Volcano Name": "Cumbal",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.88,
        0.98
      ]
    },
    "Elevation": 4764,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "25c61320-fea5-5cce-ddcf-a3194745c502"
  },
  {
    "Volcano Name": "Cumbres, Las",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.27,
        19.15
      ]
    },
    "Elevation": 3940,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ade46b84-5df1-83c5-5b71-340d99d1ae0b"
  },
  {
    "Volcano Name": "Curacoa",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -173.67,
        -15.62
      ]
    },
    "Elevation": -33,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5f7c20d3-1a24-a2ce-84e9-58a31e7251da"
  },
  {
    "Volcano Name": "Curacoa",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -173.667,
        -15.617
      ]
    },
    "Elevation": -33,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "cf0c9679-5c02-ece5-fcd3-2266446e1312"
  },
  {
    "Volcano Name": "Curtis Island",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.561,
        -30.542
      ]
    },
    "Elevation": 137,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "345441b4-1d45-75b8-643c-dd73960d3526"
  },
  {
    "Volcano Name": "C?ndor, Cerro el",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.35,
        -26.617
      ]
    },
    "Elevation": 6532,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4fa377ed-0168-6773-34ad-1032b011edad"
  },
  {
    "Volcano Name": "Dabbahu",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.48,
        12.6
      ]
    },
    "Elevation": 1442,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9d4c2dc5-97ec-bbf1-d087-3e3571bc9bfd"
  },
  {
    "Volcano Name": "Dabbayra",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.07,
        12.38
      ]
    },
    "Elevation": 1302,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "190bd2c6-53d7-8f26-3b91-c26cba804aa0"
  },
  {
    "Volcano Name": "Dacht-I-Navar Group",
    "Country": "Afghanistan",
    "Region": "Afghanistan",
    "Location": {
      "type": "Point",
      "coordinates": [
        67.92,
        33.95
      ]
    },
    "Elevation": 3800,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e512e438-fb7c-da86-10f3-75abdad21d34"
  },
  {
    "Volcano Name": "Daikoku",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.194,
        21.324
      ]
    },
    "Elevation": -323,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c283011c-b428-0e33-f477-4825239d8825"
  },
  {
    "Volcano Name": "Daisetsu",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.88,
        43.68
      ]
    },
    "Elevation": 2290,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "d648d976-abe3-6b03-c14d-fbde9cad442a"
  },
  {
    "Volcano Name": "Dakataua",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.108,
        -5.056
      ]
    },
    "Elevation": 400,
    "Type": "Caldera",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b7041333-6c52-cfc0-5045-511abf9c0392"
  },
  {
    "Volcano Name": "Dalaffilla",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.55,
        13.792
      ]
    },
    "Elevation": 613,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fa38e679-4afc-d109-a0e9-62794875a87d"
  },
  {
    "Volcano Name": "Dallol",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.3,
        14.242
      ]
    },
    "Elevation": -48,
    "Type": "Explosion crater",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "68e2ea13-88ad-48ec-a672-655c65a8a4ef"
  },
  {
    "Volcano Name": "Dama Ali",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.63,
        11.28
      ]
    },
    "Elevation": 1068,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "668076ce-754d-3e17-3d69-28cd8cb70b31"
  },
  {
    "Volcano Name": "Damavand",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        52.109,
        35.951
      ]
    },
    "Elevation": 5670,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6d59497b-abeb-3c84-48c2-b5d7486d908d"
  },
  {
    "Volcano Name": "Dana",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -161.22,
        55.62
      ]
    },
    "Elevation": 1354,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ad35f3a0-a590-6cdc-5456-32e8f7392965"
  },
  {
    "Volcano Name": "Danau Complex",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        105.97,
        -6.2
      ]
    },
    "Elevation": 1778,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9f3c23d7-286e-782c-3dc7-db65ecabd4ea"
  },
  {
    "Volcano Name": "Dar-Alages",
    "Country": "Armenia",
    "Region": "Armenia",
    "Location": {
      "type": "Point",
      "coordinates": [
        45.542,
        39.7
      ]
    },
    "Elevation": 3329,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fbaa9a58-adbd-6251-e363-7dde6463aaba"
  },
  {
    "Volcano Name": "Dariganga Volc Field",
    "Country": "Mongolia",
    "Region": "Mongolia",
    "Location": {
      "type": "Point",
      "coordinates": [
        114,
        45.33
      ]
    },
    "Elevation": 1778,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "57b4edab-357e-5063-f0cd-27ea822dd61e"
  },
  {
    "Volcano Name": "Darwin, Volcan",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.28,
        -0.18
      ]
    },
    "Elevation": 1330,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "34a6d33c-1a19-7b1d-a1d8-ddd87e338afd"
  },
  {
    "Volcano Name": "Daun, Bukit",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.37,
        -3.38
      ]
    },
    "Elevation": 2467,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b0ed04a6-6840-71ac-011b-bf5315e27299"
  },
  {
    "Volcano Name": "Davidof",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.33,
        51.97
      ]
    },
    "Elevation": 328,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "632c46df-6d75-0b30-9052-8756c3869e17"
  },
  {
    "Volcano Name": "Davis Lake",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.82,
        43.57
      ]
    },
    "Elevation": 2163,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "6465f5ca-141b-8be3-923c-f9275a523be1"
  },
  {
    "Volcano Name": "Dawson Strait Group",
    "Country": "Papua New Guinea",
    "Region": "D'Entrecasteaux Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.88,
        -9.62
      ]
    },
    "Elevation": 500,
    "Type": "Volcanic field",
    "Status": "Hydration Rind",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "7ba0112a-5e6f-39a3-bd61-e53a331a9b69"
  },
  {
    "Volcano Name": "Deception Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -60.65,
        -62.97
      ]
    },
    "Elevation": 576,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "263dc68e-7962-9601-44c8-21524f723504"
  },
  {
    "Volcano Name": "Demon",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.85,
        45.5
      ]
    },
    "Elevation": 1205,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d0f0e1a0-aa39-b253-ba48-02f73dfb8fa0"
  },
  {
    "Volcano Name": "Dempo",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.13,
        -4.03
      ]
    },
    "Elevation": 3173,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b913750f-e245-38f9-8815-1064d7f0a99e"
  },
  {
    "Volcano Name": "Denison",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.45,
        58.42
      ]
    },
    "Elevation": 2318,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "2a0646aa-8f53-3037-c348-87e36e45343c"
  },
  {
    "Volcano Name": "Descabezado Grande",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.75,
        -35.58
      ]
    },
    "Elevation": 3953,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4684f729-e53e-bcd7-adc3-d0d3bd63bee5"
  },
  {
    "Volcano Name": "Devils Garden",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -120.861,
        43.512
      ]
    },
    "Elevation": 1698,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "d2226a28-ab3b-0376-d5f4-dd2e70cb5657"
  },
  {
    "Volcano Name": "Dgida Basin",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.25,
        50.52
      ]
    },
    "Elevation": 1500,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "21fac9c9-5d82-eafe-3d6f-68f3ce68ed18"
  },
  {
    "Volcano Name": "Dhamar, Harras of",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.67,
        14.57
      ]
    },
    "Elevation": 3500,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b151cb59-9ecc-16ed-a329-49eaa8bed139"
  },
  {
    "Volcano Name": "Diable, Morne Au",
    "Country": "Dominica",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.45,
        15.62
      ]
    },
    "Elevation": 861,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "731959cc-805a-f977-9d7a-5b8fb9b3eb6c"
  },
  {
    "Volcano Name": "Diablotins, Morne",
    "Country": "Dominica",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.42,
        15.5
      ]
    },
    "Elevation": 1430,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ed01fa85-835d-11a5-3e22-aa9192358c13"
  },
  {
    "Volcano Name": "Diamond Craters",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -118.75,
        43.1
      ]
    },
    "Elevation": 1435,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "eba01b6b-b5fc-edcd-2a71-54291f7a283e"
  },
  {
    "Volcano Name": "Didicas",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.202,
        19.077
      ]
    },
    "Elevation": 244,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5ad80a5d-c3bb-bccb-7cf3-8a0aa721e982"
  },
  {
    "Volcano Name": "Dieng Volc Complex",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.92,
        -7.2
      ]
    },
    "Elevation": 2565,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3d11b24d-99da-3351-2912-0a568f47903d"
  },
  {
    "Volcano Name": "Diky Greben",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157,
        51.43
      ]
    },
    "Elevation": 1070,
    "Type": "Lava dome",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "262e5fd8-a3cf-d407-0f5a-5b93d194c587"
  },
  {
    "Volcano Name": "Dofen",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.13,
        9.35
      ]
    },
    "Elevation": 1151,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ee4d319e-4ab1-a593-1763-8ad576534acd"
  },
  {
    "Volcano Name": "Doma Peaks",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.15,
        -5.9
      ]
    },
    "Elevation": 3568,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e89d3be4-2f16-5260-c5fb-d43264c9c69b"
  },
  {
    "Volcano Name": "Domuyo",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.42,
        -36.63
      ]
    },
    "Elevation": 4709,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e3e12b2b-98ea-b1d2-e129-28712fb8e2a0"
  },
  {
    "Volcano Name": "Don Joao de Castro Bank",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -26.63,
        38.23
      ]
    },
    "Elevation": -14,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "c12d9001-5c74-e6f4-7ad1-c4e2946a282a"
  },
  {
    "Volcano Name": "Dona Juana",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.92,
        1.47
      ]
    },
    "Elevation": 4150,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "5363bde0-0262-700d-dbad-f9baee8daacb"
  },
  {
    "Volcano Name": "Dotsero",
    "Country": "United States",
    "Region": "US-Colorado",
    "Location": {
      "type": "Point",
      "coordinates": [
        -107.03,
        39.65
      ]
    },
    "Elevation": 2230,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "474998d5-16ed-552f-75f9-9b63f9a58b37"
  },
  {
    "Volcano Name": "Douglas",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -153.55,
        58.87
      ]
    },
    "Elevation": 2140,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0089e296-5303-2b40-2915-895cd0d6ebaf"
  },
  {
    "Volcano Name": "Doyo Seamount",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.8,
        27.683
      ]
    },
    "Elevation": -860,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "26144be2-1dcb-a9f0-72e8-fcd679653730"
  },
  {
    "Volcano Name": "Dubbi",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.808,
        13.58
      ]
    },
    "Elevation": 1625,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "2a0aac33-e395-b44e-8fa8-b436bb1f803d"
  },
  {
    "Volcano Name": "Dukono",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.88,
        1.68
      ]
    },
    "Elevation": 1185,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "dd524940-f424-4e02-e7cc-9c70cd563279"
  },
  {
    "Volcano Name": "Duncan Canal",
    "Country": "United States",
    "Region": "Alaska-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -133.1,
        56.5
      ]
    },
    "Elevation": 15,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ec80601c-5eba-d848-05af-93192c20d049"
  },
  {
    "Volcano Name": "Durango Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.45,
        24.15
      ]
    },
    "Elevation": 2075,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "95a5fd27-2751-48f0-ed22-34adc7b38d96"
  },
  {
    "Volcano Name": "Dutton",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -162.27,
        55.18
      ]
    },
    "Elevation": 1506,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "41c40e37-f85e-6e34-0d2b-69cfe44e78a0"
  },
  {
    "Volcano Name": "Dzenzursky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.922,
        53.637
      ]
    },
    "Elevation": 2155,
    "Type": "Compound volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "570a6486-56c0-e5c9-525a-80236cd2fc91"
  },
  {
    "Volcano Name": "E-san",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.17,
        41.8
      ]
    },
    "Elevation": 618,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "8f0a02ba-ae38-6d44-0734-42212f1d582e"
  },
  {
    "Volcano Name": "Eagle Lake Field",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -120.83,
        40.63
      ]
    },
    "Elevation": 1652,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "44cd5f5d-4933-a728-ae77-6de7e0c0c693"
  },
  {
    "Volcano Name": "Easter Island",
    "Country": "Chile",
    "Region": "Chile-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -109.383,
        -27.15
      ]
    },
    "Elevation": 511,
    "Type": "Shield volcanoes",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7a62e340-734d-623c-cd3c-ce8590f1e2a1"
  },
  {
    "Volcano Name": "Eastern Gemini Seamount",
    "Country": "Pacific Ocean",
    "Region": "SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        170.28,
        -20.98
      ]
    },
    "Elevation": -80,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a364deeb-a977-7331-615a-a1c6d61f58f1"
  },
  {
    "Volcano Name": "Ebeko",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.02,
        50.68
      ]
    },
    "Elevation": 1156,
    "Type": "Somma volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1970b607-193b-e6f4-f147-66689a2a5578"
  },
  {
    "Volcano Name": "Ebulobo",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.18,
        -8.808
      ]
    },
    "Elevation": 2124,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "64b5e159-b88a-6e73-8890-0ae905bd4c86"
  },
  {
    "Volcano Name": "Eburru, Ol Doinyo",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.23,
        -0.63
      ]
    },
    "Elevation": 2856,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "65111700-37e2-4042-0842-91e90fa0cfac"
  },
  {
    "Volcano Name": "Ecuador, Volcan",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.546,
        -0.02
      ]
    },
    "Elevation": 790,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e45cbde8-8469-3a40-21fc-47613d5b9e2a"
  },
  {
    "Volcano Name": "Edgecumbe",
    "Country": "United States",
    "Region": "Alaska-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -135.75,
        57.05
      ]
    },
    "Elevation": 976,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "5a9af8e4-19db-b9c7-13ce-d507332d028a"
  },
  {
    "Volcano Name": "Edziza",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.63,
        57.72
      ]
    },
    "Elevation": 2786,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "ff2d1076-ec5a-7f5a-9cd5-e6620adb8a22"
  },
  {
    "Volcano Name": "Eggella",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.52,
        56.57
      ]
    },
    "Elevation": 1046,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "de40d201-b057-2bb8-cfbe-6835ddf9c0cd"
  },
  {
    "Volcano Name": "Egmont",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        174.07,
        -39.3
      ]
    },
    "Elevation": 2518,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "afb5d3eb-2c57-660b-3e8e-58dbd5777573"
  },
  {
    "Volcano Name": "Egon",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.45,
        -8.67
      ]
    },
    "Elevation": 1703,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "60d30dd6-5c56-0b29-57bc-28e9acb2f55a"
  },
  {
    "Volcano Name": "Ekarma",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.93,
        48.958
      ]
    },
    "Elevation": 1170,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "649d6c97-f666-b2b2-46c0-8594b790ba54"
  },
  {
    "Volcano Name": "Elbrus",
    "Country": "Russia",
    "Region": "Russia-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.45,
        43.33
      ]
    },
    "Elevation": 5633,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "7b899e8a-1674-6187-7888-a687dac6556a"
  },
  {
    "Volcano Name": "Elmenteita Badlands",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.27,
        -0.52
      ]
    },
    "Elevation": 2126,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a3cfbea4-9226-9d5d-0b29-2ac66dd59d92"
  },
  {
    "Volcano Name": "Elovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.53,
        57.53
      ]
    },
    "Elevation": 1381,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dae80a69-d873-83e2-f250-0e396a83cee3"
  },
  {
    "Volcano Name": "Emmons Lake",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -162.07,
        55.33
      ]
    },
    "Elevation": 1465,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f834517a-61fa-a0d3-426f-32a7179c4101"
  },
  {
    "Volcano Name": "Emperor of China",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.22,
        -6.62
      ]
    },
    "Elevation": -2850,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "f81258f0-4dd4-4479-2c4d-d91a6faab037"
  },
  {
    "Volcano Name": "Emuruangogolak",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.33,
        1.5
      ]
    },
    "Elevation": 1328,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "cb0d3ff3-0eb1-c24c-b0f1-34735061a8d0"
  },
  {
    "Volcano Name": "Endeavour Ridge",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -129.1,
        47.95
      ]
    },
    "Elevation": -2400,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "972cd9c1-a614-97e3-24da-dbe2cdf067f8"
  },
  {
    "Volcano Name": "Epi",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.37,
        -16.68
      ]
    },
    "Elevation": 833,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "720b9c39-22d3-e970-7b1e-d54629804519"
  },
  {
    "Volcano Name": "Erciyes Dagi",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        35.48,
        38.52
      ]
    },
    "Elevation": 3916,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dc8d2277-d902-72eb-5520-3d0f2c64b637"
  },
  {
    "Volcano Name": "Erebus",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.17,
        -77.53
      ]
    },
    "Elevation": 3794,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "011fe1ad-75f8-3381-4ea3-1d1ca40c4017"
  },
  {
    "Volcano Name": "Erta Ale",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.67,
        13.6
      ]
    },
    "Elevation": 613,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "07551466-1770-0e59-f684-392fe98b041e"
  },
  {
    "Volcano Name": "Es Safa",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.15,
        33.08
      ]
    },
    "Elevation": 979,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "4f90c7b5-9548-c0d0-295f-fa52106a38b3"
  },
  {
    "Volcano Name": "Escanaba Segment",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -127.5,
        40.983
      ]
    },
    "Elevation": -1700,
    "Type": "Submarine volcano",
    "Status": "Uranium-series",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "84088d00-1b39-6eee-2feb-2dadea3d2fc8"
  },
  {
    "Volcano Name": "Escorial, Cerro",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.37,
        -25.08
      ]
    },
    "Elevation": 5447,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "98d7d3d6-f0df-abff-7914-9f59ceff15d0"
  },
  {
    "Volcano Name": "Esjufjoll",
    "Country": "Iceland",
    "Region": "Iceland-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.65,
        64.27
      ]
    },
    "Elevation": 1760,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "dcab4c27-bf0b-7d6e-acf8-54fbbf992290"
  },
  {
    "Volcano Name": "Esmeralda Bank",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.25,
        15
      ]
    },
    "Elevation": -43,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a05aad6c-a12d-9a95-b451-dadd01492be1"
  },
  {
    "Volcano Name": "Espenberg",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -164.33,
        66.35
      ]
    },
    "Elevation": 243,
    "Type": "Volcanic field",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "481c209c-0064-f447-1bea-dff806c75396"
  },
  {
    "Volcano Name": "Est, Ile de l'",
    "Country": "French Southern & Antarctic Lands",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        52.2,
        -46.43
      ]
    },
    "Elevation": 1090,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "b3659db8-795e-115f-a0b7-a4fff857cd08"
  },
  {
    "Volcano Name": "Esteli",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.4,
        13.17
      ]
    },
    "Elevation": 899,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "25ccb048-4a9d-2bc3-74ae-af611f3059bf"
  },
  {
    "Volcano Name": "Etna",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        15.004,
        37.734
      ]
    },
    "Elevation": 3350,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4a57d676-98e4-52d3-f248-662056579edd"
  },
  {
    "Volcano Name": "Eyjafjallajokull",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.62,
        63.63
      ]
    },
    "Elevation": 1666,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "56ed25f9-7371-e98d-5922-646819238c9b"
  },
  {
    "Volcano Name": "Falcon Island",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.42,
        -20.32
      ]
    },
    "Elevation": -17,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "fe1a5e7c-db94-f642-44f3-674dddcef45b"
  },
  {
    "Volcano Name": "Farallon de Pajaros",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.9,
        20.53
      ]
    },
    "Elevation": 360,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "75e87766-a74b-bdfa-a82a-329e9aba7a72"
  },
  {
    "Volcano Name": "Fayal",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -28.73,
        38.6
      ]
    },
    "Elevation": 1043,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "887f5ba3-325b-08fc-f5e9-46f76c137cca"
  },
  {
    "Volcano Name": "Fedotych",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.4,
        57.13
      ]
    },
    "Elevation": 965,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c6dde8d3-beee-49c9-8f3d-efef0fca870d"
  },
  {
    "Volcano Name": "Fentale",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.93,
        8.975
      ]
    },
    "Elevation": 2007,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "d586cbaf-1281-0b7f-3f1d-4450b1bb3ad9"
  },
  {
    "Volcano Name": "Fernandina",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.55,
        -0.37
      ]
    },
    "Elevation": 1495,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ccb55772-1c51-a90a-5c5e-6a9cc014c1a3"
  },
  {
    "Volcano Name": "Firura, Nevados",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.63,
        -15.23
      ]
    },
    "Elevation": 5498,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "882ee23f-d41c-9d10-e25b-cd23b1024f17"
  },
  {
    "Volcano Name": "Fisher",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -164.35,
        54.67
      ]
    },
    "Elevation": 1094,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "9817f7b4-f618-17da-b24a-4be71fae0390"
  },
  {
    "Volcano Name": "Flores",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90,
        14.3
      ]
    },
    "Elevation": 1600,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2d6830b0-fd4d-7866-5023-a71084252850"
  },
  {
    "Volcano Name": "Flores",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -31.17,
        39.4
      ]
    },
    "Elevation": 915,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "6e4a2dd3-e382-f6eb-fcfc-087ae04d95e2"
  },
  {
    "Volcano Name": "Fogo",
    "Country": "Cape Verde",
    "Region": "Cape Verde Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -24.35,
        14.95
      ]
    },
    "Elevation": 2829,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "fcda704f-0afd-8792-69e5-d6b8bbd00118"
  },
  {
    "Volcano Name": "Fonualei",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.325,
        -18.02
      ]
    },
    "Elevation": 200,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "900bb06a-44bb-1632-8fbd-c5283469c31f"
  },
  {
    "Volcano Name": "Forecast Seamount",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.917,
        13.4
      ]
    },
    "Elevation": null,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "27ed8b85-a339-1aaa-42bc-75bd7e15ce4c"
  },
  {
    "Volcano Name": "Fort Portal Field",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        30.25,
        0.7
      ]
    },
    "Elevation": 1524,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1db54383-772e-f9a4-4418-5779d6ae330c"
  },
  {
    "Volcano Name": "Fort Selkirk",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -137.38,
        62.93
      ]
    },
    "Elevation": 1239,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0b5aa5c4-60e3-a558-ac92-65676951a8e6"
  },
  {
    "Volcano Name": "Four Craters Lava Field",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -120.669,
        43.361
      ]
    },
    "Elevation": 1501,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "c738cf80-f02d-f51b-0c96-d76076964783"
  },
  {
    "Volcano Name": "Fournaise, Piton de la",
    "Country": "Reunion",
    "Region": "Indian O-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        55.713,
        -21.229
      ]
    },
    "Elevation": 2631,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "38c084c5-3b86-6096-ad1c-647317c89b06"
  },
  {
    "Volcano Name": "Fourpeaked",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -153.68,
        58.77
      ]
    },
    "Elevation": 2104,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "656cc2e1-a64e-d7a1-1ebf-b11a3ae3271d"
  },
  {
    "Volcano Name": "Fremrinamur",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.65,
        65.43
      ]
    },
    "Elevation": 939,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "94434feb-2b43-66d3-35a6-582b76ae124b"
  },
  {
    "Volcano Name": "Frosty",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -162.82,
        55.07
      ]
    },
    "Elevation": 1920,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "22304720-6ad4-d8a5-278f-b16547d2e2c9"
  },
  {
    "Volcano Name": "Fuego",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.88,
        14.473
      ]
    },
    "Elevation": 3763,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "25a131b8-9cc5-01e8-12e8-7ee600a0b193"
  },
  {
    "Volcano Name": "Fuerteventura",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -14.02,
        28.358
      ]
    },
    "Elevation": 529,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "544ca23b-105a-9615-7d3f-1b436ada736f"
  },
  {
    "Volcano Name": "Fuji",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.73,
        35.35
      ]
    },
    "Elevation": 3776,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "8b4c7cdd-a6c1-2398-494e-98755176dd57"
  },
  {
    "Volcano Name": "Fukue-jima",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        128.851,
        32.653
      ]
    },
    "Elevation": 317,
    "Type": "Shield volcanoes",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fa911d78-d424-3974-0999-dd96bd4bf778"
  },
  {
    "Volcano Name": "Fukujin",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.442,
        21.925
      ]
    },
    "Elevation": -217,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "94beb056-56d1-3bbd-2c1e-7b0ef06f5fce"
  },
  {
    "Volcano Name": "Fukutoku-okanoba",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.52,
        24.28
      ]
    },
    "Elevation": -14,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "562298ef-204e-b831-da00-eaef2eeabaa1"
  },
  {
    "Volcano Name": "Furnas",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.32,
        37.77
      ]
    },
    "Elevation": 805,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "1137b48b-82e8-ac39-9c9c-c190e11e32e8"
  },
  {
    "Volcano Name": "Fuss Peak",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.25,
        50.27
      ]
    },
    "Elevation": 1772,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "d0a8183a-2306-c2bb-cde7-3fde56a293cb"
  },
  {
    "Volcano Name": "Gabillema",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.27,
        11.08
      ]
    },
    "Elevation": 1459,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "91c758bb-21dc-c262-e824-339e5e4d11bf"
  },
  {
    "Volcano Name": "Gada Ale",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.408,
        13.975
      ]
    },
    "Elevation": 287,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "73fbc363-8a83-54bb-2446-f7ddf050f5e2"
  },
  {
    "Volcano Name": "Galapagos Rift",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.15,
        0.792
      ]
    },
    "Elevation": -2430,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b4408f71-9e87-fb23-6a9e-ecd031262241"
  },
  {
    "Volcano Name": "Galapagos Rift",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.15,
        0.792
      ]
    },
    "Elevation": -2430,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a9171404-f07b-7fe8-7ac2-a811df25fd95"
  },
  {
    "Volcano Name": "Galeras",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.37,
        1.22
      ]
    },
    "Elevation": 4276,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1384c4c9-86f5-0cc7-56f7-52ccf3568f26"
  },
  {
    "Volcano Name": "Gallego",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.73,
        -9.35
      ]
    },
    "Elevation": 1000,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "a096e074-fbc5-8ce0-efc2-32cb1261be2f"
  },
  {
    "Volcano Name": "Galunggung",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        108.05,
        -7.25
      ]
    },
    "Elevation": 2168,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d6d5f258-3751-179c-ea44-8a8d285a3617"
  },
  {
    "Volcano Name": "Gamalama",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.325,
        0.8
      ]
    },
    "Elevation": 1715,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "bf4eef17-6148-977a-27a6-69e04f443574"
  },
  {
    "Volcano Name": "Gamchen",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.702,
        54.973
      ]
    },
    "Elevation": 2576,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7971d874-c976-f3ad-c9ed-414a336c73b2"
  },
  {
    "Volcano Name": "Gamkonora",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.52,
        1.375
      ]
    },
    "Elevation": 1635,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "def3422a-5c3a-3682-f7ef-9e0c964dd194"
  },
  {
    "Volcano Name": "Garbuna Group",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.03,
        -5.45
      ]
    },
    "Elevation": 564,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "17afe285-269b-1373-de8d-5227c1c4c844"
  },
  {
    "Volcano Name": "Gareloi",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.8,
        51.78
      ]
    },
    "Elevation": 1573,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9e045a53-621d-c8ca-d832-5ef2e8234788"
  },
  {
    "Volcano Name": "Garibaldi Lake",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -123.03,
        49.92
      ]
    },
    "Elevation": 2316,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "abe3beb5-5fc3-0855-c692-7d87e7e5021c"
  },
  {
    "Volcano Name": "Garibaldi, Mt.",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -123,
        49.85
      ]
    },
    "Elevation": 2678,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ac277c83-cc3c-e407-afc4-4305224bbc9e"
  },
  {
    "Volcano Name": "Garove",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.5,
        -4.692
      ]
    },
    "Elevation": 368,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0ae7e3ca-840b-2d72-c329-a92411abce64"
  },
  {
    "Volcano Name": "Garua Harbour",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.088,
        -5.269
      ]
    },
    "Elevation": 565,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "49cb4a5b-7e6e-0a4f-814b-214c4a72188b"
  },
  {
    "Volcano Name": "Gaua",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.5,
        -14.27
      ]
    },
    "Elevation": 797,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "48efa96a-566a-3c33-054d-2273832b51c7"
  },
  {
    "Volcano Name": "Gedamsa Caldera",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.18,
        8.35
      ]
    },
    "Elevation": 1984,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8425c8f6-77a0-2e1b-ef37-60cc809c2fa8"
  },
  {
    "Volcano Name": "Gede",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.98,
        -6.78
      ]
    },
    "Elevation": 2958,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "70c16a35-d51d-521a-dfe7-da4f3372ef17"
  },
  {
    "Volcano Name": "Genovesa",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.958,
        0.32
      ]
    },
    "Elevation": 64,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "187e868f-c1a6-896f-15d7-e66c5e4597ff"
  },
  {
    "Volcano Name": "Geodesistoy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.67,
        56.33
      ]
    },
    "Elevation": 1170,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6a485035-efe6-271c-9f7a-508a1c10d202"
  },
  {
    "Volcano Name": "Giggenbach",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.713,
        -30.036
      ]
    },
    "Elevation": -65,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dd04df04-b0ab-240f-f795-584bff6b5c6d"
  },
  {
    "Volcano Name": "Girekol",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        43.33,
        39.17
      ]
    },
    "Elevation": 0,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3ef3678f-1ced-0795-1475-0a0b3ae99caf"
  },
  {
    "Volcano Name": "Glacier Peak",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.113,
        48.112
      ]
    },
    "Elevation": 3213,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "a694e8e1-1bf9-de9b-7402-63ea2a009ddc"
  },
  {
    "Volcano Name": "Gloria, La",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.25,
        19.33
      ]
    },
    "Elevation": 3500,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e0bf94ed-120b-05e5-932e-26d9f976143a"
  },
  {
    "Volcano Name": "Golden Trout Creek",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -118.32,
        36.358
      ]
    },
    "Elevation": 2886,
    "Type": "Volcanic field",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "cc0df283-24b0-10dc-9d2a-b34ea1d8f887"
  },
  {
    "Volcano Name": "Golets-Tornyi Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.35,
        45.25
      ]
    },
    "Elevation": 442,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "356aecbf-abfc-71fa-1f42-564031eec994"
  },
  {
    "Volcano Name": "Gollu Dag",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        34.57,
        38.25
      ]
    },
    "Elevation": 2143,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e32b4f23-ec6f-c663-8d3c-28e81ed7cde2"
  },
  {
    "Volcano Name": "Golovnin",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.53,
        43.85
      ]
    },
    "Elevation": 541,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "06a18b4b-cc99-84f6-6fb5-53194da9d753"
  },
  {
    "Volcano Name": "Goodenough",
    "Country": "Papua New Guinea",
    "Region": "D'Entrecasteaux Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.35,
        -9.48
      ]
    },
    "Elevation": 220,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5039ae55-6bb4-c3bc-fe8d-fe41781f1aff"
  },
  {
    "Volcano Name": "Gordon",
    "Country": "United States",
    "Region": "Alaska-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -143.08,
        62.13
      ]
    },
    "Elevation": 2755,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6b5a5a79-2ebc-0ff6-46ed-86ce5bbb42ff"
  },
  {
    "Volcano Name": "Gorely",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.03,
        52.558
      ]
    },
    "Elevation": 1829,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "23bc6bb8-76f9-9a44-2fe6-13b5f831e668"
  },
  {
    "Volcano Name": "Goriaschaia Sopka",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.75,
        46.83
      ]
    },
    "Elevation": 891,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "cd5a84c0-d09a-f7b7-a6f4-ba837cd52790"
  },
  {
    "Volcano Name": "Gorny Institute",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.2,
        57.33
      ]
    },
    "Elevation": 2125,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b7a939a5-725b-db5f-3aa8-7a2c729e725a"
  },
  {
    "Volcano Name": "Graciosa",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.97,
        39.02
      ]
    },
    "Elevation": 402,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2da41f60-76eb-cfa0-0e51-46d0145813aa"
  },
  {
    "Volcano Name": "Gran Canaria",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -15.58,
        28
      ]
    },
    "Elevation": 1950,
    "Type": "Fissure vent",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f0ca7ff1-d73f-e55e-a5bf-c355d544c9ab"
  },
  {
    "Volcano Name": "Great Sitkin",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -176.13,
        52.08
      ]
    },
    "Elevation": 1740,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9ac62fc1-68b0-086c-f597-cdd3a97f6504"
  },
  {
    "Volcano Name": "Griggs",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.1,
        58.35
      ]
    },
    "Elevation": 2317,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4092d249-97a8-4450-e6ca-23a748107b45"
  },
  {
    "Volcano Name": "Grille, La",
    "Country": "Comoros",
    "Region": "Indian O-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        43.33,
        -11.47
      ]
    },
    "Elevation": 1087,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a456172b-5cab-070a-c7ba-8a48a9ceeb74"
  },
  {
    "Volcano Name": "Grimsnes",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -20.87,
        64.03
      ]
    },
    "Elevation": 214,
    "Type": "Crater rows",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "7f302f01-8293-94be-bbdf-96ebc33024a4"
  },
  {
    "Volcano Name": "Grimsvotn",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.33,
        64.42
      ]
    },
    "Elevation": 1725,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "de103f68-ced0-5120-85cc-59c2a08cf22b"
  },
  {
    "Volcano Name": "Groppo",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.25,
        11.73
      ]
    },
    "Elevation": 930,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "74b34a4b-fd77-ecb7-6f92-f557c8ce5821"
  },
  {
    "Volcano Name": "Grozny Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.87,
        45.02
      ]
    },
    "Elevation": 1211,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7dd7661a-8bf3-3244-9f11-6ff199e3d9c9"
  },
  {
    "Volcano Name": "Guadalupe",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -118.28,
        29.07
      ]
    },
    "Elevation": 1100,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "175246e4-0254-bf3a-8f78-fbcb6b50e75b"
  },
  {
    "Volcano Name": "Guagua Pichincha",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.598,
        -0.171
      ]
    },
    "Elevation": 4784,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c21c00dd-2ac7-ebdb-d488-19cb1fbbc35b"
  },
  {
    "Volcano Name": "Guallatiri",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.17,
        -18.42
      ]
    },
    "Elevation": 6071,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ef2d132d-0487-82dd-2a9a-d7480e2c3982"
  },
  {
    "Volcano Name": "Guayaques",
    "Country": "Bolivia",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.58,
        -22.88
      ]
    },
    "Elevation": 5598,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d62dd6ce-7c32-470e-6d12-f446e0e606d0"
  },
  {
    "Volcano Name": "Guazapa",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.12,
        13.9
      ]
    },
    "Elevation": 1438,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5c8194c6-c2e9-3a7d-5449-bf8f6ba71840"
  },
  {
    "Volcano Name": "Gufa",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.53,
        12.55
      ]
    },
    "Elevation": 600,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "56cc02e7-9f1c-abd9-7524-ac814e88927d"
  },
  {
    "Volcano Name": "Guguan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.85,
        17.32
      ]
    },
    "Elevation": 287,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "74177208-395b-80b4-8d9c-00f880df7f63"
  },
  {
    "Volcano Name": "Guntur",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.83,
        -7.13
      ]
    },
    "Elevation": 2249,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "8a4b0688-ef75-ebb6-53fb-cd536491205b"
  },
  {
    "Volcano Name": "Gunungapi Wetar",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        126.65,
        -6.642
      ]
    },
    "Elevation": 282,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "b7ad1931-950b-1628-daa9-a4ad94c48bdd"
  },
  {
    "Volcano Name": "Hachijo-jima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.77,
        33.13
      ]
    },
    "Elevation": 854,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "447548c3-dad6-5ac2-4927-dba3d74b085e"
  },
  {
    "Volcano Name": "Hachimantai",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.85,
        39.95
      ]
    },
    "Elevation": 1614,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "325f62ed-1804-98de-a311-f436cad36fb7"
  },
  {
    "Volcano Name": "Hainan Dao",
    "Country": "China",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.1,
        19.7
      ]
    },
    "Elevation": null,
    "Type": "Pyroclastic cones",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b45a8ed3-f4d6-8e7d-89dc-fedfea531f45"
  },
  {
    "Volcano Name": "Hakkoda Group",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.88,
        40.65
      ]
    },
    "Elevation": 1585,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "51e3efc0-f675-2bdf-0c37-569215326aef"
  },
  {
    "Volcano Name": "Hakone",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.02,
        35.22
      ]
    },
    "Elevation": 1438,
    "Type": "Complex volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "792b345e-9457-34a3-0452-3b1dfe809d8d"
  },
  {
    "Volcano Name": "Haku-san",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        136.78,
        36.15
      ]
    },
    "Elevation": 2702,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "a1856263-4d3d-e568-bcd0-7723a5577394"
  },
  {
    "Volcano Name": "Haleakala",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -156.25,
        20.708
      ]
    },
    "Elevation": 3055,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "397feef7-0e2f-4044-5059-2f0215e51034"
  },
  {
    "Volcano Name": "Halla",
    "Country": "South Korea",
    "Region": "Korea",
    "Location": {
      "type": "Point",
      "coordinates": [
        126.53,
        33.37
      ]
    },
    "Elevation": 1950,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "f4be2e5d-e514-4e6c-3049-05772047803e"
  },
  {
    "Volcano Name": "Hanish",
    "Country": "Yemen",
    "Region": "Red Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.73,
        13.72
      ]
    },
    "Elevation": 422,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ca83a9bd-2f45-8fa3-c890-bacc27927aac"
  },
  {
    "Volcano Name": "Hargy",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.1,
        -5.33
      ]
    },
    "Elevation": 1148,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "333287ed-6359-3feb-8b11-9a43f0ee276b"
  },
  {
    "Volcano Name": "Harrah, Al",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.417,
        31.083
      ]
    },
    "Elevation": 1100,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1fe56045-b447-ecec-21fc-69bda3a7a90b"
  },
  {
    "Volcano Name": "Haruj",
    "Country": "Libya",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        17.5,
        27.25
      ]
    },
    "Elevation": 1200,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4fe63904-eb08-286d-1a4a-b75e18c870f5"
  },
  {
    "Volcano Name": "Haruna",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.88,
        36.47
      ]
    },
    "Elevation": 1449,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "f1a7ab94-dacb-e262-8818-01b77ffbc640"
  },
  {
    "Volcano Name": "Hasan Dagi",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        34.17,
        38.13
      ]
    },
    "Elevation": 3253,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e4a0b771-1257-a27e-e7aa-ad79f42f19a1"
  },
  {
    "Volcano Name": "Haut Dong Nai",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        108.2,
        11.6
      ]
    },
    "Elevation": 1000,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "09f2651c-21d4-fcf8-8bbd-79851f54c7f4"
  },
  {
    "Volcano Name": "Hayes",
    "Country": "United States",
    "Region": "Alaska-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -152.48,
        61.62
      ]
    },
    "Elevation": 2788,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "9f7ea03c-1552-82b6-e86f-4a2080942928"
  },
  {
    "Volcano Name": "Haylan, Jabal",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.78,
        15.43
      ]
    },
    "Elevation": 1550,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "de4ded30-ae5e-1caa-b61c-e315351686a0"
  },
  {
    "Volcano Name": "Hayli Gubbi",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.72,
        13.5
      ]
    },
    "Elevation": 521,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ba016fff-34e1-b9c4-7981-3fedd327fef2"
  },
  {
    "Volcano Name": "Healy",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.973,
        -35.004
      ]
    },
    "Elevation": 980,
    "Type": "Submarine volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "58f6f8b6-807a-c9ba-186a-7d55e25cafe8"
  },
  {
    "Volcano Name": "Heard",
    "Country": "Heard I. & McDonald Is.",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        73.513,
        -53.106
      ]
    },
    "Elevation": 2745,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "06e92060-1b66-282d-f603-52953b16265f"
  },
  {
    "Volcano Name": "Heart Peaks",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -131.97,
        58.6
      ]
    },
    "Elevation": 2012,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2690642c-5ed6-027a-642a-402a4e707316"
  },
  {
    "Volcano Name": "Hekla",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.7,
        63.98
      ]
    },
    "Elevation": 1491,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f63ca1de-8f28-930e-c36a-1e1528048c22"
  },
  {
    "Volcano Name": "Hell's Half Acre",
    "Country": "United States",
    "Region": "US-Idaho",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.45,
        43.5
      ]
    },
    "Elevation": 1631,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "8c5a3a58-707a-3556-9796-26fe8278ecd6"
  },
  {
    "Volcano Name": "Hengill",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.33,
        64.18
      ]
    },
    "Elevation": 803,
    "Type": "Crater rows",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "57e66369-28ab-a58f-3a49-285a1097cede"
  },
  {
    "Volcano Name": "Herbert",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -170.12,
        52.75
      ]
    },
    "Elevation": 1290,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d78b649c-8b82-ed5e-45fb-ee28e47240fe"
  },
  {
    "Volcano Name": "Hertali",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.33,
        9.78
      ]
    },
    "Elevation": 900,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "401d407e-ca1b-be08-455a-740c666cad3b"
  },
  {
    "Volcano Name": "Hibok-Hibok",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.673,
        9.203
      ]
    },
    "Elevation": 1332,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d0ad0835-e536-e4b1-f154-c996611b525b"
  },
  {
    "Volcano Name": "Hierro",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -18.03,
        27.73
      ]
    },
    "Elevation": 1500,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "16818d30-f0a8-a8cd-b73a-ffc8450ab298"
  },
  {
    "Volcano Name": "Hijiori",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.18,
        38.6
      ]
    },
    "Elevation": 516,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "8169c205-bdf7-1be8-ab73-a1adc6aa4539"
  },
  {
    "Volcano Name": "Hiri",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.32,
        0.88
      ]
    },
    "Elevation": 630,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5e24a8d5-0a20-0d0b-07b4-bbdcaae64f4e"
  },
  {
    "Volcano Name": "Hiuchi",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.28,
        36.95
      ]
    },
    "Elevation": 2346,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "e17d439a-e5ea-4570-268c-fa1baffe5c6e"
  },
  {
    "Volcano Name": "Hobicha Caldera",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.83,
        6.78
      ]
    },
    "Elevation": 1800,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "4fa2955f-ca5f-6635-f698-c08b815c5d2b"
  },
  {
    "Volcano Name": "Hodson",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.15,
        -56.7
      ]
    },
    "Elevation": 1005,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5f408f02-8e78-50a6-4e4f-b03a516ee117"
  },
  {
    "Volcano Name": "Hofsjokull",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -18.92,
        64.78
      ]
    },
    "Elevation": 1782,
    "Type": "Subglacial volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7a45129f-995f-0fe2-5be6-7a72515edca6"
  },
  {
    "Volcano Name": "Homa Mountain",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        34.5,
        -0.38
      ]
    },
    "Elevation": 1751,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4ed54344-f9df-fb2b-b69d-b5f25b1c5c9d"
  },
  {
    "Volcano Name": "Home Reef",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.775,
        -18.992
      ]
    },
    "Elevation": -2,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ea7781fa-8257-9c4f-74d3-df5dd65a29e0"
  },
  {
    "Volcano Name": "Honggeertu",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        113,
        41.47
      ]
    },
    "Elevation": 1700,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "6e67f0b2-7fd4-a343-2d29-631b6fc93a25"
  },
  {
    "Volcano Name": "Hood",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.694,
        45.374
      ]
    },
    "Elevation": 3426,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "acf67a43-d1db-0d15-4b73-ffd267d6ac9f"
  },
  {
    "Volcano Name": "Hoodoo Mountain",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -131.28,
        56.78
      ]
    },
    "Elevation": 1820,
    "Type": "Subglacial volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1da15b2e-ea17-f53e-d97e-3005a4e702db"
  },
  {
    "Volcano Name": "Hornopiren",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.431,
        -41.874
      ]
    },
    "Elevation": 1572,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "fe09cc71-24ec-f848-0484-47443f7385a3"
  },
  {
    "Volcano Name": "Hromundartindur",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.202,
        64.073
      ]
    },
    "Elevation": 540,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "d7331d85-c44f-da2a-5763-85bb89acaad1"
  },
  {
    "Volcano Name": "Hualalai",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.87,
        19.692
      ]
    },
    "Elevation": 2523,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "474bb23c-b5f0-07cd-2297-2e24dd097ad8"
  },
  {
    "Volcano Name": "Huanquihue Group",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.55,
        -39.87
      ]
    },
    "Elevation": 1300,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ae750650-5adb-f316-cfbf-695bd8aa4f08"
  },
  {
    "Volcano Name": "Huaynaputina",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.85,
        -16.608
      ]
    },
    "Elevation": 4850,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "91ae7e00-ef98-2836-df7a-0b15752a89a5"
  },
  {
    "Volcano Name": "Hudson Mountains",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -99.42,
        -74.33
      ]
    },
    "Elevation": 749,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "2f645bc3-0db4-ebbf-fb18-84fd12ab0dd7"
  },
  {
    "Volcano Name": "Hudson, Cerro",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.97,
        -45.9
      ]
    },
    "Elevation": 1905,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "deeb11ae-6a5c-ffb9-571b-9db71210150c"
  },
  {
    "Volcano Name": "Huequi",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.578,
        -42.377
      ]
    },
    "Elevation": 1318,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3355d95e-c21c-8422-e673-ad81e491d72a"
  },
  {
    "Volcano Name": "Huila",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.05,
        2.92
      ]
    },
    "Elevation": 5365,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "47a62f09-5060-c7a7-ed76-630ec2805d0e"
  },
  {
    "Volcano Name": "Hulubelu",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        104.6,
        -5.35
      ]
    },
    "Elevation": 1040,
    "Type": "Caldera",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ced7956e-2f8f-208d-fc65-af4f041d43e1"
  },
  {
    "Volcano Name": "Humeros, Los",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.45,
        19.68
      ]
    },
    "Elevation": 3150,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "4de6a24c-ea82-4fa7-782b-d7bb2f3925f6"
  },
  {
    "Volcano Name": "Hunga Tonga-Hunga Ha'apai",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.38,
        -20.57
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8d3f877d-c425-f420-2a67-35fad2654cf1"
  },
  {
    "Volcano Name": "Hunter Island",
    "Country": "Pacific Ocean",
    "Region": "SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        172.05,
        -22.4
      ]
    },
    "Elevation": 297,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "36009604-14ca-710d-b768-657473160ecf"
  },
  {
    "Volcano Name": "Hutapanjang",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        101.6,
        -2.33
      ]
    },
    "Elevation": 2021,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fd8d7782-71f0-190d-62de-ce6b9d2fc2d5"
  },
  {
    "Volcano Name": "Hutapanjang",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        101.6,
        -2.333
      ]
    },
    "Elevation": 2021,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "558f3ba7-b81d-fac4-7c8d-540f07adbe01"
  },
  {
    "Volcano Name": "Hydrographers Range",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.37,
        -9
      ]
    },
    "Elevation": 1915,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a55add03-b15d-1b86-c4f4-ea3bbcfafe8f"
  },
  {
    "Volcano Name": "Iamalele",
    "Country": "Papua New Guinea",
    "Region": "D'Entrecasteaux Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.53,
        -9.52
      ]
    },
    "Elevation": 200,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "22c908af-f879-9453-c70d-7c055700fa92"
  },
  {
    "Volcano Name": "Ibu",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.642,
        1.475
      ]
    },
    "Elevation": 1325,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ab8942b4-bfd8-3a71-1a36-aec7d5a34a67"
  },
  {
    "Volcano Name": "Ibusuki Volc Field",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.57,
        31.22
      ]
    },
    "Elevation": 922,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "f0be97cd-217d-4d5c-664d-b4727aad6eaa"
  },
  {
    "Volcano Name": "Ichinsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.73,
        55.68
      ]
    },
    "Elevation": 3621,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d1d7b238-e526-9cd8-aa61-a0de88821ccc"
  },
  {
    "Volcano Name": "Iettunup",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        161.08,
        58.4
      ]
    },
    "Elevation": 1340,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9cba2cac-dad0-af5b-0935-63b64e00531d"
  },
  {
    "Volcano Name": "Igwisi Hills",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        31.92,
        -4.87
      ]
    },
    "Elevation": 0,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e9fecaf3-0bbb-7b57-f658-af65267dadb5"
  },
  {
    "Volcano Name": "Ijen",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        114.242,
        -8.058
      ]
    },
    "Elevation": 2799,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2e727c7b-18e9-0bba-902f-7144ed927cbf"
  },
  {
    "Volcano Name": "Iktunup",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.77,
        58.08
      ]
    },
    "Elevation": 2300,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e1312c07-987a-a95f-efae-9a507b9aa4f3"
  },
  {
    "Volcano Name": "Iliamna",
    "Country": "United States",
    "Region": "Alaska-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -153.08,
        60.03
      ]
    },
    "Elevation": 3053,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4789e7e5-d98e-007d-5bbd-a7a49b699cb9"
  },
  {
    "Volcano Name": "Iliboleng",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.258,
        -8.342
      ]
    },
    "Elevation": 1659,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7a2088a8-07b1-cbc7-ff7e-fa81831838be"
  },
  {
    "Volcano Name": "Ililabalekan",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.42,
        -8.53
      ]
    },
    "Elevation": 1018,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "264e730b-19b0-72b7-8d40-1ed495ecae80"
  },
  {
    "Volcano Name": "Ilimuda",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.671,
        -8.478
      ]
    },
    "Elevation": 1100,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e2b9596c-ee0e-60b8-6365-394587fb7227"
  },
  {
    "Volcano Name": "Iliwerung",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.59,
        -8.54
      ]
    },
    "Elevation": 1018,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f3ac9099-41a2-f6e7-3942-3565d3df1dbd"
  },
  {
    "Volcano Name": "Illiniza",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.714,
        -0.659
      ]
    },
    "Elevation": 5248,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d796a898-8cc0-6b24-e218-ac4783d29c8b"
  },
  {
    "Volcano Name": "Ilopango",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.053,
        13.672
      ]
    },
    "Elevation": 450,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "17fbe175-6915-272c-f951-3910bcf74f9a"
  },
  {
    "Volcano Name": "Ilyinsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.2,
        51.49
      ]
    },
    "Elevation": 1578,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "393e9735-e913-5bf1-541f-2b400fbd32de"
  },
  {
    "Volcano Name": "Imun",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.93,
        2.15
      ]
    },
    "Elevation": 1505,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "a293faee-92a9-c180-d4b1-6ea5deb66d26"
  },
  {
    "Volcano Name": "Imuruk Lake",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.92,
        65.6
      ]
    },
    "Elevation": 610,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "9aa7c9de-1765-a29d-9fba-490a26dda262"
  },
  {
    "Volcano Name": "In Ezzane Volc Field",
    "Country": "Algeria",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        10.833,
        23
      ]
    },
    "Elevation": null,
    "Type": "Volcanic field",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "66c8208f-e211-3930-eb7c-59b9f9334ac0"
  },
  {
    "Volcano Name": "Indian Heaven",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.82,
        45.93
      ]
    },
    "Elevation": 1513,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "99255b22-75b3-e0bd-5cde-7cec02434c13"
  },
  {
    "Volcano Name": "Ingakslugwat Hills",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -164.47,
        61.43
      ]
    },
    "Elevation": 190,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "219e2a2e-499f-93d5-3cb5-3469b5e31afc"
  },
  {
    "Volcano Name": "Inielika",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.98,
        -8.73
      ]
    },
    "Elevation": 1559,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ff0b5e03-5add-2b4b-cf69-68bdf000b798"
  },
  {
    "Volcano Name": "Inierie",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.95,
        -8.875
      ]
    },
    "Elevation": 2245,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5149537d-1ae6-3e0a-7034-fe4c1c5c1031"
  },
  {
    "Volcano Name": "Inyo Craters",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -119.02,
        37.692
      ]
    },
    "Elevation": 2629,
    "Type": "Lava dome",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "108e1ae6-8b73-24fe-928f-45b5dc6a0aa6"
  },
  {
    "Volcano Name": "Ipala Volc Field",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.63,
        14.55
      ]
    },
    "Elevation": 1650,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d0aab6da-15cd-9b39-f20a-fafec3e9b2dd"
  },
  {
    "Volcano Name": "Iraya",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.01,
        20.469
      ]
    },
    "Elevation": 1009,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "287f5d06-78e8-b642-5d2e-18fd2a17337d"
  },
  {
    "Volcano Name": "Irazu",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -83.852,
        9.979
      ]
    },
    "Elevation": 3432,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0c31648e-ca74-a25d-7c7c-23264f95f22a"
  },
  {
    "Volcano Name": "Iriga",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.457,
        13.457
      ]
    },
    "Elevation": 1196,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "df71eb12-db1f-9821-3c95-dd581421449b"
  },
  {
    "Volcano Name": "Iriomote-jima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        124,
        24.558
      ]
    },
    "Elevation": -200,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "1de8a602-dd71-85b9-59a4-f949459b2856"
  },
  {
    "Volcano Name": "Irruputuncu",
    "Country": "Bolivia",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.55,
        -20.73
      ]
    },
    "Elevation": 5163,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "da527aff-2c3e-52aa-f010-d09c4a5fc575"
  },
  {
    "Volcano Name": "Isanotski",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.73,
        54.75
      ]
    },
    "Elevation": 2446,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "87754149-093e-0a24-bad1-c008e5085e09"
  },
  {
    "Volcano Name": "Isarog",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.37,
        13.658
      ]
    },
    "Elevation": 1966,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b8eac86f-55cb-80f6-af3b-43329cd13888"
  },
  {
    "Volcano Name": "Ischia",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        13.897,
        40.73
      ]
    },
    "Elevation": 789,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "ac77573a-d3ea-424b-e8c9-65d36d12af6e"
  },
  {
    "Volcano Name": "Iskut-Unuk River Cones",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.55,
        56.58
      ]
    },
    "Elevation": 1880,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "4901f67c-cf21-e7a2-16d8-7ca368f628eb"
  },
  {
    "Volcano Name": "Isluga",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.83,
        -19.15
      ]
    },
    "Elevation": 5050,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "8e21610e-e77a-c53a-5995-fc6df7ddb75a"
  },
  {
    "Volcano Name": "Itasy Volc Field",
    "Country": "Madagascar",
    "Region": "Madagascar",
    "Location": {
      "type": "Point",
      "coordinates": [
        46.77,
        -19
      ]
    },
    "Elevation": 1800,
    "Type": "Scoria cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "d55e10e6-187c-2e7e-345e-87257241ee34"
  },
  {
    "Volcano Name": "Ithnayn, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.2,
        26.58
      ]
    },
    "Elevation": 1625,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bea75385-2eee-ce28-7bfd-74c62ecd5c14"
  },
  {
    "Volcano Name": "Ivao Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.68,
        45.77
      ]
    },
    "Elevation": 1426,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2ac707cc-cd51-17a9-fcbe-9adc5798720f"
  },
  {
    "Volcano Name": "Iwaki",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.3,
        40.65
      ]
    },
    "Elevation": 1625,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "3c871676-39ee-703e-7751-b2900bc56732"
  },
  {
    "Volcano Name": "Iwate",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141,
        39.85
      ]
    },
    "Elevation": 2041,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2e98e7b4-41e7-f817-c21b-34355a874aeb"
  },
  {
    "Volcano Name": "Iwo-Tori-shima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        128.25,
        27.85
      ]
    },
    "Elevation": 217,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d7ff5070-87cf-b71f-037f-157d87997d2b"
  },
  {
    "Volcano Name": "Iwo-jima",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.33,
        24.75
      ]
    },
    "Elevation": 161,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ccc74302-6c31-b0fb-70d4-dc60447bbccd"
  },
  {
    "Volcano Name": "Ixtepeque",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.68,
        14.42
      ]
    },
    "Elevation": 1292,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7df68cd7-02de-750f-207c-3555fa1d9444"
  },
  {
    "Volcano Name": "Iya",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.63,
        -8.88
      ]
    },
    "Elevation": 637,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ca64ab03-7782-33af-24dd-8355594a3f4c"
  },
  {
    "Volcano Name": "Iyang-Argapura",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        113.57,
        -7.97
      ]
    },
    "Elevation": 3088,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ad88627b-c2c6-025f-ee21-d78cae59433a"
  },
  {
    "Volcano Name": "Izalco",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.633,
        13.813
      ]
    },
    "Elevation": 1950,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "36ba3684-d17b-c22d-e1f6-e721b61c09f9"
  },
  {
    "Volcano Name": "Iztaccihuatl",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -98.642,
        19.179
      ]
    },
    "Elevation": 5230,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2a48f411-fbb9-9394-565b-a64f2e510033"
  },
  {
    "Volcano Name": "Izu-Tobu",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.12,
        34.92
      ]
    },
    "Elevation": 1406,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b8adf9bb-d05f-1145-8e03-de5970f3b0a6"
  },
  {
    "Volcano Name": "Izumbwe-Mpoli",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.4,
        -8.93
      ]
    },
    "Elevation": 1568,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "138f3eeb-6deb-d2ef-021b-fcb997d3547f"
  },
  {
    "Volcano Name": "Jailolo",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.42,
        1.08
      ]
    },
    "Elevation": 1130,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "efdf00e1-8e90-d34a-6b76-5a4947133f80"
  },
  {
    "Volcano Name": "Jalajala",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.33,
        14.35
      ]
    },
    "Elevation": 743,
    "Type": "Fumarole field",
    "Status": "Fumarolic",
    "Last Known Eruption": "Unknown",
    "id": "9e6fca9f-b968-3523-f92b-9e8e87d9582c"
  },
  {
    "Volcano Name": "Jalua",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.82,
        15.042
      ]
    },
    "Elevation": 713,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "281ad876-5cf2-a297-1a48-1b2644dc90a8"
  },
  {
    "Volcano Name": "Jan Mayen",
    "Country": "Jan Mayen",
    "Region": "Atlantic-N-Jan Mayen",
    "Location": {
      "type": "Point",
      "coordinates": [
        -8.17,
        71.08
      ]
    },
    "Elevation": 2277,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8161fb84-bd0c-4c54-61e8-6bf020d578ef"
  },
  {
    "Volcano Name": "Jaraguay Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -114.5,
        29.33
      ]
    },
    "Elevation": 960,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b8cab54e-f289-b187-877f-a97ccd61913a"
  },
  {
    "Volcano Name": "Jayu Khota, Laguna",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.417,
        -19.45
      ]
    },
    "Elevation": 3650,
    "Type": "Maars",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "35b85345-6610-90db-ae08-813c6619b43d"
  },
  {
    "Volcano Name": "Jefferson",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.8,
        44.692
      ]
    },
    "Elevation": 3199,
    "Type": "Stratovolcano",
    "Status": "Varve Count",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "20c1add8-2e8d-b171-355a-be6121c0c096"
  },
  {
    "Volcano Name": "Jingpohu",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        128.83,
        44.08
      ]
    },
    "Elevation": 500,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "671e7fcf-d10a-96cc-a487-c980510fbc63"
  },
  {
    "Volcano Name": "Jocotitlan",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -99.758,
        19.733
      ]
    },
    "Elevation": 3900,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "4df61ed7-9715-9cfe-9db4-4ed46e3ee3f8"
  },
  {
    "Volcano Name": "Jocotitlan",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -99.757,
        19.724
      ]
    },
    "Elevation": 3900,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "e8aa0ab4-9a60-4017-191a-6dc27f44360f"
  },
  {
    "Volcano Name": "Jordan Craters",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -117.42,
        43.03
      ]
    },
    "Elevation": 1473,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d887e927-6ce0-2941-6435-5cd01c6b1133"
  },
  {
    "Volcano Name": "Joya, La",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.98,
        11.92
      ]
    },
    "Elevation": 300,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8ea91f88-d068-1916-c25d-419dff2a4d59"
  },
  {
    "Volcano Name": "Kaba",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.62,
        -3.52
      ]
    },
    "Elevation": 1952,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "7826d84d-253c-0c61-9380-8ae740de5695"
  },
  {
    "Volcano Name": "Kabargin Oth Group",
    "Country": "Georgia",
    "Region": "Georgia",
    "Location": {
      "type": "Point",
      "coordinates": [
        44,
        42.55
      ]
    },
    "Elevation": 3650,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "93364406-0014-9c6e-c445-2b9509ea905b"
  },
  {
    "Volcano Name": "Kadovar",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.62,
        -3.62
      ]
    },
    "Elevation": 365,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5e4e8888-846d-72a2-41e6-272aed0b63b5"
  },
  {
    "Volcano Name": "Kagamil",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.72,
        52.97
      ]
    },
    "Elevation": 893,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e666926c-9a9b-5fa4-0730-efc2e8f2b027"
  },
  {
    "Volcano Name": "Kaguyak",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.05,
        58.62
      ]
    },
    "Elevation": 901,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "edaf347b-33fa-fb9c-2c07-b3b1e5dcaf06"
  },
  {
    "Volcano Name": "Kaikata Seamount",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141,
        26.667
      ]
    },
    "Elevation": -162,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ae69bfd0-d172-8565-2668-ffa7ae165925"
  },
  {
    "Volcano Name": "Kaikohe-Bay of Islands",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        173.9,
        -35.3
      ]
    },
    "Elevation": 388,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "c226a60b-a701-21b3-aebc-dee5e8bd96e4"
  },
  {
    "Volcano Name": "Kaileney",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.67,
        57.8
      ]
    },
    "Elevation": 1582,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e70b615c-a9ef-e19e-3d3e-40bce5f6cab6"
  },
  {
    "Volcano Name": "Kaitoku Seamount",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.102,
        26.122
      ]
    },
    "Elevation": -10,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9d6198b0-a174-9d30-7ae3-0897f20efc19"
  },
  {
    "Volcano Name": "Kalatungan",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.8,
        7.95
      ]
    },
    "Elevation": 2824,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "95c7df63-7a1a-83c8-8c50-16a69c34cac0"
  },
  {
    "Volcano Name": "Kambalny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.87,
        51.3
      ]
    },
    "Elevation": 2156,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "690dd742-7c5d-88dd-7003-617a53416af8"
  },
  {
    "Volcano Name": "Kamen",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.593,
        56.02
      ]
    },
    "Elevation": 4585,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1261822b-af7f-7c8d-6505-a67ebbde65d6"
  },
  {
    "Volcano Name": "Kana Keoki",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.03,
        -8.75
      ]
    },
    "Elevation": -700,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "04cede81-4f07-a8a8-6b79-8b69577111b5"
  },
  {
    "Volcano Name": "Kanaga",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.17,
        51.92
      ]
    },
    "Elevation": 1307,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "61e03287-a7ea-ba72-02a4-08fc5269c2ce"
  },
  {
    "Volcano Name": "Kao",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.033,
        -19.667
      ]
    },
    "Elevation": 1030,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6bf2cdfc-4d8c-d6de-c37a-4c720e38bbad"
  },
  {
    "Volcano Name": "Karacalidag",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.83,
        37.67
      ]
    },
    "Elevation": 1957,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "52e3db05-a997-3e75-1fd6-3b4bac81d7cb"
  },
  {
    "Volcano Name": "Karaha, Kawah",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        108.08,
        -7.12
      ]
    },
    "Elevation": 1155,
    "Type": "Fumarole field",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5c664bcf-c920-670f-598b-8057d2eab90b"
  },
  {
    "Volcano Name": "Karang",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.042,
        -6.27
      ]
    },
    "Elevation": 1778,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "a6221fd9-9cb4-5e3f-e9ed-aca213751e78"
  },
  {
    "Volcano Name": "Karangetang [Api Siau]",
    "Country": "Indonesia",
    "Region": "Sangihe Is-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.48,
        2.78
      ]
    },
    "Elevation": 1784,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f8c9e04c-5516-17e4-8887-95686057743f"
  },
  {
    "Volcano Name": "Karapinar Field",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.65,
        37.67
      ]
    },
    "Elevation": 1302,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "36d85630-6abf-837f-3767-ea960a5c2932"
  },
  {
    "Volcano Name": "Karisimbi",
    "Country": "Congo, DRC",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.45,
        -1.5
      ]
    },
    "Elevation": 4507,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6ba0f02a-4a1c-8449-f94a-f7552aa86054"
  },
  {
    "Volcano Name": "Karkar",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.964,
        -4.649
      ]
    },
    "Elevation": 1839,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "68f8e731-07a7-9f5a-1340-736ccdb4d3e7"
  },
  {
    "Volcano Name": "Karpinsky Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.37,
        50.13
      ]
    },
    "Elevation": 1345,
    "Type": "Cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "5db100f4-99ed-f672-0ebc-f7f8a5a1c1f7"
  },
  {
    "Volcano Name": "Kars Plateau",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.9,
        40.75
      ]
    },
    "Elevation": 3000,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "9b2f1334-688d-933e-608b-04b8cc321a31"
  },
  {
    "Volcano Name": "Karthala",
    "Country": "Comoros",
    "Region": "Indian O-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        43.38,
        -11.75
      ]
    },
    "Elevation": 2361,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0fa5b841-c64a-87a7-8342-49f12d303662"
  },
  {
    "Volcano Name": "Karymsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.43,
        54.05
      ]
    },
    "Elevation": 1536,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "89f4b531-ac6c-a5cb-98ed-f77e4e837345"
  },
  {
    "Volcano Name": "Kasatochi",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.5,
        52.18
      ]
    },
    "Elevation": 314,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5720a0d4-9220-80ed-b9c0-96e0a45d7a7f"
  },
  {
    "Volcano Name": "Kasbek",
    "Country": "Georgia",
    "Region": "Georgia",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.5,
        42.7
      ]
    },
    "Elevation": 5050,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "4f38dc2f-04ad-726b-0ea1-003ee097e323"
  },
  {
    "Volcano Name": "Kasuga",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.72,
        21.77
      ]
    },
    "Elevation": -558,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d3f30eda-a8cd-377d-ca06-7a967e77b244"
  },
  {
    "Volcano Name": "Katla",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.05,
        63.63
      ]
    },
    "Elevation": 1512,
    "Type": "Subglacial volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a3d05095-68ee-3ff4-e8db-0c447dc431d1"
  },
  {
    "Volcano Name": "Katmai",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.98,
        58.27
      ]
    },
    "Elevation": 2047,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "174814ec-1eea-aeea-4063-8012c195aa72"
  },
  {
    "Volcano Name": "Katunga",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        30.18,
        -0.47
      ]
    },
    "Elevation": 1707,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2c3a8c72-fa89-c342-0e66-50b2653cf47c"
  },
  {
    "Volcano Name": "Katwe-Kikorongo Field",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.92,
        -0.08
      ]
    },
    "Elevation": 1067,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c6c30124-13ee-1617-5b7b-91d24540c583"
  },
  {
    "Volcano Name": "Kavachi",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.95,
        -9.02
      ]
    },
    "Elevation": -20,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "34532cb5-066d-f63b-ed93-c7b9b789ef77"
  },
  {
    "Volcano Name": "Kawi-Butak",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.45,
        -7.92
      ]
    },
    "Elevation": 2651,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b8cc8dd9-ded5-c67c-ca02-26fbff086e9e"
  },
  {
    "Volcano Name": "Kebeney",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.93,
        57.1
      ]
    },
    "Elevation": 1527,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "46df7ed0-7c8e-b4b9-2fa5-119a5f6cd4ba"
  },
  {
    "Volcano Name": "Kekurny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.85,
        56.4
      ]
    },
    "Elevation": 1377,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "334448f2-d9f9-8f51-1c2a-f151d22df00a"
  },
  {
    "Volcano Name": "Kelimutu",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.83,
        -8.758
      ]
    },
    "Elevation": 1640,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2b5e62b3-52b4-e60e-8085-c00105aa2eba"
  },
  {
    "Volcano Name": "Kell",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.35,
        51.65
      ]
    },
    "Elevation": 900,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fb6d18ab-e555-77cf-df43-c29d3d30da49"
  },
  {
    "Volcano Name": "Keluo Group",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.92,
        49.37
      ]
    },
    "Elevation": 670,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cb55020e-f3f9-040a-f779-7b18c01d74a3"
  },
  {
    "Volcano Name": "Kelut",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.308,
        -7.93
      ]
    },
    "Elevation": 1731,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b24eeb5e-6670-ab0c-9ccb-e0b3d93f1d2d"
  },
  {
    "Volcano Name": "Kendang",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.72,
        -7.23
      ]
    },
    "Elevation": 2608,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eaa1175c-fab7-d2f4-0798-f591db1043ca"
  },
  {
    "Volcano Name": "Kerguelen Islands",
    "Country": "French Southern & Antarctic Lands",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        69.5,
        -49.58
      ]
    },
    "Elevation": 1840,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b76cf1fb-6e4e-ca6d-6997-2896284af6fe"
  },
  {
    "Volcano Name": "Kerinci",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        101.264,
        -1.814
      ]
    },
    "Elevation": 3800,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c67ff524-72e9-0f0d-9ff7-823afd70ce48"
  },
  {
    "Volcano Name": "Ketoi",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.475,
        47.35
      ]
    },
    "Elevation": 1172,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "6f806c0b-e11b-97bd-93c9-0fd434779390"
  },
  {
    "Volcano Name": "Khangar",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.38,
        54.75
      ]
    },
    "Elevation": 2000,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "b6440936-78f4-c46b-c599-b36544d73f9b"
  },
  {
    "Volcano Name": "Khanuy Gol",
    "Country": "Mongolia",
    "Region": "Mongolia",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.75,
        48.67
      ]
    },
    "Elevation": 1886,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6a3c80b6-b330-f328-2158-be392c6dadec"
  },
  {
    "Volcano Name": "Kharimkotan",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.508,
        49.12
      ]
    },
    "Elevation": 1145,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "bb82697d-735c-a13a-06a7-16edff71f383"
  },
  {
    "Volcano Name": "Khaybar, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.92,
        25
      ]
    },
    "Elevation": 2093,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "af93e01f-1087-21ac-b0c3-d19624b52885"
  },
  {
    "Volcano Name": "Khodutka",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.703,
        52.063
      ]
    },
    "Elevation": 2090,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "583fb085-584f-ede8-77c9-b09dd6d0b782"
  },
  {
    "Volcano Name": "Kialagvik",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -156.75,
        57.38
      ]
    },
    "Elevation": 1575,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8abeeb74-5546-a002-02db-a6459216e05b"
  },
  {
    "Volcano Name": "Kiaraberes-Gagak",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.65,
        -6.73
      ]
    },
    "Elevation": 1511,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "098a8f85-c625-6367-ff0a-f1e90e688cb8"
  },
  {
    "Volcano Name": "Kick-'em-Jenny",
    "Country": "Netherlands",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.63,
        12.3
      ]
    },
    "Elevation": -177,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3bc61029-73a3-fcf4-ad08-39e63e671a79"
  },
  {
    "Volcano Name": "Kieyo",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.78,
        -9.23
      ]
    },
    "Elevation": 2175,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "1a845b54-05b9-4318-1a31-00bea20e2ee6"
  },
  {
    "Volcano Name": "Kikai",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.28,
        30.78
      ]
    },
    "Elevation": 717,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "630566fc-e75b-da55-885f-a72071096265"
  },
  {
    "Volcano Name": "Kikhpinych",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.253,
        54.487
      ]
    },
    "Elevation": 1552,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "2bd5fafd-9504-6c2c-c2e4-c9c17607fb25"
  },
  {
    "Volcano Name": "Kilauea",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.292,
        19.425
      ]
    },
    "Elevation": 1222,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "27222e2d-1478-8b75-eb09-08073fcb175a"
  },
  {
    "Volcano Name": "Kilimanjaro",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.35,
        -3.07
      ]
    },
    "Elevation": 5895,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eb51144b-dcea-2280-c6a1-3da07167998e"
  },
  {
    "Volcano Name": "Kinenin",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.967,
        57.35
      ]
    },
    "Elevation": 583,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "387231d5-32c6-c4e9-be1b-00e79d311d1c"
  },
  {
    "Volcano Name": "Kirishima",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.87,
        31.93
      ]
    },
    "Elevation": 1700,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c6956fb8-92ac-335f-9e24-3f09a3bc3b60"
  },
  {
    "Volcano Name": "Kishb, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.38,
        22.8
      ]
    },
    "Elevation": 1475,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e7c69ec8-ddff-23d1-36db-7ef849cd5770"
  },
  {
    "Volcano Name": "Kiska",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        177.6,
        52.1
      ]
    },
    "Elevation": 1220,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "89d2bfda-f6b3-dddc-d18d-24a177c3ccb6"
  },
  {
    "Volcano Name": "Kita-Fukutokutai",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.419,
        24.414
      ]
    },
    "Elevation": -73,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "U1",
    "id": "d2d8bac3-71ac-1a60-f314-e7039ad1ad55"
  },
  {
    "Volcano Name": "Kita-Iwo-jima",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.23,
        25.43
      ]
    },
    "Elevation": 792,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ba51f4e8-e570-0b97-4207-9fdcd34f3e1b"
  },
  {
    "Volcano Name": "Kizimen",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.32,
        55.13
      ]
    },
    "Elevation": 2485,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "605babdc-3f8b-fe8f-21f2-9a0cc21a50a5"
  },
  {
    "Volcano Name": "Klabat",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.03,
        1.47
      ]
    },
    "Elevation": 1995,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a6d96517-6dee-ee99-1af4-b07b0ad201b4"
  },
  {
    "Volcano Name": "Kliuchevskoi",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.638,
        56.057
      ]
    },
    "Elevation": 4835,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c2a6b021-75c8-2192-1f32-687483f14b73"
  },
  {
    "Volcano Name": "Kolbeinsey Ridge",
    "Country": "Iceland",
    "Region": "Iceland-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        -18.5,
        66.67
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "296aeab5-ca1f-e81d-18e7-e35862b36d47"
  },
  {
    "Volcano Name": "Kolkhozhny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.77,
        55.07
      ]
    },
    "Elevation": 2161,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "13c41159-c356-1a71-a930-92bc4fee55a8"
  },
  {
    "Volcano Name": "Kolokol Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.05,
        46.042
      ]
    },
    "Elevation": 1328,
    "Type": "Somma volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "17408717-f2c1-aefe-19be-c001ba6ec142"
  },
  {
    "Volcano Name": "Komaga-take",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.68,
        42.07
      ]
    },
    "Elevation": 1140,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7424c1d6-029a-f128-df34-4e87ca010452"
  },
  {
    "Volcano Name": "Komarov",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.72,
        55.032
      ]
    },
    "Elevation": 2070,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0f619a4f-fb68-9158-d731-75bb15ca8085"
  },
  {
    "Volcano Name": "Kone",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.692,
        8.8
      ]
    },
    "Elevation": 1619,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "880c5b8f-a7dc-c96a-5c07-0f0491d6e571"
  },
  {
    "Volcano Name": "Koniuji",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.13,
        52.22
      ]
    },
    "Elevation": 272,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "e7f96bd1-6c6f-5003-28f4-979a3e78a794"
  },
  {
    "Volcano Name": "Kookooligit Mountains",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -170.43,
        63.6
      ]
    },
    "Elevation": 673,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "008747cd-fbd7-b3b1-bbe9-6a99467473ec"
  },
  {
    "Volcano Name": "Koranga",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.708,
        -7.33
      ]
    },
    "Elevation": 0,
    "Type": "Explosion crater",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "79f81f39-f76f-a7c4-7dd9-ce615063c73d"
  },
  {
    "Volcano Name": "Korath Range",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        35.88,
        5.1
      ]
    },
    "Elevation": 912,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "c061a91d-1372-21d3-2fe6-6d616ed678be"
  },
  {
    "Volcano Name": "Koro",
    "Country": "Fiji",
    "Region": "Fiji Is-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        179.4,
        -17.32
      ]
    },
    "Elevation": 522,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "cb9c5d1c-d3bd-39f0-47ee-3ebc13e401a8"
  },
  {
    "Volcano Name": "Korosi",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.12,
        0.77
      ]
    },
    "Elevation": 1446,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6c005884-acfe-3a26-a447-ef926914670d"
  },
  {
    "Volcano Name": "Koryaksky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.688,
        53.32
      ]
    },
    "Elevation": 3456,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "519a6ddd-0ed2-d7b3-0769-a2696052fb1e"
  },
  {
    "Volcano Name": "Koshelev",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.75,
        51.357
      ]
    },
    "Elevation": 1812,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "614e9eee-f46d-2a2e-bff5-22dd8031f5d7"
  },
  {
    "Volcano Name": "Koussi, Emi",
    "Country": "Chad",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        18.53,
        19.8
      ]
    },
    "Elevation": 3415,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "db9a39f9-b9d9-106a-73b1-65149779a644"
  },
  {
    "Volcano Name": "Kozu-shima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.15,
        34.22
      ]
    },
    "Elevation": 574,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "03984a86-ca1e-f5d5-20ca-c9c3ca1241dd"
  },
  {
    "Volcano Name": "Kozyrevsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.38,
        55.58
      ]
    },
    "Elevation": 2016,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4cb0f884-03a6-5720-0638-017e1cd81977"
  },
  {
    "Volcano Name": "Krafla",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.78,
        65.73
      ]
    },
    "Elevation": 650,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "20718521-3ee9-59f6-7b27-d832bf2c98da"
  },
  {
    "Volcano Name": "Krainy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.03,
        56.37
      ]
    },
    "Elevation": 1554,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bdd83650-d590-ae23-1515-73ad3c26bb55"
  },
  {
    "Volcano Name": "Krakagiger",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.7,
        63.98
      ]
    },
    "Elevation": 1491,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3714b4d9-c7f5-93cb-1330-93b158ca22c1"
  },
  {
    "Volcano Name": "Krakatau",
    "Country": "Indonesia",
    "Region": "Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        105.423,
        -6.102
      ]
    },
    "Elevation": 813,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2cf0ae99-d89f-a508-af40-ccb381a2c1f8"
  },
  {
    "Volcano Name": "Krasheninnikov",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.273,
        54.593
      ]
    },
    "Elevation": 1856,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "22386618-0699-a0f0-d079-7a27428c69a5"
  },
  {
    "Volcano Name": "Kristnitokugigar",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.42,
        63.98
      ]
    },
    "Elevation": 301,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Unknown",
    "id": "9fd22841-46e5-5d6a-72ad-253843bbb190"
  },
  {
    "Volcano Name": "Krisuvik",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -22.1,
        63.93
      ]
    },
    "Elevation": 379,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "20303d61-a31f-cd77-920b-f09934df9956"
  },
  {
    "Volcano Name": "Kronotsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.527,
        54.753
      ]
    },
    "Elevation": 3528,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ca1faf68-1d36-4b1f-cc75-e127791ba7bf"
  },
  {
    "Volcano Name": "Ksudach",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.53,
        51.8
      ]
    },
    "Elevation": 1079,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b602de7c-6506-90c9-21f5-67324c28d5ba"
  },
  {
    "Volcano Name": "Kuchino-shima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.928,
        29.964
      ]
    },
    "Elevation": 628,
    "Type": "Stratovolcanoes",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "2edae0a6-9347-3bbb-4c80-f0e5ff3095b1"
  },
  {
    "Volcano Name": "Kuchino-shima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.93,
        29.97
      ]
    },
    "Elevation": 627,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "9fa3b0dd-e496-0129-52e1-1eada76de747"
  },
  {
    "Volcano Name": "Kuchinoerabu-jima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.22,
        30.43
      ]
    },
    "Elevation": 649,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "fcefa069-5316-2eef-ea22-6f4105e92538"
  },
  {
    "Volcano Name": "Kuei-Shan-Tao",
    "Country": "Taiwan",
    "Region": "Taiwan",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.92,
        24.85
      ]
    },
    "Elevation": 401,
    "Type": "Stratovolcano",
    "Status": "Pleistocene-Fumarol",
    "Last Known Eruption": "Quaternary eruption(s) with the only known Holocene activity being hydrothermal",
    "id": "8ea23526-a900-1738-9872-23048e75534f"
  },
  {
    "Volcano Name": "Kuju Group",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        131.25,
        33.08
      ]
    },
    "Elevation": 1788,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ba654d9e-6992-25e0-12e3-cef895908f79"
  },
  {
    "Volcano Name": "Kukak",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.35,
        58.47
      ]
    },
    "Elevation": 2040,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "47af4699-c134-f263-485c-38b52d9afb79"
  },
  {
    "Volcano Name": "Kula",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        28.52,
        38.58
      ]
    },
    "Elevation": 750,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "823fddd9-dacc-5bc6-3c8b-7135c1918791"
  },
  {
    "Volcano Name": "Kulkev",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.37,
        56.37
      ]
    },
    "Elevation": 915,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "796b8a47-01ae-a33c-4bf0-93c185bb502d"
  },
  {
    "Volcano Name": "Kunlun Volc Group",
    "Country": "China",
    "Region": "China-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        80.2,
        35.52
      ]
    },
    "Elevation": 5808,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e600f9e3-ef3a-cd10-cd2b-23c74bca2454"
  },
  {
    "Volcano Name": "Kupreanof",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -159.8,
        56.02
      ]
    },
    "Elevation": 1895,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "859d19cb-d950-2a97-fdf8-12017c02361c"
  },
  {
    "Volcano Name": "Kurikoma",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.78,
        38.95
      ]
    },
    "Elevation": 1628,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "751dc523-c298-4af6-69a9-25bfc2f2fd35"
  },
  {
    "Volcano Name": "Kurile Lake",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.12,
        51.45
      ]
    },
    "Elevation": 104,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "a2d27942-02fe-2e80-256d-edb86f01a31f"
  },
  {
    "Volcano Name": "Kurose Hole",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.68,
        33.4
      ]
    },
    "Elevation": -107,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "087eacd7-82df-75be-650b-07ba39f473f2"
  },
  {
    "Volcano Name": "Kurose Hole",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.683,
        33.4
      ]
    },
    "Elevation": -107,
    "Type": "Submarine volcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "8c829648-9723-2969-1527-8b05797cc76c"
  },
  {
    "Volcano Name": "Kurub",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.208,
        11.88
      ]
    },
    "Elevation": 625,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "99935db2-17cf-7b59-0c00-a557867a8c0e"
  },
  {
    "Volcano Name": "Kusatsu-Shirane",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.55,
        36.62
      ]
    },
    "Elevation": 2176,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "87df62cf-f233-1044-2c82-ae071341f935"
  },
  {
    "Volcano Name": "Kutcharo",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.43,
        43.55
      ]
    },
    "Elevation": 1000,
    "Type": "Caldera",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "e665cd22-058b-32ac-1e5a-0c7a8726dacc"
  },
  {
    "Volcano Name": "Kuttara",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.18,
        42.5
      ]
    },
    "Elevation": 581,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "447c49cf-c251-835e-d98d-0ead45376eef"
  },
  {
    "Volcano Name": "Kutum Volc Field",
    "Country": "Sudan",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        25.8,
        14.5
      ]
    },
    "Elevation": 0,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "cf0b0de1-aaa5-dbdb-1629-3b0ef5211c47"
  },
  {
    "Volcano Name": "Kuwae",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.536,
        -16.829
      ]
    },
    "Elevation": -2,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c1aa52bb-8cfa-5076-fe42-83aa39153881"
  },
  {
    "Volcano Name": "Kverkfjoll",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.72,
        64.65
      ]
    },
    "Elevation": 1920,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "cce5fc17-96d4-e9f7-1849-47c44180e131"
  },
  {
    "Volcano Name": "Kyatwa Volc Field",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        30.25,
        0.45
      ]
    },
    "Elevation": 1430,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "6c12f1be-ec6f-c45c-f052-e049417c5b48"
  },
  {
    "Volcano Name": "La Palma",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.83,
        28.58
      ]
    },
    "Elevation": 2426,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "57aea83b-66e3-18a3-6115-286ac3dd6241"
  },
  {
    "Volcano Name": "Lajas, Las",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.73,
        12.3
      ]
    },
    "Elevation": 926,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "bd08724a-ead7-4aee-268e-ecdd96d0106a"
  },
  {
    "Volcano Name": "Lakagigar",
    "Country": "Sweden",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        17.33,
        64.42
      ]
    },
    "Elevation": 1725,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ea6af622-ce46-1275-155b-8e93c6961142"
  },
  {
    "Volcano Name": "Lambafit",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.4,
        64.08
      ]
    },
    "Elevation": 550,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "317a3644-64f7-3591-942f-42e7c540d9f9"
  },
  {
    "Volcano Name": "Lamington",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.15,
        -8.95
      ]
    },
    "Elevation": 1680,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c0e8d256-d274-174c-6630-5764c72d5167"
  },
  {
    "Volcano Name": "Lamongan",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        113.342,
        -8
      ]
    },
    "Elevation": 1651,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "761d214e-471b-af1d-1246-b9382da212f0"
  },
  {
    "Volcano Name": "Langila",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.42,
        -5.525
      ]
    },
    "Elevation": 1330,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d3bb4beb-35d3-c2d3-e965-caf3ac933c6c"
  },
  {
    "Volcano Name": "Langjokull",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.98,
        64.75
      ]
    },
    "Elevation": 1360,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "f74aea70-5f3d-1faa-802c-8620a6b119e3"
  },
  {
    "Volcano Name": "Lanin",
    "Country": "Argentina",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.5,
        -39.633
      ]
    },
    "Elevation": 3747,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "38e7eed2-5c50-ff00-b019-45cde447e9ce"
  },
  {
    "Volcano Name": "Lanzarote",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -13.63,
        29.03
      ]
    },
    "Elevation": 670,
    "Type": "Fissure vent",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "3a306294-2a5c-48ec-5887-d90ff579c0aa"
  },
  {
    "Volcano Name": "Larderello",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        10.87,
        43.25
      ]
    },
    "Elevation": 500,
    "Type": "Explosion crater",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "ef193a8a-4ef3-6af0-a752-e7bfb832ea1d"
  },
  {
    "Volcano Name": "Lascar",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.73,
        -23.37
      ]
    },
    "Elevation": 5592,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d3dee368-cab1-86da-934d-bf663b09d106"
  },
  {
    "Volcano Name": "Lassen Volc Center",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.508,
        40.492
      ]
    },
    "Elevation": 3187,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "082433e1-b671-a375-db63-fa02c0aea5bd"
  },
  {
    "Volcano Name": "Lastarria",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.5,
        -25.17
      ]
    },
    "Elevation": 5697,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "776f95e1-f57a-918b-8aa5-1bbd28f56f11"
  },
  {
    "Volcano Name": "Late",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.65,
        -18.806
      ]
    },
    "Elevation": 518,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "837d910d-da6d-018c-cb1c-33165666afe7"
  },
  {
    "Volcano Name": "Latukan",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.47,
        7.65
      ]
    },
    "Elevation": 2158,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4ae39e29-bcd8-2231-e9b7-7f600b66f9a2"
  },
  {
    "Volcano Name": "Lautaro",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.55,
        -49.02
      ]
    },
    "Elevation": 3380,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "27515483-26ce-a3a7-4a52-4247f23c6642"
  },
  {
    "Volcano Name": "Lavic Lake",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -116.625,
        34.75
      ]
    },
    "Elevation": 1495,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7ba9061a-9ca5-3431-83d7-07872cc29e8c"
  },
  {
    "Volcano Name": "Lawu",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        111.192,
        -7.625
      ]
    },
    "Elevation": 3265,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3d55cad6-856e-147c-5278-07427634ff4e"
  },
  {
    "Volcano Name": "Leizhou Bandao",
    "Country": "China",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.78,
        20.83
      ]
    },
    "Elevation": 259,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ba3aca86-4db2-c77e-63fe-852455185d2a"
  },
  {
    "Volcano Name": "Lengai, Ol Doinyo",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        35.902,
        -2.751
      ]
    },
    "Elevation": 2890,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ca0a7bbf-f210-6dcc-e6dd-4744eb5257aa"
  },
  {
    "Volcano Name": "Leonard Range",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        126.397,
        7.393
      ]
    },
    "Elevation": 1190,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "47765e24-ac30-b126-8e31-0aab5c49a9e3"
  },
  {
    "Volcano Name": "Leroboleng",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.842,
        -8.358
      ]
    },
    "Elevation": 1117,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "6ece41ca-b87c-1730-e8c4-a8ae61b10b97"
  },
  {
    "Volcano Name": "Leskov Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -28.13,
        -56.67
      ]
    },
    "Elevation": 190,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dc4afd1a-a51a-d2ba-e604-112d10b4fc5a"
  },
  {
    "Volcano Name": "Leutongey",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.83,
        57.3
      ]
    },
    "Elevation": 1333,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e6e2e89c-e60b-dd91-5390-b937090c178b"
  },
  {
    "Volcano Name": "Level Mountain",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -131.35,
        58.42
      ]
    },
    "Elevation": 2190,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "909e3054-3a80-97aa-4feb-48f2727f97c9"
  },
  {
    "Volcano Name": "Lewotobi",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.775,
        -8.53
      ]
    },
    "Elevation": 1703,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8f672352-b3a0-c152-d95c-83b0e0fff72d"
  },
  {
    "Volcano Name": "Lewotobi Perempuan",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.78,
        -8.575
      ]
    },
    "Elevation": 1703,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c05e4f26-1743-29ae-1c6b-80150e03d9d6"
  },
  {
    "Volcano Name": "Lewotolo",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.505,
        -8.272
      ]
    },
    "Elevation": 1423,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "cad7a81a-0f13-dbbf-ace9-e915b8ac7e14"
  },
  {
    "Volcano Name": "Lexone",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.48,
        -17.87
      ]
    },
    "Elevation": 5340,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "eb268f4f-d12c-4511-98e3-ad6d5945cb73"
  },
  {
    "Volcano Name": "Liado Hayk Field",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.28,
        9.57
      ]
    },
    "Elevation": 878,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "2d70cc4c-855a-2f93-d4b6-f00e2405952c"
  },
  {
    "Volcano Name": "Liamuiga",
    "Country": "St. Kitts & Nevis",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -62.8,
        17.37
      ]
    },
    "Elevation": 1156,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "a5888da6-c94c-0a7c-82e5-8d3969332ab5"
  },
  {
    "Volcano Name": "Licancabur",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.88,
        -22.83
      ]
    },
    "Elevation": 5916,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "36afbd16-8ec8-be83-d2ca-8b57313cadd2"
  },
  {
    "Volcano Name": "Lihir",
    "Country": "Papua New Guinea",
    "Region": "New Ireland-SW Pacif",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.642,
        -3.125
      ]
    },
    "Elevation": 700,
    "Type": "Volcanic complex",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8d8e2897-b2cf-1ec2-0541-c899fe4953d3"
  },
  {
    "Volcano Name": "Lindenberg Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -60.05,
        -65.03
      ]
    },
    "Elevation": 368,
    "Type": "Pyroclastic cone",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "9d9b0ae7-f521-564b-83a1-734ab2f8c6b1"
  },
  {
    "Volcano Name": "Lipari",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.95,
        38.48
      ]
    },
    "Elevation": 602,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "ffdd3a5b-0ebd-40f6-97e2-f2e833a42af9"
  },
  {
    "Volcano Name": "Lipari",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.95,
        38.483
      ]
    },
    "Elevation": 602,
    "Type": "Stratovolcanoes",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "984b0298-275a-a938-1007-e60cb9102bba"
  },
  {
    "Volcano Name": "Little Sitkin",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.53,
        51.95
      ]
    },
    "Elevation": 1188,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "42656642-8622-873b-2fe3-310aafc4e87a"
  },
  {
    "Volcano Name": "Ljosufjoll",
    "Country": "Iceland",
    "Region": "Iceland-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -22.23,
        64.87
      ]
    },
    "Elevation": 988,
    "Type": "Fissure vent",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "59391b18-b742-cdbe-e31c-4a34b5dbface"
  },
  {
    "Volcano Name": "Llaima",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.729,
        -38.692
      ]
    },
    "Elevation": 3125,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6ff85caf-49d9-0b69-a221-290219dd2406"
  },
  {
    "Volcano Name": "Llullaillaco",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.53,
        -24.72
      ]
    },
    "Elevation": 6739,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "99b44d85-b0f8-8ebb-99f8-e1fa5454f50c"
  },
  {
    "Volcano Name": "Loihi",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.27,
        18.92
      ]
    },
    "Elevation": -975,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f2e86e92-ae35-d3bc-2c9e-1620a9fcbf88"
  },
  {
    "Volcano Name": "Lokon-Empung",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.792,
        1.358
      ]
    },
    "Elevation": 1580,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c2409c07-382f-12c8-b12d-fcab8ac6a258"
  },
  {
    "Volcano Name": "Lolo",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.5,
        -5.47
      ]
    },
    "Elevation": 805,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "fe7baec3-fa7d-f9ae-bcd1-88b97eb77b3b"
  },
  {
    "Volcano Name": "Lolobau",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.158,
        -4.92
      ]
    },
    "Elevation": 858,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a40c48cc-73be-fbca-d6c8-b94b10492295"
  },
  {
    "Volcano Name": "Loloru",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.62,
        -6.52
      ]
    },
    "Elevation": 1887,
    "Type": "Pyroclastic shield",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fd5fb4cb-cf07-4ace-5b4e-6b24ada3890c"
  },
  {
    "Volcano Name": "Lomonosov Group",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.43,
        50.25
      ]
    },
    "Elevation": 1681,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5c237d80-b2c2-f94c-4a27-0baa07dc0e34"
  },
  {
    "Volcano Name": "Long Island",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.12,
        -5.358
      ]
    },
    "Elevation": 1280,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c2cb9f81-8211-4d0d-d0b1-d4814261b12c"
  },
  {
    "Volcano Name": "Longavi, Nevado de",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.161,
        -36.193
      ]
    },
    "Elevation": 3242,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6f4f7a28-fb75-2df4-4b35-5a000afd4f8a"
  },
  {
    "Volcano Name": "Longgang Group",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        126.5,
        42.33
      ]
    },
    "Elevation": 1000,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "7f60d890-68db-23f9-bba7-fa084968c062"
  },
  {
    "Volcano Name": "Longonot",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.45,
        -0.92
      ]
    },
    "Elevation": 2776,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "17aa69f4-b4b5-8d26-ce21-54e78fc0ed7b"
  },
  {
    "Volcano Name": "Lonquimay",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.58,
        -38.377
      ]
    },
    "Elevation": 2865,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "541e3b04-7f2e-39e3-be3a-b5640aea5d31"
  },
  {
    "Volcano Name": "Lopevi",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.346,
        -16.507
      ]
    },
    "Elevation": 1413,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3c2867dd-3896-6fff-cfd1-a19362fa1935"
  },
  {
    "Volcano Name": "Lower Chindwin",
    "Country": "Myanmar",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        95.1,
        22.28
      ]
    },
    "Elevation": 385,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "bf8a6d64-ba30-9e20-8347-76db3439289d"
  },
  {
    "Volcano Name": "Lubukraya",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.209,
        1.478
      ]
    },
    "Elevation": 1862,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "a8cd0435-8cca-0766-51f6-3b4d52f8c3ae"
  },
  {
    "Volcano Name": "Lumut Balai, Bukit",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.62,
        -4.22
      ]
    },
    "Elevation": 2055,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2a11cc51-e9bd-76e1-850f-d95c8d51a50d"
  },
  {
    "Volcano Name": "Lunayyir, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.75,
        25.17
      ]
    },
    "Elevation": 1370,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "5571b55c-4d0e-8868-9833-827050db6932"
  },
  {
    "Volcano Name": "Lurus",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        113.58,
        -7.7
      ]
    },
    "Elevation": 539,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "7b6879f2-e36d-4a97-e4c4-906da3a448b9"
  },
  {
    "Volcano Name": "Lvinaya Past",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147,
        44.62
      ]
    },
    "Elevation": 528,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fcfeb0ae-5a19-72cb-d5dd-e1fdb1d15c84"
  },
  {
    "Volcano Name": "Lysuholl",
    "Country": "Iceland",
    "Region": "Iceland-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -23.25,
        64.87
      ]
    },
    "Elevation": 540,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "23775ca8-2d85-fe20-47c8-c37614f132c3"
  },
  {
    "Volcano Name": "Ma Alalta",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.2,
        13.02
      ]
    },
    "Elevation": 1815,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f9a3a8ee-4e49-b4bf-cd3f-ac897e2e5b3d"
  },
  {
    "Volcano Name": "Maca",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.2,
        -45.1
      ]
    },
    "Elevation": 2960,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d01a0ccf-c58f-e102-a6e5-402affaca20f"
  },
  {
    "Volcano Name": "Macauley Island",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.47,
        -30.2
      ]
    },
    "Elevation": 238,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1811838e-66b9-4abf-1fe6-4ee387b4b96e"
  },
  {
    "Volcano Name": "Macdonald",
    "Country": "Antarctica",
    "Region": "Austral Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -140.25,
        -28.98
      ]
    },
    "Elevation": -50,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "60f0387e-e7c3-9fb2-2f56-ccba8509f20e"
  },
  {
    "Volcano Name": "Machin, Cerro",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.4,
        4.48
      ]
    },
    "Elevation": 2650,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "8a0e14e3-3bdb-a340-6015-147dff3bd80c"
  },
  {
    "Volcano Name": "Madeira",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.967,
        32.733
      ]
    },
    "Elevation": 1862,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "603f22a6-f553-7ba8-04ac-410b58620ecf"
  },
  {
    "Volcano Name": "Maderas",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.515,
        11.446
      ]
    },
    "Elevation": 1394,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4f3754b8-8a0f-5ac2-84b4-2fe6635cf0f6"
  },
  {
    "Volcano Name": "Madilogo",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.57,
        -9.2
      ]
    },
    "Elevation": 850,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5821a355-e6c9-ab13-ff5e-424caee18c54"
  },
  {
    "Volcano Name": "Magaso",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.175,
        9.258
      ]
    },
    "Elevation": 1904,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "50ea680a-1cdf-d124-210f-bf4f4224f754"
  },
  {
    "Volcano Name": "Mageik",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.25,
        58.2
      ]
    },
    "Elevation": 2165,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c744a7f3-841f-1cb3-7a96-28a18e1544be"
  },
  {
    "Volcano Name": "Mahagnoa",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.853,
        10.872
      ]
    },
    "Elevation": 800,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f4f02dda-5f32-23de-ad7a-6f86f6557b13"
  },
  {
    "Volcano Name": "Mahawu",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.858,
        1.358
      ]
    },
    "Elevation": 1324,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c5fe34b8-cf29-9d8d-fd38-386c55a27bad"
  },
  {
    "Volcano Name": "Maipo",
    "Country": "Argentina",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.833,
        -34.161
      ]
    },
    "Elevation": 5264,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2d644b1f-31c9-d346-a55e-b8605e8509dd"
  },
  {
    "Volcano Name": "Makaturing",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.32,
        7.647
      ]
    },
    "Elevation": 1940,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "403c1286-0fd9-567e-505c-3c21d080e976"
  },
  {
    "Volcano Name": "Makian",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.4,
        0.32
      ]
    },
    "Elevation": 1357,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "bb5ac646-1355-a302-35db-0f3da59d7763"
  },
  {
    "Volcano Name": "Makushin",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -166.93,
        53.9
      ]
    },
    "Elevation": 2036,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0451a6ff-3ebb-3276-0833-9b47a56e55ac"
  },
  {
    "Volcano Name": "Malabar",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.65,
        -7.13
      ]
    },
    "Elevation": 2343,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "9baefc2f-25c2-4942-307a-18843aa04ce2"
  },
  {
    "Volcano Name": "Malang Plain",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.68,
        -8.02
      ]
    },
    "Elevation": 680,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "42e8a7a1-619e-e91b-3df5-b2254baea6f8"
  },
  {
    "Volcano Name": "Malinao",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.597,
        13.422
      ]
    },
    "Elevation": 1548,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "U1",
    "id": "28d3f8f5-c0bd-4cb8-791b-122b99e4b68f"
  },
  {
    "Volcano Name": "Malinche, La",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -98.03,
        19.23
      ]
    },
    "Elevation": 4503,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "b895e97b-c19a-512d-5461-0d48a1da56b3"
  },
  {
    "Volcano Name": "Malindang",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.63,
        8.22
      ]
    },
    "Elevation": 2435,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "abb3f385-ac5d-ca82-0ca6-30af9455381d"
  },
  {
    "Volcano Name": "Malindig",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.018,
        13.24
      ]
    },
    "Elevation": 1157,
    "Type": "Stratovolcano",
    "Status": "Hot Springs",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0ea236f9-22c8-d2d0-66c2-9dfc5b04a03a"
  },
  {
    "Volcano Name": "Malintang",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.667,
        0.467
      ]
    },
    "Elevation": 1983,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4e447fa8-265d-02af-a70d-ad7bfc799cd8"
  },
  {
    "Volcano Name": "Mallahle",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.65,
        13.27
      ]
    },
    "Elevation": 1875,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e8f51479-1e02-7321-ea99-25f6a5e2d2d0"
  },
  {
    "Volcano Name": "Maly Payalpan",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.98,
        55.82
      ]
    },
    "Elevation": 1802,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5873709b-a183-ca30-f18c-1c35581b618c"
  },
  {
    "Volcano Name": "Maly Semiachik",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.67,
        54.13
      ]
    },
    "Elevation": 1560,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a3ae6ea7-38be-4cf5-8e63-af7a2e368d5a"
  },
  {
    "Volcano Name": "Managlase Plateau",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.33,
        -9.08
      ]
    },
    "Elevation": 1342,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f2e8f4d7-c488-59b2-92f7-d900ad4f21dd"
  },
  {
    "Volcano Name": "Manam",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.061,
        -4.1
      ]
    },
    "Elevation": 1807,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1d237e60-ad95-c8aa-1c75-24b5c2f7d583"
  },
  {
    "Volcano Name": "Manareyjar",
    "Country": "Iceland",
    "Region": "Iceland-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.1,
        66.3
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "03c5fd0a-7ad5-f346-ce7e-46ed9bf2996e"
  },
  {
    "Volcano Name": "Manda Gargori",
    "Country": "Ethiopia",
    "Region": "Ethiopia",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.483,
        11.75
      ]
    },
    "Elevation": null,
    "Type": "Fissure vents",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2c8a6665-0570-d0d7-26b2-88066e9703ba"
  },
  {
    "Volcano Name": "Manda Hararo",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.82,
        12.17
      ]
    },
    "Elevation": 600,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "878704af-a851-292e-3bec-f1a87eb90282"
  },
  {
    "Volcano Name": "Manda-Inakir",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.2,
        12.38
      ]
    },
    "Elevation": 600,
    "Type": "Fissure vent",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "17a08755-ff18-455d-de76-6f3690a05205"
  },
  {
    "Volcano Name": "Mandalagan",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.22,
        10.615
      ]
    },
    "Elevation": 1879,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e7865bb9-6639-eb8f-dd07-feadfea49ae6"
  },
  {
    "Volcano Name": "Manengouba",
    "Country": "Cameroon",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        9.83,
        5.03
      ]
    },
    "Elevation": 2411,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "90e05732-69da-15c6-ce08-97bf767586c1"
  },
  {
    "Volcano Name": "Manuk",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.292,
        -5.53
      ]
    },
    "Elevation": 282,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "33262bd6-d6d9-f95e-ba22-31c6022b7e58"
  },
  {
    "Volcano Name": "Manzaz Volc Field",
    "Country": "Algeria",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        5.833,
        23.917
      ]
    },
    "Elevation": 1672,
    "Type": "Scoria cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "49b0f172-9783-b66a-afce-83178f061451"
  },
  {
    "Volcano Name": "Maquiling",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.2,
        14.13
      ]
    },
    "Elevation": 1090,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a76805cf-1e31-2145-74cb-c2fd6d745dd6"
  },
  {
    "Volcano Name": "Marapi",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        100.473,
        -0.381
      ]
    },
    "Elevation": 2891,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c194b4ae-b044-3e08-03b8-2c173fdcac0e"
  },
  {
    "Volcano Name": "Marchena",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.47,
        0.33
      ]
    },
    "Elevation": 343,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "987ae518-ef22-a420-95bf-09fac2fd770a"
  },
  {
    "Volcano Name": "Mare",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.4,
        0.57
      ]
    },
    "Elevation": 308,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "490bb302-ef11-9b8e-7065-a223178cffdd"
  },
  {
    "Volcano Name": "Marha, Jabal el-",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.22,
        15.28
      ]
    },
    "Elevation": 2650,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f8a7828a-3c73-76cc-51bb-fec6eec133a5"
  },
  {
    "Volcano Name": "Marion Island",
    "Country": "South Africa",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.75,
        -46.9
      ]
    },
    "Elevation": 1230,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4461b2eb-7f38-38a6-f53e-345abb89556a"
  },
  {
    "Volcano Name": "Mariveles",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.5,
        14.5
      ]
    },
    "Elevation": 1420,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f5fa746c-53a7-fe0b-7efb-0d3566ffaba3"
  },
  {
    "Volcano Name": "Markagunt Plateau",
    "Country": "United States",
    "Region": "US-Utah",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.67,
        37.58
      ]
    },
    "Elevation": 2840,
    "Type": "Volcanic field",
    "Status": "Dendrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "7b876889-6a6c-a60b-710c-71cd34a29e3a"
  },
  {
    "Volcano Name": "Maroa",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        176.08,
        -38.42
      ]
    },
    "Elevation": 1156,
    "Type": "Caldera",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "0bacd46f-7075-9acd-30dd-319c0e13758c"
  },
  {
    "Volcano Name": "Marra, Jebel",
    "Country": "Sudan",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        24.27,
        12.95
      ]
    },
    "Elevation": 3042,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "97dd5687-8aed-4691-18af-1df21d8fac2a"
  },
  {
    "Volcano Name": "Marsabit",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.97,
        2.32
      ]
    },
    "Elevation": 1707,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "cc8359cc-b63a-a233-411a-02b638978fa5"
  },
  {
    "Volcano Name": "Martin",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.35,
        58.17
      ]
    },
    "Elevation": 1860,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b7128cb1-59d5-ae8e-01ab-6fd3139d9716"
  },
  {
    "Volcano Name": "Masaraga",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.6,
        13.32
      ]
    },
    "Elevation": 1328,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6845e027-c190-8741-e51c-7afc22ad6453"
  },
  {
    "Volcano Name": "Masaya",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.161,
        11.984
      ]
    },
    "Elevation": 635,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "befca90d-bc02-e969-3c3b-d976fb498dd3"
  },
  {
    "Volcano Name": "Mascota Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.83,
        20.62
      ]
    },
    "Elevation": 2540,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a2c1c034-b629-1913-f8e9-f661f64e03a6"
  },
  {
    "Volcano Name": "Mashkovtsev",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.72,
        51.1
      ]
    },
    "Elevation": 503,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cf9f4f1f-c66e-36be-d47b-cc393be89872"
  },
  {
    "Volcano Name": "Mashu",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.57,
        43.57
      ]
    },
    "Elevation": 855,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "b18a7ad5-beed-1661-86ff-902b8d0ecd73"
  },
  {
    "Volcano Name": "Mat Ala",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.15,
        13.1
      ]
    },
    "Elevation": 523,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "54f77376-098f-ff62-6413-797b0bb09fb3"
  },
  {
    "Volcano Name": "Matthew Island",
    "Country": "Pacific Ocean",
    "Region": "SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        171.32,
        -22.33
      ]
    },
    "Elevation": 177,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e2135da4-6050-ce1f-183b-9720542b4b3f"
  },
  {
    "Volcano Name": "Matutum",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.108,
        6.37
      ]
    },
    "Elevation": 2293,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5156e234-dbcc-87f5-c583-e7d622b94846"
  },
  {
    "Volcano Name": "Maug Islands",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.217,
        20.017
      ]
    },
    "Elevation": 227,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0d76278d-475f-e530-5bb7-591e88458ed8"
  },
  {
    "Volcano Name": "Maule, Laguna del",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.58,
        -36.02
      ]
    },
    "Elevation": 3092,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c18cc33f-f6d9-3f10-e58a-cf9af090fc7f"
  },
  {
    "Volcano Name": "Mauna Kea",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.47,
        19.82
      ]
    },
    "Elevation": 4206,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "d5722baa-e429-96fc-eef4-8415926b64d3"
  },
  {
    "Volcano Name": "Mauna Loa",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.608,
        19.475
      ]
    },
    "Elevation": 4170,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "af716693-dc25-4b11-da74-46283c877969"
  },
  {
    "Volcano Name": "May-ya-moto",
    "Country": "Congo, DRC",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.33,
        -0.93
      ]
    },
    "Elevation": 950,
    "Type": "Fumarole field",
    "Status": "Fumarolic",
    "Last Known Eruption": "Unknown",
    "id": "2b34ae4b-c3fd-578c-ba43-d63760e3f7e3"
  },
  {
    "Volcano Name": "Mayon",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.685,
        13.257
      ]
    },
    "Elevation": 2462,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1591acf2-c84f-9e22-9eb1-5a4df75b45a6"
  },
  {
    "Volcano Name": "Mayor Island",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        176.25,
        -37.28
      ]
    },
    "Elevation": 355,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "c381a003-285f-c8f3-6f60-c38940574403"
  },
  {
    "Volcano Name": "McDonald Islands",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        72.6,
        -53.03
      ]
    },
    "Elevation": 186,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0e25654d-0364-c1c7-f945-f7eb80242d42"
  },
  {
    "Volcano Name": "Meager",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -123.5,
        50.63
      ]
    },
    "Elevation": 2680,
    "Type": "Complex volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "5233a74e-e798-981b-6b8d-e3c78a3c0765"
  },
  {
    "Volcano Name": "Medicine Lake",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.57,
        41.58
      ]
    },
    "Elevation": 2412,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "4ca1a7af-549a-e736-9fea-aaf355682c9b"
  },
  {
    "Volcano Name": "Medvezhia",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.83,
        45.38
      ]
    },
    "Elevation": 1124,
    "Type": "Somma volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "348fda84-c42b-988d-0a1c-4bba9b30c739"
  },
  {
    "Volcano Name": "Mega Basalt Field",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.42,
        4.08
      ]
    },
    "Elevation": 1067,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "35494a46-1247-b902-d3b5-0f52d7bf6e49"
  },
  {
    "Volcano Name": "Megata",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.73,
        39.95
      ]
    },
    "Elevation": 291,
    "Type": "Maar",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "6ec4bccb-c616-8a7d-29fb-e79000d2ceb7"
  },
  {
    "Volcano Name": "Mehetia",
    "Country": "France",
    "Region": "Society Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -148.07,
        -17.87
      ]
    },
    "Elevation": 435,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0d5f4370-1ca0-1227-9d58-fceb8be47c2b"
  },
  {
    "Volcano Name": "Meidob Volc Field",
    "Country": "Sudan",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        26.47,
        15.32
      ]
    },
    "Elevation": 2000,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "8f633214-8756-d886-28dc-175d3e8fb4ed"
  },
  {
    "Volcano Name": "Melbourne",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        164.7,
        -74.35
      ]
    },
    "Elevation": 2732,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "30be2d53-851f-66a2-38be-3b2fccb68533"
  },
  {
    "Volcano Name": "Melimoyu",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.88,
        -44.08
      ]
    },
    "Elevation": 2400,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b39b67f8-cb37-d40e-eec2-c630512bdcb5"
  },
  {
    "Volcano Name": "Mendeleev",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.7,
        43.98
      ]
    },
    "Elevation": 887,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ebf253f8-0772-11de-853b-805a72c0002c"
  },
  {
    "Volcano Name": "Menengai",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.07,
        -0.2
      ]
    },
    "Elevation": 2278,
    "Type": "Shield volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "48a09b72-c0f8-6a3a-a987-361f42a9b972"
  },
  {
    "Volcano Name": "Mentolat",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.08,
        -44.67
      ]
    },
    "Elevation": 1660,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "5c25c1f7-2cc0-aeed-b210-6e4b8dd3e9e8"
  },
  {
    "Volcano Name": "Merapi",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.442,
        -7.542
      ]
    },
    "Elevation": 2947,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2d6b2605-bfce-f500-d40f-0dc4950b328b"
  },
  {
    "Volcano Name": "Merbabu",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.43,
        -7.45
      ]
    },
    "Elevation": 3145,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "43e9f2e1-d94b-b032-06d9-24632ff22af9"
  },
  {
    "Volcano Name": "Mere Lava",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.05,
        -14.45
      ]
    },
    "Elevation": 1028,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "14716eac-f498-8e18-2798-fac35e294e55"
  },
  {
    "Volcano Name": "Meru",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.75,
        -3.25
      ]
    },
    "Elevation": 4565,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "482fb989-c411-c6ce-fac4-c44c393d1aac"
  },
  {
    "Volcano Name": "Methana",
    "Country": "Greece",
    "Region": "Greece",
    "Location": {
      "type": "Point",
      "coordinates": [
        23.336,
        37.615
      ]
    },
    "Elevation": 760,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "3cfd7753-c84a-3191-598e-e63916eb200d"
  },
  {
    "Volcano Name": "Metis Shoal",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.87,
        -19.18
      ]
    },
    "Elevation": 43,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3c374e6f-c682-0878-e662-c411d0e7d0b5"
  },
  {
    "Volcano Name": "Mezhdusopochny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.25,
        57.47
      ]
    },
    "Elevation": 1641,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "87ac627e-4d3c-b021-047e-9982a8a6e2d2"
  },
  {
    "Volcano Name": "Michael",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -26.45,
        -57.78
      ]
    },
    "Elevation": 990,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "da58141e-5d6e-9bb7-669a-00f08038c6b6"
  },
  {
    "Volcano Name": "Michoacan-Guanajuato",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -102.25,
        19.48
      ]
    },
    "Elevation": 3860,
    "Type": "Cinder cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "f959c838-777e-94a9-d9b6-e8fa07c38a61"
  },
  {
    "Volcano Name": "Micotrin",
    "Country": "Dominica",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.33,
        15.33
      ]
    },
    "Elevation": 1387,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ca19bbcf-c1c9-08b1-9e2a-ee6e0c31bc5d"
  },
  {
    "Volcano Name": "Middle Gobi",
    "Country": "Mongolia",
    "Region": "Mongolia",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.7,
        45.28
      ]
    },
    "Elevation": 1120,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "4284e40d-d759-bf86-2ded-7b9ea1deec5f"
  },
  {
    "Volcano Name": "Milbanke Sound Group",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -128.73,
        52.5
      ]
    },
    "Elevation": 335,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "56ecd09d-f3e2-78e0-787c-fda40598c8cb"
  },
  {
    "Volcano Name": "Milne",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.78,
        46.82
      ]
    },
    "Elevation": 1540,
    "Type": "Somma volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f59d6b02-97c5-cf08-030c-8120b03d8051"
  },
  {
    "Volcano Name": "Milos",
    "Country": "Greece",
    "Region": "Greece",
    "Location": {
      "type": "Point",
      "coordinates": [
        24.439,
        36.699
      ]
    },
    "Elevation": 751,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9b4d04c6-50ee-9d5e-bc65-0c2976f42c63"
  },
  {
    "Volcano Name": "Minami-Hiyoshi",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.905,
        23.507
      ]
    },
    "Elevation": -30,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ed7eb64f-d4ad-a595-7e29-cb21a1806273"
  },
  {
    "Volcano Name": "Minchinmavida",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.43,
        -42.78
      ]
    },
    "Elevation": 2404,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b9c7aac7-9f08-cb59-2953-992652f86841"
  },
  {
    "Volcano Name": "Miravalles",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.153,
        10.748
      ]
    },
    "Elevation": 2028,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "0dc27daf-6d61-988c-1351-3eef1898ad44"
  },
  {
    "Volcano Name": "Misti, El",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.409,
        -16.294
      ]
    },
    "Elevation": 5822,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "89740c7b-7a27-acc6-95c9-aab4a0d3d8a8"
  },
  {
    "Volcano Name": "Miyake-jima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.53,
        34.08
      ]
    },
    "Elevation": 815,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "34a4906e-d1bc-f57e-e06d-21cfb7094cb0"
  },
  {
    "Volcano Name": "Mocho-Choshuenco",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.027,
        -39.927
      ]
    },
    "Elevation": 2422,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "541e84f6-8207-da86-db9f-fc06b19328a1"
  },
  {
    "Volcano Name": "Moffett",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -176.75,
        51.93
      ]
    },
    "Elevation": 1196,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "689079bb-21c1-d64e-fcc6-b62107919ab5"
  },
  {
    "Volcano Name": "Mojanda",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.27,
        0.13
      ]
    },
    "Elevation": 4263,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "18e43fb5-b077-33b1-fbe1-a5cb48a80172"
  },
  {
    "Volcano Name": "Mojanda",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.267,
        0.133
      ]
    },
    "Elevation": 4263,
    "Type": "Stratovolcanoes",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "3579ca8d-97db-5fd9-10fa-2b0c4ccb0963"
  },
  {
    "Volcano Name": "Mokuyo Seamount",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.567,
        28.317
      ]
    },
    "Elevation": -920,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ffa29f5d-b84c-a59b-76b4-dd46b4f639e5"
  },
  {
    "Volcano Name": "Mombacho",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.968,
        11.826
      ]
    },
    "Elevation": 1344,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "22eafc6c-8d72-bfa7-7b8c-ab583d8566fd"
  },
  {
    "Volcano Name": "Momotombo",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.54,
        12.422
      ]
    },
    "Elevation": 1297,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "73a6f464-439d-8da0-cce7-2b802bfc16f7"
  },
  {
    "Volcano Name": "Monaco Bank",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.88,
        37.6
      ]
    },
    "Elevation": -197,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d0c54df6-55b9-f5cc-9eae-1d45ea291908"
  },
  {
    "Volcano Name": "Mondaca",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.8,
        -35.464
      ]
    },
    "Elevation": 2048,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3a2a9608-79c5-d74a-80f9-9be50650abd6"
  },
  {
    "Volcano Name": "Mono Craters",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -119,
        37.88
      ]
    },
    "Elevation": 2796,
    "Type": "Lava dome",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "7bd41501-44ef-d7a7-c01a-f8975a1464ef"
  },
  {
    "Volcano Name": "Mono Lake Volc Field",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -119.03,
        38
      ]
    },
    "Elevation": 2121,
    "Type": "Cinder cone",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "ab8d2eee-0653-a49e-92d2-48db32bf1b98"
  },
  {
    "Volcano Name": "Monowai Seamount",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.188,
        -25.887
      ]
    },
    "Elevation": -100,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b9db6567-da2f-1eb3-8dde-d0c97e92e310"
  },
  {
    "Volcano Name": "Montagu Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -26.333,
        -58.417
      ]
    },
    "Elevation": 1370,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "cbebecfd-959c-430d-a999-41fa2f71c9e4"
  },
  {
    "Volcano Name": "Morning, Mt.",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        163.533,
        -78.5
      ]
    },
    "Elevation": 2723,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7f15c357-0955-1d38-77fd-9a7dbfb1733e"
  },
  {
    "Volcano Name": "Moti",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.4,
        0.45
      ]
    },
    "Elevation": 950,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "12a3e407-ea6a-7875-5821-57466687d7fa"
  },
  {
    "Volcano Name": "Motlav",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.67,
        -13.67
      ]
    },
    "Elevation": 411,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e968e28c-bed5-c83a-066f-a58c0fb1e7a1"
  },
  {
    "Volcano Name": "Moua Pihaa",
    "Country": "France",
    "Region": "Society Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -148.67,
        -18.32
      ]
    },
    "Elevation": -180,
    "Type": "Submarine volcano",
    "Status": "Seismicity",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "13c58ff3-470d-e6a3-796c-c295d13b5a1b"
  },
  {
    "Volcano Name": "Mousa Alli",
    "Country": "Djibouti",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.4,
        12.47
      ]
    },
    "Elevation": 2028,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9585dcf4-ea35-2ffd-4aa4-68b77a53f479"
  },
  {
    "Volcano Name": "Moyuta",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.1,
        14.03
      ]
    },
    "Elevation": 1662,
    "Type": "Stratovolcano",
    "Status": "Hot Springs",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "77157c02-a50a-6730-888a-f9cac5e445b2"
  },
  {
    "Volcano Name": "Muhavura",
    "Country": "Uganda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.67,
        -1.38
      ]
    },
    "Elevation": 4127,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3a740987-eb49-3412-b8b0-5267291ec829"
  },
  {
    "Volcano Name": "Mundafell",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.7,
        63.98
      ]
    },
    "Elevation": 1491,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "767a21dc-20a0-1565-941a-adfe88e1500e"
  },
  {
    "Volcano Name": "Mundua",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.35,
        -4.63
      ]
    },
    "Elevation": 179,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "50cd07e7-d293-d628-20b1-7a7ebb89052d"
  },
  {
    "Volcano Name": "Muria",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.883,
        -6.617
      ]
    },
    "Elevation": 1625,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "1d2b1b61-0b23-d2ca-00dc-1ff37e3fd2c7"
  },
  {
    "Volcano Name": "Musa River",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.13,
        -9.308
      ]
    },
    "Elevation": 808,
    "Type": "Hydrothermal field",
    "Status": "Hot Springs",
    "Last Known Eruption": "Unknown",
    "id": "0baf524c-af36-897f-37cc-63f69e704538"
  },
  {
    "Volcano Name": "Mutnovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.195,
        52.453
      ]
    },
    "Elevation": 2322,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3516105d-4b2f-5f13-963e-7a456c3538fc"
  },
  {
    "Volcano Name": "Myojun Knoll",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.85,
        32.1
      ]
    },
    "Elevation": 360,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "3069023b-78c5-cbe4-6fad-d6063b533685"
  },
  {
    "Volcano Name": "Myoko",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.12,
        36.88
      ]
    },
    "Elevation": 2446,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "523d7a44-b099-a90a-705e-fbdab4991ee2"
  },
  {
    "Volcano Name": "NW Rota-1",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.775,
        14.601
      ]
    },
    "Elevation": -517,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8fb4cb51-1f03-a3a8-aeea-ab9b69982207"
  },
  {
    "Volcano Name": "Nabro",
    "Country": "Eritrea",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.7,
        13.37
      ]
    },
    "Elevation": 2218,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "65965f14-3c7f-4cf3-f4de-4bc3dd1cb9b9"
  },
  {
    "Volcano Name": "Nabukelevu",
    "Country": "Tonga",
    "Region": "Fiji Is-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        177.983,
        -19.117
      ]
    },
    "Elevation": 805,
    "Type": "Lava domes",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "fed86c36-00b5-c209-7a02-8f119799e4c9"
  },
  {
    "Volcano Name": "Nakano-shima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.87,
        29.85
      ]
    },
    "Elevation": 979,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "47ec2c28-4ba9-80d9-a97f-e43b9c2d3e3b"
  },
  {
    "Volcano Name": "Namarunu",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.27,
        1.9
      ]
    },
    "Elevation": 817,
    "Type": "Shield volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "3ce8c4fd-e440-b328-2d0d-a17c18665800"
  },
  {
    "Volcano Name": "Nantai",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.5,
        36.77
      ]
    },
    "Elevation": 2484,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0383aba4-ffb7-ddd4-7004-3dda29dae8b8"
  },
  {
    "Volcano Name": "Naolinco Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -96.75,
        19.67
      ]
    },
    "Elevation": 2000,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "80689c46-9bce-3d09-317f-a7bd5876cbdd"
  },
  {
    "Volcano Name": "Narcondum",
    "Country": "India",
    "Region": "Andaman Is-Indian O",
    "Location": {
      "type": "Point",
      "coordinates": [
        94.25,
        13.43
      ]
    },
    "Elevation": 710,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2be02ada-5286-f920-3df4-d1229d9f77de"
  },
  {
    "Volcano Name": "Narugo",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.73,
        38.73
      ]
    },
    "Elevation": 462,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "7cf9210b-a979-a0df-68a4-56a244b6eaf7"
  },
  {
    "Volcano Name": "Nasu",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.97,
        37.12
      ]
    },
    "Elevation": 1917,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e4e43986-d84c-4cfb-a178-f7c2f6064312"
  },
  {
    "Volcano Name": "Natib",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.4,
        14.705
      ]
    },
    "Elevation": 1287,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "d84bf915-b7b0-1b22-8377-4d04a6eb57a4"
  },
  {
    "Volcano Name": "Nazko",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -123.73,
        52.9
      ]
    },
    "Elevation": 1230,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "9d156ac4-5dfc-1ad5-f0dd-bc79f2aab5f7"
  },
  {
    "Volcano Name": "Ndete Napu",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.78,
        -8.72
      ]
    },
    "Elevation": 750,
    "Type": "Fumarole field",
    "Status": "Fumarolic",
    "Last Known Eruption": "Unknown",
    "id": "3bdf0e29-51d4-869c-fbca-2dfb19a1d222"
  },
  {
    "Volcano Name": "Negra, Sierra",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.17,
        -0.83
      ]
    },
    "Elevation": 1490,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d7a0040b-a592-dcf4-0bb3-2aff96a6837a"
  },
  {
    "Volcano Name": "Negrillar, El",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.25,
        -24.18
      ]
    },
    "Elevation": 3500,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "f3a5e478-cdcd-987d-aba0-b0cd974e92ed"
  },
  {
    "Volcano Name": "Negrillar, La",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.6,
        -24.28
      ]
    },
    "Elevation": 4109,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "bc6c13a1-b28a-a8b0-886f-d28d27c4e404"
  },
  {
    "Volcano Name": "Negro de Mayasquer, Cerro",
    "Country": "Ecuador",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.964,
        0.828
      ]
    },
    "Elevation": 4445,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ed0d2695-f3e4-34c6-07e6-42ebdc5e6f0e"
  },
  {
    "Volcano Name": "Negro, Cerro",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.702,
        12.506
      ]
    },
    "Elevation": 728,
    "Type": "Cinder cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5e29b39a-cc8b-3a36-5afb-01e9d402dd2b"
  },
  {
    "Volcano Name": "Nejapa-Miraflores",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.32,
        12.12
      ]
    },
    "Elevation": 360,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8817bb82-ef8d-dc2b-8efc-13f8149e82c0"
  },
  {
    "Volcano Name": "Nemo Peak",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.808,
        49.57
      ]
    },
    "Elevation": 1018,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "459227ab-83d7-a3b9-d08a-f75cc5fbf749"
  },
  {
    "Volcano Name": "Nemrut Dagi",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.02,
        38.65
      ]
    },
    "Elevation": 3050,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "8e8b79b3-ae1b-ee3b-6d55-d77b109e1167"
  },
  {
    "Volcano Name": "Nevada, Sierra",
    "Country": "Argentina",
    "Region": "Chile",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.58,
        -26.48
      ]
    },
    "Elevation": 6127,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0b3a4fec-4528-e0a3-657a-8f72b3355427"
  },
  {
    "Volcano Name": "Nevis Peak",
    "Country": "Netherlands",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -62.58,
        17.15
      ]
    },
    "Elevation": 985,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d85784d9-37c0-2cbd-2cea-bb9f746e1cb8"
  },
  {
    "Volcano Name": "Newberry Volcano",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.229,
        43.722
      ]
    },
    "Elevation": 2434,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "3c4f8199-b148-25f1-d477-7957a6576b1a"
  },
  {
    "Volcano Name": "Newer Volcanics Prov",
    "Country": "Australia",
    "Region": "Australia",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.5,
        -37.77
      ]
    },
    "Elevation": 1011,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f767ec88-96ae-beb6-2d26-b946ffd9b390"
  },
  {
    "Volcano Name": "Ngaoundere Plateau",
    "Country": "Cameroon",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        13.67,
        7.25
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "43270dc5-d3d5-1375-141d-b4c67432e427"
  },
  {
    "Volcano Name": "Ngauruhoe",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        175.63,
        -39.158
      ]
    },
    "Elevation": 2291,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "08a0f8ac-106d-3e55-8b6b-db5bb963b237"
  },
  {
    "Volcano Name": "Ngozi",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.57,
        -8.97
      ]
    },
    "Elevation": 2622,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d156a67f-b12b-f969-7423-5ff3c2574b6e"
  },
  {
    "Volcano Name": "Nicholson, Cerro",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.753,
        -16.258
      ]
    },
    "Elevation": 2520,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dd0cdc7b-e958-2ee1-ad5b-39b56d08e4f7"
  },
  {
    "Volcano Name": "Nieuwerkerk",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.675,
        -6.6
      ]
    },
    "Elevation": -2285,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "ac23d84c-e6f0-3d8f-e52d-dfac8ab92821"
  },
  {
    "Volcano Name": "Nightingale Island",
    "Country": "United Kingdom",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -12.483,
        -37.417
      ]
    },
    "Elevation": 365,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "444fb125-2d2d-6545-01ce-be017d0a3d1a"
  },
  {
    "Volcano Name": "Nii-jima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.27,
        34.37
      ]
    },
    "Elevation": 432,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "6ab2802c-9947-64bf-765b-7a9a43687951"
  },
  {
    "Volcano Name": "Niigata-Yake-yama",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.03,
        36.92
      ]
    },
    "Elevation": 2400,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "50a500b2-67bf-f5c8-e6a3-f04aa2c63199"
  },
  {
    "Volcano Name": "Nikko",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.308,
        23.075
      ]
    },
    "Elevation": -391,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "8f5bb76c-e632-cb9c-67fc-5f5f1c88d8b1"
  },
  {
    "Volcano Name": "Nikko-Shirane",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.38,
        36.8
      ]
    },
    "Elevation": 2578,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "d2a3c9d8-ccb9-c458-637b-eef41c7322ff"
  },
  {
    "Volcano Name": "Nila",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.5,
        -6.73
      ]
    },
    "Elevation": 781,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ccb57821-54ad-7fb1-66fb-ff3975cfda34"
  },
  {
    "Volcano Name": "Nipesotsu-Upepesanke",
    "Country": "Japan",
    "Region": "Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.03,
        43.45
      ]
    },
    "Elevation": 2013,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "330ac867-5379-c49c-bd5e-915e80006109"
  },
  {
    "Volcano Name": "Niseko",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.63,
        42.88
      ]
    },
    "Elevation": 1154,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "417c8c5e-d3d5-b5e2-35a0-64f2e2a22c4e"
  },
  {
    "Volcano Name": "Nishino-shima",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.882,
        27.274
      ]
    },
    "Elevation": 38,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4a21058b-4d04-0d68-a5dc-09c4dfb7ab94"
  },
  {
    "Volcano Name": "Nisyros",
    "Country": "Greece",
    "Region": "Greece",
    "Location": {
      "type": "Point",
      "coordinates": [
        27.18,
        36.58
      ]
    },
    "Elevation": 698,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "2ffc0c6e-2eb4-355f-6395-8a8fdacc19b8"
  },
  {
    "Volcano Name": "Niuafo'ou",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.63,
        -15.6
      ]
    },
    "Elevation": 260,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "136fe63a-4b66-62f6-bb79-242f3f0e5151"
  },
  {
    "Volcano Name": "Norikura",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        137.55,
        36.12
      ]
    },
    "Elevation": 3026,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "5215b23e-2934-5f40-165f-58f46de9b8dc"
  },
  {
    "Volcano Name": "North Gorda Ridge",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -126.783,
        42.667
      ]
    },
    "Elevation": -3000,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "7d316a3b-bcd2-602d-6ab2-09b33466b193"
  },
  {
    "Volcano Name": "North Gorda Ridge",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -126.78,
        42.67
      ]
    },
    "Elevation": -3000,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5b2a1a94-172e-66ba-1017-280f1a9b28a6"
  },
  {
    "Volcano Name": "North Island",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.05,
        4.07
      ]
    },
    "Elevation": 520,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3e83a3de-1144-dee0-71ad-6cf3541a4c12"
  },
  {
    "Volcano Name": "North Sister Field",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.77,
        44.17
      ]
    },
    "Elevation": 3074,
    "Type": "Complex volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "2983c045-f01b-174d-1cc2-110fcd26cc38"
  },
  {
    "Volcano Name": "North Vate",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.33,
        -17.45
      ]
    },
    "Elevation": 594,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6d3dec64-f66f-bd2c-58c6-ca4f47d500b9"
  },
  {
    "Volcano Name": "Northern EPR-Segment RO3",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -105.433,
        15.833
      ]
    },
    "Elevation": -2300,
    "Type": "Submarine volcano",
    "Status": "Magnetism",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ac7ac4c4-8015-a4f5-f298-4cbe2043bc07"
  },
  {
    "Volcano Name": "Nosy-Be",
    "Country": "Madagascar",
    "Region": "Madagascar",
    "Location": {
      "type": "Point",
      "coordinates": [
        48.48,
        -13.32
      ]
    },
    "Elevation": 214,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "aad3f19a-dd1f-0cbc-5a03-c1e518cba148"
  },
  {
    "Volcano Name": "Novarupta",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.16,
        58.27
      ]
    },
    "Elevation": 841,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "17d2c5d2-e886-18fb-25e1-d1e6ecf1d60d"
  },
  {
    "Volcano Name": "Nuevo Mundo",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -66.483,
        -19.783
      ]
    },
    "Elevation": 5438,
    "Type": "Lava domes",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "78977bbe-c489-62c7-e113-78735501e282"
  },
  {
    "Volcano Name": "Numazawa",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.58,
        37.43
      ]
    },
    "Elevation": 1100,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ee207b9f-becc-7354-d4e9-ac611c157039"
  },
  {
    "Volcano Name": "Nunivak Island",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -166.33,
        60.02
      ]
    },
    "Elevation": 511,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "530626db-1fce-0912-b781-6706d0c80c72"
  },
  {
    "Volcano Name": "Nyambeni Hills",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.87,
        0.23
      ]
    },
    "Elevation": 750,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "88f0bc1c-3d04-6504-7a6b-67874568138f"
  },
  {
    "Volcano Name": "Nyamuragira",
    "Country": "Congo, DRC",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.2,
        -1.408
      ]
    },
    "Elevation": 3058,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3bc0272e-8095-7f5e-879e-50ab7dc097a7"
  },
  {
    "Volcano Name": "Nyiragongo",
    "Country": "Congo, DRC",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.25,
        -1.52
      ]
    },
    "Elevation": 3470,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5fe5e648-f81f-2d4d-82d6-a262121c24ac"
  },
  {
    "Volcano Name": "Ofu-Olosega",
    "Country": "United States",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.618,
        -14.175
      ]
    },
    "Elevation": 639,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ba703c62-9f65-293d-c0cb-085fa7263198"
  },
  {
    "Volcano Name": "Ojos del Salado, Nevados",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.53,
        -27.12
      ]
    },
    "Elevation": 6887,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f0694974-a731-dece-51dc-9d1b3acafbc0"
  },
  {
    "Volcano Name": "Oka Volc Field",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.98,
        52.7
      ]
    },
    "Elevation": 2077,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7e411d04-b2ac-91fe-4fa0-eeb1bd6abac2"
  },
  {
    "Volcano Name": "Okataina",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        176.5,
        -38.12
      ]
    },
    "Elevation": 1111,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "bf8d0c91-c8d3-ed46-1e03-dedcbb316c34"
  },
  {
    "Volcano Name": "Oki-Dogo",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        133.33,
        36.17
      ]
    },
    "Elevation": 151,
    "Type": "Shield volcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "34ff9833-8827-a67f-76d3-15cff30a5c3a"
  },
  {
    "Volcano Name": "Okmok",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -168.13,
        53.42
      ]
    },
    "Elevation": 1073,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "cfaafa62-955f-5045-2ac4-2cb38f0a927a"
  },
  {
    "Volcano Name": "Oku Volc Field",
    "Country": "Cameroon",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        10.5,
        6.25
      ]
    },
    "Elevation": 3011,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "4b9549de-5f4d-e391-7987-cddc14340d7d"
  },
  {
    "Volcano Name": "Ol Kokwe",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.08,
        0.63
      ]
    },
    "Elevation": 1130,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6e4337dc-aab5-c2e9-7c2f-537e7786d4da"
  },
  {
    "Volcano Name": "Olca-Paruma",
    "Country": "Bolivia",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.48,
        -20.93
      ]
    },
    "Elevation": 5407,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "c00f42d8-3e32-d491-b289-991b72c7b33f"
  },
  {
    "Volcano Name": "Olkaria",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.292,
        -0.904
      ]
    },
    "Elevation": 2434,
    "Type": "Pumice cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "6cd10847-e3ac-20e1-8397-c41ed17e6d26"
  },
  {
    "Volcano Name": "Olkoviy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.477,
        52.077
      ]
    },
    "Elevation": 636,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f52be475-5d89-312f-e9d7-bbb383f4a040"
  },
  {
    "Volcano Name": "Ollague",
    "Country": "Bolivia",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.18,
        -21.3
      ]
    },
    "Elevation": 5868,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2179e1dd-1e1d-9dee-2574-87c846d629c8"
  },
  {
    "Volcano Name": "Olot Volc Field",
    "Country": "Spain",
    "Region": "Spain",
    "Location": {
      "type": "Point",
      "coordinates": [
        2.53,
        42.17
      ]
    },
    "Elevation": 893,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "69f9d955-94bc-6773-9181-6b6603a24f1d"
  },
  {
    "Volcano Name": "Omachi Seamount",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.8,
        29.22
      ]
    },
    "Elevation": -1700,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d0b24eb7-b1a9-8a47-f3b1-7c77c1e7b23e"
  },
  {
    "Volcano Name": "Omanago Group",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.5,
        36.78
      ]
    },
    "Elevation": 2367,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "56df6714-aad5-afc8-5d68-bda44f2e2fed"
  },
  {
    "Volcano Name": "On-take",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        137.48,
        35.9
      ]
    },
    "Elevation": 3063,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "036361b6-3161-28ca-291c-4e9a0b50d1bb"
  },
  {
    "Volcano Name": "Opala",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.335,
        52.543
      ]
    },
    "Elevation": 2475,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "5f39c182-c271-83a7-64ea-594ed28b7696"
  },
  {
    "Volcano Name": "Oraefajokull",
    "Country": "Iceland",
    "Region": "Iceland-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.65,
        64
      ]
    },
    "Elevation": 2119,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "f6370b57-799e-1266-b725-7bbf51123b45"
  },
  {
    "Volcano Name": "Orizaba, Pico de",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.268,
        19.03
      ]
    },
    "Elevation": 5675,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "4fcd5994-5546-b682-db68-00200f92b50f"
  },
  {
    "Volcano Name": "Ormus Islands",
    "Country": "Indian Ocean",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        57,
        26
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b2dcf179-77fb-4f65-7930-09a2ba6a782b"
  },
  {
    "Volcano Name": "Orosi",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.473,
        10.98
      ]
    },
    "Elevation": 1659,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "fd2150c0-7c81-fff7-ea4a-e458c94e381e"
  },
  {
    "Volcano Name": "Oshima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.38,
        34.73
      ]
    },
    "Elevation": 758,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8e72f771-c55d-b457-a58f-149b2c3f13d8"
  },
  {
    "Volcano Name": "Oshima-Oshima",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.37,
        41.5
      ]
    },
    "Elevation": 737,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "cf04cc8c-fbea-ae28-a04d-200240db1374"
  },
  {
    "Volcano Name": "Osore-yama",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.08,
        41.32
      ]
    },
    "Elevation": 879,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "de1a17ad-ceba-d57a-055c-df1d6e9b2a8f"
  },
  {
    "Volcano Name": "Osorno",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.493,
        -41.1
      ]
    },
    "Elevation": 2652,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "e9570342-4c1c-f8de-f4fa-d15a187af983"
  },
  {
    "Volcano Name": "Ostanets",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.322,
        52.146
      ]
    },
    "Elevation": 719,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f597e7ab-eeaf-c6c5-383f-44a9e5e04c50"
  },
  {
    "Volcano Name": "Ostry",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.82,
        58.18
      ]
    },
    "Elevation": 2552,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8dc3a0be-06d0-4a0b-10ac-d42f931bb5c5"
  },
  {
    "Volcano Name": "Otdelniy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.428,
        52.22
      ]
    },
    "Elevation": 791,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d42a3b87-3189-a9af-c738-e6ba49352eaf"
  },
  {
    "Volcano Name": "Overo, Cerro",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.67,
        -23.35
      ]
    },
    "Elevation": 4555,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1dd67061-c163-18da-a81b-539006163f18"
  },
  {
    "Volcano Name": "Ozernoy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.38,
        51.88
      ]
    },
    "Elevation": 562,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fd7d702f-35d0-c0a8-107b-bc3f655e5681"
  },
  {
    "Volcano Name": "Pacaya",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.601,
        14.381
      ]
    },
    "Elevation": 2552,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c836d6bd-18d2-cd1c-4675-ea6df6426a06"
  },
  {
    "Volcano Name": "Paco",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.52,
        9.593
      ]
    },
    "Elevation": 524,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1118108e-895e-2d4f-baf1-b9aef28f5292"
  },
  {
    "Volcano Name": "Pagan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.8,
        18.13
      ]
    },
    "Elevation": 570,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "614cd82a-090f-716e-6fb3-05fb6d6f97fa"
  },
  {
    "Volcano Name": "Pago",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.52,
        -5.58
      ]
    },
    "Elevation": 742,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c1c9c2e2-cf98-750f-a576-e8727e4e03ac"
  },
  {
    "Volcano Name": "Paka",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.18,
        0.92
      ]
    },
    "Elevation": 1697,
    "Type": "Shield volcano",
    "Status": "Ar/Ar",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "58d1f1e3-1249-cf96-1de6-22df863ea5cd"
  },
  {
    "Volcano Name": "Palei-Aike Volc Field",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70,
        -52
      ]
    },
    "Elevation": 250,
    "Type": "Cinder cone",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "af9e9b47-9de5-9c84-c822-0d5eabf280e3"
  },
  {
    "Volcano Name": "Palena Volc Group",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.5,
        -43.68
      ]
    },
    "Elevation": 0,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "010cc4cd-d67a-81f4-c66c-3fd41e09a08b"
  },
  {
    "Volcano Name": "Palinuro",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.833,
        39.483
      ]
    },
    "Elevation": -70,
    "Type": "Submarine volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "de41ad7f-cbf8-b933-c78c-375bdcf1462a"
  },
  {
    "Volcano Name": "Palomo",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.295,
        -34.608
      ]
    },
    "Elevation": 4860,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "df16c7cb-dc59-7962-185c-9066d98c925e"
  },
  {
    "Volcano Name": "Paluweh",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.708,
        -8.32
      ]
    },
    "Elevation": 875,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c399c999-d4e7-d647-dc09-ee602a086830"
  },
  {
    "Volcano Name": "Pampa Luxsar",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.2,
        -20.85
      ]
    },
    "Elevation": 5543,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0a0a290e-8db7-63f9-8d99-a80a9cbe3335"
  },
  {
    "Volcano Name": "Pantelleria",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        12.02,
        36.77
      ]
    },
    "Elevation": 836,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "07787e1b-3a96-afe6-aee0-97648252a1f3"
  },
  {
    "Volcano Name": "Pantoja, Cerro",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.95,
        -40.77
      ]
    },
    "Elevation": 2112,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "49a3af88-a9e3-78d8-8770-fe33a3a8cd8c"
  },
  {
    "Volcano Name": "Papandayan",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.73,
        -7.32
      ]
    },
    "Elevation": 2665,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3362ab5a-42fe-73f7-de16-90cea52043dc"
  },
  {
    "Volcano Name": "Papayo",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -98.7,
        19.308
      ]
    },
    "Elevation": 3600,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b9e04506-54c9-5bc0-c67f-9e003fdf0768"
  },
  {
    "Volcano Name": "Paricutin Volcanic Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -102.25,
        19.48
      ]
    },
    "Elevation": 3860,
    "Type": "Cinder cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "85a65f92-0864-a7c1-08d2-d53a6bd5c927"
  },
  {
    "Volcano Name": "Parker",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.892,
        6.12
      ]
    },
    "Elevation": 1824,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "6a40117b-d765-bb8b-852b-374135cc9ada"
  },
  {
    "Volcano Name": "Patah",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.3,
        -4.27
      ]
    },
    "Elevation": 2817,
    "Type": "Unknown",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "cfb49a90-0bc6-d3e0-969a-aeb3011a852b"
  },
  {
    "Volcano Name": "Patates, Morne",
    "Country": "Dominica",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.37,
        15.22
      ]
    },
    "Elevation": 960,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "7edae044-13ec-67c2-dafc-cec123d40fdd"
  },
  {
    "Volcano Name": "Patilla Pata",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.03,
        -18.05
      ]
    },
    "Elevation": 5300,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "f63a0a99-13ed-43eb-4caa-800a8b6ac463"
  },
  {
    "Volcano Name": "Patoc",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.98,
        17.147
      ]
    },
    "Elevation": 1865,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9fd258a2-cf18-c6cb-d8a1-4cbddbe2a03a"
  },
  {
    "Volcano Name": "Patuha",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.37,
        -7.15
      ]
    },
    "Elevation": 2434,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2699f3a3-efdb-eb11-1c19-edbfb5694a26"
  },
  {
    "Volcano Name": "Paulet",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -55.77,
        -63.58
      ]
    },
    "Elevation": 353,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "daa82bb6-277e-c14d-21b0-81ce398c2b33"
  },
  {
    "Volcano Name": "Pavlof",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -161.9,
        55.42
      ]
    },
    "Elevation": 2519,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4f3ee63d-23f1-227f-1689-f5e8d0bc187e"
  },
  {
    "Volcano Name": "Pavlof Sister",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -161.87,
        55.45
      ]
    },
    "Elevation": 2142,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "a26a9206-53d5-9d38-f255-01a13e007e5b"
  },
  {
    "Volcano Name": "Payun Matru, Cerro",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.2,
        -36.42
      ]
    },
    "Elevation": 3691,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c69a6ae8-fe99-7739-27d8-68af97c367c6"
  },
  {
    "Volcano Name": "Peinado",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.15,
        -26.617
      ]
    },
    "Elevation": 5740,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6fe082b4-ae22-7e92-8914-a91f461b7a40"
  },
  {
    "Volcano Name": "Pelee",
    "Country": "Martinique",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.17,
        14.82
      ]
    },
    "Elevation": 1397,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b7cff751-f5dd-41e4-abcc-42848dc15bab"
  },
  {
    "Volcano Name": "Penanggungan",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.63,
        -7.62
      ]
    },
    "Elevation": 1653,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4213ecab-3f3e-890b-31e8-7e4563bef46f"
  },
  {
    "Volcano Name": "Pendan",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.02,
        -2.82
      ]
    },
    "Elevation": 0,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "177299e7-c66f-dcbb-8d5c-92972e11a7cb"
  },
  {
    "Volcano Name": "Penguin Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -57.93,
        -62.1
      ]
    },
    "Elevation": 180,
    "Type": "Stratovolcano",
    "Status": "Lichenometry",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "e088ec06-86b0-8ff9-0364-793fc4ed25e2"
  },
  {
    "Volcano Name": "Perbakti",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.68,
        -6.75
      ]
    },
    "Elevation": 1699,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fea6cf74-2b53-ddda-b8b7-54a5af56a788"
  },
  {
    "Volcano Name": "Petacas",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.78,
        1.57
      ]
    },
    "Elevation": 4054,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "39c73c78-1547-cf32-b492-1e8916e68ec1"
  },
  {
    "Volcano Name": "Peter I Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.58,
        -68.85
      ]
    },
    "Elevation": 1640,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e39f1ccd-6336-76dd-c89d-b67f67a96e2d"
  },
  {
    "Volcano Name": "Peuet Sague",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        96.329,
        4.914
      ]
    },
    "Elevation": 2801,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "09e977e7-7dd4-1629-3aa0-8fe1758ddc51"
  },
  {
    "Volcano Name": "Pico",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -28.4,
        38.47
      ]
    },
    "Elevation": 2351,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "a5f1b348-7c7a-d964-ad68-95d0b13cfed0"
  },
  {
    "Volcano Name": "Piip",
    "Country": "Russia",
    "Region": "Kamchatka-E of",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.33,
        55.42
      ]
    },
    "Elevation": -300,
    "Type": "Submarine volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fc346fa5-e874-4519-6d5d-f023dfbeefad"
  },
  {
    "Volcano Name": "Pilas, Las",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.688,
        12.495
      ]
    },
    "Elevation": 1088,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "fe5575c5-e141-03a3-10d0-f1c5b5bd63a0"
  },
  {
    "Volcano Name": "Pina, Cerro",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.65,
        -19.492
      ]
    },
    "Elevation": 4037,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "12e08a3a-155f-ac49-764a-e42403787ad1"
  },
  {
    "Volcano Name": "Pinacate",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.498,
        31.772
      ]
    },
    "Elevation": 1200,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2ff11201-75a3-51fa-6216-0c9cc6730b58"
  },
  {
    "Volcano Name": "Pinatubo",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.35,
        15.13
      ]
    },
    "Elevation": 1486,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a6520c8c-0ec1-07fa-88d6-8c9c26cabc0f"
  },
  {
    "Volcano Name": "Pinta",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.75,
        0.58
      ]
    },
    "Elevation": 780,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "19406e3a-da34-a420-96a3-c5a9fa48a71a"
  },
  {
    "Volcano Name": "Piparo",
    "Country": "Trinidad",
    "Region": "Trinidad",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61,
        10
      ]
    },
    "Elevation": 140,
    "Type": "Mud volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d5a5583d-4e26-d3f2-3b9e-4fc995afad00"
  },
  {
    "Volcano Name": "Piratkovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.849,
        52.113
      ]
    },
    "Elevation": 1322,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "000c3896-a4cd-8a90-61f2-5b1115fabfeb"
  },
  {
    "Volcano Name": "Planchon-Peteroa",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.57,
        -35.24
      ]
    },
    "Elevation": 4107,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d46ba647-f3e8-3309-5fd4-40d6fa09cbbb"
  },
  {
    "Volcano Name": "Platanar",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -84.366,
        10.3
      ]
    },
    "Elevation": 2267,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "de6b6ef4-cec9-0f07-731b-be460ac37a3f"
  },
  {
    "Volcano Name": "Pleiades, The",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        165.5,
        -72.67
      ]
    },
    "Elevation": 3040,
    "Type": "Stratovolcano",
    "Status": "K-Ar",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fff9579c-c6f9-d385-1d23-3e6e33106ac9"
  },
  {
    "Volcano Name": "Plosky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.47,
        55.2
      ]
    },
    "Elevation": 1236,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1ef9d275-0fe6-241d-5ae3-3a582249a086"
  },
  {
    "Volcano Name": "Plosky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.25,
        57.83
      ]
    },
    "Elevation": 1255,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "941ff97c-0c12-3ee8-efc5-3a40d2db790a"
  },
  {
    "Volcano Name": "Plosky Volc Group",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.53,
        52.02
      ]
    },
    "Elevation": 681,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c9788caa-6c6b-63a4-46e4-e5a90c6afcb9"
  },
  {
    "Volcano Name": "Poas",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -84.233,
        10.2
      ]
    },
    "Elevation": 2708,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "539aef08-9d9e-812b-b440-20a487b0a815"
  },
  {
    "Volcano Name": "Pocdol Mountains",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.958,
        13.05
      ]
    },
    "Elevation": 1102,
    "Type": "Compound volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ae7a837f-00f9-1393-8ed0-6a82e766a5c6"
  },
  {
    "Volcano Name": "Poco Leok",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.48,
        -8.68
      ]
    },
    "Elevation": 1675,
    "Type": "Unknown",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d47ef0f8-26c5-d09c-f539-66d2611d6be5"
  },
  {
    "Volcano Name": "Pogranychny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.8,
        56.85
      ]
    },
    "Elevation": 1427,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "29af2f7b-4d04-6a15-54b5-197a120ed20c"
  },
  {
    "Volcano Name": "Popa",
    "Country": "Myanmar",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        95.23,
        20.87
      ]
    },
    "Elevation": 1518,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "fd02ff0e-018e-cdf8-6659-40ef31a81d15"
  },
  {
    "Volcano Name": "Popocatepetl",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -98.622,
        19.023
      ]
    },
    "Elevation": 5426,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7182da7a-b446-3843-da02-da05e47e7b38"
  },
  {
    "Volcano Name": "Porak",
    "Country": "Armenia",
    "Region": "Armenia",
    "Location": {
      "type": "Point",
      "coordinates": [
        45.783,
        40.017
      ]
    },
    "Elevation": 2800,
    "Type": "Stratovolcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f2dfef3f-47d4-f87b-506b-8bfd132acabc"
  },
  {
    "Volcano Name": "Possession, Ile de la",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        51.63,
        -46.42
      ]
    },
    "Elevation": 934,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "72c23d17-1751-46f0-37ad-7710af5f6160"
  },
  {
    "Volcano Name": "Prestahnukur",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -20.58,
        64.6
      ]
    },
    "Elevation": 1390,
    "Type": "Subglacial volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "2fa828de-9348-f7f8-819d-bbc4bfdb7db8"
  },
  {
    "Volcano Name": "Prevo Peak",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.12,
        47.02
      ]
    },
    "Elevation": 1360,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b1143c6e-8f28-3aa9-b9e4-6d9e83dc41f5"
  },
  {
    "Volcano Name": "Prieto, Cerro",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -115.305,
        32.418
      ]
    },
    "Elevation": 223,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "39cb29be-bc9d-f97c-51f8-537f300e097e"
  },
  {
    "Volcano Name": "Prince Edward Island",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.95,
        -46.63
      ]
    },
    "Elevation": 672,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cc3b0f06-5e9f-b0af-8e97-e2c1f4c4b7c7"
  },
  {
    "Volcano Name": "Protector Shoal",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -28.08,
        -55.92
      ]
    },
    "Elevation": -27,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "33acd0c8-0430-6b02-cb3d-76ca75604b36"
  },
  {
    "Volcano Name": "Puesto Cortaderas",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.62,
        -37.55
      ]
    },
    "Elevation": 970,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f415364b-6702-2c31-1305-260e7f543728"
  },
  {
    "Volcano Name": "Puesto Cortaderas",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.617,
        -37.567
      ]
    },
    "Elevation": 970,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "75405b8c-debc-6bd7-5af4-11beeaf101b3"
  },
  {
    "Volcano Name": "Pular",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.05,
        -24.18
      ]
    },
    "Elevation": 6233,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7fe98b3f-b0ca-0e3f-ee27-fc51bf845d08"
  },
  {
    "Volcano Name": "Pululagua",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.463,
        0.038
      ]
    },
    "Elevation": 3356,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ccfb30cb-648e-2003-4565-2a4c88360aa4"
  },
  {
    "Volcano Name": "Puntiagudo-Cordon Cenizo",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.264,
        -40.969
      ]
    },
    "Elevation": 2493,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "5dc53c53-d890-c6a2-8d41-2d8f5ed4a38c"
  },
  {
    "Volcano Name": "Purace",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.4,
        2.32
      ]
    },
    "Elevation": 4650,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "439e5ff6-ab39-fb36-6fe8-699b2efd2160"
  },
  {
    "Volcano Name": "Purico Complex",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.75,
        -23
      ]
    },
    "Elevation": 5703,
    "Type": "Pyroclastic shield",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8c56e835-cb62-fc47-f19a-9c526e89b7b2"
  },
  {
    "Volcano Name": "Putana",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.87,
        -22.57
      ]
    },
    "Elevation": 5890,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "936f06d7-d65e-1dc5-5b4f-78c540722b41"
  },
  {
    "Volcano Name": "Puyehue",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.117,
        -40.59
      ]
    },
    "Elevation": 2236,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2a389ccc-d386-20f1-2cab-c8712a882dd8"
  },
  {
    "Volcano Name": "Puyuhuapi",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.53,
        -44.3
      ]
    },
    "Elevation": 255,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cd5b1ac2-eb9b-dcb4-65b5-1c51f2780132"
  },
  {
    "Volcano Name": "Qal'eh Hasan Ali",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        57.57,
        29.4
      ]
    },
    "Elevation": 0,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "1291ee35-0595-e040-79d8-7833fe81b42c"
  },
  {
    "Volcano Name": "Qualibou",
    "Country": "St. Lucia",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.05,
        13.83
      ]
    },
    "Elevation": 777,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "89add59d-eaa3-c5f8-a0e5-d7b35733df49"
  },
  {
    "Volcano Name": "Quetrupillan",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.7,
        -39.5
      ]
    },
    "Elevation": 2360,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "64bfdb00-cdb3-1021-7f43-3548a6e24e46"
  },
  {
    "Volcano Name": "Quezaltepeque",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.35,
        14.65
      ]
    },
    "Elevation": 1200,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fe7efee0-ff92-64ee-f2de-6ad61958ce46"
  },
  {
    "Volcano Name": "Quill, The",
    "Country": "Netherlands",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -62.95,
        17.48
      ]
    },
    "Elevation": 601,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "1653d40c-2542-a939-39a2-c9f538fcefa5"
  },
  {
    "Volcano Name": "Quilotoa",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.9,
        -0.85
      ]
    },
    "Elevation": 3914,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "d634d0bd-4563-9619-213b-8189da694500"
  },
  {
    "Volcano Name": "Quimsachata",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.17,
        -14.37
      ]
    },
    "Elevation": 3923,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "608d1780-bd3c-b261-d69f-dce128a3d54c"
  },
  {
    "Volcano Name": "Rabaul",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.203,
        -4.271
      ]
    },
    "Elevation": 688,
    "Type": "Pyroclastic shield",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "602569d4-2f7f-3090-bf10-f03a1d21ee2f"
  },
  {
    "Volcano Name": "Ragang",
    "Country": "Philippines",
    "Region": "Mindanao-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.5,
        7.67
      ]
    },
    "Elevation": 2815,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "73db5100-acdc-90b5-0e62-5b63370c9c10"
  },
  {
    "Volcano Name": "Rahah, Harrat ar",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.17,
        27.8
      ]
    },
    "Elevation": 1660,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ff8d1049-6a53-71db-e738-7c382dd713b4"
  },
  {
    "Volcano Name": "Rahat, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.78,
        23.08
      ]
    },
    "Elevation": 1744,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "133900b5-0617-3bab-0c47-97775b78be5f"
  },
  {
    "Volcano Name": "Raikoke",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.25,
        48.292
      ]
    },
    "Elevation": 551,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "11fe9a8a-a006-cda7-4c74-1248b708442a"
  },
  {
    "Volcano Name": "Rainier",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.758,
        46.87
      ]
    },
    "Elevation": 4392,
    "Type": "Stratovolcano",
    "Status": "Dendrochronology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "682fe1d3-1e2a-c135-d47f-f3351afd03e3"
  },
  {
    "Volcano Name": "Rajabasa",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        105.625,
        -5.78
      ]
    },
    "Elevation": 1281,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "862140fb-b724-a968-3c7d-54ebf3c098b1"
  },
  {
    "Volcano Name": "Ranakah",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.52,
        -8.62
      ]
    },
    "Elevation": 2100,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8dbc5c0f-df6f-e66e-e438-4d446a5ba885"
  },
  {
    "Volcano Name": "Ranau",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        103.92,
        -4.83
      ]
    },
    "Elevation": 1881,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "67c36e40-a3d0-8772-a87d-78f3e2e16a1d"
  },
  {
    "Volcano Name": "Raoul Island",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.92,
        -29.27
      ]
    },
    "Elevation": 516,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2a78eaef-5c8e-03ae-24d1-635920b57017"
  },
  {
    "Volcano Name": "Rasshua",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.02,
        47.77
      ]
    },
    "Elevation": 956,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c051f292-a23b-5669-8058-bf8bce8879a5"
  },
  {
    "Volcano Name": "Raung",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        114.042,
        -8.125
      ]
    },
    "Elevation": 3332,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7416b57d-6905-1fa3-4002-e7abfa27d994"
  },
  {
    "Volcano Name": "Rausu",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.125,
        44.073
      ]
    },
    "Elevation": 1660,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "64472b91-5cd8-6743-b6e6-1377b2c43de6"
  },
  {
    "Volcano Name": "Recheschnoi",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -168.55,
        53.15
      ]
    },
    "Elevation": 1984,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "91560aa5-8f65-7510-471f-3d3acc5c8850"
  },
  {
    "Volcano Name": "Reclus",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.7,
        -50.98
      ]
    },
    "Elevation": 0,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ced76b04-662c-9d3b-355b-5db6aaa53b12"
  },
  {
    "Volcano Name": "Red Cones",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -119.05,
        37.58
      ]
    },
    "Elevation": 2748,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "061c9abc-cdde-b096-818c-619ae8de9f6b"
  },
  {
    "Volcano Name": "Redoubt",
    "Country": "United States",
    "Region": "Alaska-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -152.75,
        60.48
      ]
    },
    "Elevation": 3108,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f5036e12-95d9-d0a0-8bc5-f3e578056455"
  },
  {
    "Volcano Name": "Reporoa",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        176.33,
        -38.42
      ]
    },
    "Elevation": 592,
    "Type": "Caldera",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "21f9e16c-fa57-d68e-ba4f-4802da7d3205"
  },
  {
    "Volcano Name": "Resago, Volcan",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.92,
        -36.45
      ]
    },
    "Elevation": 1550,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8e4cd6db-b132-70ef-0552-d085a2bfdd0b"
  },
  {
    "Volcano Name": "Reventador",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.656,
        -0.078
      ]
    },
    "Elevation": 3562,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5399c58d-098a-d313-823f-f0d051e8c30f"
  },
  {
    "Volcano Name": "Reykjanes",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -22.5,
        63.88
      ]
    },
    "Elevation": 230,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "4fd75473-54ce-c94d-bbf0-707659dfb997"
  },
  {
    "Volcano Name": "Reykjaneshryggur",
    "Country": "Iceland",
    "Region": "Iceland-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -23.33,
        63.67
      ]
    },
    "Elevation": 80,
    "Type": "Submarine volcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c81bc785-6c7e-c00d-ff9e-07091dbcaa5c"
  },
  {
    "Volcano Name": "Rincon de la Vieja",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.324,
        10.83
      ]
    },
    "Elevation": 1916,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8e0c9711-d25c-01e2-2384-8d421a792366"
  },
  {
    "Volcano Name": "Rinjani",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        116.47,
        -8.42
      ]
    },
    "Elevation": 3726,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "697d7b9b-b392-a3cd-cd41-7b7dd33d371a"
  },
  {
    "Volcano Name": "Risco Plateado",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70,
        -34.933
      ]
    },
    "Elevation": 4999,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "a9edddb1-67b0-7ec1-1853-1fd6e0f6f091"
  },
  {
    "Volcano Name": "Rishiri",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.25,
        45.18
      ]
    },
    "Elevation": 1719,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "dde6f535-d2f6-6407-a86e-6d8b309d915a"
  },
  {
    "Volcano Name": "Ritter Island",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.121,
        -5.52
      ]
    },
    "Elevation": 140,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "264329c7-bcfe-0a57-e791-6891452a2b4c"
  },
  {
    "Volcano Name": "Robinson Crusoe",
    "Country": "Chile",
    "Region": "Chile-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.85,
        -33.658
      ]
    },
    "Elevation": 922,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "4c8e3f85-7489-ad19-3586-03528c4c9a6e"
  },
  {
    "Volcano Name": "Robledo",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.717,
        -26.767
      ]
    },
    "Elevation": 4400,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "949734d2-d8c4-92c1-06ba-bde8aaf5f0e1"
  },
  {
    "Volcano Name": "Rocard",
    "Country": "France",
    "Region": "Society Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -148.6,
        -17.642
      ]
    },
    "Elevation": -2100,
    "Type": "Submarine volcano",
    "Status": "Seismicity",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "514d99cd-6bae-ae3a-8c3b-ba11c89ec611"
  },
  {
    "Volcano Name": "Romanovka",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.8,
        55.65
      ]
    },
    "Elevation": 1442,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5ffe208e-8413-779a-22b8-4c75d3e7a7d7"
  },
  {
    "Volcano Name": "Rota",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.75,
        12.55
      ]
    },
    "Elevation": 832,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fdb3fed3-517c-4f23-8861-9cb099e73552"
  },
  {
    "Volcano Name": "Roundtop",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.6,
        54.8
      ]
    },
    "Elevation": 1871,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c3b09838-9b3d-78d3-e7b4-40019c3f9cb1"
  },
  {
    "Volcano Name": "Royal Society Range",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        163.6,
        -78.25
      ]
    },
    "Elevation": 3000,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "7c639630-b075-f179-67b5-dd1c21367e61"
  },
  {
    "Volcano Name": "Ruang",
    "Country": "Indonesia",
    "Region": "Sangihe Is-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.425,
        2.28
      ]
    },
    "Elevation": 725,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "ddc9678d-307c-eeaf-ec4e-b13e8d54e094"
  },
  {
    "Volcano Name": "Ruapehu",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        175.57,
        -39.28
      ]
    },
    "Elevation": 2797,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "521c04f4-7081-76e8-e105-218cea725c85"
  },
  {
    "Volcano Name": "Ruby",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.567,
        15.617
      ]
    },
    "Elevation": -230,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4f3ca660-9e5e-8ead-bce5-298f4c9fc9b5"
  },
  {
    "Volcano Name": "Ruby",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.57,
        15.62
      ]
    },
    "Elevation": -230,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "1335d0cc-6cdb-bbfb-17b2-2015b1fd120a"
  },
  {
    "Volcano Name": "Ruby Mountain",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -133.32,
        59.68
      ]
    },
    "Elevation": 1523,
    "Type": "Cinder cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "225391dc-1b30-3123-a545-e4681c36c7e2"
  },
  {
    "Volcano Name": "Rudakov",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.83,
        45.88
      ]
    },
    "Elevation": 542,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "418af57f-a1b9-120c-2633-c43baed6fe0a"
  },
  {
    "Volcano Name": "Ruiz",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.322,
        4.895
      ]
    },
    "Elevation": 5321,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "27668b04-eb85-f2d6-98c0-b49406f7887c"
  },
  {
    "Volcano Name": "Rumble II West",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.527,
        -35.353
      ]
    },
    "Elevation": 1200,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "522f0343-a675-c11a-8c43-8fc1c4417c99"
  },
  {
    "Volcano Name": "Rumble III",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.478,
        -35.745
      ]
    },
    "Elevation": -140,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4ddacb07-1149-d209-13f9-e2ef9462a562"
  },
  {
    "Volcano Name": "Rumble IV",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.05,
        -36.13
      ]
    },
    "Elevation": -450,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9125c390-5abc-33f5-a96a-4bc25431e64c"
  },
  {
    "Volcano Name": "Rumble V",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.197,
        -36.139
      ]
    },
    "Elevation": -700,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c0d545f6-746f-0a14-3bce-d525daa2da0c"
  },
  {
    "Volcano Name": "Rungwe",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.67,
        -9.13
      ]
    },
    "Elevation": 2961,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0009bbf3-b686-a196-dd7b-40bb6190a998"
  },
  {
    "Volcano Name": "R?o Murta",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.667,
        -46.167
      ]
    },
    "Elevation": null,
    "Type": "Pyroclastic cones",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "2d7ef78e-9029-7fdf-6522-6d08253a77d7"
  },
  {
    "Volcano Name": "SW Usangu Basin",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.8,
        -8.75
      ]
    },
    "Elevation": 2179,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2c08fdbd-e42f-3bc0-65c8-c4e8fe81b72a"
  },
  {
    "Volcano Name": "Saba",
    "Country": "Netherlands",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -63.23,
        17.63
      ]
    },
    "Elevation": 887,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "c4654066-6887-c00e-5439-7b0e3ee0d1f9"
  },
  {
    "Volcano Name": "Sabalan",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        47.917,
        38.25
      ]
    },
    "Elevation": 4811,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3d665c7b-2f55-b352-22a5-3bcaa5330961"
  },
  {
    "Volcano Name": "Sabancaya",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.85,
        -15.783
      ]
    },
    "Elevation": 5967,
    "Type": "Stratovolcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3b1e5164-8ac8-6199-e1e2-e4490cf4b436"
  },
  {
    "Volcano Name": "Sabancaya",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.85,
        -15.78
      ]
    },
    "Elevation": 5967,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "bb362c21-749f-8c99-d40b-10a5daedd2cb"
  },
  {
    "Volcano Name": "Sahand",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        46.433,
        37.75
      ]
    },
    "Elevation": 3707,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5a7c7f37-5748-ab2d-66fb-7f0b66a33657"
  },
  {
    "Volcano Name": "Sairecabur",
    "Country": "Bolivia",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.88,
        -22.73
      ]
    },
    "Elevation": 5971,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0ed1504f-3f99-179c-0f94-db572ad8963d"
  },
  {
    "Volcano Name": "Sakar",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.094,
        -5.414
      ]
    },
    "Elevation": 992,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "7a3480cf-d551-176f-94ea-55926990b10a"
  },
  {
    "Volcano Name": "Sakura-jima",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.67,
        31.58
      ]
    },
    "Elevation": 1117,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "e5ffed21-e74f-bec6-5dd5-fc231dbe31d2"
  },
  {
    "Volcano Name": "Salak",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        106.73,
        -6.72
      ]
    },
    "Elevation": 2211,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d18b78c0-d575-9443-7fda-d99b927186fe"
  },
  {
    "Volcano Name": "San Borja Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.75,
        28.5
      ]
    },
    "Elevation": 1360,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "71520785-6273-09c0-7fd2-90cb3b5e6ffb"
  },
  {
    "Volcano Name": "San Carlos",
    "Country": "Equatorial Guinea",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        8.52,
        3.35
      ]
    },
    "Elevation": 2260,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e4d821f0-c439-b16b-7467-1db69a959d15"
  },
  {
    "Volcano Name": "San Cristobal",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.5,
        -0.88
      ]
    },
    "Elevation": 759,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "dad175ed-818d-7a59-464a-135fb78580af"
  },
  {
    "Volcano Name": "San Cristobal",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.004,
        12.702
      ]
    },
    "Elevation": 1745,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9e390da0-aaa3-291b-36fe-6d46136cbae7"
  },
  {
    "Volcano Name": "San Diego",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.48,
        14.27
      ]
    },
    "Elevation": 781,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "47630028-ba68-0187-c5b6-0289e66cf511"
  },
  {
    "Volcano Name": "San Felix",
    "Country": "Chile",
    "Region": "Chile-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -80.12,
        -26.27
      ]
    },
    "Elevation": 183,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1cc1ddf0-4cd4-5854-a4d4-42fcd8bc419b"
  },
  {
    "Volcano Name": "San Joaquin",
    "Country": "Equatorial Guinea",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        8.63,
        3.35
      ]
    },
    "Elevation": 2009,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4bc0269b-c512-ebd0-12f4-d3a1fe703eb8"
  },
  {
    "Volcano Name": "San Jorge",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -28.08,
        38.65
      ]
    },
    "Elevation": 1053,
    "Type": "Fissure vent",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "95bb7d2d-8eb2-686e-85cf-b3b9810fd63f"
  },
  {
    "Volcano Name": "San Jose",
    "Country": "Argentina",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.897,
        -33.782
      ]
    },
    "Elevation": 5856,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "1937dd2c-d5bd-5896-330e-ce37c84b4a94"
  },
  {
    "Volcano Name": "San Luis, Isla",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -114.384,
        29.814
      ]
    },
    "Elevation": 180,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "01a5ccb5-a2bd-d3fe-7ccb-f6503bc91b13"
  },
  {
    "Volcano Name": "San Marcelino",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.63,
        13.853
      ]
    },
    "Elevation": 2381,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d0547de5-111e-351c-b1b5-840ebf41a8ab"
  },
  {
    "Volcano Name": "San Martin",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -95.169,
        18.572
      ]
    },
    "Elevation": 1650,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "e40a3d12-a9ea-546e-64ea-ec876380a32e"
  },
  {
    "Volcano Name": "San Miguel",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.272,
        13.431
      ]
    },
    "Elevation": 2130,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d0ed5831-6590-311f-78b6-1a8352413901"
  },
  {
    "Volcano Name": "San Pedro",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.4,
        -21.88
      ]
    },
    "Elevation": 6145,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "f557df83-514b-6619-c724-3741ce85c508"
  },
  {
    "Volcano Name": "San Pedro-Pellado",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.849,
        -35.989
      ]
    },
    "Elevation": 3621,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d6a7d604-48e3-1de1-c914-801453e3dc4d"
  },
  {
    "Volcano Name": "San Quintin Volc Field",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -115.996,
        30.468
      ]
    },
    "Elevation": 260,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5557f362-db18-96b1-d715-2a2addac2593"
  },
  {
    "Volcano Name": "San Salvador",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.286,
        13.736
      ]
    },
    "Elevation": 1893,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "91fde530-98eb-f8be-9742-c859078f16c5"
  },
  {
    "Volcano Name": "San Vicente",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.852,
        13.623
      ]
    },
    "Elevation": 2182,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bd58c6bd-a705-9047-05a4-048bd1068ad3"
  },
  {
    "Volcano Name": "Sanbe",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        132.62,
        35.13
      ]
    },
    "Elevation": 1126,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "71b1ecb4-6934-b717-cb89-61208ae45ab5"
  },
  {
    "Volcano Name": "Sand Mountain Field",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.93,
        44.38
      ]
    },
    "Elevation": 1664,
    "Type": "Cinder cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ce968961-53be-a806-1e0c-38e683e8e1b6"
  },
  {
    "Volcano Name": "Sanford",
    "Country": "United States",
    "Region": "Alaska-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -144.13,
        62.22
      ]
    },
    "Elevation": 4949,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "94186c0d-8c50-0fe9-c243-4f801b1231d9"
  },
  {
    "Volcano Name": "Sanganguey",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.73,
        21.45
      ]
    },
    "Elevation": 2340,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8174ea31-149e-0ee8-5d73-073bdc73e4ee"
  },
  {
    "Volcano Name": "Sangang?ey",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.733,
        21.45
      ]
    },
    "Elevation": 2340,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5c4e1c23-b9f9-dac7-3041-b42660049432"
  },
  {
    "Volcano Name": "Sangay",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.33,
        -2.03
      ]
    },
    "Elevation": 5230,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3cd22d6e-eba6-ea82-39b3-2504c37fb547"
  },
  {
    "Volcano Name": "Sangeang Api",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        119.058,
        -8.18
      ]
    },
    "Elevation": 1949,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "4c364e52-135d-a360-cde6-b43ac5d534e2"
  },
  {
    "Volcano Name": "Sano, Wai",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.025,
        -8.68
      ]
    },
    "Elevation": 903,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1c2670de-0e68-f70a-4855-e7eb4d13d954"
  },
  {
    "Volcano Name": "Santa Ana",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.63,
        13.853
      ]
    },
    "Elevation": 2365,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "fa99df00-e569-dcdd-1a7c-368c7aba9114"
  },
  {
    "Volcano Name": "Santa Clara",
    "Country": "United States",
    "Region": "US-Utah",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.625,
        37.257
      ]
    },
    "Elevation": 1465,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4b082e59-f76f-63a0-a015-6596bd090751"
  },
  {
    "Volcano Name": "Santa Cruz",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.33,
        -0.62
      ]
    },
    "Elevation": 864,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7e6b3427-b92a-78b0-7d81-93a71bacb452"
  },
  {
    "Volcano Name": "Santa Isabel",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.37,
        4.82
      ]
    },
    "Elevation": 4950,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "8c363aed-f035-a1c9-15ce-1969e68dde82"
  },
  {
    "Volcano Name": "Santa Isabel",
    "Country": "Equatorial Guinea",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        8.75,
        3.58
      ]
    },
    "Elevation": 2972,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "83e2bf8c-c20e-7288-940c-b66d978ef702"
  },
  {
    "Volcano Name": "Santa Maria",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.552,
        14.756
      ]
    },
    "Elevation": 3772,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "df2e33fb-39fe-0a43-bc23-c186ec127186"
  },
  {
    "Volcano Name": "Santiago",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.77,
        -0.22
      ]
    },
    "Elevation": 920,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "0a12bd2a-ffe7-eadf-46d7-501f0959af59"
  },
  {
    "Volcano Name": "Santiago, Cerro",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.87,
        14.33
      ]
    },
    "Elevation": 1192,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6008e3d9-4682-ccd0-9ab4-e01f815ab893"
  },
  {
    "Volcano Name": "Santo Antao",
    "Country": "Cape Verde",
    "Region": "Cape Verde Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.17,
        17.07
      ]
    },
    "Elevation": 1979,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e8f52659-7ca6-25da-62d7-45a5209011d7"
  },
  {
    "Volcano Name": "Santo Tomas",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.55,
        16.33
      ]
    },
    "Elevation": 2260,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "791d22c0-0ed0-9d42-d237-4bb59ab074df"
  },
  {
    "Volcano Name": "Santorini",
    "Country": "Greece",
    "Region": "Greece",
    "Location": {
      "type": "Point",
      "coordinates": [
        25.396,
        36.404
      ]
    },
    "Elevation": 329,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "f4c17035-c341-7f5d-a014-dc08698c2e81"
  },
  {
    "Volcano Name": "Sao Tome",
    "Country": "Sao Tome & Principe",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        6.72,
        0.32
      ]
    },
    "Elevation": 2024,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "ee3fb375-711c-67db-1dc6-ac893c50b995"
  },
  {
    "Volcano Name": "Sarigan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.78,
        16.708
      ]
    },
    "Elevation": 538,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9e537dbb-595f-eb17-4175-e497b8621b63"
  },
  {
    "Volcano Name": "Sarigan",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.783,
        16.708
      ]
    },
    "Elevation": 538,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f3c3060d-4016-6e4d-9371-95cbdea04790"
  },
  {
    "Volcano Name": "Sarik-Gajah",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        100.2,
        0.08
      ]
    },
    "Elevation": 0,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "9a36ec2b-23fd-d346-b092-7f140bdd5672"
  },
  {
    "Volcano Name": "Sarychev Peak",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.2,
        48.092
      ]
    },
    "Elevation": 1496,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6b28dba7-078f-bb41-9f29-4ad0fb888284"
  },
  {
    "Volcano Name": "Satah Mountain",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -124.7,
        52.47
      ]
    },
    "Elevation": 1921,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f44c3b58-047a-2b95-b563-b6486be9840a"
  },
  {
    "Volcano Name": "Savai'i",
    "Country": "Samoa",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -172.525,
        -13.612
      ]
    },
    "Elevation": 1858,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d4033ea9-6457-e3cb-8e38-697a106948d5"
  },
  {
    "Volcano Name": "Savo",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.82,
        -9.13
      ]
    },
    "Elevation": 510,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "8d88b6d9-ce5b-ad70-b134-7aff8db4c431"
  },
  {
    "Volcano Name": "Sawad, Harra Es-",
    "Country": "Yemen",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        46.12,
        13.58
      ]
    },
    "Elevation": 1737,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "d1abe856-b23f-44ce-2f99-b7c9b477183a"
  },
  {
    "Volcano Name": "Schmidt",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.633,
        54.917
      ]
    },
    "Elevation": 2020,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "130491e9-58c8-2646-5007-04d74f0c2f0e"
  },
  {
    "Volcano Name": "Seal Nunataks Group",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -60.05,
        -65.03
      ]
    },
    "Elevation": 368,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "f34f3e15-b5d1-cb04-ca1d-d86e503ba23c"
  },
  {
    "Volcano Name": "Seamount X",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.017,
        13.25
      ]
    },
    "Elevation": -1230,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3e707c89-5de7-4f02-743c-8f88057c40a5"
  },
  {
    "Volcano Name": "Sedankinsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.08,
        57.23
      ]
    },
    "Elevation": 1241,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7c59463f-332e-4efa-5d73-3c23730ea183"
  },
  {
    "Volcano Name": "Segererua Plateau",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.9,
        1.57
      ]
    },
    "Elevation": 699,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "623f362a-a8c8-b30e-bb4e-780538c3f60d"
  },
  {
    "Volcano Name": "Seguam",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -172.52,
        52.32
      ]
    },
    "Elevation": 1054,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "66ed22ab-7793-5a10-2af1-562b08bea258"
  },
  {
    "Volcano Name": "Segula",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.13,
        52.02
      ]
    },
    "Elevation": 1153,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b7025871-e6b3-07e3-8fb6-08e9242cd247"
  },
  {
    "Volcano Name": "Sekincau Belirang",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        104.32,
        -5.12
      ]
    },
    "Elevation": 1719,
    "Type": "Caldera",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4fc77398-a12e-b00d-0db6-9129cff686ed"
  },
  {
    "Volcano Name": "Semeru",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.92,
        -8.108
      ]
    },
    "Elevation": 3676,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6f563cc7-6e9f-2e80-26a2-3c9d2bf747cc"
  },
  {
    "Volcano Name": "Semisopochnoi",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        179.62,
        51.95
      ]
    },
    "Elevation": 1221,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5751c71d-c061-7d53-2437-3adf0c8226fc"
  },
  {
    "Volcano Name": "Sempu",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.73,
        1.142
      ]
    },
    "Elevation": 1549,
    "Type": "Caldera",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "20f620bd-086f-4441-2f7d-0ffb691fbbe1"
  },
  {
    "Volcano Name": "Serdan-Oriental",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -97.47,
        19.27
      ]
    },
    "Elevation": 3485,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f66fdd06-5bf2-014e-a562-a7fec7bf9dcb"
  },
  {
    "Volcano Name": "Sergief",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.93,
        52.03
      ]
    },
    "Elevation": 560,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "3de6d3fe-8b52-559a-b5e3-0f3e4337e5e4"
  },
  {
    "Volcano Name": "Serua",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        130,
        -6.3
      ]
    },
    "Elevation": 641,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b3b6c1b9-777d-d993-fb1e-fe08d24831aa"
  },
  {
    "Volcano Name": "Sessagara",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.13,
        -9.48
      ]
    },
    "Elevation": 370,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fb587c4c-2e84-44bb-9a92-39b0e81be08f"
  },
  {
    "Volcano Name": "Sete Cidades",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.78,
        37.87
      ]
    },
    "Elevation": 856,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "dd2183e7-0363-8763-bb2b-e1d15e1e3c12"
  },
  {
    "Volcano Name": "Seulawah Agam",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        95.658,
        5.448
      ]
    },
    "Elevation": 1810,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "d12f2103-77ad-59be-b1d9-b0ad30c0f9e5"
  },
  {
    "Volcano Name": "Severny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.87,
        58.28
      ]
    },
    "Elevation": 1936,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9812dd46-4baa-fbdb-5ef4-43beca11910a"
  },
  {
    "Volcano Name": "Shala",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.55,
        7.47
      ]
    },
    "Elevation": 2075,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f6d72108-ffd7-51f8-72aa-c8f251063446"
  },
  {
    "Volcano Name": "Sharat Kovakab",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        40.85,
        36.53
      ]
    },
    "Elevation": 534,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3e91018a-c870-fab2-676e-73e4a603da44"
  },
  {
    "Volcano Name": "Shasta",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.2,
        41.42
      ]
    },
    "Elevation": 4317,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "433e1d69-3c68-df2d-a739-280fa858430f"
  },
  {
    "Volcano Name": "Shiga",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.52,
        36.7
      ]
    },
    "Elevation": 2036,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8eccbdee-d149-f8bd-8087-248e4f1cdcbf"
  },
  {
    "Volcano Name": "Shikaribetsu Group",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.08,
        43.28
      ]
    },
    "Elevation": 1430,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b7406e31-3bb5-90ea-33dd-5b5763bb75e9"
  },
  {
    "Volcano Name": "Shikotsu",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        141.33,
        42.7
      ]
    },
    "Elevation": 1320,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "02e88429-f94f-00be-1f4c-434774d55266"
  },
  {
    "Volcano Name": "Shiretoko-Iwo-zan",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.17,
        44.13
      ]
    },
    "Elevation": 1563,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "bf097085-88f3-b2bd-2aa2-0280c742285a"
  },
  {
    "Volcano Name": "Shirinki",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.98,
        50.2
      ]
    },
    "Elevation": 761,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "54044a1f-c676-9de3-1073-b45e930c4453"
  },
  {
    "Volcano Name": "Shishaldin",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.97,
        54.75
      ]
    },
    "Elevation": 2857,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "41f3dbe5-c3a8-129f-0986-00022b2ac317"
  },
  {
    "Volcano Name": "Shisheika",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        161.083,
        57.15
      ]
    },
    "Elevation": 379,
    "Type": "Lava dome",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "4fc4f7ae-8729-d279-ed80-06180e23d02d"
  },
  {
    "Volcano Name": "Shishel",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.37,
        57.45
      ]
    },
    "Elevation": 2525,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "525cc92d-ccf9-8ff7-10cd-00cc9e3e62c2"
  },
  {
    "Volcano Name": "Shiveluch",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        161.36,
        56.653
      ]
    },
    "Elevation": 3283,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3e398590-5fee-212b-15f4-50cc357104ba"
  },
  {
    "Volcano Name": "Shoshone Lava Field",
    "Country": "United States",
    "Region": "US-Idaho",
    "Location": {
      "type": "Point",
      "coordinates": [
        -114.43,
        43.07
      ]
    },
    "Elevation": 1525,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "30605087-bdee-e65c-bae1-602fe216fb0c"
  },
  {
    "Volcano Name": "Sibayak",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.52,
        3.2
      ]
    },
    "Elevation": 2212,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "7c88a50d-8f71-5662-189b-04862ca3b47e"
  },
  {
    "Volcano Name": "Sibualbuali",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.255,
        1.556
      ]
    },
    "Elevation": 1819,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "8cce99f9-df9e-e915-8540-e78503976506"
  },
  {
    "Volcano Name": "Silali",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.23,
        1.15
      ]
    },
    "Elevation": 1528,
    "Type": "Shield volcano",
    "Status": "Ar/Ar",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "87d6779a-da1a-cfda-1d22-9eca7abbf29e"
  },
  {
    "Volcano Name": "Silay",
    "Country": "Philippines",
    "Region": "Philippines-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.23,
        10.77
      ]
    },
    "Elevation": 1535,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bd971b56-21a6-d775-cb3f-916830951640"
  },
  {
    "Volcano Name": "Silverthrone",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -126.3,
        51.43
      ]
    },
    "Elevation": 3160,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3d0c96fb-b0cc-e459-abd4-579f08739d3e"
  },
  {
    "Volcano Name": "Simbo",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.52,
        -8.292
      ]
    },
    "Elevation": 335,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "7fbf13d1-4842-2502-9a8a-43ce8314078a"
  },
  {
    "Volcano Name": "Sinabung",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.392,
        3.17
      ]
    },
    "Elevation": 2460,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "11736eaf-7ad2-3f8c-e078-5fd5c3dbefd3"
  },
  {
    "Volcano Name": "Sinarka",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.175,
        48.875
      ]
    },
    "Elevation": 934,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "989cc91a-5ee3-b9a3-6c74-f850db843e76"
  },
  {
    "Volcano Name": "Singu Plateau",
    "Country": "Myanmar",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        95.98,
        22.7
      ]
    },
    "Elevation": 507,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b6bc5506-e9c6-d4bb-bf08-c7617c0a7f73"
  },
  {
    "Volcano Name": "Singuil, Cerro",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.7,
        14.03
      ]
    },
    "Elevation": 957,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b0878eb1-64fd-ee4f-d226-d5c05107c787"
  },
  {
    "Volcano Name": "Siple",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -126.67,
        -73.43
      ]
    },
    "Elevation": 3110,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "f49672a1-3bec-cd93-f08d-79d18ab12336"
  },
  {
    "Volcano Name": "Sirung",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.148,
        -8.51
      ]
    },
    "Elevation": 862,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2e004034-b1b8-3f9f-7ae5-a627c5ef6aa0"
  },
  {
    "Volcano Name": "Slamet",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.208,
        -7.242
      ]
    },
    "Elevation": 3432,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "27aa29af-8753-ee12-2fdf-e008b7d68e14"
  },
  {
    "Volcano Name": "Smirnov",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.13,
        44.43
      ]
    },
    "Elevation": 1189,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9938c5b3-f3f1-210d-c89a-120693508819"
  },
  {
    "Volcano Name": "Smith Rock",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.05,
        31.32
      ]
    },
    "Elevation": 136,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "80100fee-f18a-8038-1448-6f46120d03ca"
  },
  {
    "Volcano Name": "Smith Volcano",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.94,
        19.523
      ]
    },
    "Elevation": 2090,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "27c8e2be-b2ea-e947-9f7b-09c86f9bb03b"
  },
  {
    "Volcano Name": "Snaefellsjokull",
    "Country": "Iceland",
    "Region": "Iceland-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -23.78,
        64.8
      ]
    },
    "Elevation": 1448,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "637dc84f-386b-aab9-d793-dd9185de5267"
  },
  {
    "Volcano Name": "Snegovoy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.97,
        58.2
      ]
    },
    "Elevation": 2169,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b3abd5f6-7adb-534f-4925-b62e1958c622"
  },
  {
    "Volcano Name": "Snezhniy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.75,
        58.02
      ]
    },
    "Elevation": 2169,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d58ab5e3-e60e-912a-32ec-607cbafb757c"
  },
  {
    "Volcano Name": "Snowy",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.68,
        58.33
      ]
    },
    "Elevation": 2161,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "db14e9d7-be3b-c5ba-f0cb-33115463e1bb"
  },
  {
    "Volcano Name": "Soche",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.58,
        0.552
      ]
    },
    "Elevation": 3955,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "bd61b19f-a9f9-0eec-3bb0-9a97cbff25d6"
  },
  {
    "Volcano Name": "Socompa",
    "Country": "Argentina",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.25,
        -24.4
      ]
    },
    "Elevation": 6051,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "beca2ee1-f316-cc41-f51a-280f5a2dfab7"
  },
  {
    "Volcano Name": "Socorro",
    "Country": "Mexico",
    "Region": "Mexico-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -110.95,
        18.78
      ]
    },
    "Elevation": 1050,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9e44c80c-b49e-0cdd-78bd-822a4f0f6085"
  },
  {
    "Volcano Name": "Sodore",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.35,
        8.43
      ]
    },
    "Elevation": 1765,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "13bdd575-43ab-cead-58a6-f2aadee6a22e"
  },
  {
    "Volcano Name": "Sollipulli",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.52,
        -38.97
      ]
    },
    "Elevation": 2282,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "abf20d29-6e19-9b33-0fca-a6faa0c098f4"
  },
  {
    "Volcano Name": "Soputan",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.725,
        1.108
      ]
    },
    "Elevation": 1784,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "914fe974-ddcd-45a7-3a23-4b358da9cfc2"
  },
  {
    "Volcano Name": "Sorikmarapi",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.539,
        0.686
      ]
    },
    "Elevation": 2145,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "547f855b-39ad-5779-2b79-624f821ea12c"
  },
  {
    "Volcano Name": "Sorkale",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.725,
        13.18
      ]
    },
    "Elevation": 1611,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "d0660f3b-09b6-2d4b-30a4-945dc7fa0def"
  },
  {
    "Volcano Name": "Sotara",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.58,
        2.12
      ]
    },
    "Elevation": 4400,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5407f62e-4867-8a2a-cbae-8e936c3f0ee9"
  },
  {
    "Volcano Name": "Soufriere Guadeloupe",
    "Country": "Guadeloupe",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.67,
        16.05
      ]
    },
    "Elevation": 1467,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "80f3503c-734f-b6be-a0ce-8f02e25cb0de"
  },
  {
    "Volcano Name": "Soufriere Hills",
    "Country": "Montserrat",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -62.18,
        16.72
      ]
    },
    "Elevation": 915,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c12f41ac-14a2-4311-0c01-f9dd358cadcc"
  },
  {
    "Volcano Name": "Soufriere St. Vincent",
    "Country": "St. Vincent & the Grenadines",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.18,
        13.33
      ]
    },
    "Elevation": 1220,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5855d96d-23e0-cd70-7fbe-8ba5eed33a56"
  },
  {
    "Volcano Name": "South Island",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.6,
        2.63
      ]
    },
    "Elevation": 700,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "c1b0e14e-d2ae-186e-b665-1661b3cd51f2"
  },
  {
    "Volcano Name": "South Sister",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.77,
        44.1
      ]
    },
    "Elevation": 3157,
    "Type": "Complex volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "310bd621-0267-793f-9a1d-51021cab8e36"
  },
  {
    "Volcano Name": "Southern EPR-Segment I",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.417,
        -18.533
      ]
    },
    "Elevation": -2600,
    "Type": "Submarine volcano",
    "Status": "Magnetism",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3fd587c7-109f-9e8d-dc99-57e6440914a6"
  },
  {
    "Volcano Name": "Southern EPR-Segment J",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.35,
        -18.175
      ]
    },
    "Elevation": -2650,
    "Type": "Submarine volcano",
    "Status": "Magnetism",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "f310be4c-de6b-ba10-6543-81315188b6e6"
  },
  {
    "Volcano Name": "Southern EPR-Segment K",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.206,
        -17.436
      ]
    },
    "Elevation": -2566,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "12f4847f-3618-c829-63b2-cdabfdbef0cb"
  },
  {
    "Volcano Name": "Southern Sikhote-Alin",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        135.5,
        44.5
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d4b73004-5de2-b181-463f-ff1cd86196db"
  },
  {
    "Volcano Name": "Spectrum Range",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.68,
        57.43
      ]
    },
    "Elevation": 2430,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1b127712-dead-e36e-62d9-924996c07a37"
  },
  {
    "Volcano Name": "Spokoiny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.817,
        58.133
      ]
    },
    "Elevation": 2171,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "10d6402c-c5d5-624c-5489-8ec46cce325e"
  },
  {
    "Volcano Name": "Spurr",
    "Country": "United States",
    "Region": "Alaska-SW",
    "Location": {
      "type": "Point",
      "coordinates": [
        -152.25,
        61.3
      ]
    },
    "Elevation": 3374,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0176dcfc-b214-a3d4-ee64-840cf5680391"
  },
  {
    "Volcano Name": "Squaw Ridge Field",
    "Country": "United States",
    "Region": "US-Oregon",
    "Location": {
      "type": "Point",
      "coordinates": [
        -120.754,
        43.472
      ]
    },
    "Elevation": 1711,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "485523f1-cca4-3b8f-9175-6c025760508f"
  },
  {
    "Volcano Name": "Srednii",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.92,
        47.6
      ]
    },
    "Elevation": 36,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8fec45e1-ea50-fe4a-55c8-b7b54e6a80bf"
  },
  {
    "Volcano Name": "St. Andrew Strait",
    "Country": "Papua New Guinea",
    "Region": "Admiralty Is-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.35,
        -2.38
      ]
    },
    "Elevation": 270,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3f285c53-818b-550d-d82c-c331bce8f199"
  },
  {
    "Volcano Name": "St. Catherine",
    "Country": "Grenada",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.67,
        12.15
      ]
    },
    "Elevation": 840,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7f9d5733-659f-3935-eee2-dabbfb080910"
  },
  {
    "Volcano Name": "St. Helens",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.18,
        46.2
      ]
    },
    "Elevation": 2549,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "061fc6b9-7c77-7ed5-084c-5d3fe4e4f0a8"
  },
  {
    "Volcano Name": "St. Michael",
    "Country": "United States",
    "Region": "Alaska-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        -162.12,
        63.45
      ]
    },
    "Elevation": 715,
    "Type": "Shield volcano",
    "Status": "Anthropology",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8754f914-b955-367c-6fb4-e9f72bf8066a"
  },
  {
    "Volcano Name": "St. Paul",
    "Country": "Indian Ocean",
    "Region": "Indian O-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        77.53,
        -38.72
      ]
    },
    "Elevation": 268,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "081cc635-e728-35a7-d28f-cb923d0b6ced"
  },
  {
    "Volcano Name": "Steamboat Springs",
    "Country": "United States",
    "Region": "US-Nevada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -119.72,
        39.375
      ]
    },
    "Elevation": 1415,
    "Type": "Lava dome",
    "Status": "Pleistocene-Fumarol",
    "Last Known Eruption": "Quaternary eruption(s) with the only known Holocene activity being hydrothermal",
    "id": "c66cef0a-6d7a-6e61-73bf-9a1200aa3a55"
  },
  {
    "Volcano Name": "Steller",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -154.4,
        58.43
      ]
    },
    "Elevation": 2272,
    "Type": "Unknown",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "262a0189-4f4a-ba84-043f-599c600d5663"
  },
  {
    "Volcano Name": "Stepovak Bay 3",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -160.002,
        55.929
      ]
    },
    "Elevation": 1555,
    "Type": "Cinder cone",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "0fbe6377-99f3-5016-4c79-157d67e9d09b"
  },
  {
    "Volcano Name": "Stepovak Bay 4",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -159.954,
        55.954
      ]
    },
    "Elevation": 1557,
    "Type": "Stratovolcano",
    "Status": "Pleistocene",
    "Last Known Eruption": "P",
    "id": "555d9c1b-fb24-f647-58ab-4e1128ca2e7f"
  },
  {
    "Volcano Name": "Stromboli",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        15.213,
        38.789
      ]
    },
    "Elevation": 926,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "98dda97a-7d56-8cec-091a-5223795156ed"
  },
  {
    "Volcano Name": "Sturge Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        164.83,
        -67.4
      ]
    },
    "Elevation": 1167,
    "Type": "Stratovolcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "4ada4a28-ad21-c94a-aefa-89cb94454e8d"
  },
  {
    "Volcano Name": "Suchitan Volc Field",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.78,
        14.4
      ]
    },
    "Elevation": 2042,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b4d064a5-d27b-bf2f-8b09-89d11526dd5b"
  },
  {
    "Volcano Name": "Sukaria Caldera",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.77,
        -8.792
      ]
    },
    "Elevation": 1500,
    "Type": "Caldera",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d5ea057b-ed62-bb76-b811-3d1cf9b9770d"
  },
  {
    "Volcano Name": "Sumaco",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -77.626,
        -0.538
      ]
    },
    "Elevation": 3990,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "f07349b2-b5a6-7285-a52a-fc185e33b3e4"
  },
  {
    "Volcano Name": "Sumbing",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        101.728,
        -2.414
      ]
    },
    "Elevation": 2507,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a03ebbd2-a426-407b-ab36-0b29fe5c7eaa"
  },
  {
    "Volcano Name": "Sumbing",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.058,
        -7.38
      ]
    },
    "Elevation": 3371,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "116ea5e6-a4a6-b595-ebd7-405046040fdd"
  },
  {
    "Volcano Name": "Sumiyoshi-ike",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.594,
        31.768
      ]
    },
    "Elevation": 100,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "9a765fbf-fa57-cc84-6616-e40678c9aa42"
  },
  {
    "Volcano Name": "Sundoro",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.992,
        -7.3
      ]
    },
    "Elevation": 3136,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "40b9ce75-a9bf-686b-d8f1-01b12c083fb6"
  },
  {
    "Volcano Name": "Sunset Crater",
    "Country": "United States",
    "Region": "US-Arizona",
    "Location": {
      "type": "Point",
      "coordinates": [
        -111.5,
        35.37
      ]
    },
    "Elevation": 2447,
    "Type": "Cinder cone",
    "Status": "Dendrochronology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "4492fe70-2448-7cdc-82e9-9fa0e6ec5bf2"
  },
  {
    "Volcano Name": "Suoh",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        104.27,
        -5.25
      ]
    },
    "Elevation": 1000,
    "Type": "Maar",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b49fb06c-3998-dffc-e0e8-58443645763c"
  },
  {
    "Volcano Name": "Suphan Dagi",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.82,
        38.92
      ]
    },
    "Elevation": 4434,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9dfaf251-be03-a32e-4c77-eba0cfdd7e35"
  },
  {
    "Volcano Name": "Supply Reef",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.1,
        20.13
      ]
    },
    "Elevation": -8,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "42a3769c-3f87-39fb-9a7d-7a06f0bd5261"
  },
  {
    "Volcano Name": "Suretamatai",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        167.47,
        -13.8
      ]
    },
    "Elevation": 921,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7d2e9c51-87de-65f2-950b-a7c26b13139f"
  },
  {
    "Volcano Name": "Suswa",
    "Country": "Kenya",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.35,
        -1.175
      ]
    },
    "Elevation": 2356,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c8b7f8e9-667e-4cee-2e45-06c40a23429b"
  },
  {
    "Volcano Name": "Suwanose-jima",
    "Country": "Japan",
    "Region": "Ryukyu Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.72,
        29.53
      ]
    },
    "Elevation": 799,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "070280ab-b0f2-08da-3ab8-e3a44467fe19"
  },
  {
    "Volcano Name": "Ta'u",
    "Country": "United States",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.454,
        -14.23
      ]
    },
    "Elevation": 931,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e58327e1-94a5-1279-0bdf-3acefafbbf53"
  },
  {
    "Volcano Name": "Taal",
    "Country": "Philippines",
    "Region": "Luzon-Philippines",
    "Location": {
      "type": "Point",
      "coordinates": [
        120.993,
        14.002
      ]
    },
    "Elevation": 400,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6abc6a36-0ed3-eed1-97c9-557c3cdfc3ab"
  },
  {
    "Volcano Name": "Taburete",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.532,
        13.435
      ]
    },
    "Elevation": 1172,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "996e22e3-bb09-eba2-1274-366e4f604580"
  },
  {
    "Volcano Name": "Tacana",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -92.112,
        15.13
      ]
    },
    "Elevation": 4110,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "659fc494-8750-2101-54b6-721596ab723a"
  },
  {
    "Volcano Name": "Tacora",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.77,
        -17.72
      ]
    },
    "Elevation": 5980,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ec5b6f84-eb69-1662-5ac6-746164b827e5"
  },
  {
    "Volcano Name": "Taftan",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        61.6,
        28.6
      ]
    },
    "Elevation": 4050,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7e11d6d9-9247-0a47-a77d-3f401ebccf3a"
  },
  {
    "Volcano Name": "Tafu-Maka",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.233,
        -15.367
      ]
    },
    "Elevation": -1400,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ccff61d9-a9d9-13d8-501e-cfd3c3bd3315"
  },
  {
    "Volcano Name": "Tahalra Volc Field",
    "Country": "Algeria",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        5,
        22.667
      ]
    },
    "Elevation": 1467,
    "Type": "Pyroclastic cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4f07a3c9-b81d-7229-b900-62a022bdfeef"
  },
  {
    "Volcano Name": "Tahual",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -89.9,
        14.433
      ]
    },
    "Elevation": 1716,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c8af76ba-1b5a-75d9-21e7-9150ab24e4ad"
  },
  {
    "Volcano Name": "Tair, Jebel at",
    "Country": "Yemen",
    "Region": "Red Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.742,
        15.7
      ]
    },
    "Elevation": 244,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "af1342a2-37c0-b935-ed29-b219aeb1d2a8"
  },
  {
    "Volcano Name": "Tajumulco",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.903,
        15.034
      ]
    },
    "Elevation": 4220,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "840cb3ce-e95f-2b6b-9d74-77da5414a96f"
  },
  {
    "Volcano Name": "Takahara",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.78,
        36.9
      ]
    },
    "Elevation": 1795,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "U7",
    "id": "19158c67-ffa4-4ab5-c890-fca4154057a7"
  },
  {
    "Volcano Name": "Takahe",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.08,
        -76.28
      ]
    },
    "Elevation": 3460,
    "Type": "Shield volcano",
    "Status": "Ice Core",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "84f053a9-9682-7186-7089-da0138bdbb74"
  },
  {
    "Volcano Name": "Takawangha",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.02,
        51.87
      ]
    },
    "Elevation": 1449,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "afb46aa0-1611-480b-6415-ffd882962546"
  },
  {
    "Volcano Name": "Takuan Group",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.608,
        -6.442
      ]
    },
    "Elevation": 2210,
    "Type": "Volcanic complex",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b0f2c4fb-b831-2db2-4a1b-9649d4ff8467"
  },
  {
    "Volcano Name": "Talagabodas",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        108.07,
        -7.208
      ]
    },
    "Elevation": 2201,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3333482a-4d02-8d72-5984-fdfbf7af6a3d"
  },
  {
    "Volcano Name": "Talakmau",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.98,
        0.079
      ]
    },
    "Elevation": 2919,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2a00aacf-0b7b-f2b7-f0c1-dfef671cc565"
  },
  {
    "Volcano Name": "Talang",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        100.679,
        -0.978
      ]
    },
    "Elevation": 2597,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8f1bcb73-f9b1-75e7-f38b-84c45b8c8bbb"
  },
  {
    "Volcano Name": "Tambora",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        118,
        -8.25
      ]
    },
    "Elevation": 2850,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a80e466a-6921-6637-bacf-36548269ac9c"
  },
  {
    "Volcano Name": "Tampomas",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.95,
        -6.77
      ]
    },
    "Elevation": 1684,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "87bb4ac8-795c-a370-98b4-1541c52b7378"
  },
  {
    "Volcano Name": "Tana",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.767,
        52.833
      ]
    },
    "Elevation": 1170,
    "Type": "Stratovolcanoes",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "94b58ef7-4231-1b52-1375-344b9fe6fed9"
  },
  {
    "Volcano Name": "Tanaga",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -178.13,
        51.88
      ]
    },
    "Elevation": 1806,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "6507c256-1acc-d9f1-7fd7-994fd2f9c022"
  },
  {
    "Volcano Name": "Tandikat",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        100.317,
        -0.433
      ]
    },
    "Elevation": 2438,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "3a143008-b1c8-c370-52df-1dba0a032958"
  },
  {
    "Volcano Name": "Tangaroa",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        178.028,
        -36.321
      ]
    },
    "Elevation": 600,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c9074e38-6108-fd03-1644-18f16abdf03c"
  },
  {
    "Volcano Name": "Tangkubanparahu",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.6,
        -6.77
      ]
    },
    "Elevation": 2084,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ca2bec54-924f-aae3-9154-18680709aa6f"
  },
  {
    "Volcano Name": "Tao-Rusyr Caldera",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.7,
        49.35
      ]
    },
    "Elevation": 1325,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d7a2bef5-4022-6f3c-5885-a1c31add755d"
  },
  {
    "Volcano Name": "Tara, Batu",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.579,
        -7.792
      ]
    },
    "Elevation": 748,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ad988c74-9ac8-7743-c592-a39856fdf9df"
  },
  {
    "Volcano Name": "Tarakan",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.792,
        1.78
      ]
    },
    "Elevation": 318,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "78405a5d-2cc6-4002-a61e-adf64554dd4c"
  },
  {
    "Volcano Name": "Taryatu-Chulutu",
    "Country": "Mongolia",
    "Region": "Mongolia",
    "Location": {
      "type": "Point",
      "coordinates": [
        99.7,
        48.17
      ]
    },
    "Elevation": 2400,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "807d3265-7bb2-879f-0fdf-aeadce4ecde3"
  },
  {
    "Volcano Name": "Tat Ali",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.07,
        13.28
      ]
    },
    "Elevation": 700,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6179beb9-e4a9-4b76-1b06-eb213059f762"
  },
  {
    "Volcano Name": "Tate-yama",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        137.6,
        36.57
      ]
    },
    "Elevation": 2621,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "a9c87a40-5fbe-2398-a8d4-4d95ff3e3ddc"
  },
  {
    "Volcano Name": "Tateshina",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        138.3,
        36.1
      ]
    },
    "Elevation": 2530,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9d91ec2b-8975-38b6-838a-af1c55c63887"
  },
  {
    "Volcano Name": "Tatun Group",
    "Country": "China",
    "Region": "Taiwan",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.52,
        25.17
      ]
    },
    "Elevation": 1120,
    "Type": "Stratovolcano",
    "Status": "Pleistocene-Fumarol",
    "Last Known Eruption": "Quaternary eruption(s) with the only known Holocene activity being hydrothermal",
    "id": "cc1078e1-9af7-9452-6642-7109eab9e084"
  },
  {
    "Volcano Name": "Taunshits",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.8,
        54.53
      ]
    },
    "Elevation": 2353,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "0f76c38a-df7b-c399-d2ba-e0436b0d773e"
  },
  {
    "Volcano Name": "Taupo",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        176,
        -38.82
      ]
    },
    "Elevation": 760,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "737c31bf-432c-1c42-a198-efa7b91d1c17"
  },
  {
    "Volcano Name": "Taveuni",
    "Country": "Fiji",
    "Region": "Fiji Is-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -179.97,
        -16.82
      ]
    },
    "Elevation": 1241,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "2d02dab0-bc2e-169f-2a57-aab07ca17ed9"
  },
  {
    "Volcano Name": "Tavui",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.2,
        -4.117
      ]
    },
    "Elevation": 200,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "42a9740d-17fe-7c8e-6ba8-45da557a02a4"
  },
  {
    "Volcano Name": "Teahitia",
    "Country": "France",
    "Region": "Society Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -148.85,
        -17.57
      ]
    },
    "Elevation": -1600,
    "Type": "Submarine volcano",
    "Status": "Seismicity",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c106c0dd-67d7-6a2c-d344-d56c2be7e2f2"
  },
  {
    "Volcano Name": "Tecapa",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.502,
        13.494
      ]
    },
    "Elevation": 1593,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c13b2b07-97c0-c9a7-a634-5e6fe8a4cefd"
  },
  {
    "Volcano Name": "Tecuamburro",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -90.407,
        14.156
      ]
    },
    "Elevation": 1845,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "969c5630-945a-c3bb-9f20-8e0dfc676c1d"
  },
  {
    "Volcano Name": "Telica",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.845,
        12.602
      ]
    },
    "Elevation": 1061,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d2c07944-4635-d55d-5ada-1127ddbcc404"
  },
  {
    "Volcano Name": "Telomoyo",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.4,
        -7.37
      ]
    },
    "Elevation": 1894,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e80286c7-8491-c1dd-d138-3ad609151b55"
  },
  {
    "Volcano Name": "Telong, Bur Ni",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        96.808,
        4.77
      ]
    },
    "Elevation": 2624,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a427e831-e40f-23fb-9d56-1da91d9193a8"
  },
  {
    "Volcano Name": "Tenduruk Dagi",
    "Country": "Turkey",
    "Region": "Turkey",
    "Location": {
      "type": "Point",
      "coordinates": [
        43.83,
        39.33
      ]
    },
    "Elevation": 3584,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "313af8d2-447e-fb61-e7d7-dfe2d9742f2a"
  },
  {
    "Volcano Name": "Tenerife",
    "Country": "Spain",
    "Region": "Canary Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.641,
        28.271
      ]
    },
    "Elevation": 3715,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2e74774f-a58c-4f62-f87f-37e939d16c67"
  },
  {
    "Volcano Name": "Tengchong",
    "Country": "China",
    "Region": "China-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.47,
        25.32
      ]
    },
    "Elevation": 2865,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "fa98aaeb-bc32-c6ea-cc3e-2925d8629cc2"
  },
  {
    "Volcano Name": "Tengger Caldera",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        112.95,
        -7.942
      ]
    },
    "Elevation": 2329,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c13dfc07-f578-b490-7ed3-5def6434d909"
  },
  {
    "Volcano Name": "Tenorio",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.015,
        10.673
      ]
    },
    "Elevation": 1916,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1392a4e8-da69-f567-b0fa-72caebe804f4"
  },
  {
    "Volcano Name": "Teon",
    "Country": "Pacific Ocean",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        129.125,
        -6.92
      ]
    },
    "Elevation": 655,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "f4b7fc30-2075-5436-b21a-0f3e045b9421"
  },
  {
    "Volcano Name": "Tepi",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        35.43,
        7.42
      ]
    },
    "Elevation": 2728,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bc6ea179-d628-76d9-20ed-381366b9faeb"
  },
  {
    "Volcano Name": "Terceira",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.32,
        38.73
      ]
    },
    "Elevation": 1023,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "d54044a4-bd05-18ce-f28c-b498fa5aad0a"
  },
  {
    "Volcano Name": "Terpuk",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.833,
        57.2
      ]
    },
    "Elevation": 765,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "5566f66f-e026-8e0a-9aea-e5c05edead39"
  },
  {
    "Volcano Name": "Theistareykjarbunga",
    "Country": "Iceland",
    "Region": "Iceland",
    "Location": {
      "type": "Point",
      "coordinates": [
        -16.83,
        65.88
      ]
    },
    "Elevation": 564,
    "Type": "Shield volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "9e47b1a6-f620-58aa-a2bd-94a736871e7f"
  },
  {
    "Volcano Name": "Thompson Island",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        5.5,
        -53.93
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "5e2bfd59-586c-8df5-f7a8-fc41a01ba864"
  },
  {
    "Volcano Name": "Thordarhyrna",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.6,
        62.27
      ]
    },
    "Elevation": 1659,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "1ad49564-e458-904a-c8dd-d9f1a3797f9e"
  },
  {
    "Volcano Name": "Thule Islands",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.37,
        -59.45
      ]
    },
    "Elevation": 1075,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "13636bbf-74e3-6490-0193-da9f62b17b5d"
  },
  {
    "Volcano Name": "Tianshan Volc Group",
    "Country": "China",
    "Region": "China-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        82.5,
        42.5
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "de482919-06fc-fed4-33c8-46a51d95092e"
  },
  {
    "Volcano Name": "Tiatia",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.27,
        44.358
      ]
    },
    "Elevation": 1819,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "dfcecbe5-a3ad-7194-b16e-e9fb97ad1b58"
  },
  {
    "Volcano Name": "Ticsani",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.595,
        -16.755
      ]
    },
    "Elevation": 5408,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b9651b79-ade8-ec43-7637-e11a049c36b1"
  },
  {
    "Volcano Name": "Tidore",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.4,
        0.65
      ]
    },
    "Elevation": 1730,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e6032116-96ab-ab9d-e499-9eb677dd07a4"
  },
  {
    "Volcano Name": "Tigalalu",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.42,
        0.07
      ]
    },
    "Elevation": 422,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e8b7cb0c-04ce-d853-c6e3-aec67204b7ae"
  },
  {
    "Volcano Name": "Tigre, El",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.433,
        13.467
      ]
    },
    "Elevation": 1640,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1a531f3a-2686-8f13-a90a-33d3018d5bae"
  },
  {
    "Volcano Name": "Tigre, Isla El",
    "Country": "Honduras",
    "Region": "Honduras",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.63,
        13.27
      ]
    },
    "Elevation": 760,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a1d9bf50-da49-3ee4-6ff4-03099335c0ea"
  },
  {
    "Volcano Name": "Tin Zaouatene Volc Field",
    "Country": "Mali",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        2.833,
        19.833
      ]
    },
    "Elevation": null,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "da881451-0068-9511-56da-acd4db781101"
  },
  {
    "Volcano Name": "Tinakula",
    "Country": "Solomon Is.",
    "Region": "Santa Cruz Is-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        165.8,
        -10.38
      ]
    },
    "Elevation": 851,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "38b2b33b-258b-fe8a-d133-47c3ce8b7c90"
  },
  {
    "Volcano Name": "Tindfjallajokull",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.57,
        63.78
      ]
    },
    "Elevation": 1463,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "bf77fb39-dca4-ff1d-0567-477c1025c037"
  },
  {
    "Volcano Name": "Tinguiririca",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.352,
        -34.814
      ]
    },
    "Elevation": 4280,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "40ec290c-2333-a65e-4046-c46e0d131ad8"
  },
  {
    "Volcano Name": "Tipas",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.55,
        -27.2
      ]
    },
    "Elevation": 6660,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "523edf17-95cf-f884-f672-9e74f6610d96"
  },
  {
    "Volcano Name": "Tipas",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.55,
        -27.2
      ]
    },
    "Elevation": 6660,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "18c93625-d814-09ad-962a-1495f6803305"
  },
  {
    "Volcano Name": "Titila",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.1,
        57.4
      ]
    },
    "Elevation": 1559,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "4a59c853-92d7-160f-4fe3-f290c823adc6"
  },
  {
    "Volcano Name": "Tjornes Fracture Zone",
    "Country": "Iceland",
    "Region": "Iceland-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.1,
        66.3
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "0cf06905-12f4-adfa-1537-77beb4a96294"
  },
  {
    "Volcano Name": "Tlevak Strait-Suemez Is.",
    "Country": "United States",
    "Region": "Alaska-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -133.3,
        55.25
      ]
    },
    "Elevation": 50,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7b5cce9a-20de-b9ad-1b86-a10d7c47520e"
  },
  {
    "Volcano Name": "To-shima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        139.28,
        34.52
      ]
    },
    "Elevation": 508,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "89747892-f84b-2b6d-359d-fd445f7bcdf0"
  },
  {
    "Volcano Name": "Toba",
    "Country": "Indonesia",
    "Region": "Sumatra",
    "Location": {
      "type": "Point",
      "coordinates": [
        98.83,
        2.58
      ]
    },
    "Elevation": 2157,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "68aa281d-75e1-ee2a-eebf-4588a0684037"
  },
  {
    "Volcano Name": "Tobaru",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.675,
        1.625
      ]
    },
    "Elevation": 1035,
    "Type": "Unknown",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ca3c107c-f803-9e17-6598-fdfb13a61140"
  },
  {
    "Volcano Name": "Todoko-Ranu",
    "Country": "Indonesia",
    "Region": "Halmahera-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        127.43,
        1.22
      ]
    },
    "Elevation": 979,
    "Type": "Caldera",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a530fdf4-8361-2e8a-a468-04e41b753976"
  },
  {
    "Volcano Name": "Todra Volc Field",
    "Country": "Chad",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        8.5,
        17.683
      ]
    },
    "Elevation": 2000,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5becc54b-7d6d-482c-bcce-100babb36316"
  },
  {
    "Volcano Name": "Tofua",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.07,
        -19.75
      ]
    },
    "Elevation": 512,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "b59bf47d-7499-3db0-7092-e0f5efdce137"
  },
  {
    "Volcano Name": "Toh, Tarso",
    "Country": "Chad",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        16.333,
        21.333
      ]
    },
    "Elevation": 2000,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "66ad56f5-d90c-dd4c-d952-d4ca7940c255"
  },
  {
    "Volcano Name": "Tokachi",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.68,
        43.42
      ]
    },
    "Elevation": 2077,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "db24572c-bcab-ee87-3d75-f848dac13e84"
  },
  {
    "Volcano Name": "Tolbachik",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.33,
        55.83
      ]
    },
    "Elevation": 3682,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "08273ae6-dcf9-b9db-c541-af521c8235cc"
  },
  {
    "Volcano Name": "Tolguaca",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.645,
        -38.31
      ]
    },
    "Elevation": 2806,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "cdb26f8c-cce6-c099-a0c4-cafb73a9a67b"
  },
  {
    "Volcano Name": "Tolguaca",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.645,
        -38.31
      ]
    },
    "Elevation": 2806,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3de7c29c-ae32-1d28-a1f6-5a7b2c365f5a"
  },
  {
    "Volcano Name": "Tolima",
    "Country": "Colombia",
    "Region": "Colombia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -75.33,
        4.67
      ]
    },
    "Elevation": 5200,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "8f6989d1-ce6d-019a-e737-6066dc00614d"
  },
  {
    "Volcano Name": "Toliman",
    "Country": "Guatemala",
    "Region": "Guatemala",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.189,
        14.612
      ]
    },
    "Elevation": 3158,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b5aa3959-a50c-7fe8-4828-259320221a29"
  },
  {
    "Volcano Name": "Tolmachev Dol",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.58,
        52.63
      ]
    },
    "Elevation": 1021,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2a42d2c4-5bd2-2daf-5c63-24020741bbbd"
  },
  {
    "Volcano Name": "Toluca, Nevado de",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -99.758,
        19.108
      ]
    },
    "Elevation": 4690,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "b1728613-03f0-dd5e-412c-48263a81be19"
  },
  {
    "Volcano Name": "Tombel Graben",
    "Country": "Cameroon",
    "Region": "Africa-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        9.667,
        4.75
      ]
    },
    "Elevation": 500,
    "Type": "Cinder cones",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "984efc7c-a62f-c0a5-7152-4f445e812353"
  },
  {
    "Volcano Name": "Tondano Caldera",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.83,
        1.23
      ]
    },
    "Elevation": 1202,
    "Type": "Caldera",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7424f856-e670-1b83-aba4-01a8e7db7ef5"
  },
  {
    "Volcano Name": "Toney Mountain",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -115.83,
        -75.8
      ]
    },
    "Elevation": 3595,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "67d4fde3-a07b-948e-74ee-c906cc472bab"
  },
  {
    "Volcano Name": "Tongariro",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        175.642,
        -39.13
      ]
    },
    "Elevation": 1978,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "868b3e91-8af5-e8ad-a344-275bd4bf37bb"
  },
  {
    "Volcano Name": "Tongkoko",
    "Country": "Indonesia",
    "Region": "Sulawesi-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        125.2,
        1.52
      ]
    },
    "Elevation": 1149,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "88f7ee31-78c7-d554-3fb6-5cd9e7428a95"
  },
  {
    "Volcano Name": "Tore",
    "Country": "Papua New Guinea",
    "Region": "Bougainville-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        154.93,
        -5.83
      ]
    },
    "Elevation": 2200,
    "Type": "Lava cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f195c1bd-d820-a769-7189-084bcca986d8"
  },
  {
    "Volcano Name": "Torfajokull",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -19.17,
        63.92
      ]
    },
    "Elevation": 1259,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "1ea0cbe7-bbe2-4dc7-08ba-d047a454ebc9"
  },
  {
    "Volcano Name": "Tori-shima",
    "Country": "Japan",
    "Region": "Izu Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.32,
        30.48
      ]
    },
    "Elevation": 403,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "940716f5-6849-0806-f8a5-1d7206bc308e"
  },
  {
    "Volcano Name": "Toroeng Prong",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        108,
        14.93
      ]
    },
    "Elevation": 800,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "34e44830-63d4-1a5f-3130-6014280ceea0"
  },
  {
    "Volcano Name": "Tortuga, Isla",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -111.858,
        27.392
      ]
    },
    "Elevation": 210,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "932a016f-0193-e1d8-daf6-e67599a83cfc"
  },
  {
    "Volcano Name": "Tosa Sucha",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.57,
        5.92
      ]
    },
    "Elevation": 1650,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b7508c21-fb48-49bc-3b8f-d78a557819df"
  },
  {
    "Volcano Name": "Tousside, Tarso",
    "Country": "Chad",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        16.45,
        21.03
      ]
    },
    "Elevation": 3265,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "279fc354-c949-1365-d45e-cf7aad075dea"
  },
  {
    "Volcano Name": "Towada",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.92,
        40.47
      ]
    },
    "Elevation": 1159,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "fc4fd53e-9c70-ebe9-01d6-e4c2ed36096c"
  },
  {
    "Volcano Name": "Traitor's Head",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        169.23,
        -18.75
      ]
    },
    "Elevation": 837,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "e284a777-557d-e8ef-ff46-8dccd2d1de94"
  },
  {
    "Volcano Name": "Tres Virgenes",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -112.591,
        27.47
      ]
    },
    "Elevation": 1940,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "2e1d2c2b-14f8-7ba0-1b2e-67518f1ff980"
  },
  {
    "Volcano Name": "Tri Sestry",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.92,
        45.93
      ]
    },
    "Elevation": 998,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "3e60226b-b3c1-0775-b123-a30e375c1669"
  },
  {
    "Volcano Name": "Trident",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.08,
        58.23
      ]
    },
    "Elevation": 1864,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "c5bdc228-5611-1aae-8daa-9ae3f1448df6"
  },
  {
    "Volcano Name": "Trindade",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -29.331,
        -20.514
      ]
    },
    "Elevation": 600,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "548975f3-8311-3656-6084-49ce3182bbdb"
  },
  {
    "Volcano Name": "Tristan da Cunha",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -12.28,
        -37.092
      ]
    },
    "Elevation": 2060,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "8665cf73-f337-f7c2-a973-98e0bb06a57e"
  },
  {
    "Volcano Name": "Trocon",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.9,
        -37.733
      ]
    },
    "Elevation": 2500,
    "Type": "Lava domes",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "14a961b0-bfa4-6251-ae1b-ff39d101bfc2"
  },
  {
    "Volcano Name": "Trollagigar",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -18.13,
        64.43
      ]
    },
    "Elevation": 1000,
    "Type": "Crater rows",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "da553fdb-629b-774b-54f2-3f9d1ab12970"
  },
  {
    "Volcano Name": "Tromen",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.03,
        -37.142
      ]
    },
    "Elevation": 3978,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "0283131a-32ab-e81d-3b27-cb60faba4e1d"
  },
  {
    "Volcano Name": "Tronador",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.885,
        -41.157
      ]
    },
    "Elevation": 3491,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "02e035c7-7b06-4fef-8b9a-d7259274f64c"
  },
  {
    "Volcano Name": "Tseax River Cone",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -128.9,
        55.12
      ]
    },
    "Elevation": 609,
    "Type": "Pyroclastic cone",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "ef766f1d-e046-0178-ef6b-0ad89ccc23b6"
  },
  {
    "Volcano Name": "Tshibinda",
    "Country": "Congo, DRC",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        28.75,
        -2.32
      ]
    },
    "Elevation": 1460,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c3904a43-d9c2-6aff-b778-681de800463a"
  },
  {
    "Volcano Name": "Tskhouk-Karckar",
    "Country": "Armenia",
    "Region": "Armenia",
    "Location": {
      "type": "Point",
      "coordinates": [
        46.017,
        39.733
      ]
    },
    "Elevation": 3000,
    "Type": "Pyroclastic cones",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "07638532-5163-f91e-59aa-67b35d30366f"
  },
  {
    "Volcano Name": "Tsurumi",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        131.43,
        33.28
      ]
    },
    "Elevation": 1584,
    "Type": "Lava dome",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "b105d2e8-7e34-d6f6-3c95-4591cf99ee93"
  },
  {
    "Volcano Name": "Tujle, Cerro",
    "Country": "Chile",
    "Region": "Chile-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.95,
        -23.83
      ]
    },
    "Elevation": 3550,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c9baf586-cfc2-d095-533f-a69d063cf8ab"
  },
  {
    "Volcano Name": "Tulabug",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.613,
        -1.78
      ]
    },
    "Elevation": 3336,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "999daba7-c21d-f0d0-cf8a-5b1efeee174c"
  },
  {
    "Volcano Name": "Tullu Moje",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.13,
        8.158
      ]
    },
    "Elevation": 2349,
    "Type": "Pumice cone",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d13f506c-39a7-8ef5-016e-bf51b99a78ed"
  },
  {
    "Volcano Name": "Tumble Buttes",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.55,
        40.68
      ]
    },
    "Elevation": 2191,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "bac1f2b6-690a-cba9-c592-946e11c766de"
  },
  {
    "Volcano Name": "Tundroviy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.6,
        52.25
      ]
    },
    "Elevation": 739,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e41da98e-e5fe-8057-aa76-4819622543db"
  },
  {
    "Volcano Name": "Tungnafellsjokull",
    "Country": "Iceland",
    "Region": "Iceland-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -17.92,
        64.73
      ]
    },
    "Elevation": 1535,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "82aa45a5-43d9-7d91-ba6b-497e9c889d74"
  },
  {
    "Volcano Name": "Tungurahua",
    "Country": "Ecuador",
    "Region": "Ecuador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.442,
        -1.467
      ]
    },
    "Elevation": 5023,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "bba99db2-950b-14da-a4c2-32255d3e8901"
  },
  {
    "Volcano Name": "Tunkin Depression",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        102.5,
        51.5
      ]
    },
    "Elevation": 1200,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d7f46710-521a-aabe-ae1c-985e9c54adb2"
  },
  {
    "Volcano Name": "Tupungatito",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -69.8,
        -33.4
      ]
    },
    "Elevation": 6000,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ad621ecc-667f-90a5-2b60-fed960c18f21"
  },
  {
    "Volcano Name": "Turfan",
    "Country": "China",
    "Region": "China-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        89.25,
        42.9
      ]
    },
    "Elevation": 0,
    "Type": "Cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "301f3fb1-a731-0a11-8e04-0e38edb268a4"
  },
  {
    "Volcano Name": "Turrialba",
    "Country": "Costa Rica",
    "Region": "Costa Rica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -83.77,
        10.03
      ]
    },
    "Elevation": 3340,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "36b6bbbd-9848-4098-af4f-be828842117b"
  },
  {
    "Volcano Name": "Tutuila",
    "Country": "American Samoa",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -170.7,
        -14.295
      ]
    },
    "Elevation": 653,
    "Type": "Tuff cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "96e3cd03-1a7d-7a82-a0c0-950da8cbe884"
  },
  {
    "Volcano Name": "Tutupaca",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.358,
        -17.025
      ]
    },
    "Elevation": 5815,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "4914ef4b-8887-5e71-fdd1-3a573295b449"
  },
  {
    "Volcano Name": "Tuya Volc Field",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -130.583,
        59.367
      ]
    },
    "Elevation": 2123,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7482ec96-6984-d39a-2a17-6c9bb9c0d8c8"
  },
  {
    "Volcano Name": "Tuzgle, Cerro",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -66.483,
        -24.05
      ]
    },
    "Elevation": 5500,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "a1bc9a1e-bc6d-40aa-e945-fe3ea9db8c16"
  },
  {
    "Volcano Name": "Tuzovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.97,
        57.32
      ]
    },
    "Elevation": 1533,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7c98758e-fdc9-f271-31a7-238ef0a9c4f7"
  },
  {
    "Volcano Name": "Twin Buttes",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -121.6,
        40.78
      ]
    },
    "Elevation": 1631,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "56959757-5c15-bef1-bb2f-63eb897a2681"
  },
  {
    "Volcano Name": "Ubehebe Craters",
    "Country": "United States",
    "Region": "US-California",
    "Location": {
      "type": "Point",
      "coordinates": [
        -117.45,
        37.02
      ]
    },
    "Elevation": 752,
    "Type": "Maar",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9ba3e1b8-08ef-59d2-913b-2305b5dcdbbd"
  },
  {
    "Volcano Name": "Ubinas",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.903,
        -16.355
      ]
    },
    "Elevation": 5672,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "fd6fce01-cf9e-5df8-5563-db8500dc571a"
  },
  {
    "Volcano Name": "Udina",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.527,
        55.755
      ]
    },
    "Elevation": 2923,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c4a18b9b-f00b-43cf-0819-3bc0368ee35c"
  },
  {
    "Volcano Name": "Udokan Volc Field",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        117.47,
        56.18
      ]
    },
    "Elevation": 1980,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "3c52b992-b2d5-af57-d342-48e89883813a"
  },
  {
    "Volcano Name": "Ugashik-Peulik",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -156.37,
        57.75
      ]
    },
    "Elevation": 1474,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "2e62020e-2585-b3cd-af9c-b90720411c4b"
  },
  {
    "Volcano Name": "Uinkaret Field",
    "Country": "United States",
    "Region": "US-Arizona",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.13,
        36.38
      ]
    },
    "Elevation": 1555,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6c0501cf-0744-163e-0210-5b170efbd2e5"
  },
  {
    "Volcano Name": "Uka",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.58,
        57.7
      ]
    },
    "Elevation": 1643,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "9c8fcf66-c2cb-2c81-598d-2749bd2e1f51"
  },
  {
    "Volcano Name": "Ukinrek Maars",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -156.52,
        57.83
      ]
    },
    "Elevation": 91,
    "Type": "Maar",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "9c8dc7de-8a23-c4da-7874-4c801f6aed08"
  },
  {
    "Volcano Name": "Uksichan",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.38,
        56.08
      ]
    },
    "Elevation": 1692,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f7a9522b-a266-5ce9-385f-70613852b74d"
  },
  {
    "Volcano Name": "Ulawun",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.33,
        -5.05
      ]
    },
    "Elevation": 2334,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0224de39-100e-73a5-57ee-94ebd2c92300"
  },
  {
    "Volcano Name": "Uliaga",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.77,
        53.07
      ]
    },
    "Elevation": 888,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "b90bbc19-8dea-b7b3-dad7-9348ae413a32"
  },
  {
    "Volcano Name": "Ulreung",
    "Country": "South Korea",
    "Region": "Korea",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.87,
        37.5
      ]
    },
    "Elevation": 984,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f0555b60-7025-6009-6fc4-58f023d0fdcd"
  },
  {
    "Volcano Name": "Ulug-Arginsky",
    "Country": "Russia",
    "Region": "Russia-SE",
    "Location": {
      "type": "Point",
      "coordinates": [
        98,
        52.33
      ]
    },
    "Elevation": 1800,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "299f0614-73a8-a3df-f43d-9406c5adb6a4"
  },
  {
    "Volcano Name": "Umboi",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.875,
        -5.589
      ]
    },
    "Elevation": 1548,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "64a4f61d-0cc0-8536-db59-b68afcaa7d70"
  },
  {
    "Volcano Name": "Umm Arafieb, Jebel",
    "Country": "Sudan",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.83,
        18.17
      ]
    },
    "Elevation": 0,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e4cf93eb-35d2-ed02-ad0d-7ea5c7e3999b"
  },
  {
    "Volcano Name": "Ungaran",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        110.33,
        -7.18
      ]
    },
    "Elevation": 2050,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "1bbdbec3-0f91-d65e-2e8f-3b54a30cafb1"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.67,
        5.65
      ]
    },
    "Elevation": 1200,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5877a168-a14d-1cc7-b7e8-29077c8fb027"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Chile",
    "Region": "Chile-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -76.83,
        -33.62
      ]
    },
    "Elevation": -642,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "860bccf4-2365-d84c-2cf8-12f4aed0da32"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Arctic Ocean",
    "Region": "Arctic Ocean",
    "Location": {
      "type": "Point",
      "coordinates": [
        -65.6,
        88.27
      ]
    },
    "Elevation": -1500,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "cc0b6c25-4976-f342-379e-26e2e7d15cb2"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -38.08,
        38.75
      ]
    },
    "Elevation": -4200,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "2f7d40ea-43f7-d21e-1960-98a9de97b55a"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.83,
        39.95
      ]
    },
    "Elevation": -2835,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "216a13bf-69ba-a63d-07d2-0a03bf2d710e"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Portugal",
    "Region": "Azores",
    "Location": {
      "type": "Point",
      "coordinates": [
        -25.67,
        37.78
      ]
    },
    "Elevation": 350,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "912a6237-e382-e771-cf01-1c4e8315db84"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -163.83,
        23.58
      ]
    },
    "Elevation": -4000,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c9f2663c-2078-7337-38c5-ca3720748808"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Mexico",
    "Region": "Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -115,
        28
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "45a6d8e8-beda-3f25-4f9e-67e882d77b0e"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.48,
        26.13
      ]
    },
    "Elevation": -3200,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d846d20d-46ba-f75e-d7a9-eb4b401402c2"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.85,
        -4.75
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d3a2573c-7051-dc43-c7d9-f71de4ff4fac"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        161.75,
        55.92
      ]
    },
    "Elevation": 0,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "20191177-5046-f9e9-975a-f7086102f8fb"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Hawaiian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -158.75,
        21.75
      ]
    },
    "Elevation": -3000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "3e0d4edf-e708-2105-9ee3-7e5dcb0b57a3"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.9,
        21
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "1cf594bf-e6ae-04dd-2d30-e424e6094508"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Philippines",
    "Region": "Luzon-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.75,
        20.33
      ]
    },
    "Elevation": -24,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "e83a5fc1-b944-66d9-8269-7892ed49e1be"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.256,
        -4.311
      ]
    },
    "Elevation": -2000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "e27243e0-f1de-d0bb-7b17-f5e160311296"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -124.25,
        31.75
      ]
    },
    "Elevation": -2533,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "f5d3f598-c522-5c28-8353-cee7466d96d4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.28,
        46.47
      ]
    },
    "Elevation": -502,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "5f18253d-b767-c0ec-4a10-e862965f301c"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tanzania",
    "Region": "Africa-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        33.57,
        -8.63
      ]
    },
    "Elevation": 0,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c54db535-dc4b-bf08-2017-4e7caeb449cb"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -160,
        55.93
      ]
    },
    "Elevation": 1555,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5147a1c7-e960-7f8d-fc07-bb4395cd2907"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "China",
    "Region": "China-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        91.7,
        35.85
      ]
    },
    "Elevation": 5400,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "700d134b-01d7-9052-46a6-4bd8e929609b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.3,
        52.88
      ]
    },
    "Elevation": 700,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "267756b2-5eab-3a6b-a319-024b8c8cf8f8"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.5,
        46.1
      ]
    },
    "Elevation": -100,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "847efbb0-5495-30b5-5765-d6f3d54a83f6"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.208,
        45.03
      ]
    },
    "Elevation": -930,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "56226a49-b1a8-d132-1fe7-bba1384a5888"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.592,
        -16.992
      ]
    },
    "Elevation": 216,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "5023ed33-4d2f-63be-2e7c-dcf4815b095b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.2,
        20.3
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "03d219c0-ac6c-af98-fbd8-73c04158e163"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        156.55,
        51.6
      ]
    },
    "Elevation": 298,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a3aefd12-3acc-f488-4463-e37ac04b6f78"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -155.42,
        57.87
      ]
    },
    "Elevation": 300,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "721d7b9d-a4d7-a9d6-7dc4-49bb35dab376"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Indian Ocean",
    "Region": "Indian O-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        80.75,
        11.75
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "23a23b2b-1c55-5794-725a-5751d1dbdef3"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Indian Ocean",
    "Region": "Arabia-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        45,
        12.25
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "aac2986b-6972-895c-fb42-e285889fd1c7"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.52,
        52.92
      ]
    },
    "Elevation": 450,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f2eb01e3-d29b-26da-694f-a5000c8e0e1b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Admiralty Is-SW Paci",
    "Location": {
      "type": "Point",
      "coordinates": [
        147.78,
        -3.03
      ]
    },
    "Elevation": -1300,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ceb36dfa-8cba-5c67-6e7f-73316c37f198"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        60.67,
        28.17
      ]
    },
    "Elevation": 0,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "427ce1cb-e70b-c8b2-22ad-b867d29339ce"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Taiwan",
    "Region": "Taiwan-E of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.83,
        24
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ccf70f3c-c4c2-4f32-ca45-fdb84031c84c"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.425,
        32.658
      ]
    },
    "Elevation": 1436,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b77cb294-ee13-b305-ccb6-c5cce52235de"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Taiwan",
    "Region": "Taiwan-E of",
    "Location": {
      "type": "Point",
      "coordinates": [
        132.25,
        19.17
      ]
    },
    "Elevation": -10,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d9ee4480-d4a6-d8fc-e05b-faff4dd7a59d"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.925,
        33.308
      ]
    },
    "Elevation": 945,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e1659731-0de5-2822-8165-caf3464991a4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        37,
        36.67
      ]
    },
    "Elevation": 0,
    "Type": "Unknown",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "1c088041-d944-b61e-a3fd-c81bd15f0b8b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.02,
        52.57
      ]
    },
    "Elevation": 610,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3be8143f-8cac-ab04-db93-a87213cb4e84"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.33,
        52.33
      ]
    },
    "Elevation": 638,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ed28d17f-b86a-eccf-bd5d-49121ddd2569"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.87,
        -29.18
      ]
    },
    "Elevation": -560,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "02e5d3e8-8051-a426-8cca-4e3303954495"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -169.07,
        -14.23
      ]
    },
    "Elevation": -650,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "8d78af2d-e9e1-bf85-036b-6c269f5d4a0a"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Taiwan",
    "Region": "Taiwan-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.33,
        25.42
      ]
    },
    "Elevation": -100,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "69ce5586-37e0-bbb6-7a5d-caeed06504e1"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -34.5,
        49
      ]
    },
    "Elevation": -1650,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b96d63f0-d55c-d0e0-48f9-ab93c2e4815a"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        153.33,
        48.08
      ]
    },
    "Elevation": -150,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "712b0c4b-976b-736b-2f79-13b7f71dde5f"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        148.57,
        -5.2
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "00ddc7d6-7092-4aa7-77ce-79e9789547e0"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Syria",
    "Region": "Syria",
    "Location": {
      "type": "Point",
      "coordinates": [
        36.258,
        33.15
      ]
    },
    "Elevation": 1197,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "431d3a33-d7b0-d7b8-027c-65a7f365d7b0"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.65,
        -21.38
      ]
    },
    "Elevation": -500,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "94f5c23a-a85d-afb4-9e6a-fbbc4775c0a4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Indonesia",
    "Region": "Sangihe Is-Indonesia",
    "Location": {
      "type": "Point",
      "coordinates": [
        124.17,
        3.97
      ]
    },
    "Elevation": -5000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "d5165949-0f0c-a5d5-5447-ee51f934d7db"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.95,
        56.82
      ]
    },
    "Elevation": 1185,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "868994e9-0072-c333-ee6d-760b66a062f4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -140.3,
        -53.9
      ]
    },
    "Elevation": -1000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "3dfcbf7a-9151-6f8d-40f4-2b429373988d"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        168.63,
        -25.78
      ]
    },
    "Elevation": -2400,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "70bc9453-2fd6-1b88-965a-efe4f87080c4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Solomon Is.",
    "Region": "Solomon Is-SW Pacifi",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.03,
        -8.92
      ]
    },
    "Elevation": -240,
    "Type": "Submarine volcanoes",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "a33be3ee-c8ea-4295-f39c-9478f34fc19e"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.3,
        9.82
      ]
    },
    "Elevation": -2500,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2213ba55-587e-7455-4caa-70d8bc4df9ab"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Taiwan",
    "Region": "Taiwan-E of",
    "Location": {
      "type": "Point",
      "coordinates": [
        121.18,
        21.83
      ]
    },
    "Elevation": -115,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "6664352e-1009-21db-aa7e-9b2bb53e7965"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Taiwan",
    "Region": "Taiwan-E of",
    "Location": {
      "type": "Point",
      "coordinates": [
        134.75,
        20.93
      ]
    },
    "Elevation": -6000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "042647fc-f9f2-c993-646d-f8d536668de6"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.53,
        -20.85
      ]
    },
    "Elevation": -13,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "dcdb3c27-c81c-c52d-9f7f-6448f16c3ad3"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Iran",
    "Region": "Iran",
    "Location": {
      "type": "Point",
      "coordinates": [
        45.167,
        39.25
      ]
    },
    "Elevation": null,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "f9b7035c-8afe-be7d-31e2-8d0c435205ce"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -177.017,
        -24.8
      ]
    },
    "Elevation": -385,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "05b8c229-18da-593d-bab2-86859e5187e0"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.75,
        -21.15
      ]
    },
    "Elevation": -65,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "d878ebbe-2469-56bd-d6fc-92d7d4329355"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -174.365,
        -18.325
      ]
    },
    "Elevation": -40,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "0c355c2e-622a-0148-46ab-fc14aaf3c73c"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Japan",
    "Region": "Volcano Is-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        144.483,
        26.133
      ]
    },
    "Elevation": -3200,
    "Type": "Submarine volcano?",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "3d7ad5b4-3187-b195-ffb2-dffe4e3b2d76"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        142.9,
        21
      ]
    },
    "Elevation": null,
    "Type": "Submarine volcano?",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "d26b35a2-5008-9a60-544b-045b3041026c"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Mariana Is-C Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        143.2,
        20.3
      ]
    },
    "Elevation": null,
    "Type": "Submarine volcano?",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "12d0d7ff-f8a5-c7ac-00f0-6e9cabad5a1c"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        -124.25,
        31.75
      ]
    },
    "Elevation": -2533,
    "Type": "Submarine volcano?",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "39401597-f080-9645-7cdd-f8b1bf1b18a4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -103.583,
        10.733
      ]
    },
    "Elevation": null,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "e56701c6-9bec-5138-5c64-90a02d6be396"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -104.3,
        9.833
      ]
    },
    "Elevation": -2500,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "2e6afcc2-2043-d0cf-9ea6-f56537ee1afc"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -107.95,
        -8.267
      ]
    },
    "Elevation": -2800,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "215178d6-f9f1-9ea5-6fdf-b7ffe9e4d115"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -140.3,
        -53.9
      ]
    },
    "Elevation": -1000,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "5fbcb83d-df74-5913-93e8-99597eab966f"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Pacific Ocean",
    "Region": "Pacific-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -143.167,
        -55.967
      ]
    },
    "Elevation": null,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "5a4aea70-7ec1-3e7e-33b7-8e71d108e17d"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -68.267,
        -25.1
      ]
    },
    "Elevation": null,
    "Type": "Pyroclastic cone",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "154c2e67-4c6a-0e1f-7a3f-04b6aa6c1b4a"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Arctic Ocean",
    "Region": "Arctic Ocean",
    "Location": {
      "type": "Point",
      "coordinates": [
        85,
        85.583
      ]
    },
    "Elevation": -3800,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "71c45ab9-75aa-d9c8-9cf4-b3e689b7b216"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Georgia",
    "Region": "Georgia",
    "Location": {
      "type": "Point",
      "coordinates": [
        44.25,
        42.45
      ]
    },
    "Elevation": 3750,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fa3e7457-5b92-a270-8dc5-55d84964e2c1"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.83,
        7
      ]
    },
    "Elevation": -1415,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "6d4a36ae-b6b2-f884-40d5-de92a4a2199d"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.95,
        8.62
      ]
    },
    "Elevation": 1800,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b7388959-df5c-637c-e0e0-c4267b6edab9"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -15.83,
        -0.58
      ]
    },
    "Elevation": -1528,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "c7e0c203-69d0-d0c1-8dbd-d8f3474c4aa5"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        164.58,
        -73.45
      ]
    },
    "Elevation": 2987,
    "Type": "Scoria cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "d5a2ebae-c28e-d969-a0d5-f4aa77ff4c50"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.63,
        8.7
      ]
    },
    "Elevation": 1300,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "be1420d3-9457-270a-f7a4-39da268863d2"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        163,
        -76.83
      ]
    },
    "Elevation": -500,
    "Type": "Submarine volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "ce238462-8e64-76b9-f495-64d9c7c68827"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        39.07,
        8.07
      ]
    },
    "Elevation": 1800,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c570f56f-6ee7-9e38-487a-6ac5dc8d6343"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        38.93,
        7.95
      ]
    },
    "Elevation": 1889,
    "Type": "Fissure vent",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "2d38ad47-5362-c7ad-c292-709c7ebee996"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -24.5,
        -3.5
      ]
    },
    "Elevation": -5300,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "74a45712-6c31-3783-5166-e1f42f55c189"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -21.45,
        4.2
      ]
    },
    "Elevation": -2900,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "1c7f4c46-54e0-d711-a5ca-fdc883aa721b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Georgia",
    "Region": "Georgia",
    "Location": {
      "type": "Point",
      "coordinates": [
        43.6,
        41.55
      ]
    },
    "Elevation": 3400,
    "Type": "Lava cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "58f021cf-c9b5-7530-5ec3-d9467feb865f"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -175.33,
        -21.07
      ]
    },
    "Elevation": 0,
    "Type": "Not Volcanic",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "8bfa69b5-2c10-2815-5afd-3439e80cc26b"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -173.5,
        52
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Hydrophonic",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "785caf22-a9b0-ba35-5fe8-cc8feba07857"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Chile",
    "Region": "Chile-Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -78.78,
        -33.622
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "595d79ae-c899-6f26-e8f9-532712a09fb4"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Netherlands",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.28,
        16.13
      ]
    },
    "Elevation": -7,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "8bef5210-7ba2-58e7-6ff0-97df8869adf8"
  },
  {
    "Volcano Name": "Unnamed",
    "Country": "Atlantic Ocean",
    "Region": "Atlantic-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.75,
        66
      ]
    },
    "Elevation": -108,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "eb4332fa-4e5f-688f-340f-bd7043dfc07a"
  },
  {
    "Volcano Name": "Unzen",
    "Country": "Japan",
    "Region": "Kyushu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        130.3,
        32.75
      ]
    },
    "Elevation": 1500,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "37f99e99-195a-262b-b384-27dfde2f9835"
  },
  {
    "Volcano Name": "Upolu",
    "Country": "Samoa",
    "Region": "Samoa-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -171.72,
        -13.935
      ]
    },
    "Elevation": 1100,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "059b89f9-c74a-6130-de2e-111080aef2cc"
  },
  {
    "Volcano Name": "Urataman",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.23,
        47.12
      ]
    },
    "Elevation": 678,
    "Type": "Somma volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "729c9e88-348a-87d7-3a80-10f4fcb17f47"
  },
  {
    "Volcano Name": "Ushishur",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        152.8,
        47.52
      ]
    },
    "Elevation": 401,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "96a72858-c84f-82c6-b929-fdfe3e1293bf"
  },
  {
    "Volcano Name": "Ushkovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.47,
        56.105
      ]
    },
    "Elevation": 3943,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "b2a3e073-ae84-e7d2-c91f-c5fde3a48591"
  },
  {
    "Volcano Name": "Usu",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.83,
        42.53
      ]
    },
    "Elevation": 731,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "7a9474bd-a8fa-07e1-5133-bfd954cc29a5"
  },
  {
    "Volcano Name": "Usulutan",
    "Country": "El Salvador",
    "Region": "El Salvador",
    "Location": {
      "type": "Point",
      "coordinates": [
        -88.471,
        13.419
      ]
    },
    "Elevation": 1449,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5e4baa7b-7f7d-ce94-1704-d2a4bff13333"
  },
  {
    "Volcano Name": "Utila Island",
    "Country": "Honduras",
    "Region": "Honduras",
    "Location": {
      "type": "Point",
      "coordinates": [
        -86.9,
        16.1
      ]
    },
    "Elevation": 90,
    "Type": "Pyroclastic cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "38a07bbd-9c15-7693-d704-05d08a8cbc47"
  },
  {
    "Volcano Name": "Uwayrid, Harrat",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        37.25,
        27.08
      ]
    },
    "Elevation": 1900,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption from A.D. 1-1499, inclusive",
    "id": "bcfbc5ae-609c-6d0a-a1eb-30c31cecae12"
  },
  {
    "Volcano Name": "Uzon",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.97,
        54.5
      ]
    },
    "Elevation": 1617,
    "Type": "Caldera",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "06631956-4ae2-e761-76d4-4e9caf92fd4b"
  },
  {
    "Volcano Name": "Vakak Group",
    "Country": "Afghanistan",
    "Region": "Afghanistan",
    "Location": {
      "type": "Point",
      "coordinates": [
        67.97,
        34.25
      ]
    },
    "Elevation": 3190,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "aa0e6565-234b-7640-62d0-76fd63ddb081"
  },
  {
    "Volcano Name": "Valle, El",
    "Country": "Panama",
    "Region": "Panama",
    "Location": {
      "type": "Point",
      "coordinates": [
        -80.167,
        8.583
      ]
    },
    "Elevation": 1185,
    "Type": "Stratovolcano",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "081b9e46-aa6f-a1b1-dd71-7ab5eecc9032"
  },
  {
    "Volcano Name": "Veniaminof",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -159.38,
        56.17
      ]
    },
    "Elevation": 2507,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8fdbe0a8-9bda-f088-6ff6-3714e0f474ef"
  },
  {
    "Volcano Name": "Verkhovoy",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.53,
        56.52
      ]
    },
    "Elevation": 1400,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "b2cfbe27-d687-df3e-397b-c9be2de99236"
  },
  {
    "Volcano Name": "Vernadskii Ridge",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        155.97,
        50.55
      ]
    },
    "Elevation": 1183,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "defc7a1c-261a-35c6-df8f-65777adb6bc5"
  },
  {
    "Volcano Name": "Vestmannaeyjar",
    "Country": "Iceland",
    "Region": "Iceland-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -20.28,
        63.43
      ]
    },
    "Elevation": 279,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "b419946b-2001-6d86-9afb-a78f80aef29d"
  },
  {
    "Volcano Name": "Vesuvius",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.426,
        40.821
      ]
    },
    "Elevation": 1281,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "56576153-4967-6a1a-dae7-9e5ce33a241f"
  },
  {
    "Volcano Name": "Veteran",
    "Country": "Vietnam",
    "Region": "SE Asia",
    "Location": {
      "type": "Point",
      "coordinates": [
        109.05,
        9.83
      ]
    },
    "Elevation": 0,
    "Type": "Submarine volcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3d46f0b6-cc76-d3a2-b0fe-83bb754edc89"
  },
  {
    "Volcano Name": "Victory",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.07,
        -9.2
      ]
    },
    "Elevation": 1925,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "67f89b76-9344-7fb7-4d53-8a5156ba3b09"
  },
  {
    "Volcano Name": "Viedma, Volcan",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -73.28,
        -49.358
      ]
    },
    "Elevation": 1300,
    "Type": "Subglacial volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "a7fe5a71-c446-a7f9-e3ee-b63d8702dc38"
  },
  {
    "Volcano Name": "Villarrica",
    "Country": "Chile",
    "Region": "Chile-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.93,
        -39.42
      ]
    },
    "Elevation": 2847,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f025cb73-763a-003e-802c-16a3cffb0437"
  },
  {
    "Volcano Name": "Vilyuchik",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.3,
        52.68
      ]
    },
    "Elevation": 2173,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ce248c20-f1ac-a1a4-39ef-4ea409a52cd7"
  },
  {
    "Volcano Name": "Visoke",
    "Country": "Rwanda",
    "Region": "Africa-C",
    "Location": {
      "type": "Point",
      "coordinates": [
        29.492,
        -1.47
      ]
    },
    "Elevation": 3711,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "35e5d984-077b-9814-7cd5-1d309bbe043e"
  },
  {
    "Volcano Name": "Volcanico, Cerro",
    "Country": "Argentina",
    "Region": "Argentina",
    "Location": {
      "type": "Point",
      "coordinates": [
        -71.65,
        -42.07
      ]
    },
    "Elevation": 0,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "c24c527f-9ef2-fce5-1597-8b4825fe1987"
  },
  {
    "Volcano Name": "Volcano W",
    "Country": "New Zealand",
    "Region": "Kermadec Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -179.183,
        -31.85
      ]
    },
    "Elevation": -900,
    "Type": "Submarine volcanoes",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e4925f82-a4c9-18ce-92c0-3c57ceea5f2e"
  },
  {
    "Volcano Name": "Voon, Tarso",
    "Country": "Chad",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        17.28,
        20.92
      ]
    },
    "Elevation": 3100,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "639aed71-4e83-c2b6-589a-622d030681eb"
  },
  {
    "Volcano Name": "Voyampolsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.62,
        58.37
      ]
    },
    "Elevation": 1225,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "45039cd4-1304-40fc-ed8d-29eb730c4290"
  },
  {
    "Volcano Name": "Vsevidof",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -168.68,
        53.13
      ]
    },
    "Elevation": 2149,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "f9b0668a-f2d0-ab21-9c2b-240b9b19a0c1"
  },
  {
    "Volcano Name": "Vulcano",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        14.962,
        38.404
      ]
    },
    "Elevation": 500,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "457a90ab-bbb3-9487-e583-aa27adfdd4ed"
  },
  {
    "Volcano Name": "Vulsini",
    "Country": "Italy",
    "Region": "Italy",
    "Location": {
      "type": "Point",
      "coordinates": [
        11.93,
        42.6
      ]
    },
    "Elevation": 800,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "6a6957a1-4994-35f1-e272-7170eddf7d75"
  },
  {
    "Volcano Name": "Waesche",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -126.88,
        -77.17
      ]
    },
    "Elevation": 3292,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "5781259a-3390-4d9e-ba85-421244837fc2"
  },
  {
    "Volcano Name": "Waiowa",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        149.075,
        -9.57
      ]
    },
    "Elevation": 640,
    "Type": "Pyroclastic cone",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "2bb49be5-2bbd-774c-089c-cbd5e32b0087"
  },
  {
    "Volcano Name": "Wallis Islands",
    "Country": "Wallis & Futuna",
    "Region": "SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -176.17,
        -13.3
      ]
    },
    "Elevation": 143,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "6995bf3a-ce0a-4080-2054-dc1651bf0ee8"
  },
  {
    "Volcano Name": "Walo",
    "Country": "Papua New Guinea",
    "Region": "New Britain-SW Pac",
    "Location": {
      "type": "Point",
      "coordinates": [
        150.9,
        -5.53
      ]
    },
    "Elevation": 15,
    "Type": "Hydrothermal field",
    "Status": "Hot Springs",
    "Last Known Eruption": "Unknown",
    "id": "67f32431-0715-a5df-408a-e051a5587c0a"
  },
  {
    "Volcano Name": "Wapi Lava Field",
    "Country": "United States",
    "Region": "US-Idaho",
    "Location": {
      "type": "Point",
      "coordinates": [
        -113.22,
        42.88
      ]
    },
    "Elevation": 1604,
    "Type": "Shield volcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "4854b971-35ac-620d-bbf3-0f50f854374f"
  },
  {
    "Volcano Name": "Washiba-Kumonotaira",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        137.5,
        36.4
      ]
    },
    "Elevation": 2924,
    "Type": "Shield volcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "f3d86fa1-fb37-66ed-c217-56012ef20b51"
  },
  {
    "Volcano Name": "Watt, Morne",
    "Country": "Dominica",
    "Region": "W Indies",
    "Location": {
      "type": "Point",
      "coordinates": [
        -61.305,
        15.307
      ]
    },
    "Elevation": 1224,
    "Type": "Stratovolcanoes",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "56295d8f-eada-5c1c-8d80-cd7b1cc72df0"
  },
  {
    "Volcano Name": "Wau-en-Namus",
    "Country": "Libya",
    "Region": "Africa-N",
    "Location": {
      "type": "Point",
      "coordinates": [
        17.55,
        25.05
      ]
    },
    "Elevation": 547,
    "Type": "Caldera",
    "Status": "Holocene?",
    "Last Known Eruption": "Uncertain Holocene eruption",
    "id": "421d2b16-b993-b777-344c-2719a528299b"
  },
  {
    "Volcano Name": "Wayang-Windu",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        107.63,
        -7.208
      ]
    },
    "Elevation": 2182,
    "Type": "Lava dome",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "ccf8a5fe-ba32-b12d-6183-100c2f3c9575"
  },
  {
    "Volcano Name": "Wells Gray-Clearwater",
    "Country": "Canada",
    "Region": "Canada",
    "Location": {
      "type": "Point",
      "coordinates": [
        -120.57,
        52.33
      ]
    },
    "Elevation": 2015,
    "Type": "Cinder cone",
    "Status": "Dendrochronology",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "ab479adc-6931-e0cf-18c4-043c58613600"
  },
  {
    "Volcano Name": "West Crater",
    "Country": "United States",
    "Region": "US-Washington",
    "Location": {
      "type": "Point",
      "coordinates": [
        -122.08,
        45.88
      ]
    },
    "Elevation": 1329,
    "Type": "Volcanic field",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "735bb47a-4573-dfe6-1aa6-38f15c510b45"
  },
  {
    "Volcano Name": "West Eifel Volc Field",
    "Country": "Germany",
    "Region": "Germany",
    "Location": {
      "type": "Point",
      "coordinates": [
        6.85,
        50.17
      ]
    },
    "Elevation": 600,
    "Type": "Maar",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "2b0d37cb-0f38-d94e-c7b3-758bb9b10acd"
  },
  {
    "Volcano Name": "West Mata",
    "Country": "Tonga",
    "Region": "Tonga-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        -173.75,
        -15.1
      ]
    },
    "Elevation": -1174,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "f18e6940-d34d-4988-faf7-8f594acca595"
  },
  {
    "Volcano Name": "Westdahl",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -164.65,
        54.52
      ]
    },
    "Elevation": 1654,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "8a762c62-105e-ce96-32e7-83e2608e6d50"
  },
  {
    "Volcano Name": "Whangarei",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        174.27,
        -35.75
      ]
    },
    "Elevation": 397,
    "Type": "Cinder cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "e79b90dc-328c-46a6-46a1-a596adc46b70"
  },
  {
    "Volcano Name": "White Island",
    "Country": "New Zealand",
    "Region": "New Zealand",
    "Location": {
      "type": "Point",
      "coordinates": [
        177.18,
        -37.52
      ]
    },
    "Elevation": 321,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6337e403-e785-757a-788d-19424ef19404"
  },
  {
    "Volcano Name": "Wilis",
    "Country": "Indonesia",
    "Region": "Java",
    "Location": {
      "type": "Point",
      "coordinates": [
        111.758,
        -7.808
      ]
    },
    "Elevation": 2563,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "7e094d32-3366-5a88-1d99-7b556f93f6b7"
  },
  {
    "Volcano Name": "Wolf, Volcan",
    "Country": "Ecuador",
    "Region": "Galapagos",
    "Location": {
      "type": "Point",
      "coordinates": [
        -91.35,
        0.02
      ]
    },
    "Elevation": 1710,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "6380806e-7555-3f0d-4abd-dbbf52820e72"
  },
  {
    "Volcano Name": "Wrangell",
    "Country": "United States",
    "Region": "Alaska-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        -144.02,
        62
      ]
    },
    "Elevation": 4317,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "fe6758bb-1af3-c69f-76ad-2eacd4feb85a"
  },
  {
    "Volcano Name": "Wudalianchi",
    "Country": "China",
    "Region": "China-E",
    "Location": {
      "type": "Point",
      "coordinates": [
        126.12,
        48.72
      ]
    },
    "Elevation": 597,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "d3d69e33-ad89-c08e-4edf-475e0f736537"
  },
  {
    "Volcano Name": "Wurlali",
    "Country": "Indonesia",
    "Region": "Banda Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        128.675,
        -7.125
      ]
    },
    "Elevation": 868,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "62e6db65-22e1-313c-4d9c-3ad96ae312f8"
  },
  {
    "Volcano Name": "Xianjindao",
    "Country": "North Korea",
    "Region": "Korea",
    "Location": {
      "type": "Point",
      "coordinates": [
        128,
        41.33
      ]
    },
    "Elevation": 0,
    "Type": "Unknown",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1500-1699, inclusive",
    "id": "078e5ba7-f0ce-a2b4-2760-ae5e0f8eff64"
  },
  {
    "Volcano Name": "Yake-dake",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        137.58,
        36.22
      ]
    },
    "Elevation": 2455,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "ff9ae17b-63ca-29fe-5026-7cbbf3052657"
  },
  {
    "Volcano Name": "Yali",
    "Country": "Greece",
    "Region": "Greece",
    "Location": {
      "type": "Point",
      "coordinates": [
        27.1,
        36.63
      ]
    },
    "Elevation": 176,
    "Type": "Lava dome",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "474176f8-ebdb-d88b-775b-1649484e0afb"
  },
  {
    "Volcano Name": "Yangudi",
    "Country": "Ethiopia",
    "Region": "Africa-NE",
    "Location": {
      "type": "Point",
      "coordinates": [
        41.042,
        10.58
      ]
    },
    "Elevation": 1383,
    "Type": "Complex volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "8664b806-a7c4-5a62-88f7-3b7db81b0a28"
  },
  {
    "Volcano Name": "Yantarni",
    "Country": "United States",
    "Region": "Alaska Peninsula",
    "Location": {
      "type": "Point",
      "coordinates": [
        -157.18,
        57.02
      ]
    },
    "Elevation": 1336,
    "Type": "Stratovolcano",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "e28f89f3-c163-1772-d266-bc9b41d3987a"
  },
  {
    "Volcano Name": "Yanteles",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.8,
        -43.5
      ]
    },
    "Elevation": 2042,
    "Type": "Stratovolcanoes",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "371eac75-9c3c-c07b-add6-17da4a88a029"
  },
  {
    "Volcano Name": "Yanteles, Cerro",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.83,
        -43.42
      ]
    },
    "Elevation": 2050,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "11539b95-b86d-a724-c73c-9044b546eb77"
  },
  {
    "Volcano Name": "Yar, Jabal",
    "Country": "Saudi Arabia",
    "Region": "Arabia-W",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.83,
        17.05
      ]
    },
    "Elevation": 305,
    "Type": "Volcanic field",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "0b48590c-0ce2-6fdf-986c-34c2ac1cfa1d"
  },
  {
    "Volcano Name": "Yasur",
    "Country": "Vanuatu",
    "Region": "Vanuatu-SW Pacific",
    "Location": {
      "type": "Point",
      "coordinates": [
        169.425,
        -19.52
      ]
    },
    "Elevation": 361,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "3f7b6334-dbb2-a6f7-011f-9041f90ac420"
  },
  {
    "Volcano Name": "Yate",
    "Country": "Chile",
    "Region": "Chile-S",
    "Location": {
      "type": "Point",
      "coordinates": [
        -72.396,
        -41.755
      ]
    },
    "Elevation": 2187,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "12d7ab60-6ed8-a4e8-4ab6-2829424ddcfe"
  },
  {
    "Volcano Name": "Yelia",
    "Country": "Papua New Guinea",
    "Region": "New Guinea",
    "Location": {
      "type": "Point",
      "coordinates": [
        145.858,
        -7.05
      ]
    },
    "Elevation": 3384,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "a958af91-4caa-f068-5b3e-b41991c88da3"
  },
  {
    "Volcano Name": "Yellowstone",
    "Country": "United States",
    "Region": "US-Wyoming",
    "Location": {
      "type": "Point",
      "coordinates": [
        -110.67,
        44.43
      ]
    },
    "Elevation": 2805,
    "Type": "Caldera",
    "Status": "Tephrochronology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ac3d8b26-6dbc-0ad9-df91-13052bd73281"
  },
  {
    "Volcano Name": "Yersey",
    "Country": "Indonesia",
    "Region": "Lesser Sunda Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        123.95,
        -7.53
      ]
    },
    "Elevation": -3800,
    "Type": "Submarine volcano",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "0626980d-0c69-23e7-88d6-0f5bd39df7ee"
  },
  {
    "Volcano Name": "Yojoa, Lake",
    "Country": "Honduras",
    "Region": "Honduras",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.98,
        14.98
      ]
    },
    "Elevation": 1090,
    "Type": "Volcanic field",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "fd2a6b39-8c55-6272-6531-9558a2751610"
  },
  {
    "Volcano Name": "Yomba",
    "Country": "Papua New Guinea",
    "Region": "New Guinea-NE of",
    "Location": {
      "type": "Point",
      "coordinates": [
        146.75,
        -4.92
      ]
    },
    "Elevation": 0,
    "Type": "Unknown",
    "Status": "Uncertain",
    "Last Known Eruption": "Unknown",
    "id": "6a4a54f6-bae2-26cc-f01b-1f7b75d547d7"
  },
  {
    "Volcano Name": "Yotei",
    "Country": "Japan",
    "Region": "Hokkaido-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.82,
        42.83
      ]
    },
    "Elevation": 1893,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "ba34402c-0f5d-03d1-2b26-f4973116f1bb"
  },
  {
    "Volcano Name": "Young Island",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        162.45,
        -66.42
      ]
    },
    "Elevation": 1340,
    "Type": "Stratovolcano",
    "Status": "Fumarolic",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3ae7194e-f642-86dc-a2de-b46b5dbe47a6"
  },
  {
    "Volcano Name": "Yucamane",
    "Country": "Peru",
    "Region": "Peru",
    "Location": {
      "type": "Point",
      "coordinates": [
        -70.2,
        -17.18
      ]
    },
    "Elevation": 5550,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1700-1799, inclusive",
    "id": "23a49a77-a6e2-9178-6162-fa71afdd18e5"
  },
  {
    "Volcano Name": "Yumia, Cerro",
    "Country": "Bolivia",
    "Region": "Bolivia",
    "Location": {
      "type": "Point",
      "coordinates": [
        -67.5,
        -21.5
      ]
    },
    "Elevation": 4050,
    "Type": "Cone",
    "Status": "Holocene",
    "Last Known Eruption": "Unknown",
    "id": "797005e7-9909-0f1b-a22c-d8bc71bbc7bd"
  },
  {
    "Volcano Name": "Yunaska",
    "Country": "United States",
    "Region": "Aleutian Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        -170.63,
        52.63
      ]
    },
    "Elevation": 550,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "a8ba1c5f-60eb-a12a-ac23-6878d017b0ee"
  },
  {
    "Volcano Name": "Zacate Grande, Isla",
    "Country": "Honduras",
    "Region": "Honduras",
    "Location": {
      "type": "Point",
      "coordinates": [
        -87.63,
        13.33
      ]
    },
    "Elevation": 600,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "5091d78e-1217-194d-9001-d0af92bc0124"
  },
  {
    "Volcano Name": "Zao",
    "Country": "Japan",
    "Region": "Honshu-Japan",
    "Location": {
      "type": "Point",
      "coordinates": [
        140.45,
        38.15
      ]
    },
    "Elevation": 1841,
    "Type": "Complex volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "da58c21a-96db-7e49-be0f-953a7e66923f"
  },
  {
    "Volcano Name": "Zaozerny",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.95,
        56.88
      ]
    },
    "Elevation": 1349,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "e9394bb1-6bb4-aeca-412a-79fe1b431883"
  },
  {
    "Volcano Name": "Zapatera",
    "Country": "Nicaragua",
    "Region": "Nicaragua",
    "Location": {
      "type": "Point",
      "coordinates": [
        -85.82,
        11.73
      ]
    },
    "Elevation": 629,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3b25ab8d-6d05-7d88-c63a-a96cf1c2d7c8"
  },
  {
    "Volcano Name": "Zavaritsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.385,
        53.905
      ]
    },
    "Elevation": 1567,
    "Type": "Cinder cones",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "51489dad-3030-6dfa-4e27-6e3dfdc61976"
  },
  {
    "Volcano Name": "Zavaritsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        158.385,
        53.905
      ]
    },
    "Elevation": 1567,
    "Type": "Stratovolcano",
    "Status": "Radiocarbon",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "8711961f-9ed8-548f-949d-688d99204c78"
  },
  {
    "Volcano Name": "Zavaritzki Caldera",
    "Country": "Russia",
    "Region": "Kuril Is",
    "Location": {
      "type": "Point",
      "coordinates": [
        151.95,
        46.925
      ]
    },
    "Elevation": 624,
    "Type": "Caldera",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "c467bfc1-f06f-e019-c4bc-e2c21eea4257"
  },
  {
    "Volcano Name": "Zavodovski",
    "Country": "Antarctica",
    "Region": "Antarctica",
    "Location": {
      "type": "Point",
      "coordinates": [
        -27.57,
        -56.3
      ]
    },
    "Elevation": 551,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "ab10b214-cec0-51a2-d42d-b68f50eb8f8e"
  },
  {
    "Volcano Name": "Zengyu",
    "Country": "Taiwan",
    "Region": "Taiwan-N of",
    "Location": {
      "type": "Point",
      "coordinates": [
        122.458,
        26.18
      ]
    },
    "Elevation": -418,
    "Type": "Submarine volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "d07e3467-2d79-b600-ecba-cf2e67363ae0"
  },
  {
    "Volcano Name": "Zheltovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        157.323,
        51.57
      ]
    },
    "Elevation": 1953,
    "Type": "Stratovolcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption in 1964 or later",
    "id": "e8b677bc-c1c6-2078-22a5-8e9c94f7e097"
  },
  {
    "Volcano Name": "Zhupanovsky",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        159.147,
        53.59
      ]
    },
    "Elevation": 2958,
    "Type": "Compound volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1900-1963, inclusive",
    "id": "be8fb244-5c69-9685-b580-66eff4ccf8e9"
  },
  {
    "Volcano Name": "Zimina",
    "Country": "Russia",
    "Region": "Kamchatka",
    "Location": {
      "type": "Point",
      "coordinates": [
        160.603,
        55.862
      ]
    },
    "Elevation": 3081,
    "Type": "Stratovolcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3dfece75-962b-5271-2d55-28accbfae80b"
  },
  {
    "Volcano Name": "Zubair, Jebel",
    "Country": "Yemen",
    "Region": "Red Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.18,
        15.05
      ]
    },
    "Elevation": 191,
    "Type": "Shield volcano",
    "Status": "Historical",
    "Last Known Eruption": "Last known eruption from 1800-1899, inclusive",
    "id": "24f4da9c-7afb-985d-950a-1c5b402c38f6"
  },
  {
    "Volcano Name": "Zukur",
    "Country": "Yemen",
    "Region": "Red Sea",
    "Location": {
      "type": "Point",
      "coordinates": [
        42.75,
        14.02
      ]
    },
    "Elevation": 624,
    "Type": "Shield volcano",
    "Status": "Holocene",
    "Last Known Eruption": "Undated, but probable Holocene eruption",
    "id": "3e8037d2-8638-61fa-e5a0-20fbcd3122e0"
  },
  {
    "Volcano Name": "Zuni-Bandera",
    "Country": "United States",
    "Region": "US-New Mexico",
    "Location": {
      "type": "Point",
      "coordinates": [
        -108,
        34.8
      ]
    },
    "Elevation": 2550,
    "Type": "Volcanic field",
    "Status": "Anthropology",
    "Last Known Eruption": "Last known eruption B.C. (Holocene)",
    "id": "a3a04d36-ca18-6995-24bf-39c4279251e2"
  },
  {
    "id": "washington-polygon",
    "country": "United States of America",
    "stateCode": "WA",
    "geometry": {
      "type": "Polygon",
      "coordinates": [
        [
          [
            -124.63,
            48.36
          ],
          [
            -123.87,
            46.14
          ],
          [
            -122.23,
            45.54
          ],
          [
            -119.17,
            45.95
          ],
          [
            -116.92,
            45.96
          ],
          [
            -116.99,
            49
          ],
          [
            -123.05,
            49.02
          ],
          [
            -123.15,
            48.31
          ],
          [
            -124.63,
            48.36
          ]
        ]
      ]
    }
  },
  {
    "id": "india-polygon",
    "geometry": {
      "type": "Polygon",
      "coordinates": [
        [
          [
            77.8374507994746,
            35.4940095077878
          ],
          [
            78.9122689147132,
            34.3219363469758
          ],
          [
            78.8110864602857,
            33.5061980250324
          ],
          [
            79.2088916360686,
            32.9943946396137
          ],
          [
            79.1761287779955,
            32.4837798121377
          ],
          [
            78.458446486326,
            32.6181643743127
          ],
          [
            78.738894484374,
            31.5159060735271
          ],
          [
            79.7213668151071,
            30.8827147486547
          ],
          [
            81.1112561380293,
            30.1834809433134
          ],
          [
            80.4767212259174,
            29.7298652206553
          ],
          [
            80.0884245136763,
            28.7944701197401
          ],
          [
            81.057202589852,
            28.416095282499
          ],
          [
            81.999987420585,
            27.92547923432
          ],
          [
            83.3042488951995,
            27.3645057235756
          ],
          [
            84.6750179381738,
            27.2349012313875
          ],
          [
            85.2517785989834,
            26.7261984319063
          ],
          [
            86.0243929381792,
            26.6309846054086
          ],
          [
            87.2274719583663,
            26.3978980575561
          ],
          [
            88.0602376647498,
            26.4146153834025
          ],
          [
            88.1748043151409,
            26.810405178326
          ],
          [
            88.0431327656612,
            27.4458185897868
          ],
          [
            88.1204407083699,
            27.8765416529396
          ],
          [
            88.7303259622786,
            28.0868647323675
          ],
          [
            88.8142484883206,
            27.2993159042394
          ],
          [
            88.8356425312894,
            27.0989663762438
          ],
          [
            89.7445276224389,
            26.71940298106
          ],
          [
            90.3732747741341,
            26.8757241887429
          ],
          [
            91.2175126484864,
            26.808648179628
          ],
          [
            92.0334835143751,
            26.8383104517636
          ],
          [
            92.1037117858597,
            27.4526140406332
          ],
          [
            91.6966565286967,
            27.7717418482517
          ],
          [
            92.5031189310436,
            27.8968763290464
          ],
          [
            93.4133476094327,
            28.6406293808072
          ],
          [
            94.5659904317029,
            29.27743805594
          ],
          [
            95.4048022806646,
            29.0317166203921
          ],
          [
            96.117678664131,
            29.4528020289225
          ],
          [
            96.5865906107475,
            28.8309795191543
          ],
          [
            96.2488334492878,
            28.4110309921344
          ],
          [
            97.32711388549,
            28.2615827499463
          ],
          [
            97.4025614766361,
            27.8825361190854
          ],
          [
            97.0519885599681,
            27.6990589462332
          ],
          [
            97.1339990580153,
            27.08377350515
          ],
          [
            96.419365675851,
            27.2645893417392
          ],
          [
            95.124767694075,
            26.5735720891323
          ],
          [
            95.1551534362626,
            26.0013072779321
          ],
          [
            94.6032491393854,
            25.1624954289704
          ],
          [
            94.5526579121716,
            24.6752383488903
          ],
          [
            94.1067419779251,
            23.8507408716735
          ],
          [
            93.3251876159428,
            24.0785564234322
          ],
          [
            93.2863269388593,
            23.043658352139
          ],
          [
            93.0602942240146,
            22.7031106633356
          ],
          [
            93.1661275573484,
            22.2784595809771
          ],
          [
            92.6727209818256,
            22.0412389185413
          ],
          [
            92.1460347839068,
            23.6274986841726
          ],
          [
            91.8699276061713,
            23.6243464218028
          ],
          [
            91.7064750508321,
            22.9852639836492
          ],
          [
            91.1589632506997,
            23.5035269231044
          ],
          [
            91.4677299336437,
            24.0726394719348
          ],
          [
            91.9150928079944,
            24.1304137232371
          ],
          [
            92.3762016133348,
            24.976692816665
          ],
          [
            91.7995959818221,
            25.1474317489573
          ],
          [
            90.8722107279121,
            25.1326006128895
          ],
          [
            89.9206925801219,
            25.2697498641922
          ],
          [
            89.8324809101996,
            25.9650820988955
          ],
          [
            89.3550940286873,
            26.0144072535181
          ],
          [
            88.5630493509498,
            26.4465255803427
          ],
          [
            88.2097892598025,
            25.7680657007827
          ],
          [
            88.9315539896231,
            25.2386923283848
          ],
          [
            88.306372511756,
            24.8660794133442
          ],
          [
            88.0844222350624,
            24.5016572128219
          ],
          [
            88.6999402200909,
            24.2337149113886
          ],
          [
            88.5297697285538,
            23.6311418726492
          ],
          [
            88.8763118835031,
            22.8791464299378
          ],
          [
            89.0319612975662,
            22.055708319583
          ],
          [
            88.8887659036854,
            21.6905884872247
          ],
          [
            88.2084973489952,
            21.7031716984878
          ],
          [
            86.9757043802403,
            21.4955616317552
          ],
          [
            87.0331685729489,
            20.7433078068824
          ],
          [
            86.4993510273738,
            20.1516384953566
          ],
          [
            85.0602657409097,
            19.4785788029711
          ],
          [
            83.9410058939,
            18.3020097925497
          ],
          [
            83.1892171569178,
            17.671221421779
          ],
          [
            82.1927921894659,
            17.0166360539378
          ],
          [
            82.1912418964972,
            16.5566641301078
          ],
          [
            81.6927193541775,
            16.3102192245079
          ],
          [
            80.7919991393301,
            15.9519723576445
          ],
          [
            80.3248958678439,
            15.8991848820584
          ],
          [
            80.0250692076864,
            15.1364149032141
          ],
          [
            80.2332735533904,
            13.83577077886
          ],
          [
            80.2862935729219,
            13.0062606877108
          ],
          [
            79.8625468281285,
            12.0562153182409
          ],
          [
            79.8579993020868,
            10.3572750919971
          ],
          [
            79.340511509116,
            10.3088542749396
          ],
          [
            78.8853454934892,
            9.54613597252772
          ],
          [
            79.1897196796883,
            9.21654368737015
          ],
          [
            78.2779407083305,
            8.93304677981693
          ],
          [
            77.9411653990844,
            8.25295909263974
          ],
          [
            77.5398979023379,
            7.96553477623233
          ],
          [
            76.5929789570217,
            8.89927623131419
          ],
          [
            76.1300614765511,
            10.2996300317755
          ],
          [
            75.7464673196485,
            11.3082506372483
          ],
          [
            75.3961011087096,
            11.7812450220158
          ],
          [
            74.8648157083168,
            12.7419357365379
          ],
          [
            74.6167171568835,
            13.9925829126497
          ],
          [
            74.4438594908672,
            14.6172217879777
          ],
          [
            73.5341992532334,
            15.990652167215
          ],
          [
            73.1199092955494,
            17.9285700545925
          ],
          [
            72.8209094583086,
            19.2082335474362
          ],
          [
            72.8244751321368,
            20.4195032821415
          ],
          [
            72.6305334817454,
            21.356009426351
          ],
          [
            71.175273471974,
            20.7574413111142
          ],
          [
            70.4704586119451,
            20.8773306340314
          ],
          [
            69.1641300800388,
            22.0892980005727
          ],
          [
            69.6449276060824,
            22.4507746444543
          ],
          [
            69.3495967955344,
            22.8431796330627
          ],
          [
            68.1766451353734,
            23.6919650334567
          ],
          [
            68.8425993183188,
            24.3591336125609
          ],
          [
            71.0432401874682,
            24.3565239527302
          ],
          [
            70.8446993346028,
            25.2151020370435
          ],
          [
            70.2828731627256,
            25.7222287053398
          ],
          [
            70.168926629522,
            26.4918716496788
          ],
          [
            69.5143929381131,
            26.9409656845114
          ],
          [
            70.6164962096019,
            27.9891962753359
          ],
          [
            71.7776656432003,
            27.9131802434345
          ],
          [
            72.8237516620847,
            28.9615917017721
          ],
          [
            73.4506384622174,
            29.9764134791199
          ],
          [
            74.4213802428203,
            30.9798147649312
          ],
          [
            74.405928989565,
            31.6926394719653
          ],
          [
            75.2586417988132,
            32.2711054550405
          ],
          [
            74.4515592792787,
            32.7648996038055
          ],
          [
            74.1042936542773,
            33.4414732935869
          ],
          [
            73.749948358052,
            34.3176988795279
          ],
          [
            74.240202671205,
            34.7488870305713
          ],
          [
            75.7570609882683,
            34.5049225937213
          ],
          [
            76.871721632804,
            34.6535440129927
          ],
          [
            77.8374507994746,
            35.4940095077878
          ]
        ]
      ]
    }
  },
  {
    "id": "polygon",
    "footprint": {
      "type": "Polygon",
      "coordinates": [
        [
          [
            -74.780678,
            10.973175375207564
          ],
          [
            -74.773293919212719,
            10.972856728561238
          ],
          [
            -74.765966082779485,
            10.971903215715052
          ],
          [
            -74.7587503055614,
            10.970322099414361
          ],
          [
            -74.751701546746688,
            10.968125422600975
          ],
          [
            -74.744873490270649,
            10.965329916461481
          ],
          [
            -74.738318135068255,
            10.961956872690292
          ],
          [
            -74.732085398315675,
            10.958031980948769
          ],
          [
            -74.726222734709339,
            10.953585132777587
          ],
          [
            -74.72077477470674,
            10.948650193468669
          ],
          [
            -74.715782984499569,
            10.943264743654186
          ],
          [
            -74.711285350317851,
            10.937469792599721
          ],
          [
            -74.707316089470993,
            10.931309465405118
          ],
          [
            -74.703905390319591,
            10.924830666514849
          ],
          [
            -74.701079183145225,
            10.91808272211917
          ],
          [
            -74.698858943641113,
            10.911117004187272
          ],
          [
            -74.6972615304966,
            10.903986539009535
          ],
          [
            -74.696299058273979,
            10.896745603244831
          ],
          [
            -74.695978806514148,
            10.889449310556859
          ],
          [
            -74.696303165719954,
            10.882153191995066
          ],
          [
            -74.697269620586312,
            10.87491277331819
          ],
          [
            -74.698870770562053,
            10.867783152478559
          ],
          [
            -74.701094387543449,
            10.860818580480377
          ],
          [
            -74.703923510218317,
            10.854072048795539
          ],
          [
            -74.70733657430749,
            10.847594886468892
          ],
          [
            -74.711307577671761,
            10.841436369964686
          ],
          [
            -74.715806279004767,
            10.83564334871058
          ],
          [
            -74.720798428571911,
            10.830259889170497
          ],
          [
            -74.726246029223745,
            10.825326940136572
          ],
          [
            -74.732107625686879,
            10.820882021767437
          ],
          [
            -74.73833861992803,
            10.816958940718891
          ],
          [
            -74.744891610195864,
            10.813587533514497
          ],
          [
            -74.751716751171386,
            10.81079344008922
          ],
          [
            -74.758762132505623,
            10.808597909212722
          ],
          [
            -74.765974172886487,
            10.807017637252722
          ],
          [
            -74.7732980266679,
            10.80606464149605
          ],
          [
            -74.780678,
            10.805746168976489
          ],
          [
            -74.7880579733321,
            10.80606464149605
          ],
          [
            -74.795381827113516,
            10.807017637252722
          ],
          [
            -74.80259386749438,
            10.808597909212722
          ],
          [
            -74.809639248828617,
            10.81079344008922
          ],
          [
            -74.816464389804139,
            10.813587533514497
          ],
          [
            -74.823017380071974,
            10.816958940718893
          ],
          [
            -74.829248374313124,
            10.820882021767437
          ],
          [
            -74.835109970776259,
            10.825326940136572
          ],
          [
            -74.840557571428093,
            10.830259889170497
          ],
          [
            -74.845549720995223,
            10.835643348710581
          ],
          [
            -74.850048422328229,
            10.841436369964686
          ],
          [
            -74.854019425692513,
            10.84759488646889
          ],
          [
            -74.857432489781687,
            10.854072048795539
          ],
          [
            -74.860261612456554,
            10.860818580480377
          ],
          [
            -74.862485229437951,
            10.867783152478559
          ],
          [
            -74.864086379413678,
            10.87491277331819
          ],
          [
            -74.86505283428005,
            10.882153191995066
          ],
          [
            -74.865377193485855,
            10.889449310556859
          ],
          [
            -74.865056941726024,
            10.896745603244831
          ],
          [
            -74.8640944695034,
            10.903986539009535
          ],
          [
            -74.86249705635889,
            10.911117004187272
          ],
          [
            -74.860276816854778,
            10.918082722119168
          ],
          [
            -74.857450609680413,
            10.924830666514849
          ],
          [
            -74.854039910529,
            10.931309465405118
          ],
          [
            -74.850070649682138,
            10.937469792599721
          ],
          [
            -74.845573015500435,
            10.943264743654186
          ],
          [
            -74.840581225293263,
            10.948650193468669
          ],
          [
            -74.835133265290651,
            10.953585132777585
          ],
          [
            -74.829270601684328,
            10.958031980948769
          ],
          [
            -74.823037864931749,
            10.961956872690294
          ],
          [
            -74.816482509729354,
            10.965329916461481
          ],
          [
            -74.809654453253316,
            10.968125422600977
          ],
          [
            -74.8026056944386,
            10.970322099414361
          ],
          [
            -74.795389917220518,
            10.971903215715052
          ],
          [
            -74.788062080787284,
            10.972856728561238
          ],
          [
            -74.780678,
            10.973175375207564
          ]
        ]
      ]
    }
  },
  {
    "metadata": {
      "wsiProductNumber": "1140",
      "messageType": "AIRMET",
      "productId": "AIRMET_GOVT",
      "messageId": "AIRMET_20180119010212aa0cb8d6a84e3769719d07643cd8beb1",
      "expiry": "2018-01-19T04:00:00Z",
      "messageSize": 2904
    },
    "data": {
      "rawType": "SFC WND",
      "issueTime": "2018-01-18T21:44:00Z",
      "changeType": "ROU",
      "phenomenon": "WINDS",
      "rawData": "HNLT WA 182200 \nAIRMET TANGO UPDATE 3 FOR TURB VALID UNTIL 190400 \nAIRMET STG SFC WND...HI \nOVER MTN...THRU VALLEYS...AND NEAR HEADLANDS ALL ISLANDS. \nSTG SFC WIND 30 KT OR GREATER. \nCOND CONT BEYOND 0400Z",
      "type": "WINDS",
      "expireAt": "2018-01-19T04:00:00Z",
      "rowId": 42009265343914,
      "reportType": "AIRMET",
      "activeAt": "2018-01-18T22:00:00Z",
      "siteId": "PHNL",
      "id": "2",
      "region": {
        "coordinates": [
          [
            [
              [
                -156.017,
                19.645
              ],
              [
                -156.074,
                19.734
              ],
              [
                -156.048,
                19.798
              ],
              [
                -155.934,
                19.879
              ],
              [
                -155.841,
                19.996
              ],
              [
                -155.917,
                20.2
              ],
              [
                -155.899,
                20.262
              ],
              [
                -155.837,
                20.281
              ],
              [
                -155.595,
                20.134
              ],
              [
                -155.437,
                20.105
              ],
              [
                -155.208,
                19.985
              ],
              [
                -155.077,
                19.859
              ],
              [
                -155.082,
                19.741
              ],
              [
                -154.999,
                19.747
              ],
              [
                -154.973,
                19.645
              ],
              [
                -154.799,
                19.518
              ],
              [
                -154.977,
                19.34
              ],
              [
                -155.157,
                19.259
              ],
              [
                -155.294,
                19.256
              ],
              [
                -155.5,
                19.132
              ],
              [
                -155.594,
                18.965
              ],
              [
                -155.686,
                18.904
              ],
              [
                -155.732,
                18.961
              ],
              [
                -155.889,
                19.03
              ],
              [
                -155.93,
                19.113
              ],
              [
                -155.899,
                19.347
              ],
              [
                -156.017,
                19.645
              ]
            ],
            [
              [
                -156.539,
                20.987
              ],
              [
                -156.52,
                20.998
              ],
              [
                -156.472,
                20.906
              ],
              [
                -156.326,
                20.96
              ],
              [
                -156.238,
                20.948
              ],
              [
                -156.109,
                20.836
              ],
              [
                -156,
                20.805
              ],
              [
                -155.967,
                20.727
              ],
              [
                -156.109,
                20.632
              ],
              [
                -156.298,
                20.577
              ],
              [
                -156.443,
                20.592
              ],
              [
                -156.474,
                20.775
              ],
              [
                -156.631,
                20.802
              ],
              [
                -156.7,
                20.883
              ],
              [
                -156.674,
                21.019
              ],
              [
                -156.589,
                21.044
              ],
              [
                -156.539,
                20.987
              ]
            ],
            [
              [
                -156.501,
                20.646
              ],
              [
                -156.486,
                20.636
              ],
              [
                -156.5,
                20.623
              ],
              [
                -156.512,
                20.635
              ],
              [
                -156.501,
                20.646
              ]
            ],
            [
              [
                -156.593,
                20.612
              ],
              [
                -156.533,
                20.583
              ],
              [
                -156.541,
                20.507
              ],
              [
                -156.677,
                20.494
              ],
              [
                -156.712,
                20.521
              ],
              [
                -156.593,
                20.612
              ]
            ],
            [
              [
                -156.714,
                21.142
              ],
              [
                -156.698,
                21.145
              ],
              [
                -156.697,
                21.127
              ],
              [
                -156.715,
                21.126
              ],
              [
                -156.714,
                21.142
              ]
            ],
            [
              [
                -157.114,
                21.098
              ],
              [
                -157.324,
                21.105
              ],
              [
                -157.257,
                21.237
              ],
              [
                -157.017,
                21.196
              ],
              [
                -156.973,
                21.229
              ],
              [
                -156.941,
                21.186
              ],
              [
                -156.703,
                21.162
              ],
              [
                -156.772,
                21.078
              ],
              [
                -156.875,
                21.038
              ],
              [
                -157.114,
                21.098
              ]
            ],
            [
              [
                -157.006,
                20.828
              ],
              [
                -157.072,
                20.904
              ],
              [
                -157.027,
                20.939
              ],
              [
                -156.902,
                20.928
              ],
              [
                -156.799,
                20.819
              ],
              [
                -156.838,
                20.753
              ],
              [
                -156.968,
                20.725
              ],
              [
                -157.006,
                20.828
              ]
            ],
            [
              [
                -158.116,
                21.592
              ],
              [
                -158.048,
                21.691
              ],
              [
                -157.962,
                21.721
              ],
              [
                -157.903,
                21.656
              ],
              [
                -157.92,
                21.644
              ],
              [
                -157.88,
                21.568
              ],
              [
                -157.836,
                21.55
              ],
              [
                -157.836,
                21.472
              ],
              [
                -157.802,
                21.442
              ],
              [
                -157.782,
                21.475
              ],
              [
                -157.712,
                21.48
              ],
              [
                -157.733,
                21.414
              ],
              [
                -157.693,
                21.401
              ],
              [
                -157.706,
                21.381
              ],
              [
                -157.65,
                21.34
              ],
              [
                -157.645,
                21.297
              ],
              [
                -157.702,
                21.253
              ],
              [
                -157.726,
                21.282
              ],
              [
                -157.794,
                21.248
              ],
              [
                -157.969,
                21.319
              ],
              [
                -158.113,
                21.291
              ],
              [
                -158.294,
                21.579
              ],
              [
                -158.116,
                21.592
              ]
            ],
            [
              [
                -157.812,
                21.477
              ],
              [
                -157.805,
                21.489
              ],
              [
                -157.789,
                21.482
              ],
              [
                -157.806,
                21.456
              ],
              [
                -157.812,
                21.477
              ]
            ],
            [
              [
                -159.562,
                21.883
              ],
              [
                -159.769,
                21.975
              ],
              [
                -159.795,
                22.074
              ],
              [
                -159.732,
                22.161
              ],
              [
                -159.587,
                22.236
              ],
              [
                -159.405,
                22.249
              ],
              [
                -159.346,
                22.226
              ],
              [
                -159.286,
                22.145
              ],
              [
                -159.33,
                22.046
              ],
              [
                -159.326,
                21.957
              ],
              [
                -159.444,
                21.862
              ],
              [
                -159.562,
                21.883
              ]
            ],
            [
              [
                -160.168,
                21.947
              ],
              [
                -160.107,
                22.038
              ],
              [
                -160.045,
                21.994
              ],
              [
                -160.068,
                21.891
              ],
              [
                -160.155,
                21.857
              ],
              [
                -160.208,
                21.77
              ],
              [
                -160.258,
                21.849
              ],
              [
                -160.168,
                21.947
              ]
            ],
            [
              [
                -160.546,
                21.64
              ],
              [
                -160.562,
                21.659
              ],
              [
                -160.552,
                21.672
              ],
              [
                -160.536,
                21.653
              ],
              [
                -160.546,
                21.64
              ]
            ]
          ]
        ],
        "type": "MultiPolygon"
      },
      "objectId": 42009265343914
    },
    "id": "boeing"
  },
  {
    "UCATS": 0,
    "URTS": 0,
    "UTS": 0,
    "AC": null,
    "ASC": 5,
    "ESC": 0,
    "FSC": 73,
    "RSC": 0,
    "RTSC": 0,
    "SC": 112,
    "SSC": 34,
    "LN": null,
    "PLID": "ChIJA9KNRIL-1BIRb15jJFz1LOI",
    "PID": null,
    "NAME": "Italy",
    "PPID": "ChIJhdqtz4aI7UYRefD8s-aZ73I",
    "LOC": {
      "type": "Point",
      "coordinates": [
        12.7673821,
        41.9719447
      ]
    },
    "LT": 1,
    "HCA": true,
    "A2": "IT",
    "id": "CRI",
    "CID": "tJ",
    "DT": 9,
    "PK": "tJA1"
  }]
`

/*----------------------------------------------------------------------*/

func _initDataNutrition(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	dataListNutrition := make([]gocosmos.DocInfo, 0)
	dataMapNutrition := sync.Map{}
	err := json.Unmarshal([]byte(_testDataNutrition), &dataListNutrition)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	fmt.Printf("\tDataset: %#v / Number of records: %#v\n", "Nutrition", len(dataListNutrition))

	numWorkers := 4
	buff := make(chan gocosmos.DocInfo, numWorkers*4)
	wg := &sync.WaitGroup{}
	wg.Add(numWorkers)
	numDocWritten := int64(0)
	start := time.Now()
	for id := 0; id < numWorkers; id++ {
		go func(id int, wg *sync.WaitGroup, buff <-chan gocosmos.DocInfo) {
			defer wg.Done()
			for doc := range buff {
				docId := doc["id"].(string)
				dataMapNutrition.Store(docId, doc)
				if result := client.CreateDocument(gocosmos.DocumentSpec{DbName: db, CollName: container, PartitionKeyValues: []interface{}{docId}, DocumentData: doc}); result.Error() != nil {
					t.Fatalf("%s failed: (%#v) %s", testName, id, result.Error())
				}
				atomic.AddInt64(&numDocWritten, 1)
				for {
					now := time.Now()
					d := now.Sub(start)
					r := float64(numDocWritten) / (d.Seconds() + 0.001)
					if r <= 81.19 {
						break
					}
					fmt.Printf("\t[DEBUG] too fast, slowing down...(Id: %d / NumDocs: %d / Dur: %.3f / Rate: %.3f)\n", id, numDocWritten, d.Seconds(), r)
					time.Sleep(1*time.Second + time.Duration(rand.Intn(1234))*time.Millisecond)
				}
			}
			fmt.Printf("\t\tWorker %#v: %#v docs written\n", id, numDocWritten)
		}(id, wg, buff)
	}
	for _, doc := range dataListNutrition {
		buff <- doc
	}
	close(buff)
	wg.Wait()
	{
		now := time.Now()
		d := now.Sub(start)
		r := float64(numDocWritten) / (d.Seconds() + 0.001)
		fmt.Printf("\t[DEBUG] Dur: %.3f / Rate: %.3f\n", d.Seconds(), r)
		time.Sleep(1*time.Second + time.Duration(rand.Intn(1234))*time.Millisecond)
	}
	count := 0
	dataMapNutrition.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	fmt.Printf("\tDataset: %#v / (checksum) Number of records: %#v\n", "Nutrition", count)
}

func _initDataNutritionSmallRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 400})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		Ru:               400,
	})
	_initDataNutrition(t, testName, client, db, container)
}

func _initDataNutritionLargeRU(t *testing.T, testName string, client *gocosmos.RestClient, db, container string) {
	_ensureDatabase(client, gocosmos.DatabaseSpec{Id: db, Ru: 20000})
	_ensureCollection(client, gocosmos.CollectionSpec{
		DbName:           db,
		CollName:         container,
		PartitionKeyInfo: map[string]interface{}{"paths": []string{"/id"}, "kind": "Hash"},
		UniqueKeyPolicy:  map[string]interface{}{"uniqueKeys": []map[string]interface{}{{"paths": []string{"/email"}}}},
		Ru:               20000,
	})
	_initDataNutrition(t, testName, client, db, container)
}
