// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package verification

import (
	"fmt"
	"reflect"

	httppool "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// DebugInterfaces analyzes interfaces and their implementations
func DebugInterfaces() {
	// Get the interface type
	interfaceType := reflect.TypeOf((*interfaces.ConnectionPoolManager)(nil)).Elem()
	fmt.Printf("Interface: %s has %d methods\n", interfaceType.Name(), interfaceType.NumMethod())

	for i := 0; i < interfaceType.NumMethod(); i++ {
		method := interfaceType.Method(i)
		fmt.Printf("  - Method: %s\n", method.Name)
	}

	// Get the concrete type
	concreteType := reflect.TypeOf(&httppool.HttpConnectionPoolManager{})
	fmt.Printf("\nConcrete type: %s has %d methods\n", concreteType.Elem().Name(), concreteType.NumMethod())

	for i := 0; i < concreteType.NumMethod(); i++ {
		method := concreteType.Method(i)
		fmt.Printf("  - Method: %s\n", method.Name)
	}

	// Check each interface method against the concrete type
	fmt.Println("\nChecking interface methods against concrete type:")
	for i := 0; i < interfaceType.NumMethod(); i++ {
		method := interfaceType.Method(i)
		_, ok := concreteType.MethodByName(method.Name)
		fmt.Printf("  - Method %s: %s\n", method.Name, map[bool]string{true: "implemented", false: "NOT IMPLEMENTED"}[ok])
	}

	// Try the direct type assertion
	fmt.Println("\nTrying direct type assertion:")
	manager := httppool.NewHttpConnectionPoolManager(nil)
	_, ok := interface{}(manager).(interfaces.ConnectionPoolManager)
	fmt.Printf("Type assertion result: %t\n", ok)

	// Interface direct check
	fmt.Println("\nUsing direct interface check:")
	fmt.Printf("Does concrete type implement interface? %t\n",
		reflect.TypeOf(&httppool.HttpConnectionPoolManager{}).Elem().Implements(
			reflect.TypeOf((*interfaces.ConnectionPoolManager)(nil)).Elem(),
		),
	)

	// Check signatures for each method
	fmt.Println("\nChecking method signatures:")
	for i := 0; i < interfaceType.NumMethod(); i++ {
		ifaceMethod := interfaceType.Method(i)
		ifaceType := ifaceMethod.Type

		concMethod, ok := concreteType.MethodByName(ifaceMethod.Name)
		if !ok {
			fmt.Printf("  - Method %s: NOT IMPLEMENTED\n", ifaceMethod.Name)
			continue
		}

		concType := concMethod.Type
		fmt.Printf("  - Method %s:\n", ifaceMethod.Name)
		fmt.Printf("    Interface: %s\n", ifaceType)
		fmt.Printf("    Concrete:  %s\n", concType)

		// Adjust for receiver parameter
		ifaceNumIn := ifaceType.NumIn()
		concNumIn := concType.NumIn() - 1 // Subtract 1 for the receiver

		// Check if method signatures match with receiver adjustment
		if ifaceNumIn != concNumIn {
			fmt.Printf("    MISMATCH: Different number of input parameters (iface: %d, concrete: %d)\n",
				ifaceNumIn, concNumIn)
			continue
		}

		if ifaceType.NumOut() != concType.NumOut() {
			fmt.Printf("    MISMATCH: Different number of return values\n")
			continue
		}

		// For inputs, start at index 0 for interface but index 1 for concrete (to skip receiver)
		mismatch := false
		for j := 0; j < ifaceNumIn; j++ {
			ifaceIn := ifaceType.In(j)
			concIn := concType.In(j + 1) // +1 to skip the receiver

			if !ifaceIn.AssignableTo(concIn) && !concIn.AssignableTo(ifaceIn) {
				fmt.Printf("    MISMATCH: Input parameter %d types don't match\n", j)
				fmt.Printf("      Interface: %s\n", ifaceIn)
				fmt.Printf("      Concrete:  %s\n", concIn)
				mismatch = true
			}
		}

		for j := 0; j < ifaceType.NumOut(); j++ {
			ifaceOut := ifaceType.Out(j)
			concOut := concType.Out(j)

			if !concOut.AssignableTo(ifaceOut) {
				fmt.Printf("    MISMATCH: Return value %d types don't match\n", j)
				fmt.Printf("      Interface: %s\n", ifaceOut)
				fmt.Printf("      Concrete:  %s\n", concOut)
				mismatch = true
			}
		}

		if !mismatch {
			fmt.Printf("    Signatures match correctly\n")
		}
	}
}
