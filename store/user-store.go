package store

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	errType        = errors.New("wrong database type")
	errEmplyConfig = errors.New("empty configuration for datastore")
)

const (
	tableUsers   = "userCollection"
	dbUsers      = "userDB"
	dbBTCMempool = "BTCMempool" // TODO: create rates store
	tableRates   = "Rates"      // and send those two fields there
)

const defaultMongoDBaddr = "192.168.0.121:27017"

type UserStore interface {
	//GetSession()
	GetUserByDevice(device bson.M, user *User)
	Update(sel, update bson.M) error
	Insert(user User) error
	Close() error
	FindUser(query bson.M, user *User) error
	UpdateUser(sel bson.M, user *User) error
	GetAllRates(sortBy string, rates *[]RatesRecord) error //add to rates store
}

type MongoUserStore struct {
	address    string
	session    *mgo.Session
	usersData  *mgo.Collection
	ratessData *mgo.Collection
}

func (mongo *MongoUserStore) UpdateUser(sel bson.M, user *User) error {
	return mongo.usersData.Update(sel, user)
}

func (mongo *MongoUserStore) GetUserByDevice(device bson.M, user *User) { // rename GetUserByToken
	mongo.usersData.Find(device).One(user)
	return // why?
}

func (mongo *MongoUserStore) Update(sel, update bson.M) error {
	return mongo.usersData.Update(sel, update)
}

func (mongo *MongoUserStore) FindUser(query bson.M, user *User) error {
	return mongo.usersData.Find(query).One(user)
}

func (mongo *MongoUserStore) Insert(user User) error {
	return mongo.usersData.Insert(user)
}
func (mongo *MongoUserStore) GetAllRates(sortBy string, rates *[]RatesRecord) error {
	return mongo.ratessData.Find(nil).Sort(sortBy).All(rates)
}

func InitUserStore(address string) (UserStore, error) {
	if address == "" {
		log.Printf("[INFO] mongo db address: will be used %s\n", defaultMongoDBaddr)
		address = defaultMongoDBaddr
	}

	uStore := &MongoUserStore{
		address: address,
	}
	session, err := mgo.Dial(address)
	if err != nil {
		return nil, err
	}
	uStore.session = session
	uStore.usersData = uStore.session.DB(dbUsers).C(tableUsers)
	uStore.ratessData = uStore.session.DB(dbBTCMempool).C(tableRates)
	return uStore, nil
}

func (mongoUserData *MongoUserStore) Close() error {
	mongoUserData.session.Close()
	return nil
}
