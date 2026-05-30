package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/krishnassh/picostatus/internal/checker"
)

func main() {
	fmt.Print("url: ")

	url, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	url = strings.TrimSpace(url)

	res := checker.HTTPChecker{URL: url}.Run(context.Background())

	fmt.Println(res)
}
