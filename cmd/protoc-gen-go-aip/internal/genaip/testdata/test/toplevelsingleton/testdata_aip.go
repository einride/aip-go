// Code generated by protoc-gen-go-aip. DO NOT EDIT.
//
// versions:
// 	protoc-gen-go-aip development
// 	protoc (unknown)
// source: test/toplevelsingleton/testdata.proto

package toplevelsingleton

import (
	resourcename "go.einride.tech/aip/resourcename"
)

type ConfigResourceName struct {
}

func (n ConfigResourceName) Validate() error {
	return nil
}

func (n ConfigResourceName) ContainsWildcard() bool {
	return false
}

func (n ConfigResourceName) String() string {
	return resourcename.Sprint(
		"config",
	)
}

func (n ConfigResourceName) MarshalString() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	return n.String(), nil
}

func (n *ConfigResourceName) UnmarshalString(name string) error {
	err := resourcename.Sscan(
		name,
		"config",
	)
	if err != nil {
		return err
	}
	return n.Validate()
}

func (n ConfigResourceName) Type() string {
	return "test1.testdata/Config"
}
