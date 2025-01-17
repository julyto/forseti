package citiz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/CanalTP/forseti/internal/freefloatings"
	"github.com/CanalTP/forseti/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixtureDir string

func TestMain(m *testing.M) {

	fixtureDir = os.Getenv("FIXTUREDIR")
	if fixtureDir == "" {
		panic("$FIXTUREDIR isn't set")
	}

	os.Exit(m.Run())
}

func TestLoadFreeFloatingsFromFile(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	uri, err := url.Parse(fmt.Sprintf("file://%s/vehicles_citiz.json", fixtureDir))
	require.Nil(err)
	reader, err := utils.GetFileWithFS(*uri)
	require.Nil(err)

	jsonData, err := ioutil.ReadAll(reader)
	require.Nil(err)

	data := &CitizData{}
	err = json.Unmarshal([]byte(jsonData), data)
	require.Nil(err)

	freeFloatings := LoadVehiclesData(*data)

	require.Nil(err)
	assert.NotEqual(len(freeFloatings), 0)
	assert.Len(freeFloatings, 6)
	assert.NotEmpty(freeFloatings[0].Id)
	assert.Equal("FG-034-YP", freeFloatings[0].PublicId)
	assert.Equal("Citiz", freeFloatings[0].ProviderName)
	assert.Equal("3842", freeFloatings[0].Id)
	assert.Equal("CAR", freeFloatings[0].Type)
	assert.Equal("COMBUSTION", freeFloatings[0].Propulsion)
	assert.Equal(48.109989166259766, freeFloatings[0].Coord.Lat)
	assert.Equal(-1.6804883480072021, freeFloatings[0].Coord.Lon)
	assert.Equal([]string{"COMBUSTION"}, freeFloatings[0].Attributes)
}

func TestNewFreeFloating(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	ff := Vehicles{
		LocationId:       1428,
		LocationName:     "Anatole France",
		LocationCityId:   313,
		LocationCity:     "Rennes",
		CategoryId:       219,
		InternalRemark:   "",
		KmTotal:          44604,
		AffectationStart: "2021-02-02T09:00:00",
		AffectationEnd:   "2050-12-30T23:00:00",
		VehiculeId:       3817,
		ProviderId:       19,
		Name:             "208 - 020 (M)",
		LicencePlate:     "FG-020-PN",
		ExternalRemark: "Véhicule essence (SP95/E10) Boite manuelle\n5 portes - 5 places - GPS - limatisation\n" +
			"Station sans Arceau de parking\nCapacité du réservoir : 50 itres\nConsommation moyenne : 4,9L/100km",
		IsFreeFloating: false,
		ElectricEngine: true,
		Category:       "M",
		GpsLatitude:    48.118404388427734,
		GpsLongitude:   -1.6892282962799072,
	}

	f := NewCitiz(ff)
	require.NotNil(f)

	assert.Equal("FG-020-PN", f.PublicId)
	assert.Equal("3817", f.Id)
	assert.Equal("Citiz", f.ProviderName)
	assert.Equal("CAR", f.Type)
	assert.Equal(48.118404388427734, f.Coord.Lat)
	assert.Equal(-1.6892282962799072, f.Coord.Lon)
	assert.Equal("ELECTRIC", f.Propulsion)
	assert.Equal("M", f.Properties["category"])
}

func TestDataManagerGetFreeFloatings(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	freeFloatingsContext := &freefloatings.FreeFloatingsContext{}

	vehicles := make([]Vehicles, 0)
	freeFloatings := make([]freefloatings.FreeFloating, 0)

	v := Vehicles{
		VehiculeId:     3817,
		ProviderId:     19,
		LicencePlate:   "FG-020-PN",
		IsFreeFloating: false,
		ElectricEngine: true,
		Category:       "M",
		GpsLatitude:    48.118404388427734,
		GpsLongitude:   -1.6892282962799072,
	}
	vehicles = append(vehicles, v)

	v = Vehicles{
		VehiculeId:     3814,
		ProviderId:     19,
		LicencePlate:   "FG-050-PN",
		IsFreeFloating: false,
		ElectricEngine: false,
		Category:       "L",
		GpsLatitude:    48.103042602539063,
		GpsLongitude:   -1.6619917154312134,
	}
	vehicles = append(vehicles, v)

	v = Vehicles{
		VehiculeId:     3829,
		ProviderId:     19,
		LicencePlate:   "FG-335-PN",
		IsFreeFloating: false,
		ElectricEngine: false,
		Category:       "M",
		GpsLatitude:    48.110034942626953,
		GpsLongitude:   -1.65944504737854,
	}
	vehicles = append(vehicles, v)

	for _, vehicle := range vehicles {
		freeFloatings = append(freeFloatings, *NewCitiz(vehicle))
	}
	freeFloatingsContext.UpdateFreeFloating(freeFloatings)
	// init parameters:
	coord := freefloatings.Coord{Lat: 48.103042602, Lon: -1.66199171543}
	p := freefloatings.FreeFloatingRequestParameter{Distance: 900, Coord: coord, Count: 10}
	free_floatings, paginate_freefloatings, err := freeFloatingsContext.GetFreeFloatings(&p)
	require.Nil(err)
	require.NotNil(paginate_freefloatings)
	require.Len(free_floatings, 2)

	assert.Equal("FG-050-PN", free_floatings[0].PublicId)
	assert.Equal("Citiz", free_floatings[0].ProviderName)
	assert.Equal("CAR", free_floatings[0].Type)
	assert.Equal("FG-335-PN", free_floatings[1].PublicId)
	assert.Equal("3829", free_floatings[1].Id)
}
