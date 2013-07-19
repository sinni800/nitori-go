package main

import (
	"crypto"
	"errors"
	"flag"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"time"
)

var MongoSession *mgo.Session
var MongoDB *mgo.Database

var fMongoAddress *string = flag.String("mongo.address", "mongodb://user:pass@host/", "Mongo Database Address")
var fMongoDatabase *string = flag.String("mongo.dbname", "dbname", "Mongo Database Name")
var fUseMongo *bool = flag.Bool("mongo.active", false, "Use Mongo Database")
var MongoAddress string
var MongoDatabase string
var UseMongo bool

func init() {
	flag.Parse()
	MongoAddress = *fMongoAddress
	MongoDatabase = *fMongoDatabase
	UseMongo = *fUseMongo
}

func connectDB() {
	if UseMongo {
		MongoSession, _ = mgo.Dial(MongoAddress)
		MongoDB = MongoSession.DB(MongoDatabase)
	}

}

func Authenticate(username string, password string) bool {
	md5 := crypto.MD5.New()
	md5.Write([]byte(password))
	passwordhash := fmt.Sprintf("%x", md5.Sum(nil))

	var Users *mgo.Collection = MongoDB.C("users")

	qry := Users.Find(bson.M{"username": username, "password": passwordhash})
	var result bson.M
	err := qry.One(&result)
	if err == nil {
		if result["username"] == username && result["password"] == passwordhash {
			return true
		}
	}
	return false
}

func AuthenticateHashedPW(username string, passwordhash string) bool {
	var Users *mgo.Collection = MongoDB.C("users")

	qry := Users.Find(bson.M{"username": username, "password": passwordhash})
	var result bson.M
	err := qry.One(&result)
	if err == nil {
		if result["username"] == username && result["password"] == passwordhash {
			return true
		}
	}
	return false
}

func SaveToDB(collection string, stuff bson.M) {
	coll := MongoDB.C(collection)
	coll.Insert(stuff)
}

func ExistsInDB(collection string, stuff bson.M) bool {
	coll := MongoDB.C(collection)
	if count, _ := coll.Find(stuff).Count(); count > 0 {
		return true
	} else {
		return false
	}
	return false
}

func GetRandomFromDB(collection string) (bson.M, error) {
	coll := MongoDB.C(collection)
	rand.Seed(time.Now().UnixNano())
	count, err := coll.Count()

	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, err
	}

	result := bson.M{}
	coll.Find(nil).Skip(rand.Intn(count - 1)).One(&result)

	return result, nil
}

func GetAndDeleteFirstFromDB(collection string) (bson.M, error) {
	coll := MongoDB.C(collection)
	if count, _ := coll.Count(); count > 0 {
		var result bson.M
		coll.Find(nil).One(&result)
		coll.Remove(result)
		return result, nil
	}
	return nil, errors.New("Nothing inside")
}

func GetNamedFromDB(collection string, name string) (bson.M, error) {
	coll := MongoDB.C(collection)
	var result bson.M
	query := coll.Find(bson.M{"name": name})
	Count, err := query.Count()

	if err != nil {
		return nil, err
	}

	if Count > 0 {
		err = query.One(&result)

		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return nil, errors.New("Nothing returned")
}

func SaveNamedToDB(collection string, stuff bson.M) (err error) {
	coll := MongoDB.C(collection)

	query := coll.Find(bson.M{"name": stuff["name"]})
	Count, err := query.Count()

	if err != nil {
		return err
	}

	if Count > 0 {
		err = coll.Update(bson.M{"name": stuff["name"]}, stuff)

		if err != nil {
			return err
		}
	} else {
		err = coll.Insert(stuff)

		if err != nil {
			return err
		}
	}

	return
}
