package main

import (
	"database/sql"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const connStr = "host=89.108.117.52 port=5432 user=postgres password=991155 dbname=wilberries sslmode=disable"

type Id struct {
	id         int
	category   string
	imagelinks []string
}
type get struct {
	imagelinks []string
}

func WriteIdToPostgreSql(id int, images []string, category string) {
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	if err != nil {
		panic(err)
	}
	_, e := db.Exec("update items set imagelinks=$2, category=$3 where id=$1",
		id, pq.Array(images), category)
	fmt.Println("YES")
	if e != nil {
		fmt.Println("Error write")
		fmt.Println(e)
	}
}

func GetDbIds() []Id {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	res, _ := db.Query("Select id, category, imagelinks from items")
	var ids []Id
	for res.Next() {
		id := Id{}
		res.Scan(&id.id, &id.category, pq.Array(&id.imagelinks))
		if len(id.imagelinks) == 0 {
			ids = append(ids, id)
		}

	}
	return ids
}

func scrapImage(id string, category string) int {
	count := 0
	var images []string
	for {
		count++
		imageLink := ""
		if len(id) == 8 {
			imageLink = "https://images.wbstatic.net/c516x688/new/" + id[0:4] + "0000/" + id + "-" + strconv.Itoa(count) + ".jpg"
		} else if len(id) == 7 {
			imageLink = "https://images.wbstatic.net/c516x688/new/" + id[0:3] + "0000/" + id + "-" + strconv.Itoa(count) + ".jpg"
		}

		resp, e := http.Get(imageLink)
		if e != nil {
			strId, _ := strconv.Atoi(id)
			WriteIdToPostgreSql(strId, images, category) // заменить запись
			return 1
		}
		if resp.StatusCode == 200 {
			images = append(images, imageLink)
		} else {
			strId, _ := strconv.Atoi(id)
			WriteIdToPostgreSql(strId, images, category) // заменить запись
			return 1
		}
	}
}

func scrapImages() {
	var wg sync.WaitGroup
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://f20597c3014e4699969af0244a66a6f8@o1108001.ingest.sentry.io/6135375",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)
	fmt.Println("[3/4] Скрипт парсера картинок запущен!")

	count := 0
	data := GetDbIds()
	for i, v := range data {
		count += 1
		wg.Add(1)
		go func(id int, category string) {
			defer wg.Done()
			scrapImage(strconv.Itoa(id), category)
		}(v.id, v.category)
		if i%50 == 0 {
			wg.Wait()
			if i%50 == 0 {
				fmt.Println("[3/4] Обработано " + strconv.Itoa(count) + " из " + strconv.Itoa(len(data)))
			}

		}
	}
	wg.Wait()
}

func main() {
	for {
		scrapImages()
		fmt.Println("FINISH [images]")
	}
}
