package configs

import (
	"errors"
	"strings"
)

var ErrInvalidConfigName = errors.New("invalid config name")

type NameTag struct {
	Name string
	Tag  string
}

func ParseName(nameTag string) (NameTag, error) {
	var res NameTag
	parts := strings.Split(nameTag, ":")
	if len(parts) > 2 {
		return res, ErrInvalidConfigName
	}
	res.Name = parts[0]
	if len(parts) >= 2 {
		res.Tag = parts[1]
	}
	return res, nil
}

func JoinName(nameTag NameTag) string {
	var tagStr string
	if nameTag.Tag != "" {
		tagStr = ":" + nameTag.Tag
	}
	return nameTag.Name + tagStr
}
