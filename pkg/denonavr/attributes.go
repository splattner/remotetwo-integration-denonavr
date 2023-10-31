package denonavr

import (
	"fmt"
	"reflect"
)

// Set an attribute and call entity Change function if changed
func (d *DenonAVR) SetAttribute(name string, value interface{}) {

	d.attributeMutex.Lock()
	defer d.attributeMutex.Unlock()

	changed := d.attributes[name] == nil || !reflect.DeepEqual(d.attributes[name], value)

	d.attributes[name] = value

	if changed {
		d.callEntityChangeFunction(name, d.attributes[name])
	}

}

func (d *DenonAVR) GetAttribute(name string) (interface{}, error) {

	d.attributeMutex.Lock()
	defer d.attributeMutex.Unlock()

	if d.attributes[name] == nil {
		return nil, fmt.Errorf("Attribute not Found")
	}

	return d.attributes[name], nil

}
