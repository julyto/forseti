package vehiclepositions

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/CanalTP/forseti/google_transit"
	"github.com/CanalTP/forseti/internal/connectors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_VehiclePositionsAPI(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	connector, err := ConnectorFactory(string(connectors.Connector_GRFS_RT))
	require.Nil(err)
	gtfsRtContext, ok := connector.(*GtfsRtContext)
	require.True(ok)
	pVehiclePositions := gtfsRtContext.GetAllVehiclePositions()
	require.NotNil(pVehiclePositions)

	c, engine := gin.CreateTestContext(httptest.NewRecorder())
	AddVehiclePositionsEntryPoint(engine, gtfsRtContext)

	// Request without locations data
	response := VehiclePositionsResponse{}
	c.Request = httptest.NewRequest("GET", "/vehicle_positions", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(503, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.Nil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 0)
	assert.Equal(response.Error, "No data loaded")

	// Add locations data
	pVehiclePositions.vehiclePositions = vehiclePositionsMap
	assert.Equal(len(pVehiclePositions.vehiclePositions), 2)

	// Request without any parameter
	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest("GET", "/vehicle_positions", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 1)
	assert.Empty(response.Error)

	c.Request = httptest.NewRequest("GET", "/vehicle_positions?date=20210127", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 2)
	assert.Empty(response.Error)

	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest(
		"GET", "/vehicle_positions?date=20210118&vehicle_journey_code[]=653397", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 1)
	assert.Empty(response.Error)

	assert.Equal(response.VehiclePositions[0].VehicleJourneyCode, "653397")
	assert.Equal(response.VehiclePositions[0].Latitude, float32(45.413333892822266))
	assert.Equal(response.VehiclePositions[0].Longitude, float32(-71.87944793701172))
	assert.Equal(response.VehiclePositions[0].Bearing, float32(254))
	assert.Equal(response.VehiclePositions[0].Speed, float32(10))
	assert.Equal(response.VehiclePositions[0].Occupancy, google_transit.VehiclePosition_OccupancyStatus_name[1])

	// TEST DISTANCE PARAMETER
	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest(
		"GET", "/vehicle_positions?vehicle_journey_code[]=653397&distance=500&coord=-71.87944%3B45.413333", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 1)
	assert.Empty(response.Error)

	assert.Equal(response.VehiclePositions[0].VehicleJourneyCode, "653397")
	assert.Equal(response.VehiclePositions[0].Latitude, float32(45.413333892822266))
	assert.Equal(response.VehiclePositions[0].Longitude, float32(-71.87944793701172))
	assert.Equal(response.VehiclePositions[0].Bearing, float32(254))
	assert.Equal(response.VehiclePositions[0].Speed, float32(10))
	assert.Equal(response.VehiclePositions[0].Occupancy, google_transit.VehiclePosition_OccupancyStatus_name[1])
	assert.Empty(response.VehiclePositions[0].Distance, 0)

	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest(
		"GET", "/vehicle_positions?distance=500&coord=-71.87944%3B45.414333", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 1)
	assert.Empty(response.Error)

	assert.Equal(response.VehiclePositions[0].VehicleJourneyCode, "653397")
	assert.Equal(response.VehiclePositions[0].Latitude, float32(45.413333892822266))
	assert.Equal(response.VehiclePositions[0].Longitude, float32(-71.87944793701172))
	assert.Equal(response.VehiclePositions[0].Bearing, float32(254))
	assert.Equal(response.VehiclePositions[0].Speed, float32(10))
	assert.Equal(response.VehiclePositions[0].Occupancy, google_transit.VehiclePosition_OccupancyStatus_name[1])
	assert.Equal(response.VehiclePositions[0].Distance, float64(111))

	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest(
		"GET", "/vehicle_positions?coord=-71.87944%3B45.413333", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.NotNil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 1)
	assert.Empty(response.Error)

	assert.Equal(response.VehiclePositions[0].VehicleJourneyCode, "653397")
	assert.Equal(response.VehiclePositions[0].Latitude, float32(45.413333892822266))
	assert.Equal(response.VehiclePositions[0].Longitude, float32(-71.87944793701172))
	assert.Equal(response.VehiclePositions[0].Bearing, float32(254))
	assert.Equal(response.VehiclePositions[0].Speed, float32(10))
	assert.Equal(response.VehiclePositions[0].Occupancy, google_transit.VehiclePosition_OccupancyStatus_name[1])
	assert.Equal(response.VehiclePositions[0].Distance, float64(1))

	response = VehiclePositionsResponse{}
	c.Request = httptest.NewRequest(
		"GET", "/vehicle_positions?distance=500", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, c.Request)
	require.Equal(503, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(err)
	require.Nil(response.VehiclePositions)
	assert.Len(response.VehiclePositions, 0)
	assert.Equal("Bad request: coord is mandatory", response.Error)
}
