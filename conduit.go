package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/context"
)

const connectionString = "mongodb://localhost/"

type userBasicData struct {
	UID            string `bson:"uid"`
	Name           string `bson:"name"`
	Sex            string `bson:"sex"`
	Email          string `bson:"email"`
	ProfilePicture string `bson:"profilepicture"`
}
type userProjects struct {
	UID         string `bson:"uid"`
	ProjectName string `bson:"projectname"`
	ProjectDesc string `bson:"projectdesc"`
}
type joinedProjects struct {
	UID         string `bson:"uid"`
	DUID        string `bson:"duid"`
	ProjectName string `bson:"projectname"`
	ProjectDesc string `bson:"projectdesc"`
}

func main() {
	// I don't even know if we need this registration thing, G+ can populate the profile just fine, why risk sending everything to server?
	http.HandleFunc("/registerLogin", registerLogin)
	http.HandleFunc("/addProjects", addProjects)
	http.HandleFunc("/joinProjects", joinProjects)
	http.HandleFunc("/yourJoinedProjects", yourJoinedProjects)
	http.HandleFunc("/getProjectList", getProjectList)

	//TODO:
	// Api for availableProjects, bring all projects except that user's project whose profile you're logged in with

	fmt.Println("Server listening at port 8888")
	http.ListenAndServe(":8888", context.ClearHandler(http.DefaultServeMux))
}

func registerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var user userBasicData

		user.UID = r.FormValue("uid")
		user.Name = r.FormValue("uname")
		user.Sex = r.FormValue("usex")
		user.Email = r.FormValue("umail")
		user.ProfilePicture = r.FormValue("upic")

		err := checkIfRegistered(user.UID)
		if err == nil {
			up := checkAndUpdate(user)
			log.Println(up)
			log.Println("Log In Success")

			fmt.Fprintf(w, `{"response":"Success"}`)

		} else {

			err = basicDataDb(user)

			if err != nil {
				log.Println("Failed :/")
				fmt.Fprintf(w, `{"response":"Failed :("}`)
			} else {
				log.Println("Sign Up Success")
				fmt.Fprintf(w, `{"response":"Success"}`)
			}
		}
	}

	log.Println(r.URL.Path)
}
func addProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var proj userProjects

		proj.UID = r.FormValue("uid")
		proj.ProjectName = r.FormValue("pname")
		proj.ProjectDesc = r.FormValue("pdesc")

		err := checkIfProjectExists(proj.ProjectName)
		if err != nil {
			err := projectDataDb(proj)

			if err != nil {
				fmt.Fprintf(w, `{"response":"Failed :("}`)
			} else {
				fmt.Fprintf(w, `{"response":"Success"}`)
			}
		} else {
			fmt.Fprintf(w, `{"response":"Project Name taken"}`)
		}
	}

	log.Println(r.URL.Path)
}
func joinProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var join joinedProjects

		join.UID = r.FormValue("uid")
		join.DUID = r.FormValue("duid")
		join.ProjectName = r.FormValue("pname")
		join.ProjectDesc = r.FormValue("pdesc")

		err := checkIfAlreadyJoined(join.DUID, join.ProjectName)

		if err != nil {
			err := joinProjectDataDb(join)

			if err != nil {
				fmt.Fprintf(w, `{"response":"Failed :("}`)
			} else {
				fmt.Fprintf(w, `{"response":"Success"}`)
			}
		} else {
			fmt.Fprintf(w, `{"response":"Project Already Joined"}`)
		}
	}

	log.Println(r.URL.Path)
}
func getProjectList(w http.ResponseWriter, r *http.Request) {
	var getProjects []userProjects
	if r.Method == http.MethodPost {
		uid := r.FormValue("uid")

		session, err := mgo.Dial(connectionString)

		if err != nil {
			log.Println(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("Conduit").C("UserProjects")

		c.Find(bson.M{"uid": uid}).All(&getProjects)

		respond, _ := json.MarshalIndent(getProjects, "", " ")

		fmt.Fprintf(w, `{
			"response": %s}`, string(respond))
	}

	log.Println(r.URL.Path)
}
func yourJoinedProjects(w http.ResponseWriter, r *http.Request) {
	var getProjects []userProjects
	if r.Method == http.MethodPost {
		uid := r.FormValue("uid")

		session, err := mgo.Dial(connectionString)

		if err != nil {
			log.Println(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)

		c := session.DB("Conduit").C("JoinedProjects")

		c.Find(bson.M{"uid": uid}).All(&getProjects)

		respond, _ := json.MarshalIndent(getProjects, "", " ")

		fmt.Fprintf(w, `{
			"response": %s}`, string(respond))
	}

	log.Println(r.URL.Path)
}

// Helper function(s) below
func basicDataDb(udata userBasicData) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("Conduit").C("UserBasicData")
	err = c.Insert(udata)

	return err
}
func projectDataDb(pdata userProjects) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("Conduit").C("UserProjects")
	err = c.Insert(pdata)

	return err
}
func joinProjectDataDb(jdata joinedProjects) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("Conduit").C("JoinedProjects")
	err = c.Insert(jdata)

	return err
}
func checkIfRegistered(uid string) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	result := userBasicData{}

	c := session.DB("Conduit").C("UserBasicData")
	err = c.Find(bson.M{"uid": uid}).One(&result)

	return err
}
func checkIfProjectExists(pname string) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	result := userBasicData{}

	c := session.DB("Conduit").C("UserProjects")
	err = c.Find(bson.M{"projectname": pname}).One(&result)

	return err
}
func checkIfAlreadyJoined(duid, pname string) error {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	result := userBasicData{}

	c := session.DB("Conduit").C("JoinedProjects")
	err = c.Find(bson.M{"projectname": pname, "duid": duid}).One(&result)

	return err
}
func getUserProfile(uid string) ([]byte, error) {
	session, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	result := userBasicData{}

	c := session.DB("Conduit").C("UserBasicData")
	err = c.Find(bson.M{"uid": uid}).One(&result)

	if err != nil {
		return nil, err
	} else {
		response, _ := json.MarshalIndent(result, "", " ")
		return response, err
	}
}
func checkAndUpdate(udata userBasicData) string {
	session, err := mgo.Dial(connectionString)
	if err != nil {
		log.Println(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	result := userBasicData{}

	c := session.DB("Conduit").C("UserBasicData")
	err = c.Find(bson.M{"uid": udata.UID, "name": udata.Name, "sex": udata.Sex, "email": udata.Email, "profilepicture": udata.ProfilePicture}).One(&result)
	if err != nil {
		colQuerier := bson.M{"uid": udata.UID}
		err = c.Update(colQuerier, udata)

		return "Profile was updated"
	} else {
		return "No updates"
	}
}
