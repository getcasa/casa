package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ItsJimi/casa/logger"
	"github.com/getcasa/sdk"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/lib/pq"
)

// WebsocketMessage struct to format received message
type WebsocketMessage struct {
	Action string // newData
	Body   []byte
}

// ActionMessage struct to format sended message
type ActionMessage struct {
	PhysicalID string
	Plugin     string
	Call       string
	Config     string
	Params     string
}

// GatewayAddr define the gateway connected
var GatewayAddr string

// WSConn define the websocket connected between casa server and gateway
var WSConn *websocket.Conn
var queues []Datas

// Configs define plugins configuration
var Configs []sdk.Configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// InitConnection create websocket connection
func InitConnection(con echo.Context) error {
	var err error
	WSConn, err = upgrader.Upgrade(con.Response(), con.Request(), nil) // error ignored for sake of simplicity
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIC001"}).Errorf("%s", err.Error())
		return err
	}
	defer WSConn.Close()

	for {
		var wm WebsocketMessage

		_, message, err := WSConn.ReadMessage()
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDIC002"}).Errorf("%s", err.Error())
			continue
		}
		err = json.Unmarshal(message, &wm)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDIC003"}).Errorf("%s", err.Error())
			continue
		}

		switch wm.Action {
		case "newConnection":
			GatewayAddr = string(wm.Body)
			GetConfigFromGateway(GatewayAddr)
		case "newData":
			go func(data []byte) {
				var datas []Datas
				json.Unmarshal(data, &datas)
				SaveNewDatas(datas)
			}(wm.Body)
		default:
			continue
		}

		logger.WithFields(logger.Fields{}).Debugf("recv: %s", message)
	}
}

// GetConfigFromGateway get config from gateway
func GetConfigFromGateway(addr string) {
	var tmpConfigs []sdk.Configuration

	resp, err := http.Get("http://" + addr + "/v1/configs")
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCFG001"}).Errorf("%s", err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCFG002"}).Errorf("%s", err.Error())
		return
	}

	err = json.Unmarshal(body, &tmpConfigs)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGCFG003"}).Errorf("%s", err.Error())
		return
	}

	if len(Configs) == 0 {
		Configs = tmpConfigs
	} else {
		for _, tmpConf := range tmpConfigs {
			if configFromPlugin(Configs, tmpConf.Name).Name == "" {
				Configs = append(Configs, tmpConf)
			}
		}
	}
}

// GetDiscoveredDevices return an array of futur discover
func GetDiscoveredDevices(c echo.Context) error {
	var discovered []sdk.DiscoveredDevice
	logger.WithFields(logger.Fields{}).Debugf("Discover devices")
	plugin := c.Param("plugin")

	resp, err := http.Get("http://" + GatewayAddr + "/v1/discover/" + plugin)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDDG001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGDD001",
			Message: err.Error(),
		})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDDG002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGDDG002",
			Message: err.Error(),
		})
	}

	err = json.Unmarshal(body, &discovered)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDDG003"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGDDG003",
			Message: err.Error(),
		})
	}

	arrayPhysicalID := []string{}
	for _, disco := range discovered {
		arrayPhysicalID = append(arrayPhysicalID, disco.PhysicalID)
	}

	rows, err := DB.Queryx(`
		SELECT physical_id
		FROM devices
		JOIN gateways ON devices.gateway_id = gateways.id
		WHERE physical_id = ANY($1) AND gateways.home_id = $2
	`, "{"+strings.Join(arrayPhysicalID, ",")+"}", c.Param("homeId"))
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDDG004"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CSSGDDG004",
			Message: err.Error(),
		})
	}
	defer rows.Close()

	for rows.Next() {
		var device Device
		err := rows.StructScan(&device)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSSGDDG005"}).Errorf("%s", err.Error())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "CSSGDDG005",
				Message: err.Error(),
			})
		}

		for ind, disco := range discovered {
			if disco.PhysicalID == device.PhysicalID {
				discovered = append(discovered[:ind], discovered[ind+1:]...)
				continue
			}
		}
	}

	return c.JSON(http.StatusOK, discovered)
}

// SaveNewDatas save receive datas from gateway in DB
func SaveNewDatas(datas []Datas) {
	var devices []Device
	var arrayID []string

	for _, data := range datas {
		if !searchStringInArray(arrayID, data.DeviceID) {
			arrayID = append(arrayID, data.DeviceID)
		}
	}

	rows, err := DB.Queryx(`SELECT id, physical_name, physical_id, plugin FROM devices WHERE physical_id = ANY($1)`, "{"+strings.Join(arrayID, ",")+"}")
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDSND001"}).Errorf("%s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var device Device
		err := rows.StructScan(&device)
		if err != nil {
			logger.WithFields(logger.Fields{"code": "CSDSND002"}).Errorf("%s", err.Error())
		}
		devices = append(devices, device)
	}

	for _, data := range datas {
		device := findDeviceFromID(devices, data.DeviceID)
		if device != nil {
			data.DeviceID = device.ID

			if FindFieldFromName(sdk.FindDevicesFromName(configFromPlugin(Configs, device.Plugin).Devices, device.PhysicalName).Triggers, data.Field).Direct {
				queues = append(queues, data)
			}
			_, err = DB.Exec("INSERT INTO datas (id, device_id, field, value_nbr, value_str, value_bool) VALUES ($1, $2, $3, $4, $5, $6)",
				data.ID, data.DeviceID, data.Field, data.ValueNbr, data.ValueStr, data.ValueBool)
			if err != nil {
				logger.WithFields(logger.Fields{"code": "CSDSND003"}).Errorf("%s", err.Error())
				continue
			}
		}
	}
}

func searchStringInArray(array []string, str string) bool {
	for _, arr := range array {
		if arr == str {
			return true
		}
	}
	return false
}

func findDeviceFromID(devices []Device, id string) *Device {
	for _, device := range devices {
		if device.PhysicalID == id {
			return &device
		}
	}
	return nil
}

// Automations loop on automations to do actions
func Automations() {

	for range time.Tick(200 * time.Millisecond) {
		rows, err := DB.Queryx("SELECT * FROM automations")
		defer rows.Close()
		if err == nil {
			for rows.Next() {
				var auto Automation
				err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerValue), pq.Array(&auto.TriggerOperator), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.UpdatedAt, &auto.CreatorID)
				if err == nil {
					var conditions []string

					for i := 0; i < len(auto.Trigger); i++ {
						var device Device
						err = DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, auto.Trigger[i])
						field := FindFieldFromName(sdk.FindDevicesFromName(configFromPlugin(Configs, device.Plugin).Devices, device.PhysicalName).Triggers, auto.TriggerKey[i])

						if field.Direct {
							queue := FindDataFromID(queues, device.ID)
							if queue.DeviceID == device.ID {
								switch field.Type {
								case "string":
									if queue.ValueStr == auto.TriggerValue[i] {
										conditions = append(conditions, "1")
									}
								case "int":
									triggerValue, err := strconv.ParseFloat(string(auto.TriggerValue[i]), 64)
									if err == nil {
										if queue.ValueNbr == triggerValue {
											conditions = append(conditions, "1")
										}
									}
								case "bool":
								default:
								}
							}
						} else if device.ID == auto.Trigger[i] {
							var data Datas
							err = DB.Get(&data, `SELECT * FROM datas WHERE device_id = $1 AND field = $2 ORDER BY created_at DESC`, device.ID, auto.TriggerKey[i])
							switch field.Type {
							case "string":
								if data.ValueStr == auto.TriggerValue[i] {
									conditions = append(conditions, "1")
								}
							case "int":
								firstchar := string(auto.TriggerValue[i][0])
								secondchar := string(auto.TriggerValue[i][1])
								value, err := strconv.ParseFloat(string(auto.TriggerValue[i][1:]), 64)
								if err == nil {
									switch firstchar {
									case ">":
										if secondchar == "=" && data.ValueNbr >= value {
											conditions = append(conditions, "1")
											break
										}
										if data.ValueNbr > value {
											conditions = append(conditions, "1")
										}
									case "<":
										if secondchar == "=" && data.ValueNbr <= value {
											conditions = append(conditions, "1")
											break
										}
										if data.ValueNbr < value {
											conditions = append(conditions, "1")
										}
									case "=":
										if data.ValueNbr == value {
											conditions = append(conditions, "1")
										}
									case "!":
										if secondchar == "=" && data.ValueNbr != value {
											conditions = append(conditions, "1")
										}
									default:
									}
								}
							case "bool":
								triggerValueBool, err := strconv.ParseBool(auto.TriggerValue[i])
								if err == nil && data.ValueBool == triggerValueBool {
									conditions = append(conditions, "1")
								}
							default:
							}
						}
						if len(conditions) == 0 {
							conditions = append(conditions, "0")
						}
						if conditions[len(conditions)-1] != "0" && conditions[len(conditions)-1] != "1" {
							conditions = append(conditions, "0")
						}
						if len(auto.TriggerOperator) >= 1 && len(auto.TriggerOperator) > i {
							conditions = append(conditions, auto.TriggerOperator[i])
						}
					}

					if checkConditionOperator(conditions) {
						for i := 0; i < len(auto.Action); i++ {
							var device Device
							err = DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, auto.Action[i])
							if err == nil {

								act := ActionMessage{
									PhysicalID: device.PhysicalID,
									Plugin:     device.Plugin,
									Call:       auto.ActionCall[i],
									Config:     device.Config,
									Params:     auto.ActionValue[i],
								}

								marshAct, _ := json.Marshal(act)

								message := WebsocketMessage{
									Action: "callAction",
									Body:   marshAct,
								}

								marshMessage, _ := json.Marshal(message)
								if err != nil {
									logger.WithFields(logger.Fields{"code": "CSSA001"}).Errorf("%s", err.Error())
									break
								}
								logger.WithFields(logger.Fields{}).Debugf("Action sent to gateway")
								err = WSConn.WriteMessage(websocket.TextMessage, marshMessage)
								if err != nil {
									logger.WithFields(logger.Fields{"code": "CSSA002"}).Errorf("%s", err.Error())
									continue
								}
							}
						}
						_, err := DB.Exec("INSERT INTO logs (id, type, type_id, value) VALUES (generate_ulid(), $1, $2, $3)", "automation", auto.ID, "")
						if err != nil {
							logger.WithFields(logger.Fields{"code": "CSSA003"}).Errorf("%s", err.Error())
						}
					}
				}
			}
		}
		queues = nil
	}

	go Automations()
}

func checkConditionOperator(conditions []string) bool {
	index := 0
	groups := []bool{false}
	for i := 0; i < len(conditions); i++ {
		if conditions[i] == "1" && len(conditions) <= i+1 {
			groups[index] = true
		}
		if conditions[i] == "AND" {
			if conditions[i-1] == "1" && conditions[i+1] == "1" {
				groups[index] = true
			} else {
				groups[index] = false
			}
		} else if conditions[i] == "OR" {
			index++
			groups = append(groups, false)
		}
	}

	for _, group := range groups {
		if group {
			return true
		}
	}

	return false
}

func configFromPlugin(configurations []sdk.Configuration, name string) sdk.Configuration {
	for _, config := range configurations {
		if config.Name == name {
			return config
		}
	}
	return sdk.Configuration{}
}

// 	return nil
// }

// FindDataFromID find data with name ID
func FindDataFromID(datas []Datas, ID string) Datas {
	for _, data := range datas {
		if data.DeviceID == ID {
			return data
		}
	}
	return Datas{}
}

// FindFieldFromName find field with name field
func FindFieldFromName(triggers []sdk.Trigger, name string) sdk.Trigger {
	for _, trigger := range triggers {
		if trigger.Name == name {
			return trigger
		}
	}
	return sdk.Trigger{}
}
