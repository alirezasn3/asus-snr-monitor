package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	prettyterm "github.com/alirezasn3/pretty-term"
)

func getAsusToken(u string, p string) string {
	req, _ := http.NewRequest("POST", "http://192.168.1.1/login.cgi", strings.NewReader(("login_authorization=" + base64.StdEncoding.EncodeToString([]byte(u+":"+p)))))
	req.Header.Set("Referer", "http://192.168.1.1/Main_Login.asp")
	client := &http.Client{}
	resp, _ := client.Do(req)
	return strings.Split(resp.Header.Get("Set-Cookie"), ";")[0]
}

func main() {

	// get username and password
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// create ticker for polling
	ticker := time.Tick(time.Second)

	// create http client
	client := &http.Client{}

	// clear terminal
	prettyterm.ClearTerminal()

	// construct http req
	adslReq, _ := http.NewRequest("GET", "http://192.168.1.1/cgi-bin/ajax_AdslStatus.asp", nil)
	adslReq.Header.Set("Cookie", getAsusToken(username, password))

	// regexes
	floatReg, _ := regexp.Compile(`[1-9]+\.[1-9]`)
	intReg, _ := regexp.Compile(`[1-9]+`)

	// init vars
	var downstreamSNR, upstreamSNR, downstreamCRC, upstreamCRC, uptime []byte
	var d, u, dAvg, uAvg, samples float64 = 0, 0, 0, 0, 0

	// start polling
	for range ticker {
		// send http req and read response body
		res, _ := client.Do(adslReq)
		data, _ := io.ReadAll(res.Body)
		res.Body.Close()

		// find snr values
		downstreamSNR = data[bytes.Index(data, []byte("SNRMarginDown")):]
		downstreamSNR = downstreamSNR[:bytes.Index(downstreamSNR, []byte(";"))]
		downstreamSNR = floatReg.Find(downstreamSNR)
		upstreamSNR = data[bytes.Index(data, []byte("SNRMarginUp")):]
		upstreamSNR = upstreamSNR[:bytes.Index(upstreamSNR, []byte(";"))]
		upstreamSNR = floatReg.Find(upstreamSNR)
		d, _ = strconv.ParseFloat(string(downstreamSNR), 64)
		u, _ = strconv.ParseFloat(string(upstreamSNR), 64)

		// find crc counts
		downstreamCRC = data[bytes.Index(data, []byte("CRCDown")):]
		downstreamCRC = downstreamCRC[:bytes.Index(downstreamCRC, []byte(";"))]
		downstreamCRC = intReg.Find(downstreamCRC)
		upstreamCRC = data[bytes.Index(data, []byte("CRCUp")):]
		upstreamCRC = upstreamCRC[:bytes.Index(upstreamCRC, []byte(";"))]
		upstreamCRC = intReg.Find(upstreamCRC)

		// find uptime
		uptime = data[bytes.Index(data, []byte("uptimeStr")):]
		uptime = uptime[:bytes.Index(uptime, []byte(";"))]
		uptime = uptime[bytes.Index(uptime, []byte("(")):]
		uptime = uptime[:bytes.Index(uptime, []byte(")"))]
		uptime = intReg.Find(uptime)

		// calculate average snr values
		if dAvg != 0 {
			dAvg = (dAvg*samples + d) / (samples + 1)
			uAvg = (uAvg*samples + u) / (samples + 1)
		} else {
			dAvg = d
			uAvg = u
		}

		// increment sample count
		samples++

		// print results
		prettyterm.SetCursor(0, 0)
		prettyterm.SetColor(prettyterm.White)
		fmt.Printf("Uptime\t: ")
		prettyterm.SetColor(prettyterm.Yellow)
		fmt.Printf("%s seconds\n", uptime)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("samples\t: ")
		prettyterm.SetColor(prettyterm.Yellow)
		fmt.Printf("%.f\n", samples)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("\nDownstream SNR\t: ")
		prettyterm.SetColor(prettyterm.Green)
		fmt.Printf("%.1f\tdB ", d)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("| average: ")
		prettyterm.SetColor(prettyterm.Green)
		fmt.Printf("%.1f\tdB", dAvg)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("\nUpstream SNR\t: ")
		prettyterm.SetColor(prettyterm.Green)
		fmt.Printf("%.1f\tdB ", u)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("| average: ")
		prettyterm.SetColor(prettyterm.Green)
		fmt.Printf("%.1f\tdB\n", uAvg)
		prettyterm.SetColor(prettyterm.White)
		fmt.Printf("\nDownstram CRC\t: ")
		prettyterm.SetColor(prettyterm.Red)
		fmt.Printf("%s", downstreamCRC)
		prettyterm.SetColor(prettyterm.White)
		fmt.Print("\nUpstream CRC\t: ")
		prettyterm.SetColor(prettyterm.Red)
		fmt.Printf("%s\n", upstreamCRC)
	}
}
