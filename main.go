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
)

type Color string

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

type BGColor string

const (
	BGBlack         = "\033[40m"
	BGRed           = "\033[41m"
	BGGreen         = "\033[42m"
	BGYellow        = "\033[43m"
	BGBlue          = "\033[44m"
	BGMagenta       = "\033[45m"
	BGCyan          = "\033[46m"
	BGWhite         = "\033[47m"
	BGBrightBlack   = "\033[40;1m"
	BGBrightRed     = "\033[41;1m"
	BGBrightGreen   = "\033[42;1m"
	BGBrightYellow  = "\033[43;1m"
	BGBrightBlue    = "\033[44;1m"
	BGBrightMagenta = "\033[45;1m"
	BGBrightCyan    = "\033[46;1m"
	BGBrightWhite   = "\033[47;1m"
)

type Decoration string

const (
	Underlined = "\033[4m"
	Reversed   = "\033[7m"
)

func clearTerminal() {
	fmt.Print("\033[2J")
}
func setCursor(x int, y int) {
	fmt.Printf("\033[%d;%dH", x, y)
}
func setColor(c Color) {
	fmt.Print(c)
}
func setBGColor(c BGColor) {
	fmt.Print(c)
}
func setDecoration(d Decoration) {
	fmt.Print(d)
}
func resetTerminal() {
	fmt.Print("\033[0m")
}

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
	clearTerminal()

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
		setCursor(0, 0)
		setColor(White)
		fmt.Printf("Uptime\t: ")
		setColor(Yellow)
		fmt.Printf("%s seconds\n", uptime)
		setColor(White)
		fmt.Print("samples\t: ")
		setColor(Yellow)
		fmt.Printf("%.f\n", samples)
		setColor(White)
		fmt.Print("\nDownstream SNR\t: ")
		setColor(Green)
		fmt.Printf("%.1f\tdB ", d)
		setColor(White)
		fmt.Print("| average: ")
		setColor(Green)
		fmt.Printf("%.1f\tdB", dAvg)
		setColor(White)
		fmt.Print("\nUpstream SNR\t: ")
		setColor(Green)
		fmt.Printf("%.1f\tdB ", u)
		setColor(White)
		fmt.Print("| average: ")
		setColor(Green)
		fmt.Printf("%.1f\tdB\n", uAvg)
		setColor(White)
		fmt.Printf("\nDownstram CRC\t: ")
		setColor(Red)
		fmt.Printf("%s", downstreamCRC)
		setColor(White)
		fmt.Print("\nUpstream CRC\t: ")
		setColor(Red)
		fmt.Printf("%s\n", upstreamCRC)
	}
}
