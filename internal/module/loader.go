package module

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

func LoadOrder(moduleFS fs.FS) (*Order, error) {
	data, err := fs.ReadFile(moduleFS, "order.json")
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere order.json: %w", err)
	}
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("impossibile parsare order.json: %w", err)
	}
	return &order, nil
}

func LoadModules(moduleFS fs.FS, order *Order) ([]*Module, error) {
	modules := make([]*Module, 0, len(order.Order))

	for _, folder := range order.Order {
		cmdPath := folder + "/command.json"
		data, err := fs.ReadFile(moduleFS, cmdPath)
		if err != nil {
			return nil, fmt.Errorf("impossibile leggere %s: %w", cmdPath, err)
		}

		var cmd Command
		if err := json.Unmarshal(data, &cmd); err != nil {
			return nil, fmt.Errorf("impossibile parsare %s: %w", cmdPath, err)
		}

		modules = append(modules, &Module{
			FolderName: folder,
			Command:    cmd,
			Status:     StatusPending,
		})
	}

	return modules, nil
}

func TotalWeight(modules []*Module) int {
	total := 0
	for _, m := range modules {
		total += m.Command.Weight
	}
	return total
}
