/*
 * Copyright © 2018. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

 package SendUldp

import (
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"encoding/json"
)

func main() {
	fmt.Printf(toFlatText("{\"id\": \"uu\", \"op\":[1,2,3,4]}"))
}

func toFlatText(s string) (string) {
	var obj interface{}
	json.Unmarshal([]byte(s), &obj)
	if obj == nil {
		return s;
	}
	return toFlatTextRec("", obj);
}

func toFlatTextRec(prefix string, jsonNode interface{}) (string) {
	var buffer bytes.Buffer
	switch  jsonNode.(type) {
	case []interface{}:
		array := jsonNode.([]interface{})
		for i, e := range array {
			if ( e == nil ) {
				break;
			}

			str := toFlatTextRec(prefix+"["+strconv.Itoa(i)+"]", e);
			if ( buffer.Len() != 0 ) {
				buffer.WriteString(", ")
			}
			buffer.WriteString(str)
		}

	case map[string]interface{}:
		theMap := jsonNode.(map[string]interface{})
		for k, v := range theMap {

			var dot string
			if (len(prefix) != 0) {
				dot = "."
			} else {
				dot = ""
			}
			str := toFlatTextRec(prefix+dot+k, v);
			if ( buffer.Len() != 0 ) {
				buffer.WriteString(", ")
			}
			buffer.WriteString(str)
		}

	case string:
		valueStr := jsonNode.(string)
		valueStr = strings.Replace(valueStr, "\\", "\\\\", -1);
		valueStr = strings.Replace(valueStr, "\"", "\\\"", -1);
		buffer.WriteString(prefix + "=\"" + valueStr + "\"");

	case float64:
		valueFloat64 := jsonNode.(float64)
		buffer.WriteString(prefix + "=\"" + strconv.FormatFloat(valueFloat64, 'g', -1, 64) + "\"")

	case bool:
		valueBool := jsonNode.(bool)
		var value string
		if (valueBool) {
			value = "\"true\""
		} else {
			value = "\"false\""
		}
		buffer.WriteString(prefix + "=\"" + value + "\"");

	case nil:
		buffer.WriteString(prefix + "=null");
	}
	return buffer.String()
}
