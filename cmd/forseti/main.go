package main

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/CanalTP/forseti/api"
	"github.com/CanalTP/forseti/internal/connectors"
	"github.com/CanalTP/forseti/internal/departures"
	"github.com/CanalTP/forseti/internal/equipments"
	"github.com/CanalTP/forseti/internal/freefloatings"
	"github.com/CanalTP/forseti/internal/freefloatings/citiz"
	"github.com/CanalTP/forseti/internal/freefloatings/fluctuo"
	"github.com/CanalTP/forseti/internal/manager"
	"github.com/CanalTP/forseti/internal/parkings"
	vehicleoccupanciesv2 "github.com/CanalTP/forseti/internal/vehicleoccupancies_v2"
	"github.com/CanalTP/forseti/internal/vehiclepositions"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	DeparturesURIStr  string        `mapstructure:"departures-uri"`
	DeparturesRefresh time.Duration `mapstructure:"departures-refresh"`
	DeparturesURI     url.URL

	ParkingsURIStr  string        `mapstructure:"parkings-uri"`
	ParkingsRefresh time.Duration `mapstructure:"parkings-refresh"`
	ParkingsURI     url.URL

	EquipmentsURIStr  string        `mapstructure:"equipments-uri"`
	EquipmentsRefresh time.Duration `mapstructure:"equipments-refresh"`
	EquipmentsURI     url.URL

	FreeFloatingsURIStr    string        `mapstructure:"free-floatings-uri"`
	FreeFloatingsRefresh   time.Duration `mapstructure:"free-floatings-refresh"`
	FreeFloatingsURI       url.URL
	FreeFloatingsToken     string   `mapstructure:"free-floatings-token"`
	FreeFloatingsType      string   `mapstructure:"free-floatings-type"`
	FreeFloatingsUserName  string   `mapstructure:"free-floatings-username"`
	FreeFloatingsPassword  string   `mapstructure:"free-floatings-password"`
	FreeFloatingsProviders []string `mapstructure:"free-floatings-providers"`

	OccupancyFilesURIStr   string `mapstructure:"occupancy-files-uri"`
	OccupancyFilesURI      url.URL
	OccupancyNavitiaURIStr string `mapstructure:"occupancy-navitia-uri"`
	OccupancyNavitiaURI    url.URL
	OccupancyServiceURIStr string `mapstructure:"occupancy-service-uri"`
	OccupancyServiceURI    url.URL
	OccupancyNavitiaToken  string        `mapstructure:"occupancy-navitia-token"`
	OccupancyServiceToken  string        `mapstructure:"occupancy-service-token"`
	OccupancyRefresh       time.Duration `mapstructure:"occupancy-refresh"`
	OccupancyCleanVJ       time.Duration `mapstructure:"occupancy-clean-vj"`
	OccupancyCleanVO       time.Duration `mapstructure:"occupancy-clean-vo"`
	RouteScheduleRefresh   time.Duration `mapstructure:"routeschedule-refresh"`
	TimeZoneLocation       string        `mapstructure:"timezone-location"`

	PositionsFilesURIStr string `mapstructure:"positions-files-uri"`
	PositionsFilesURI    url.URL
	//PositionsNavitiaURIStr string `mapstructure:"positions-navitia-uri"`
	//PositionsNavitiaURI    url.URL
	PositionsServiceURIStr string `mapstructure:"positions-service-uri"`
	PositionsServiceURI    url.URL
	//PositionsNavitiaToken  string        `mapstructure:"positions-navitia-token"`
	PositionsServiceToken string        `mapstructure:"positions-service-token"`
	PositionsRefresh      time.Duration `mapstructure:"positions-refresh"`
	//PositionsCleanVJ       time.Duration `mapstructure:"positions-clean-vj"`
	PositionsCleanVP time.Duration `mapstructure:"positions-clean-vp"`

	LogLevel            string        `mapstructure:"log-level"`
	ConnectionTimeout   time.Duration `mapstructure:"connection-timeout"`
	JSONLog             bool          `mapstructure:"json-log"`
	FreeFloatingsActive bool          `mapstructure:"free-floatings-refresh-active"`
	OccupancyActive     bool          `mapstructure:"occupancy-service-refresh-active"`
	PositionsActive     bool          `mapstructure:"positions-service-refresh-active"`
	Connector           string        `mapstructure:"connector-type"`
}

func noneOf(args ...string) bool {
	for _, a := range args {
		if a != "" {
			return false
		}
	}
	return true
}

func GetConfig() (Config, error) {
	//Passing configurations for departures
	pflag.String("departures-uri", "",
		"format: [scheme:][//[userinfo@]host][/]path \nexample: sftp://forseti:pass@172.17.0.3:22/extract_edylic.txt")
	pflag.Duration("departures-refresh", 30*time.Second, "time between refresh of departures data")

	//Passing configurations for parkings
	pflag.String("parkings-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.Duration("parkings-refresh", 30*time.Second, "time between refresh of parkings data")

	//Passing configurations for equipments
	pflag.String("equipments-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.Duration("equipments-refresh", 30*time.Second, "time between refresh of equipments data")

	//Passing configurations for free-floatings
	pflag.String("free-floatings-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("free-floatings-token", "", "token for free floating source")
	pflag.Bool("free-floatings-refresh-active", false, "activate the periodic refresh of Fluctuo data")
	pflag.Duration("free-floatings-refresh", 30*time.Second, "time between refresh of vehicles in Fluctuo data")
	pflag.String("free-floatings-type", "fluctuo", "connector type to load data source")
	pflag.String("free-floatings-username", "", "username for getting API access tokens")
	pflag.String("free-floatings-password", "", "password for getting API access tokens")
	pflag.String("free-floatings-providers", "", "list of providers to get data \nexample: 19,127,392")

	//Passing configurations for vehicle_occupancies
	pflag.String("occupancy-files-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("occupancy-navitia-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("occupancy-navitia-token", "", "token for navitia")
	pflag.String("occupancy-service-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("occupancy-service-token", "", "token for prediction source")
	pflag.Bool("occupancy-service-refresh-active", false, "activate the periodic refresh of vehicle occupancy data")
	pflag.Duration("occupancy-refresh", 5*time.Minute, "time between refresh of predictions")
	pflag.Duration("routeschedule-refresh", 24*time.Hour, "time between refresh of RouteSchedules from navitia")
	pflag.Duration("occupancy-clean-vj", 24*time.Hour, "time between clean list of VehicleJourneys")
	pflag.Duration("occupancy-clean-vo", 2*time.Hour, "time between clean list of VehicleOccupancies")

	//Passing configurations for vehicle_positions
	pflag.String("positions-files-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("positions-service-uri", "", "format: [scheme:][//[userinfo@]host][/]path")
	pflag.String("positions-service-token", "", "token for positions source")
	pflag.Bool("positions-service-refresh-active", false, "activate the periodic refresh of vehicle positions data")
	pflag.Duration("positions-refresh", 5*time.Minute, "time between refresh of positions")
	pflag.Duration("positions-clean-vp", 2*time.Hour, "time between clean list of vehiclePositions")

	//Passing configurations for vehicle_occupancies and vehicle_positions
	pflag.String("timezone-location", "Europe/Paris", "timezone location")
	pflag.String("connector-type", "oditi", "connector type to load data source")

	//Passing globals configurations
	pflag.Duration("connection-timeout", 10*time.Second, "timeout to establish the ssh connection")
	pflag.Bool("json-log", false, "enable json logging")
	pflag.String("log-level", "debug", "log level: debug, info, warn, error")
	pflag.Parse()

	var config Config
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return config, errors.Wrap(err, "Impossible to parse flags")
	}
	viper.SetEnvPrefix("FORSETI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := viper.Unmarshal(&config); err != nil {
		return config, errors.Wrap(err, "Unmarshalling of flag failed")
	}

	if noneOf(config.DeparturesURIStr, config.ParkingsURIStr, config.EquipmentsURIStr, config.FreeFloatingsURIStr,
		config.OccupancyFilesURIStr, config.OccupancyNavitiaURIStr, config.OccupancyServiceURIStr,
		config.PositionsServiceURIStr) {
		return config, errors.New("no data provided at all. Please provide at lease one type of data")
	}

	type ConfigUri struct {
		configURIStr string
		configURI    *url.URL
	}

	for _, uri := range []ConfigUri{
		{config.DeparturesURIStr, &config.DeparturesURI},
		{config.ParkingsURIStr, &config.ParkingsURI},
		{config.EquipmentsURIStr, &config.EquipmentsURI},
		{config.FreeFloatingsURIStr, &config.FreeFloatingsURI},
		{config.OccupancyFilesURIStr, &config.OccupancyFilesURI},
		{config.OccupancyNavitiaURIStr, &config.OccupancyNavitiaURI},
		{config.OccupancyServiceURIStr, &config.OccupancyServiceURI},
		{config.PositionsServiceURIStr, &config.PositionsServiceURI},
	} {
		if url, err := url.Parse(uri.configURIStr); err != nil {
			logrus.Errorf("Unable to parse data url: %s", uri.configURIStr)
		} else {
			*uri.configURI = *url
		}
	}

	return config, nil
}

func main() {
	config, err := GetConfig()
	if err != nil {
		logrus.Fatalf("Impossible to load data at startup: %s", err)
	}

	initLog(config.JSONLog, config.LogLevel)
	manager := &manager.DataManager{}

	location, _ := time.LoadLocation(config.TimeZoneLocation)

	// create API router
	router := api.SetupRouter(manager, nil)

	// With equipments
	Equipments(manager, &config, router)

	// With freefloating
	FreeFloating(manager, &config, router)

	// With departures
	Departures(manager, &config, router)

	// With parkings
	Parkings(manager, &config, router)

	// With vehicle occupancies
	VehicleOccupancies(manager, &config, router, location)

	// With vehicle positions
	VehiclePositions(manager, &config, router, location)

	// start router
	err = router.Run()
	if err != nil {
		logrus.Fatalf("Impossible to start gin: %s", err)
	}
}

func Equipments(manager *manager.DataManager, config *Config, router *gin.Engine) {
	if len(config.EquipmentsURI.String()) == 0 || config.EquipmentsRefresh.Seconds() <= 0 {
		logrus.Debug("Equipments is disabled")
		return
	}
	equipmentsContext := &equipments.EquipmentsContext{}
	manager.SetEquipmentsContext(equipmentsContext)
	go equipments.RefreshEquipmentLoop(equipmentsContext, config.EquipmentsURI,
		config.EquipmentsRefresh, config.ConnectionTimeout)
	equipments.AddEquipmentsEntryPoint(router, equipmentsContext)
}

func FreeFloating(manager *manager.DataManager, config *Config, router *gin.Engine) {
	if len(config.FreeFloatingsURI.String()) == 0 || config.FreeFloatingsRefresh.Seconds() <= 0 {
		logrus.Debug("FreeFloating is disabled")
		return
	}

	freeFloatingsContext := &freefloatings.FreeFloatingsContext{}
	manager.SetFreeFloatingsContext(freeFloatingsContext)
	freefloatings.AddFreeFloatingsEntryPoint(router, freeFloatingsContext)
	freeFloatingsContext.ManageFreeFloatingsStatus(config.FreeFloatingsActive)

	if config.FreeFloatingsType == string(connectors.Connector_CITIZ) {
		var c = citiz.CitizContext{}

		c.InitContext(config.FreeFloatingsURI, config.FreeFloatingsRefresh, config.FreeFloatingsProviders,
			config.ConnectionTimeout, config.FreeFloatingsUserName, config.FreeFloatingsPassword)

		go c.RefreshFreeFloatingLoop(freeFloatingsContext)

	} else if config.FreeFloatingsType == string(connectors.Connector_FLUCTUO) {
		var f = fluctuo.FluctuoContext{}

		f.InitContext(config.FreeFloatingsURI, config.FreeFloatingsRefresh, config.FreeFloatingsToken,
			config.ConnectionTimeout)

		go f.RefreshFreeFloatingLoop(freeFloatingsContext)
	}
}

func Departures(manager *manager.DataManager, config *Config, router *gin.Engine) {
	if len(config.DeparturesURI.String()) == 0 || config.DeparturesRefresh.Seconds() <= 0 {
		logrus.Debug("Departures is disabled")
		return
	}
	departuresContext := &departures.DeparturesContext{}
	manager.SetDeparturesContext(departuresContext)
	go departures.RefreshDeparturesLoop(departuresContext, config.DeparturesURI,
		config.DeparturesRefresh, config.ConnectionTimeout)
	departures.AddDeparturesEntryPoint(router, departuresContext)
}

func Parkings(manager *manager.DataManager, config *Config, router *gin.Engine) {
	if len(config.ParkingsURI.String()) == 0 || config.ParkingsRefresh.Seconds() <= 0 {
		logrus.Debug("Parkings is disabled")
		return
	}
	parkingsContext := &parkings.ParkingsContext{}
	manager.SetParkingsContext(parkingsContext)
	go parkings.RefreshParkingsLoop(parkingsContext, config.ParkingsURI,
		config.ParkingsRefresh, config.ConnectionTimeout)
	parkings.AddParkingsEntryPoint(router, parkingsContext)
}

func VehicleOccupancies(manager *manager.DataManager, config *Config, router *gin.Engine, location *time.Location) {
	if len(config.OccupancyNavitiaURI.String()) == 0 || len(config.OccupancyServiceURI.String()) == 0 {
		logrus.Debug("Vehicle occupancies is disabled")
		return
	}

	var err error

	if config.Connector == string(connectors.Connector_ODITI) {
		var vehicleOccupanciesOditiContext vehicleoccupanciesv2.IVehicleOccupancy
		vehicleOccupanciesOditiContext, err = vehicleoccupanciesv2.VehicleOccupancyFactory(string(connectors.Connector_ODITI))
		if err != nil {
			logrus.Error(err)
			return
		}
		manager.SetVehicleOccupanciesContext(vehicleOccupanciesOditiContext)

		vehicleOccupanciesOditiContext.InitContext(config.OccupancyFilesURI, config.OccupancyServiceURI,
			config.OccupancyServiceToken, config.OccupancyNavitiaURI, config.OccupancyNavitiaToken,
			config.OccupancyRefresh, config.OccupancyCleanVJ, config.OccupancyCleanVO, config.ConnectionTimeout,
			location, config.OccupancyActive)

		go vehicleOccupanciesOditiContext.RefreshVehicleOccupanciesLoop(config.OccupancyServiceURI,
			config.OccupancyServiceToken, config.OccupancyNavitiaURI, config.OccupancyNavitiaToken,
			config.OccupancyRefresh, config.OccupancyCleanVJ, config.OccupancyCleanVO, config.ConnectionTimeout,
			location)
		vehicleoccupanciesv2.AddVehicleOccupanciesEntryPoint(router, vehicleOccupanciesOditiContext)

	} else if config.Connector == string(connectors.Connector_GRFS_RT) {
		var vehicleOccupanciesContext vehicleoccupanciesv2.IVehicleOccupancy
		vehicleOccupanciesContext, err = vehicleoccupanciesv2.VehicleOccupancyFactory(string(connectors.Connector_GRFS_RT))
		if err != nil {
			logrus.Error(err)
			return
		}

		manager.SetVehicleOccupanciesContext(vehicleOccupanciesContext)

		vehicleOccupanciesContext.InitContext(config.OccupancyFilesURI, config.OccupancyServiceURI,
			config.OccupancyServiceToken, config.OccupancyNavitiaURI, config.OccupancyNavitiaToken,
			config.OccupancyRefresh, config.OccupancyCleanVJ, config.OccupancyCleanVO, config.ConnectionTimeout,
			location, config.OccupancyActive)

		go vehicleOccupanciesContext.RefreshVehicleOccupanciesLoop(config.OccupancyServiceURI,
			config.OccupancyServiceToken, config.OccupancyNavitiaURI, config.OccupancyNavitiaToken,
			config.OccupancyRefresh, config.OccupancyCleanVJ, config.OccupancyCleanVO, config.ConnectionTimeout,
			location)
		vehicleoccupanciesv2.AddVehicleOccupanciesEntryPoint(router, vehicleOccupanciesContext)
	} else {
		logrus.Error("Wrong vehicleoccupancy type passed")
		return
	}
}

func VehiclePositions(manager *manager.DataManager, config *Config, router *gin.Engine, location *time.Location) {
	if len(config.PositionsServiceURI.String()) == 0 {
		logrus.Debug("Vehicle positions is disabled")
		return
	}

	var vehiclePositionsContext vehiclepositions.IConnectors
	var err error

	if config.Connector == string(connectors.Connector_GRFS_RT) {
		vehiclePositionsContext, err = vehiclepositions.ConnectorFactory(string(connectors.Connector_GRFS_RT))
		if err != nil {
			logrus.Error(err)
			return
		}
	} else {
		logrus.Error("Wrong connector type passed")
		return
	}

	manager.SetVehiclePositionsContext(vehiclePositionsContext)

	vehiclePositionsContext.InitContext(config.PositionsFilesURI, config.PositionsServiceURI,
		config.PositionsServiceToken, config.PositionsRefresh, config.PositionsCleanVP, config.ConnectionTimeout,
		location, config.PositionsActive)
	go vehiclePositionsContext.RefreshVehiclePositionsLoop()

	vehiclepositions.AddVehiclePositionsEntryPoint(router, vehiclePositionsContext)
}

func initLog(jsonLog bool, logLevel string) {
	if jsonLog {
		// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetOutput(os.Stdout)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetLevel(level)
}
