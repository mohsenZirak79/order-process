package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

type Order struct {
	Id    string `json:"Id"`
	Title string `json:"Title"`
	Desc  string `json:"desc"`
	Price string `json:"Price"`
}

var orders []Order
var ctx = context.Background()

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HomePage!")
	fmt.Println("homePage")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/all", returnAllOrders)
	myRouter.HandleFunc("/order/{id}", returnSingleOrder)
	myRouter.HandleFunc("/order", createNewOrder).Methods("POST")
	log.Fatal(http.ListenAndServe(":6379", myRouter))
}

func returnSingleOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	for _, order := range orders {
		if order.Id == key {
			json.NewEncoder(w).Encode(order)
		}
	}
}

func returnAllOrders(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllOrders")
	json.NewEncoder(w).Encode(orders)
}

func createNewOrder(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var order Order
	json.Unmarshal(reqBody, &order)
	orders = append(orders, order)
	//json.NewEncoder(w).Encode(order)

	addToRedis(order)
	_, err := w.Write([]byte("added to redis"))
	if err != nil {
		return
	}
}

func addToRedis(order Order) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	jsonOrder, err := json.Marshal(order)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsonOrder))
	if err := client.Publish(ctx, "send-user-data", jsonOrder).Err(); err != nil {
		fmt.Println("error in publish in to redis", err)
	}
}

func getFromRedis(order Order) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	subscriber := client.Subscribe(ctx, "send-user-data")

	user := Order{}

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
			panic(err)
		}

		fmt.Println("Received message from " + msg.Channel + " channel.")
		fmt.Printf("%+v\n", user)
	}

}

func main() {
	handleRequests()
}
