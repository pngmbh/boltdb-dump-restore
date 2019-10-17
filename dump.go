package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
)

func main() {
	mode := os.Args[0]
	if strings.Contains(mode, "dump") {
		dump()
	} else if strings.Contains(mode, "restore") {
		restore()
	} else {
		log.Fatal(fmt.Errorf("unknown mode %s", mode))
	}
}

func dump() {
	db, err := bolt.Open(os.Args[1], 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	out := make(map[string]interface{})
	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			out[string(name)] = map[string]string{}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				out[string(name)].(map[string]string)[string(k)] = string(v)
			}
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	j, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))
}

func restore() {
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(file), &data)
	// write into DB
	db, err := bolt.Open("OUT.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		for k, bucketContents := range data {
			b, err := tx.CreateBucket([]byte(k))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			for key, value := range bucketContents.(map[string]interface{}) {
				err = b.Put([]byte(key), []byte(value.(string)))
				if err != nil {
					return fmt.Errorf("set value: %s", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
