package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

// Return 0 if not HTTP
// Return 1 if HTTP GET
// Return 2 if HTTP NOT GET
func Is_Http_Content(buf []byte, n int) int {
	if n > 16 {
		n = 16
	}
	str := string(buf[:n])
	if strings.Contains(str, "HTTP") {
		if strings.Contains(str, "GET") {
			return 1
		} else {
			return 2
		}
	} else {
		return 0
	}
}

// Parse HTTP GET
func Http_Get_Parse(buf []byte, n int) {
	fmt.Println("DEBUG ||", string(buf[:n]))
	resp, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(buf)))
	if err != nil {
		fmt.Println("DEBUG || What the fuck is that:", err)
		return
	} else {
		fmt.Println("DEBUG || Pass the parse!")
	}
	cookie := resp.Cookies()
	lens := len(cookie)
	for i := 0; i < lens; i++ {
		fmt.Printf("DEBUG || %v %v \n", cookie[i].Name, cookie[i].Value)
	}
	return
}