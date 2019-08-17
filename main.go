package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"./models"
	"github.com/astaxie/beego/orm"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var ORM orm.Ormer
var CacheChan chan models.Users

func init() {
	// init DB
	models.ConnectToDb()
	ORM = models.GetOrmObject()
}

func main() {
	router := gin.Default()
	router.POST("/createUser", createUser)
	router.GET("/readUsers", readUsers)
	router.GET("/readUser", readUser)
	router.GET("/readCacheUser", readCacheUser)
	// router.GET("/redisCreateUser", redisCreateUser(createUser))
	// router.PUT("/updateUser", updateUser)
	// router.DELETE("/deleteUser", deleteUser)
	router.Run(":3000")

}

func createUser(c *gin.Context) {
	var newUser models.Users
	c.BindJSON(&newUser)
	_, err := ORM.Insert(&newUser)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    http.StatusOK,
			"email":     newUser.Email,
			"user_name": newUser.UserName,
			"user_id":   newUser.UserId,
		})
		redisCacheUser(newUser, newUser.UserId)
	} else {
		c.JSON(http.StatusInternalServerError,
			gin.H{"status": http.StatusInternalServerError, "error": "Failed to create the user"})
	}
}

func redisCacheUser(value interface{}, key interface{}) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	newKey := fmt.Sprintf("%v", key)
	var newValue interface{}
	if fmt.Sprintf("%T", value) == "models.Users" {
		newValue, _ = json.Marshal(value)
	} else {
		newValue = fmt.Sprintf("%v", value)
	}

	fmt.Println(newKey)
	fmt.Println(newValue)

	err = client.Set(newKey, newValue, 0).Err()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Stored in redis")
	}

	val, err := client.Get(newKey).Result()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Reading from Redis...")
		fmt.Println(val)
	}
}

func readCacheUser(c *gin.Context) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	var user models.Users
	c.BindJSON(&user)
	query := strconv.Itoa(user.UserId)
	fmt.Println(query)

	var parsedResult interface{}
	result, err := client.Get(query).Result()

	if json.Unmarshal([]byte(result), &parsedResult) == nil {
		json.Unmarshal([]byte(result), &parsedResult)
	} else {
		parsedResult = result
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": err})
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("%T", parsedResult))
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "user": parsedResult})
	}

}

func readUsers(c *gin.Context) {
	var user []models.Users
	_, err := ORM.QueryTable("users").All(&user)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "users": &user})
	} else {
		c.JSON(http.StatusInternalServerError,
			gin.H{"status": http.StatusInternalServerError, "error": "Failed to read the users"})
	}
}

func readUser(c *gin.Context) {
	var user models.Users
	c.BindJSON(&user)
	err := ORM.Read(&user)

	if err == orm.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "No Result Found"})
		fmt.Println("No result found")
	} else if err == orm.ErrMissPK {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "No Primary Key Found"})
		fmt.Println("No primary key found.")
	} else {
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "user": &user})
		fmt.Println(user.UserId, user.UserName)
	}
}
