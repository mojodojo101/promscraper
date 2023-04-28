package main 

import (
	"context"
	"time"
	"log"
	"fmt"
	"os"
	"flag"
	"database/sql"
	"io/ioutil"
	"bytes"

	"promscraper/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type OpticsData struct {
	ID        string
	Instance  string
	Interface string
	Job       string
	Name      string
	Target    string
	Value     float64
}

var (
	configFile    = flag.String("config-file", "config.yml", "Path to config file")
	logFile    = flag.String("log-file", "log.txt", "Path to config file")
	cfg           *config.Config
)

func loadConfig() (*config.Config, error) {
	log.Println("Loading config from", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}



func main() {

	flag.Parse()


	cfg , err := loadConfig()
	if err != nil {
		log.Fatalf("Could not load config file. %v", err)
	}

	file, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	
	log.SetOutput(file)

	// Create a Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: cfg.Prometheus.Endpoint,
	})

	if err != nil {
		log.Printf("%s",err)
		panic(err)
	}

	// Create a query API client
	queryClient := v1.NewAPI(client)

	// Set the PromQL query to retrieve the data
	query := `mikrotik_optics_rx_power{job="mikrotik_plc_monitor", interface=~"sfp.+", id=~".+sfp.+"}`

	// Query Prometheus for the data
	result, warnings, err := queryClient.Query(context.Background(), query, time.Now())
	if err != nil {
		log.Printf("%s",err)
		panic(err)
	}
	if warnings != nil {
		log.Println("Warnings encountered during query:")
		for _, warning := range warnings {
			log.Println(warning)
		}
	}

	// Convert the query result to a slice of OpticsData structs
	data := make([]OpticsData, 0)
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			data = append(data, OpticsData{
				ID:        string(sample.Metric["id"]),
				Instance:  string(sample.Metric["instance"]),
				Interface: string(sample.Metric["interface"]),
				Job:       string(sample.Metric["job"]),
				Name:      string(sample.Metric["name"]),
				Target:    string(sample.Metric["target"]),
				Value:     float64(sample.Value),
			})
		}
	}

	// Print the retrieved data
	for _, d := range data {
		log.Printf("Retrieved Record: %+v\n", d)
		// Extract the target IP address and interface name

		// Connect to the MySQL database
		dbConnection :=fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg.Sql.User, cfg.Sql.Password, cfg.Sql.Address , cfg.Sql.Name)	
		db, err := sql.Open("mysql", dbConnection)
		if err != nil {
			log.Printf("%s",err)
			panic(err)
		}
		defer db.Close()

		// Query the database for the monitor and SFP names
		var sfpID int
		err = db.QueryRow("SELECT s.sfp_id FROM SFPs s INNER JOIN PLC_Monitors m on s.monitor_id=m.monitor_id WHERE monitor_ip=? AND sfp_name=?", d.Target, d.Interface).Scan(&sfpID)
		if err != nil {
			log.Printf("%s",err)
			panic(err)
		}

		// Insert the light level value into the database
		_, err = db.Exec("UPDATE SFPs SET light_level=? where sfp_id=?", d.Value, sfpID)
		if err != nil {
			log.Printf("%s",err)
			panic(err)
		}


	}
}
