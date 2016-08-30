package db

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Connection struct {
	config    Config
	dialInfo  *mgo.DialInfo
	Session   *mgo.Session
	OplogChan chan bson.M
	Mutex     sync.Mutex
	Optime    bson.MongoTimestamp
	NOplog    uint64
	NDone     uint64
}

func NewConnection(config Config) (*Connection, error) {
	c := new(Connection)
	c.config = config
	c.OplogChan = make(chan bson.M, 1000)
	var err error

	if c.dialInfo, err = mgo.ParseURL(c.config.URI); err != nil {
		panic(fmt.Sprintf("cannot parse given URI %s due to error: %s", c.config.URI, err.Error()))
	}

	if c.config.SSL {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = true
		c.dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	c.dialInfo.Timeout = time.Second * 3
	c.dialInfo.Username = c.config.Creds.Username
	c.dialInfo.Password = c.config.Creds.Password
	c.dialInfo.Mechanism = c.config.Creds.Mechanism

	//session, err := mgo.DialWithTimeout(c.config.URI, time.Second*3)
	session, err := mgo.DialWithInfo(c.dialInfo)
	if err != nil {
		return nil, err
	}

	c.Session = session
	c.Session.SetMode(mgo.SecondaryPreferred, true)
	c.Session.Login(&c.config.Creds)

	return c, nil
}

func (c *Connection) Databases() ([]string, error) {
	dbnames, err := c.Session.DatabaseNames()
	if err != nil {
		return nil, err
	}

	var slice []string

	for _, dbname := range dbnames {
		if dbname != "local" && dbname != "admin" {
			slice = append(slice, dbname)
		}
	}
	return slice, nil
}

func (c *Connection) databaseRegExs() ([]bson.RegEx, error) {
	dbnames, err := c.Session.DatabaseNames()
	if err != nil {
		return nil, err
	}

	var slice []bson.RegEx

	for _, dbname := range dbnames {
		if dbname != "local" && dbname != "admin" {
			slice = append(slice, bson.RegEx{Pattern: dbname + ".*"})
		}
	}
	return slice, nil
}

func (c *Connection) Push(oplog bson.M) {
	c.OplogChan <- oplog
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.NOplog++
}

func (c *Connection) SyncOplog(dst *Connection) error {
	var (
		query  bson.M
		result bson.M
		iter   *mgo.Iter
		sec    bson.MongoTimestamp
		ord    bson.MongoTimestamp
		err    error
	)

	oplog := c.Session.DB("local").C("oplog.rs")

	query = bson.M{
		"ts": bson.M{"$gt": bson.MongoTimestamp(time.Now().Unix()<<32 + time.Now().Unix())},
	}

	if viper.GetInt("since") > 0 {
		sec = bson.MongoTimestamp(viper.GetInt("since"))
		ord = bson.MongoTimestamp(viper.GetInt("ordinal"))
		query["ts"] = bson.M{"$gt": sec<<32 + ord}
	}

	dbnames, _ := c.databaseRegExs()
	if len(dbnames) > 0 {
		query["ns"] = bson.M{"$in": dbnames}
	} else {
		return fmt.Errorf("No databases found")
	}

	iter = oplog.Find(query).Tail(1 * time.Second)
	for {
		for iter.Next(&result) {
			ts, ok := result["ts"].(bson.MongoTimestamp)
			if ok {
				query["ts"] = bson.M{"$gt": ts}

				op := result["op"]
				ns := result["ns"]

				switch op {
				case "i": // insert
					dbname := strings.Split(ns.(string), ".")[0]
					collname := strings.Split(ns.(string), ".")[1]
					//id := result["o"].(bson.M)["_id"]
					//if _, err := dst.Session.DB(dbname).C(collname).UpsertId(id, result["o"]); err != nil {
					if err := dst.Session.DB(dbname).C(collname).Insert(result["o"]); err != nil {
						log.Println("insert", err)
					}
				case "u": // update
					dbname := strings.Split(ns.(string), ".")[0]
					collname := strings.Split(ns.(string), ".")[1]
					if err := dst.Session.DB(dbname).C(collname).Update(result["o2"], result["o"]); err != nil {
						log.Println("update", err)
					}
				case "d": // delete
					dbname := strings.Split(ns.(string), ".")[0]
					collname := strings.Split(ns.(string), ".")[1]
					if err := dst.Session.DB(dbname).C(collname).Remove(result["o"]); err != nil {
						log.Println("delete", err)
					}
				case "c": // command
					//dbname := strings.Split(ns.(string), ".")[0]
					//if err := dst.Session.DB(dbname).Run(result["o"], nil); err != nil {
					//	log.Println("command", err)
					//}
				case "n": // no-op
					// do nothing
				default:
				}

				fmt.Printf("\rOn %d\n", ts)
			} else {
				panic(fmt.Sprintf("`ts` is not found in result: %v\n", result))
			}
		}

		err = iter.Err()
		if err != nil {
			iter.Close()
			return err
		}

		if iter.Timeout() {
			if viper.GetBool("fast_stop") {
				iter.Close()
				return nil
			}
			continue
		}

		iter = oplog.Find(query).Tail(1 * time.Second)
	}
}
