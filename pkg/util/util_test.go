package util

import (
	"fmt"
	"log"
	"testing"
)

func TestSnowflake(t *testing.T) {
	sf, err := NewSnowflake()
	id := sf.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
	id = sf.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
	id = sf.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
	id = sf.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
	id = sf.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)

}
