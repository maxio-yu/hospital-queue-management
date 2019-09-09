package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

//TODO: 只允许一个客户端进行连接
func main() {
	os.MkdirAll("./database", os.ModeDir|os.ModePerm)
	r := gin.Default()
	backend := NewMaster()

	r.StaticFile("/", "./web/panel/index.html")
	r.StaticFile("/controller", "./web/controller/index.html")
	r.Static("/js", "./web/js")
	r.Static("/css", "./web/css")
	r.Static("/ads/img", "./web/ads/img")

	r.GET("/patient_list", backend.GetPatientList)
	r.POST("/patient_list", backend.PostPatientList)
	r.DELETE("/patient_list", backend.DeletePatientList)

	r.DELETE("/patient_list/:id", backend.DeletePatient)
	r.PUT("/patient_list/:id/actions/call", backend.CallPatient)

	r.GET("/call_list", backend.GetCallList)

	r.GET("/ads_img", backend.GetAdvertisementsImages)
	r.Run()
}

type Backend interface {
	GetPatientList(c *gin.Context)
	PostPatientList(c *gin.Context)
	DeletePatientList(c *gin.Context)
	DeletePatient(c *gin.Context)
	CallPatient(c *gin.Context)
	GetCallList(c *gin.Context)
	GetAdvertisementsImages(c *gin.Context)
}

type Master struct {
	mutex    sync.Mutex
	db       *xorm.Engine
	callList []WaitingPatient
}

func NewMaster() Backend {
	m := &Master{}
	db, err := xorm.NewEngine("sqlite3", "./database/db.sqlite")
	if err != nil {
		fmt.Println("errrorororo: ", err)
		os.Exit(1)
	}
	m.db = db

	exist, err := db.IsTableExist(WaitingPatient{})
	if err != nil {
		fmt.Println("Errrrrr check tabel: ", err)
	}
	if !exist {
		fmt.Println("table not exist")
		err := db.CreateTables(WaitingPatient{})
		if err != nil {
			fmt.Println("Errrrrr create tabel: ", err)
		}
	}
	return m
}

func (m *Master) GetPatientList(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	patients := []WaitingPatient{}
	err := m.db.Asc("id").Find(&patients)
	if err != nil {
		fmt.Println("find patiends error: ", err)
	}
	c.JSON(200, patients)
}

func (m *Master) PostPatientList(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	newPatient := &WaitingPatient{}
	err := c.ShouldBind(newPatient)
	if err != nil {
		fmt.Println("errr binding: ", err)
	}
	fmt.Println("get new patient:", newPatient)
	n, err := m.db.Insert(newPatient)
	if err != nil {
		fmt.Println("insert err: ", err)
	}
	c.JSON(200, n)
}

func (m *Master) DeletePatientList(c *gin.Context) {
	_, err := m.db.Exec("delete from waiting_patient")
	if err != nil {
		fmt.Println("delete err: ", err)
	}
	c.JSON(200, "")
}

func (m *Master) UpdatePatient(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id := c.Param("id")
	newPatient := &WaitingPatient{}
	err := c.ShouldBind(newPatient)
	if err != nil {
		fmt.Println("errr binding: ", err)
	}
	fmt.Println("get new patient:", newPatient)
	_, err = m.db.ID(id).Update(newPatient)
	if err != nil {
		fmt.Println("delete err: ", err)
	}
	c.JSON(200, "")
}

func (m *Master) CallPatient(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id := c.Param("id")
	patient := WaitingPatient{}
	has, err := m.db.ID(id).Get(&patient)
	if err != nil {
		fmt.Println("no this patient: ", err)
		c.JSON(200, "")
		return
	}
	if !has {
		fmt.Println("no this patient")
		c.JSON(200, "")
		return
	}
	m.callList = append(m.callList, patient)
	c.JSON(200, "")
}

func (m *Master) DeletePatient(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id := c.Param("id")
	n, err := m.db.ID(id).Delete(&WaitingPatient{})
	if err != nil {
		fmt.Println("delete err: ", err)
	}
	c.JSON(200, n)
}

func (m *Master) GetCallList(c *gin.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	callList := m.callList
	m.callList = []WaitingPatient{}
	c.JSON(200, callList)
}

func (m *Master) GetAdvertisementsImages(c *gin.Context) {
	files, err := filepath.Glob("web/ads/img/*")
	if err != nil {
		c.JSON(200, "")
		return
	}
	res := []string{}
	for _, file := range files {
		fileName := strings.TrimLeft(file, "web/")
		res = append(res, fileName)
	}
	c.JSON(200, res)
}

type WaitingPatient struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	Uid        string    `json:"uid"`
	ClinicNum  string    `json:"clinic_num"`
	CreateTime time.Time `xorm:"created" json:"create_time"`
	UpdateTime time.Time `xorm:"updated" json:"update_time"`
}
