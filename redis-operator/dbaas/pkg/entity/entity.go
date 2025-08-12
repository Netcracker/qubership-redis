package entity

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

type DbCreateRequestSettings struct {
	RedisDbSettings               map[string]interface{}  `json:"redisDbSettings,omitempty" mapstructure:"redisDbSettings"`
	RedisDbResources              v1.ResourceRequirements `json:"redisDbResources,omitempty" mapstructure:"redisDbResources"`
	RedisDbNodeSelector           map[string]string       `json:"redisDbNodeSelector,omitempty" mapstructure:"redisDbNodeSelector"`
	RedisDbWaitStartServiceSecond int                     `json:"redisDbWaitStartServiceSecond,omitempty" mapstructure:"redisDbWaitStartServiceSecond"`
}

type ConnectionProperties struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Service  string `json:"service" mapstructure:"service"`
	Url      string `json:"url" mapstructure:"url"`
	Password string `json:"password" mapstructure:"password"`
	Role     string `json:"role" mapstructure:"role"`
}

type TelegrafData struct {
	Config string `json:"telegraf.conf"`
}
type PatchData struct {
	Data map[string]string `json:"data"`
}

func StructToMap(s interface{}) (map[string]interface{}, error) {
	// create a new map to store the converted values
	m := make(map[string]interface{})

	// get the type of the struct
	t := reflect.TypeOf(s)

	// iterate over the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		// get the field name and json tag
		fieldName := t.Field(i).Name
		jsonTag := t.Field(i).Tag.Get("json")

		// if the json tag is "-", skip this field
		if jsonTag == "-" {
			continue
		}

		// if the json tag is not set, use the field name as key
		if jsonTag == "" {
			jsonTag = fieldName
		}

		// get the field value
		fieldValue := reflect.ValueOf(s).Field(i).Interface()

		// encode the field value to JSON and decode it to a generic interface{}
		encoded, err := json.Marshal(fieldValue)
		if err != nil {
			return nil, err
		}
		var decoded interface{}
		err = json.Unmarshal(encoded, &decoded)
		if err != nil {
			return nil, err
		}

		// add the field to the map
		m[jsonTag] = decoded
	}

	return m, nil
}

func MapToStruct(m map[string]interface{}, s interface{}) error {
	// get the type of the struct
	t := reflect.TypeOf(s)

	// create a new value of the struct type
	v := reflect.New(t).Elem()

	// iterate over the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		// get the field name and json tag
		fieldName := t.Field(i).Name
		jsonTag := t.Field(i).Tag.Get("json")

		// if the json tag is "-", skip this field
		if jsonTag == "-" {
			continue
		}

		// if the json tag is not set, use the field name as key
		if jsonTag == "" {
			jsonTag = fieldName
		}

		// get the value from the map
		fieldValue, ok := m[jsonTag]
		if !ok {
			continue
		}

		// encode the value to JSON and decode it to the field type
		encoded, err := json.Marshal(fieldValue)
		if err != nil {
			return err
		}
		err = json.Unmarshal(encoded, v.Field(i).Addr().Interface())
		if err != nil {
			return err
		}
	}

	// set the value of the struct argument to the newly created value
	reflect.ValueOf(s).Elem().Set(v)

	return nil
}
