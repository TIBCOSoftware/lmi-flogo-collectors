/*
 * Copyright © 2018. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

 package SendSyslog

import (
	"io/ioutil"
	"testing"

	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var activityMetadata *activity.Metadata

func getActivityMetadata() *activity.Metadata {

	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}

		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}

	return activityMetadata
}

func TestCreate(t *testing.T) {

	act := NewActivity(getActivityMetadata())

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
}

func createCertificate() (certPEMBlock []byte, keyPEMBlock []byte, err error) {

	err = nil

	rsaKeyLength := 2048

	// create Key
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyLength)

	if err != nil {
		log.Errorf("Error while generating RSA key: %s", err)
		return
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	host := "localhost,127.0.0.1,::1"
	isCA := true
	notBefore := time.Now()
	duration := 10 * 365 * 24 * time.Hour // 10 years
	notAfter := notBefore.Add(duration)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Errorf("failed to generate serial number: %s", err)
		return
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Testing Co."},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		log.Errorf("Failed to create certificate: %s", err)
		return nil, nil, err
	}

	certBlock := pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	certPEMBlock = pem.EncodeToMemory(&certBlock)

	if false {
		certOut, err := os.Create("cert.pem")
		if err != nil {
			log.Errorf("failed to open cert.pem for writing: %s", err)
			return nil, nil, err
		}

		pem.Encode(certOut, &certBlock)
		certOut.Close()
		log.Info("written cert.pem")
	}

	keyBlock := pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}
	keyPEMBlock = pem.EncodeToMemory(&keyBlock)

	if false {
		keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Infof("failed to open key.pem for writing:", err)
			return nil, nil, err
		}
		pem.Encode(keyOut, &keyBlock)
		keyOut.Close()
		log.Infof("written key.pem")
	}

	return
}

var messages = make([]string, 0)

func tcpConnHandler(conn net.Conn) {
	defer conn.Close()
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Errorf("Error while reading from TLS connection: %s", err)
			break
		}
		if len(message) == 0 {
			log.Infof("Received empty data!")
			continue
		}
		log.Infof("Receive: %s", message)
		messages = append(messages, message)
	}
	log.Info("server: conn: closing")
}

func tlsConnHandler(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		// reading length, up until reaching begining of Syslog message (RFC5424)
		message, err := reader.ReadString('<')
		if err != nil {
			log.Errorf("Error while reading from TLS connection: %s", err)
			break
		}
		if len(message) == 0 {
			log.Infof("Received empty data!")
			continue
		}
		// we are supposed to have received a length
		lenStr := message[:len(message)-1]
		len, err := strconv.Atoi(lenStr)
		if err != nil {
			logger.Errorf("Cannot understand length of message: %s from %s", lenStr, message)
			break
		}
		logger.Infof("Message length is %d", len)
		reader.UnreadByte()
		if reader.Buffered() < len {
			logger.Warn("Not enough data buffered ! %d>%d", len, reader.Buffered())
		}
		line := make([]byte, len)
		n, err := reader.Read(line)
		if err != nil {
			log.Errorf("Error while reading from TLS connection: %s", err)
			break
		} else if n != len {
			log.Errorf("Cannot fully read the message !", message)
		}
		log.Infof("Receive: %s", line)
		messages = append(messages, message)
	}
	log.Info("server: conn: closing")
}

func tlsServer(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("TLS server: accept error: %s", err)
			break
		}
		log.Infof("TLS server: accepted from %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			state := tlscon.ConnectionState()
			log.Infof("Connection state: %s", state.ServerName)
			for _, v := range state.PeerCertificates {
				log.Infof("%s", v)
			}
		}

		go tlsConnHandler(conn)
	}
}

func TestEval(t *testing.T) {

	//defer func() {
	//	if r := recover(); r != nil {
	//		t.Failed()
	//		t.Errorf("panic during execution: %T:%v", r, r)
	//		}
	//	}()

	udpAddress, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0")

	if err != nil {
		t.Failed()
		t.Errorf("Error resolving UDP address: %v", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddress)

	if err != nil {
		t.Failed()
		t.Errorf("Error listening to UDP: %v\n", err)
		return
	}

	tcpAddress, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:0")

	if err != nil {
		t.Failed()
		t.Errorf("Error resolving TCP address: %v", err)
		return
	}

	tcpConn, err := net.ListenTCP("tcp", tcpAddress)

	if err != nil {
		t.Failed()
		t.Errorf("Error listening to TCP: %v\n", err)
		return
	}

	fmt.Printf("Listening on UDP %s\n", udpConn.LocalAddr().String())
	fmt.Printf("Listening on TCP %s\n", tcpConn.Addr().String())

	localAddr := strings.Split(udpConn.LocalAddr().String(), ":")
	udpPort, _ := strconv.Atoi(localAddr[len(localAddr)-1])
	fmt.Printf("UDP Port: %d \n", udpPort)

	localAddr = strings.Split(tcpConn.Addr().String(), ":")
	tcpPort, _ := strconv.Atoi(localAddr[len(localAddr)-1])
	fmt.Printf("TCP Port: %d \n", tcpPort)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	//setup attrs
	tc.SetInput("protocol", "UDP")
	tc.SetInput("host", "localhost")
	tc.SetInput("port", udpPort)
	tc.SetInput("message", "My Message")
	tc.SetInput("facility", 1)
	tc.SetInput("severity", 1)
	tc.SetInput("flowInfo", false)
	act.Eval(tc)

	var buf []byte = make([]byte, 65536)

	n, address, err := udpConn.ReadFromUDP(buf)

	s := string(buf[:n])

	fmt.Printf("Received message from %s, len %d: %s\n", address.String(), n, s)

	tc.SetInput("port", tcpPort)
	tc.SetInput("protocol", "tcp")
	act.Eval(tc)

	clientCon, _ := tcpConn.Accept()
	go tcpConnHandler(clientCon)

	fmt.Printf("Accepted TCP: %v\n", clientCon.RemoteAddr())

	for len(messages) == 0 {
		time.Sleep(time.Second)
	}

	message := messages[0]

	messages = messages[1:]

	log.Infof("TCP Received %s", message)

	if len(messages) != 0 {
		logger.Errorf("Unexpected Messages: %d!", len(messages))
		t.Fail()
		return
	}

	certPEMBlock, keyPEMBlock, err := createCertificate()

	if err != nil {
		t.Failed()
		log.Errorf("Cannot create certificate: %v", err)
		return
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)

	if err != nil {
		t.Failed()
		logger.Errorf("Error while creating X509 key pair : %s", err)
		return
	}

	config := tls.Config{}
	config.Certificates = []tls.Certificate{cert}

	tlsListener, err := tls.Listen("tcp", "0.0.0.0:0", &config)

	localAddr = strings.Split(tlsListener.Addr().String(), ":")
	tlsPort, _ := strconv.Atoi(localAddr[len(localAddr)-1])
	fmt.Printf("TLS Port: %d \n", tlsPort)

	tc.SetInput("port", tlsPort)
	tc.SetInput("protocol", "tls")

	go tlsServer(tlsListener)

	act.Eval(tc)

	for len(messages) == 0 {
		time.Sleep(time.Second)
	}

	message = messages[0]

	messages = messages[1:]

	fmt.Printf("TLS Received %s", message)
}
