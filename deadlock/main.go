// The smallest deadlock: a send on a channel that nobody will ever receive from.
package main

import "fmt"

func main() {
	results := make(chan string) // unbuffered: a send waits for a receiver
	results <- "page 0"          // nobody is receiving, so this waits forever
	fmt.Println(<-results)
}
