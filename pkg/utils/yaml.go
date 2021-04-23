package utils

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func GetYAMLNode(node *yaml.Node, field string) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.Errorf("expected mapping node, got kind=%d", node.Kind)
	}

	for i, subnode := range node.Content {
		// In a map, we need at least two elements to form a (key, value) pair.
		// So if we're at the last element, this field cannot be in this map.
		if i == len(node.Content)-1 {
			break
		}

		if subnode.Tag == "!!str" && subnode.Value == field {
			return node.Content[i+1], nil
		}
	}

	return nil, nil
}

func SetYAMLField(path, field, value string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return errors.Wrap(err, "opening task definition")
	}
	defer f.Close()

	root := yaml.Node{}
	if err := yaml.NewDecoder(f).Decode(&root); err != nil {
		return errors.Wrap(err, "unmarshalling task definition")
	}

	if len(root.Content) == 0 {
		return errors.Errorf("cannot insert %s: yaml document empty", field)
	}
	// Find the root map, which may not be the first element due to comments.
	var mapnode *yaml.Node
	for _, subnode := range root.Content {
		if subnode.Kind == yaml.MappingNode {
			mapnode = subnode
			break
		}
	}
	if mapnode == nil {
		return errors.Errorf("cannot insert %s: yaml document has map field", field)
	}

	node, err := GetYAMLNode(mapnode, field)
	if err != nil {
		return err
	}

	if node != nil {
		// This field already exists, so just update it's value.
		node.Value = value
	} else {
		// This field does not exist yet, so prepend this field into the map.
		mapnode.Content = append([]*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: field,
			},
			{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: value,
			},
		}, mapnode.Content...)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return errors.Wrap(err, "seeking to start of task definition")
	}
	if err := f.Truncate(0); err != nil {
		return errors.Wrap(err, "truncating file")
	}
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(&root); err != nil {
		return errors.Wrap(err, "marshalling task definition")
	}

	return nil
}
