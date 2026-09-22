package main

import (
	"fmt"
	"os"
	"sort"
)

type Operation string

var operationOrder = map[Operation]int{
	"sync":    0,
	"restart": 1,
}

const (
	OperationRestart Operation = "restart"
	OperationSync    Operation = "sync"
)

func (o Operation) isValid() bool {
	switch o {
	case OperationRestart, OperationSync:
		return true
	}
	return false
}

func prepareOperations(operationFlag string) []Operation {
	ops := splitCSV(operationFlag)
	if len(ops) == 0 {
		fmt.Println(Red + "no operation provided" + Reset)
		os.Exit(1)
	}

	operations := make([]Operation, 0, 2)
	for _, op := range ops {
		curOp := Operation(op)
		if !curOp.isValid() {
			fmt.Println(Red + "sorry " + op + " is not supported " + Reset)
			os.Exit(1)
		}
		operations = append(operations, curOp)
	}
	operations = dedupe(operations)
	sort.Slice(operations, func(i, j int) bool {
		return operationOrder[operations[i]] < operationOrder[operations[j]]
	})

	return operations
}
