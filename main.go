package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

type SendText struct {
	Receiver string `json:"to"`
	Message  string `json:"text"`
}

type SendImage struct {
	Receiver string `json:"to"`
	Message  string `json:"text"`
	Image    string `json:"image"`
}

var (
	wac, _       = whatsapp.NewConn(20 * time.Second)
	dir, _       = filepath.Abs(filepath.Dir(os.Args[0]))
	folder       string
	textChannel  chan SendText
	imageChannel chan SendImage
)

func init() {

	textChannel = make(chan SendText)
	imageChannel = make(chan SendImage)
	wac.SetClientVersion(2, 2021, 4)

	err := login(wac)
	if err != nil {
		panic("Error logging in: \n" + err.Error())
	}

	<-time.After(3 * time.Second)
}

func main() {

	var v SendText
	var i SendImage
	var word string
	var num string
	var opc int
	var img string
	var s int
	flag.StringVar(&word, "word", "", "a string")
	flag.StringVar(&num, "numb", "", "a string")
	flag.IntVar(&opc, "opc", 1, "an int")
	flag.StringVar(&img, "img", "", "a string")
	flag.Parse()
	/*fmt.Println("==========Word===========")
	fmt.Println(word)
	fmt.Println("=======Num==============")
	fmt.Println(num)
	fmt.Println("========RES=============")*/
	i.Message = word
	v.Message = word
	i.Receiver = num
	v.Receiver = num
	i.Image = img
	s = opc
	//fmt.Println(i)

	go func() {
		for {
			request, ok := <-textChannel
			if ok {
				log.Println(texting(request))
			}
		}
	}()

	go func() {
		for {
			request, ok := <-imageChannel
			if ok {
				log.Println(image(request))
			}
		}
	}()

	if s == 1 {

		log.Println(texting(v))
	} else if s == 2 {

		log.Println(image(i))
	} else if s == 3 {
		var file string
		fmt.Print("Enter the csv name: ")
		fmt.Scanln(&file)
		log.Println(sendBulk(file + ".csv"))
	} else if s == 4 {
		var file string
		fmt.Print("Enter the csv name: ")
		fmt.Scanln(&file)
		log.Println(sendBulkImg(file + ".csv"))
	}

	//}

}

func sendBulk(file string) string {
	csvFile, err := os.Open(dir + folder + file)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, each := range csvData {
		each[0] = strings.Replace(each[0], " ", "", -1)
		if each[0] != "" {
			v := SendText{
				Receiver: each[0],
				Message:  each[1],
			}
			textChannel <- v
		}
	}

	return "Done"
}

func sendBulkImg(file string) string {
	csvFile, err := os.Open(dir + folder + file)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, each := range csvData {
		if each[0] != "" {
			each[0] = strings.Replace(each[0], " ", "", -1)
			v := SendImage{
				Receiver: each[0],
				Message:  each[1],
				Image:    each[2],
			}
			imageChannel <- v
		}
	}

	return "Done"
}

func texting(v SendText) string {
	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: "549" + v.Receiver + "@s.whatsapp.net",
		},
		Text: v.Message,
	}

	msgId, err := wac.Send(msg)
	if err != nil {
		log.Printf("Error sending message: to %v --> %v\n", v.Receiver, err)
		return "Error"
	}

	return "Message Sent -> " + v.Receiver + " : " + msgId
}

func image(i SendImage) string {
	//var folder string
	img, err := os.Open(dir + "/" + "test.jpg")

	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		return "Error"
	}

	msg := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: "549" + i.Receiver + "@s.whatsapp.net",
		},
		Type:    "image/jpeg",
		Caption: i.Message,
		Content: img,
	}

	msgId, err := wac.Send(msg)
	if err != nil {
		log.Printf("Error sending message: to %v --> %v\n", i.Receiver, err)
		return "Error"
	}

	return "Message Sent -> " + i.Receiver + " : " + msgId
}

func login(wac *whatsapp.Conn) error {
	fmt.Print("Enter your number: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()
	fmt.Println("Logging in -> " + text)
	//load saved session
	session, err := readSession(text)
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			//terminal := qrcodeTerminal.New()
			//terminal.Get(<-qr).Print()
			obj := qrcodeTerminal.New2(qrcodeTerminal.ConsoleColors.BrightBlue, qrcodeTerminal.ConsoleColors.BrightGreen, qrcodeTerminal.QRCodeRecoveryLevels.Low)
			obj.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		} else {

			/* ESTO ES SOLO PARA CORROBORAR LA SESION INICIADA,ES DECIR EL LOGEO,LA VARIABLE QUE CONTIENE ESTE ES session
			fmt.Printf("ESTE ES EL QR !! ")
			fmt.Print(session) */
		}
	}

	//save session
	err = writeSession(session, text)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession(s string) (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(dir + folder + s + ".gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session, s string) error {
	file, err := os.Create(dir + folder + s + ".gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
