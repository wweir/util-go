package influx

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/wweir/utils/util"
)

// DB is a influxdb client wrap
type DB struct {
	client.Client
	database string
	tags     map[string]string

	count       int
	flushCount  int           // default 512
	flushExpire time.Duration // default 200 Millisecond

	once    sync.Once
	wg      sync.WaitGroup
	closeCh chan struct{}
	pointCh chan *client.Point
	errCh   chan error
}

var defaultDB *DB

// Client return the raw Client of influxdata
func Client() client.Client {
	return defaultDB.Client
}

// InitInfluxDB build InfluxDB client and set the default database
func InitInfluxDB(addr, user, passwd, database string) (err error) {
	defaultDB, err = ConnectInfluxDB(addr, user, passwd, database)
	return
}

// Ping is the same as client ping
func Ping(timeout time.Duration) (time.Duration, string, error) {
	return defaultDB.Client.Ping(timeout)
}

// SetFlushFlags set buffer flags for batch write on default DB
func SetFlushFlags(maxFlushCount int, expireTime time.Duration) {
	defaultDB.SetFlushFlags(maxFlushCount, expireTime)
}

// Write is a buffer write wrap on default influx db
func Write(table string, datas ...map[string]interface{}) error {
	return defaultDB.Write(table, datas...)
}

// Close flush and close the influxdb default client
func Close() error {
	return defaultDB.Close()
}

// Query is the wrapped query of default db
func Query(precision, command string, a ...interface{}) (*client.Response, error) {
	return defaultDB.Query(precision, command, a...)
}

// QueryAsChunk is the wrapped QueryAsChunk of default db
func QueryAsChunk(precision, command string, a ...interface{}) (*client.ChunkedResponse, error) {
	return defaultDB.QueryAsChunk(precision, command, a...)
}

// ConnectInfluxDB build InfluxDB client
func ConnectInfluxDB(addr, user, passwd, database string) (*DB, error) {
	if addr == "" {
		addr = os.Getenv("INFLUX_ADDR")
	}
	if addr == "" {
		return nil, errors.New("please set endpoint by env INFLUX_ADDR or manually setting")
	}
	if user == "" {
		user = os.Getenv("INFLUX_USER")
	}
	if passwd == "" {
		passwd = os.Getenv("INFLUX_PASSWD")
	}
	if database == "" {
		database = os.Getenv("INFLUX_DATABASE")
	}

	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: passwd,
		Timeout:  5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if _, _, err := cli.Ping(time.Second); err != nil {
		return nil, err
	}

	resp, err := cli.Query(client.NewQuery("CREATE DATABASE "+database, "", ""))
	if err := util.FirstErr(err, resp); err != nil {
		cli.Close()
		return nil, err
	}

	return &DB{
		Client:   cli,
		database: database,
		tags:     map[string]string{},

		flushCount:  512,
		flushExpire: 200 * time.Millisecond,

		closeCh: make(chan struct{}),
		pointCh: make(chan *client.Point),
		errCh:   make(chan error),
	}, nil
}

// SetFlushFlags set buffer flags for batch write
func (db *DB) SetFlushFlags(maxFlushCount int, expireTime time.Duration) {
	db.flushCount = maxFlushCount
	db.flushExpire = expireTime
}

// Write is a buffer write wrap
func (db *DB) Write(table string, datas ...map[string]interface{}) error {
	// Start the daemon goroutine
	db.once.Do(db.writeDaemon)

	select {
	case <-db.closeCh:
		return errors.New("not able to write on a closed influxdb client")
	default:
	}

	db.wg.Add(1)
	for _, data := range datas {
		point, err := client.NewPoint(table, db.tags, data, time.Now())
		if err != nil {
			return err
		}

		db.pointCh <- point
	}
	db.wg.Done()

	// return err while write
	select {
	case err := <-db.errCh:
		return err
	default:
		return nil
	}
}
func (db *DB) newBatchPoint() client.BatchPoints {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db.database,
		Precision: "us",
	})
	return bp
}

func (db *DB) writeDaemon() {
	go func() {
		tick := time.Tick(db.flushExpire)
		bp := db.newBatchPoint()

		for {
			select {
			case <-tick:
			case point, ok := <-db.pointCh:
				if ok {
					bp.AddPoint(point)
					db.count++
					if db.count < db.flushCount {
						continue
					}

					// db.pointCh is closed by Close function, and all data flushed
				} else if db.count == 0 {
					return
				}
			}

			if db.count != 0 {
				if err := db.Client.Write(bp); err != nil {
					db.errCh <- err
				}

				bp = db.newBatchPoint()
				db.count = 0
			}
		}
	}()
}

// Close flush and close the influxdb client
func (db *DB) Close() error {
	close(db.closeCh)
	db.wg.Wait() // wait for all write done
	close(db.pointCh)

	errs := []error{}
	for {
		if err, ok := <-db.errCh; ok {
			errs = append(errs, err)

		} else {
			// wait for write daemon done
			errs = append(errs, db.Client.Close())
			return util.MergeErr(errs...)
		}
	}
}

// Query is the wrapped query
func (db *DB) Query(precision, command string, a ...interface{}) (*client.Response, error) {
	q := client.NewQuery(fmt.Sprintf(command, a...), db.database, precision)
	return defaultDB.Client.Query(q)
}

// QueryAsChunk is the wrapped query
func (db *DB) QueryAsChunk(precision, command string, a ...interface{}) (*client.ChunkedResponse, error) {
	q := client.NewQuery(fmt.Sprintf(command, a...), db.database, precision)
	return defaultDB.Client.QueryAsChunk(q)
}

func RespToKeyPair(resp *client.Response, e error) (keys []string, vals [][]interface{}, err error) {
	if err = util.FirstErr(e, resp); err != nil {
		return nil, nil, err
	} else if count := len(resp.Results); count != 1 {
		return nil, nil, fmt.Errorf("influx return results should be 1 but not %d", count)
	} else if count := len(resp.Results[0].Series); count != 1 {
		return nil, nil, fmt.Errorf("influx return series should be 1 but not %d", count)
	}

	return resp.Results[0].Series[0].Columns, resp.Results[0].Series[0].Values, nil
}
