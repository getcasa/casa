package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
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

var wsconn *websocket.Conn
var queues []Datas
var configs []sdk.Configuration
var discovered []sdk.Device
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

	if len(configs) == 0 {
		configs = tmpConfigs
	} else {
		for _, tmpConf := range tmpConfigs {
			if configFromPlugin(configs, tmpConf.Name).Name == "" {
				configs = append(configs, tmpConf)
			}
		}
	}
}

// InitConnection create websocket connection
func InitConnection(con echo.Context) error {
	var err error
	wsconn, err = upgrader.Upgrade(con.Response(), con.Request(), nil) // error ignored for sake of simplicity
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDIC001"}).Errorf("%s", err.Error())
		return err
	}
	defer wsconn.Close()

	for {
		var wm WebsocketMessage

		_, message, err := wsconn.ReadMessage()
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
		case "discoverevices":
			json.Unmarshal(wm.Body, &discovered)
		case "newConnection":
			GetConfigFromGateway(string(wm.Body))
		case "newData":
			var datas Datas
			json.Unmarshal(wm.Body, &datas)
			SaveNewDatas(datas)
		default:
			continue
		}

		logger.WithFields(logger.Fields{}).Debugf("recv: %s", message)
	}
}

// GetDiscoveredDevices return an array of futur discover
func GetDiscoveredDevices(c echo.Context) error {
	c.Param("gatewayId")

	message := WebsocketMessage{
		Action: "discoverDevices",
		Body:   []byte(""),
	}

	marshMessage, _ := json.Marshal(message)
	err := wsconn.WriteMessage(websocket.TextMessage, marshMessage)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDD001"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGDD001",
			Message: err.Error(),
		})
	}

	for len(discovered) == 0 {
	}

	marshDiscovered, err := json.Marshal(discovered)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSSGDD002"}).Errorf("%s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CSSGDD002",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: string(marshDiscovered),
	})
}

// SaveNewDatas save receive datas from gateway in DB
func SaveNewDatas(queue Datas) {
	var device Device

	err := DB.Get(&device, `SELECT id, physical_name, plugin FROM devices WHERE physical_id = $1`, queue.DeviceID)
	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDSND001"}).Errorf("%s", err.Error())
		return
	}
	queue.DeviceID = device.ID

	if FindFieldFromName(sdk.FindTriggerFromName(configFromPlugin(configs, device.Plugin).Triggers, device.PhysicalName).Fields, queue.Field).Direct {
		queues = append(queues, queue)
	}

	_, err = DB.Exec("INSERT INTO datas (id, device_id, field, value_nbr, value_str, value_bool) VALUES ($1, $2, $3, $4, $5, $6)",
		queue.ID, queue.DeviceID, queue.Field, queue.ValueNbr, queue.ValueStr, queue.ValueBool)

	if err != nil {
		logger.WithFields(logger.Fields{"code": "CSDSND002"}).Errorf("%s", err.Error())
		return
	}
}

// Automations loop on automations to do actions
func Automations() {

	for range time.Tick(250 * time.Millisecond) {
		rows, err := DB.Queryx("SELECT * FROM automations")
		if err == nil {
			for rows.Next() {
				var auto Automation
				err := rows.Scan(&auto.ID, &auto.HomeID, &auto.Name, pq.Array(&auto.Trigger), pq.Array(&auto.TriggerKey), pq.Array(&auto.TriggerValue), pq.Array(&auto.TriggerOperator), pq.Array(&auto.Action), pq.Array(&auto.ActionCall), pq.Array(&auto.ActionValue), &auto.Status, &auto.CreatedAt, &auto.UpdatedAt, &auto.CreatorID)
				if err == nil {
					var conditions []string

					for i := 0; i < len(auto.Trigger); i++ {
						var device Device
						err = DB.Get(&device, `SELECT * FROM devices WHERE id = $1`, auto.Trigger[i])
						field := FindFieldFromName(sdk.FindTriggerFromName(configFromPlugin(configs, device.Plugin).Triggers, device.PhysicalName).Fields, auto.TriggerKey[i])

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
								value, err := strconv.ParseFloat(string(auto.TriggerValue[i][1:]), 64)
								if err == nil {
									switch firstchar {
									case ">":
										if data.ValueNbr > value {
											conditions = append(conditions, "1")
										}
									case "<":
										if data.ValueNbr < value {
											conditions = append(conditions, "1")
										}
									case "=":
										if data.ValueNbr == value {
											conditions = append(conditions, "1")
										}
									default:
									}
								}
							case "bool":
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
								logger.WithFields(logger.Fields{}).Debugf("Action sent to gateway")
								err = wsconn.WriteMessage(websocket.TextMessage, marshMessage)
								if err != nil {
									logger.WithFields(logger.Fields{"code": "CSSA001"}).Errorf("%s", err.Error())
									break
								}
							}
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
func FindFieldFromName(fields []sdk.Field, name string) sdk.Field {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	return sdk.Field{}
}
