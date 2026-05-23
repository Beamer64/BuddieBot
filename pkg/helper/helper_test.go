package helper

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestCheckIfStructValueISEmpty(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	var floatVal interface{} = 12.5
	var stringVal interface{} = "Test String"
	var intVal interface{} = 7
	var nilVal interface{}

	var x []interface{}

	x = append(x, floatVal)
	x = append(x, stringVal)
	x = append(x, intVal)
	x = append(x, nilVal)

	for _, face := range x {
		retVal := ""

		if face != nil {
			switch face.(type) {
			case int:
				retVal = fmt.Sprintf("%v", reflect.ValueOf(face).Int())

			case float64:
				retVal = fmt.Sprintf("%v", reflect.ValueOf(face).Float())

			case string:
				if face != "" && face != " " {
					retVal = reflect.ValueOf(face).String()
				}

			default:
				retVal = "N/A"
			}
		} else {
			retVal = "N/A"
		}

		fmt.Println(fmt.Sprintf("Type: %v Value: %s", reflect.TypeOf(face), retVal))
	}
}
