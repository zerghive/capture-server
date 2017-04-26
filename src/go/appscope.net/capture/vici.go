package capture

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bronze1man/goStrongswanVici"
	"github.com/golang/glog"
)

type VpnClient struct {
	Ip4      string `json:"IP"`
	Ip6      string `json:"-"`
	Name     string `json:"-"`
	DeviceId string
	OrgId    string
}

func getDeviceIP(name string) (string, error) {

	conn, err := goStrongswanVici.NewClientConnFromDefaultSocket()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	clients, err := getVpnClients(conn, name, nil)
	if err != nil {
		return "", err
	}

	if len(clients) > 0 {
		glog.Infof("Client %s found: %s", name, clients[0].Ip4)
		return clients[0].Ip4, nil
	} else {
		return "", fmt.Errorf("Client %s not found.", name)
	}
}

func GetOrganizationVpnClients(orgId string) ([]VpnClient, error) {

	conn, err := goStrongswanVici.NewClientConnFromDefaultSocket()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	orgPostfix := fmt.Sprintf("@%s", orgId)

	handler := func(client *VpnClient) {
		client.OrgId = orgId
		client.DeviceId = strings.Replace(client.Name, orgPostfix, "", 1)
	}

	clients, err := getVpnClients(conn, fmt.Sprintf("[\\d]+%s", orgPostfix), handler)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func getVpnClients(c *goStrongswanVici.ClientConn, expr string, handler func(client *VpnClient)) (
	clients []VpnClient, err error) {

	var eventErr error
	clients = make([]VpnClient, 0)

	err = c.RegisterEvent("list-sa", func(response map[string]interface{}) {

		nameReg, err := regexp.Compile(expr)
		if err != nil {
			eventErr = err
			return
		}

		// TODO: Add IPv6 regexp too.
		ipReg, err := regexp.Compile("([\\d]{1,3}\\.){3}[\\d]{1,3}")
		if err != nil {
			eventErr = err
			return
		}

		for _, v1 := range response {

			m2, ok := v1.(map[string]interface{})
			if !ok {
				continue
			}

			remoteId, ok := m2["remote-id"].(string)
			if !ok {
				continue
			}

			if name := nameReg.FindString(remoteId); len(name) > 0 {

				m3, ok := m2["child-sas"].(map[string]interface{})
				if !ok {
					continue
				}

				for _, v3 := range m3 {

					m4, ok := v3.(map[string]interface{})
					if !ok {
						continue
					}

					ips, ok := m4["remote-ts"].([]string)
					if !ok {
						continue
					}

					if len(ips) > 0 {
						// TODO: Add IPv6.
						client := VpnClient{
							Ip4:  ipReg.FindString(ips[0]),
							Name: strings.Replace(name, "CN=", "", 1),
						}
						if handler != nil {
							handler(&client)
						}
						clients = append(clients, client)
						break
					}
				}
			}
		}
	})

	if err != nil {
		return
	}

	_, err = c.Request("list-sas", map[string]interface{}{})
	if err != nil {
		return
	}

	if eventErr != nil {
		err = eventErr
		return
	}

	err = c.UnregisterEvent("list-sa")
	if err != nil {
		return
	}

	return
}
