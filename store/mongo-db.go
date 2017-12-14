package store

/*
type mongoDB struct {
	*mgo.Session
	config *MongoConfig
}

//MongoConfig is a config for mongo database
type MongoConfig struct {
	Type     string
	User     string
	Password string
	NameDB   string
	Address  string
}

func getMongoConfig(rawConf map[string]interface{}) *MongoConfig {
	return nil
}

func (mDB *mongoDB) AddUser(user *User) error {
	session := mDB.Copy()
	// defer session.Close()
	// TODO: check if user exists
	users := session.DB(mDB.config.NameDB).C(tableUsers)
	return users.Insert(user)
}

func (mDB *mongoDB) Close() error {
	return mDB.Close()
}

func (mDB *mongoDB) FindMember(id int) (User, error) {
	// session := ms.Copy()
	// defer session.Close()
	// personnel := session.DB("kek").C("users")
	// cm := CrewMember{}
	// err := personnel.Find(bson.M{"id": id}).One(&cm)
	return User{}, nil
}*/
