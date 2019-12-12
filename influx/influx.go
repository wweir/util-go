package influx

import (
	"errors"
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

// InitInfluxDB build InfluxDB client and set the default database
func InitInfluxDB(addr, user, passwd, database string) (err error) {
	defaultDB, err = ConnectInfluxDB(addr, user, passwd, database)
	return
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
	if err := util.FirstErr(err, resp.Error()); err != nil {
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

// SetFlushFlags set buffer flags for batch write on default DB
func SetFlushFlags(maxFlushCount int, expireTime time.Duration) {
	defaultDB.SetFlushFlags(maxFlushCount, expireTime)
}

// SetFlushFlags set buffer flags for batch write
func (db *DB) SetFlushFlags(maxFlushCount int, expireTime time.Duration) {
	db.flushCount = maxFlushCount
	db.flushExpire = expireTime
}

// Write is a buffer write wrap on default influx db
func Write(table string, datas ...map[string]interface{}) error {
	return defaultDB.Write(table, datas...)
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

// Close flush and close the influxdb default client
func Close() error {
	return defaultDB.Close()
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

func Ping(timeout time.Duration) (time.Duration, string, error) {
	return defaultDB.Client.Ping(timeout)
}
func Query(q client.Query) (*client.Response, error) {
	return defaultDB.Client.Query(q)
}
func QueryAsChunk(q client.Query) (*client.ChunkedResponse, error) {
	return defaultDB.Client.QueryAsChunk(q)
}
