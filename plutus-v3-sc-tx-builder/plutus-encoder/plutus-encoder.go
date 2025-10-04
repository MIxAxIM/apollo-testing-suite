package plutusEncoder

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/PlutusData"
)

func GetAddressPlutusData(address Address.Address) (*PlutusData.PlutusData, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic occurred: %v", err)
			return
		}
	}()

	switch address.AddressType {
	case Address.KEY_KEY:
		return &PlutusData.PlutusData{
			TagNr:          121,
			PlutusDataType: PlutusData.PlutusArray,
			Value: PlutusData.PlutusDefArray{
				PlutusData.PlutusData{
					TagNr:          121,
					PlutusDataType: PlutusData.PlutusArray,
					Value: PlutusData.PlutusDefArray{
						PlutusData.PlutusData{
							TagNr:          0,
							Value:          address.PaymentPart,
							PlutusDataType: PlutusData.PlutusBytes,
						},
					},
				},
				PlutusData.PlutusData{
					TagNr:          121,
					PlutusDataType: PlutusData.PlutusArray,
					Value: PlutusData.PlutusDefArray{
						PlutusData.PlutusData{
							TagNr:          121,
							PlutusDataType: PlutusData.PlutusArray,
							Value: PlutusData.PlutusDefArray{
								PlutusData.PlutusData{
									TagNr:          121,
									PlutusDataType: PlutusData.PlutusArray,
									Value: PlutusData.PlutusDefArray{
										PlutusData.PlutusData{
											TagNr:          0,
											Value:          address.StakingPart,
											PlutusDataType: PlutusData.PlutusBytes},
									},
								},
							},
						},
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("error: Pointer Addresses are not supported")
	}
}

func MarshalPlutus(v interface{}) (*PlutusData.PlutusData, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic occurred: %v", err)
			return
		}
	}()

	var overallContainer interface{}
	var containerConstr = uint64(0)
	types := reflect.TypeOf(v)
	values := reflect.ValueOf(v)

	//get Container type
	fields, ok := types.FieldByName("_")

	if ok {
		typeOfStruct := fields.Tag.Get("plutusType")
		Constr := fields.Tag.Get("plutusConstr")

		if Constr != "" {
			parsedConstr, err := strconv.Atoi(Constr)
			if err != nil {
				return nil, fmt.Errorf("error parsing constructor: %v", err)
			}
			if parsedConstr < 7 {
				containerConstr = 121 + uint64(parsedConstr)
			} else if 7 <= parsedConstr && parsedConstr <= 1400 {
				containerConstr = 1280 + uint64(parsedConstr-7)
			} else {
				return nil, fmt.Errorf("parsedConstr value is above 1400")
			}
		}

		switch typeOfStruct {
		case "DefList":
			overallContainer = PlutusData.PlutusDefArray{}
		default:
			return nil, fmt.Errorf("error: unknown type")
		}

		for i := 0; i < types.NumField(); i++ {
			f := types.Field(i)
			if !f.IsExported() {
				continue
			}

			tag := f.Tag
			constr := uint64(0)
			typeOfField := tag.Get("plutusType")
			constrOfField := tag.Get("plutusConstr")

			if typeOfField == "Ignore" {
				continue
			}

			if len(typeOfField) > 1 {
				omitempty := false
				for _, part := range strings.Split(typeOfField, ",") {
					switch strings.ToLower(strings.TrimSpace(part)) {
					case "omitempty":
						omitempty = true
					}
				}
				if omitempty {
					continue
				}
			}

			if constrOfField != "" {
				parsedConstr, err := strconv.Atoi(constrOfField)
				if err != nil {
					return nil, fmt.Errorf("error parsing constructor: %v", err)
				}
				if parsedConstr < 7 {
					constr = 121 + uint64(parsedConstr)
				} else if 7 <= parsedConstr && parsedConstr <= 1400 {
					constr = 1280 + uint64(parsedConstr-7)
				} else {
					return nil, fmt.Errorf("parsedConstr value is above 1400")
				}
			}

			switch typeOfField {

			case "Int":
				if values.Field(i).Kind() != reflect.Int64 && values.Field(i).Kind() != reflect.Int32 && values.Field(i).Kind() != reflect.Int16 && values.Field(i).Kind() != reflect.Int8 && values.Field(i).Kind() != reflect.Int {

					return nil, fmt.Errorf("error: Int field is not int")
				}
				var pdi PlutusData.PlutusData
				switch values.Field(i).Kind() {
				case reflect.Int64:
					pdi = PlutusData.PlutusData{
						PlutusDataType: PlutusData.PlutusInt,
						Value:          values.Field(i).Interface().(int64),
						TagNr:          constr,
					}
				case reflect.Int32:
					pdi = PlutusData.PlutusData{
						PlutusDataType: PlutusData.PlutusInt,
						Value:          values.Field(i).Interface().(int32),
						TagNr:          constr,
					}
				case reflect.Int16:
					pdi = PlutusData.PlutusData{
						PlutusDataType: PlutusData.PlutusInt,
						Value:          values.Field(i).Interface().(int16),
						TagNr:          constr,
					}
				case reflect.Int8:
					pdi = PlutusData.PlutusData{
						PlutusDataType: PlutusData.PlutusInt,
						Value:          values.Field(i).Interface().(int8),
						TagNr:          constr,
					}
				case reflect.Int:
					pdi = PlutusData.PlutusData{
						PlutusDataType: PlutusData.PlutusInt,
						Value:          values.Field(i).Interface().(int),
						TagNr:          constr,
					}
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), pdi)

			case "StringBytes":
				if values.Field(i).Kind() != reflect.String {
					return nil, fmt.Errorf("error: StringBytes field is not string")
				}
				pdsb := PlutusData.PlutusData{
					PlutusDataType: PlutusData.PlutusBytes,
					Value:          []byte(values.Field(i).Interface().(string)),
					TagNr:          constr,
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), pdsb)

			case "HexString":
				if values.Field(i).Kind() != reflect.String {
					return nil, fmt.Errorf("error: HexString field is not string")
				}
				hexString, err := hex.DecodeString(values.Field(i).Interface().(string))
				if err != nil {
					return nil, fmt.Errorf("error: HexString field is not string")
				}
				pdsb := PlutusData.PlutusData{
					PlutusDataType: PlutusData.PlutusBytes,
					Value:          hexString,
					TagNr:          constr,
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), pdsb)

			case "Address":
				addPd, err := GetAddressPlutusData(values.Field(i).Interface().(Address.Address))
				if err != nil {
					return nil, fmt.Errorf("error marshalling: %v", err)
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), *addPd)

			case "DefList":
				container := PlutusData.PlutusDefArray{}
				for j := 0; j < values.Field(i).Len(); j++ {
					pd, err := MarshalPlutus(values.Field(i).Index(j).Interface())
					if err != nil {
						return nil, fmt.Errorf("error marshalling: %v", err)
					}
					container = append(container, *pd)
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), PlutusData.PlutusData{
					PlutusDataType: PlutusData.PlutusArray,
					Value:          container,
					TagNr:          constr,
				})

			default:
				pd, err := MarshalPlutus(values.Field(i).Interface())
				if err != nil {
					return nil, fmt.Errorf("error marshalling: %v", err)
				}
				overallContainer = append(overallContainer.(PlutusData.PlutusDefArray), *pd)

			}
		}

	}

	pType := PlutusData.PlutusArray

	pd := PlutusData.PlutusData{
		PlutusDataType: pType,
		Value:          overallContainer,
		TagNr:          containerConstr,
	}
	return &pd, nil
}
