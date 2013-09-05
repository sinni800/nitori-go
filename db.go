package main

import (
	"crypto"
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"time"
)

var MongoSession *mgo.Session
var MongoDB *mgo.Database

func connectDB() {
	MongoSession, _ = mgo.Dial(conf.Mongo.MongoAddress)
	MongoDB = MongoSession.DB(conf.Mongo.MongoDatabase)
}

func reconnectDB() error {
	if MongoSession == nil || MongoSession.Ping() != nil {
		MongoSession, err := mgo.Dial(conf.Mongo.MongoAddress)
		if err != nil {
			return err
		}
		MongoDB = MongoSession.DB(conf.Mongo.MongoDatabase)
	}
	return nil
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

func SaveToDB(collection string, stuff bson.M) error {
	coll := MongoDB.C(collection)
	return coll.Insert(stuff)
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

func GetFromDB(collection string, selector bson.M) ([]bson.M, error) {
	coll := MongoDB.C(collection)
	output := make([]bson.M, 0, 0)
	return output, coll.Find(selector).All(&output)
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
	coll.Find(nil).Limit(-1).Skip(rand.Intn(count)).One(&result)

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

func DeleteNamedFromDB(collection string, name string) (err error) {
	return MongoDB.C(collection).Remove(bson.M{"name": name})
}

func DeleteFromDB(collection string, selector bson.M) (err error) {
	return MongoDB.C(collection).Remove(selector)
}
