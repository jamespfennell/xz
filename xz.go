package main

import "fmt"

/*
#include "example.c"
 */
import "C"

func main() {
	fmt.Println("Working")
	fmt.Println(C.add(C.int(3), C.int(5)))
}
