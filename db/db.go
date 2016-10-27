package db

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
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

type ApplyOpsResponse struct {
	Ok     bool   `bson:"ok"`
	ErrMsg string `bson:"errmsg"`
}

type Oplog struct {
	Timestamp bson.MongoTimestamp `bson:"ts"`
	HistoryID int64               `bson:"h"`
	Version   int                 `bson:"v"`
	Operation string              `bson:"op"`
	Namespace string              `bson:"ns"`
	Object    bson.D              `bson:"o"`
	Query     bson.D              `bson:"o2"`
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
	session.SetSocketTimeout(1 * time.Minute)
	session.SetPrefetch(1.0)

	c.Session = session
	//c.Session.SetMode(mgo.SecondaryPreferred, true)
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
		restore_query bson.M
		tail_query    bson.M
		oplogEntry    Oplog
		iter          *mgo.Iter
		sec           bson.MongoTimestamp
		ord           bson.MongoTimestamp
		err           error
	)

	oplog := c.Session.DB("local").C("oplog.rs")

	var head_result struct {
		Timestamp bson.MongoTimestamp `bson:"ts"`
	}
	err = oplog.Find(nil).Sort("-$natural").Limit(1).One(&head_result)

	restore_query = bson.M{
		"ts": bson.M{"$gt": bson.MongoTimestamp(time.Now().Unix()<<32 + time.Now().Unix())},
	}

	tail_query = bson.M{
		"ts": bson.M{"$gt": head_result.Timestamp},
	}

	if viper.GetInt("since") > 0 {
		sec = bson.MongoTimestamp(viper.GetInt("since"))
		ord = bson.MongoTimestamp(viper.GetInt("ordinal"))
		restore_query["ts"] = bson.M{"$gt": bson.MongoTimestamp(sec<<32 + ord)}
	}

	dbnames, _ := c.databaseRegExs()
	if len(dbnames) > 0 {
		restore_query["ns"] = bson.M{"$in": dbnames}
		tail_query["ns"] = bson.M{"$in": dbnames}
	} else {
		return fmt.Errorf("No databases found")
	}

	applyOpsResponse := ApplyOpsResponse{}
	opCount := 0

	if viper.GetInt("since") > 0 {
		fmt.Println("Restoring oplog...")
		iter = oplog.Find(restore_query).Iter()
		for iter.Next(&oplogEntry) {
			tail_query = bson.M{
				"ts": bson.M{"$gte": oplogEntry.Timestamp},
			}

			// skip noops
			if oplogEntry.Operation == "n" {
				log.Printf("skipping no-op for namespace `%v`", oplogEntry.Namespace)
				continue
			}
			opCount++

			// apply the operation
			opsToApply := []Oplog{oplogEntry}
			err := dst.Session.Run(bson.M{"applyOps": opsToApply}, &applyOpsResponse)
			if err != nil {
				return fmt.Errorf("error applying ops: %v", err)
			}

			// check the server's response for an issue
			if !applyOpsResponse.Ok {
				return fmt.Errorf("server gave error applying ops: %v", applyOpsResponse.ErrMsg)
			}

			fmt.Println(opCount)
		}

		err = iter.Err()
		if err != nil {
			iter.Close()
			return err
		}
	}

	fmt.Println("Tailing...")
	iter = oplog.Find(tail_query).Tail(1 * time.Second)
	for {
		for iter.Next(&oplogEntry) {
			// skip noops
			if oplogEntry.Operation == "n" {
				log.Printf("skipping no-op for namespace `%v`", oplogEntry.Namespace)
				continue
			}
			opCount++

			// apply the operation
			opsToApply := []Oplog{oplogEntry}
			err := dst.Session.Run(bson.M{"applyOps": opsToApply}, &applyOpsResponse)
			if err != nil {
				return fmt.Errorf("error applying ops: %v", err)
			}

			// check the server's response for an issue
			if !applyOpsResponse.Ok {
				return fmt.Errorf("server gave error applying ops: %v", applyOpsResponse.ErrMsg)
			}

			fmt.Println(opCount)
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

		iter = oplog.Find(tail_query).Tail(1 * time.Second)
	}
}
