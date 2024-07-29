package wsdiscovery

/*******************************************************
 * Copyright (C) 2018 Palanjyan Zhorzhik
 *
 * This file is part of ws-discovery project.
 *
 * ws-discovery can be copied and/or distributed without the express
 * permission of Palanjyan Zhorzhik
 *******************************************************/

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/net/ipv4"
)

const bufSize = 8192

//SendProbe to device
func SendProbe(interfaceName string, scopes, types []string, namespaces map[string]string) ([]string) {
	// Creating UUID Version 4
	uuidV4 := uuid.Must(uuid.NewV4())
	//fmt.Printf("UUIDv4: %s\n", uuidV4)

	probeSOAP := buildProbeMessage(uuidV4.String(), scopes, types, namespaces)
	//probeSOAP = `<?xml version="1.0" encoding="UTF-8"?>
	//<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
	//<Header>
	//<a:Action mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
	//<a:MessageID>uuid:78a2ed98-bc1f-4b08-9668-094fcba81e35</a:MessageID><a:ReplyTo>
	//<a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
	//</a:ReplyTo><a:To mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
	//</Header>
	//<Body><Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
	//<d:Types xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">dp0:NetworkVideoTransmitter</d:Types>
	//</Probe>
	//</Body>
	//</Envelope>`

	return sendUDPMulticastEx(probeSOAP.String(), interfaceName)
}

func sendUDPMulticastEx(msg string, interfaceName string) ([]string) {
	if interfaceName == "" {
		ift, err := net.Interfaces()
		if err != nil {
			return nil
		}

		var result []string
		if len(ift) == 0 {
			fmt.Println("--+--> 检索到的网卡个数为0！不绑定网卡查询！")
			temp := sendUDPMulticastBase(msg, nil)
			for _,item := range temp {
				result = append(result, item)
			}
		}else {
			for _, ifi := range ift {
				fmt.Println("--+--> 查询网卡：", ifi.Name)
				temp := sendUDPMulticastBase(msg, &ifi)
				for _,item := range temp {
					result = append(result, item)
				}
			}
		}

		return result
	} else {
		iface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			fmt.Println(err)
		}
		return sendUDPMulticastBase(msg, iface)
	}
}

func sendUDPMulticast(msg string, interfaceName string) []string {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		fmt.Println(err)
	}

	return sendUDPMulticastBase(msg, iface)
}

func sendUDPMulticastBase(msg string, iface *net.Interface) []string {
	var result []string
	c, err := net.ListenPacket("udp4", "0.0.0.0:36909")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)
	group := net.IPv4(239, 255, 255, 250)
	if err := p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
		fmt.Println(err)
	}

	dst := &net.UDPAddr{IP: group, Port: 3702}
	data := []byte(msg)
	if err := p.SetMulticastInterface(iface); err != nil {
		fmt.Println(err)
		return result
	}

	p.SetMulticastTTL(2)
	if _, err := p.WriteTo(data, nil, dst); err != nil {
		fmt.Println(err)
		return result
	}

	if err := p.SetReadDeadline(time.Now().Add(time.Second * 1)); err != nil {
		fmt.Println(err)
		return result
	}

	for {
		b := make([]byte, bufSize)
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println(err)
			}
			break
		}
		result = append(result, string(b[0:n]))
	}
	return result
}
