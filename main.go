package main

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	// Import the releaseTree package
	"lca/releaseTree"
)

func main() {
	fmt.Println("--- Release Tree Test Driver ---")

	// --- Define All Sample Data using Exported types from releaseTree package ---
	release11 := releaseTree.ReleaseInput{Ver: "11", FromVer: "", Changes: []releaseTree.Chg{}}
	release21 := releaseTree.ReleaseInput{Ver: "21", FromVer: "11", Changes: []releaseTree.Chg{{ID: "1"}}}
	release31 := releaseTree.ReleaseInput{Ver: "31", FromVer: "21", Changes: []releaseTree.Chg{{ID: "2"}, {ID: "3"}, {ID: "4"}}}
	release22 := releaseTree.ReleaseInput{Ver: "22", FromVer: "21", Changes: []releaseTree.Chg{{ID: "5"}}}
	release32 := releaseTree.ReleaseInput{Ver: "32", FromVer: "31", Changes: []releaseTree.Chg{{ID: "5"}, {ID: "6"}, {ID: "7"}, {ID: "8"}}}
	release24 := releaseTree.ReleaseInput{Ver: "24", FromVer: "22", Changes: []releaseTree.Chg{{ID: "6"}, {ID: "7"}}}
	release33 := releaseTree.ReleaseInput{Ver: "33", FromVer: "31", Changes: []releaseTree.Chg{{ID: "5"}, {ID: "6"}, {ID: "7"}, {ID: "10"}}}
	release23 := releaseTree.ReleaseInput{Ver: "23", FromVer: "21", Changes: []releaseTree.Chg{{ID: "10"}}}

	// --- Build Initial Tree (Batch Load) ---
	fmt.Println("\n--- Building Initial Tree (Batch) ---")
	initialInputs := []releaseTree.ReleaseInput{ // Use package type
		release11,
		release21,
		release31,
		release22,
	}
	// Call exported NewReleaseTree function from the package
	tree, err := releaseTree.NewReleaseTree(initialInputs)
	if err != nil {
		fmt.Printf("Error building initial tree: %v\n", err)
		return
	}
	fmt.Println("Initial tree built successfully.")
	// Cannot directly access tree.root or tree.nodes anymore.
	// We could add getter methods to releaseTree if needed for verification,
	// but for now we'll rely on FindLCA/CalcChgs results.
	fmt.Printf("Tree initialized (verification requires exported methods or getters).\n")

	// --- Insert Remaining Nodes Concurrently ---
	fmt.Println("\n--- Inserting Remaining Nodes Concurrently ---")
	nodesToInsert := []releaseTree.ReleaseInput{ // Use package type
		release32,
		release24,
		release33,
		release23,
	}
	var insertWg sync.WaitGroup
	for _, nodeInput := range nodesToInsert {
		insertWg.Add(1)
		inputToInsert := nodeInput
		go func() {
			defer insertWg.Done()
			// Call exported InsertNode method
			err := tree.InsertNode(inputToInsert)
			if err != nil {
				fmt.Printf("!!! Error inserting %s: %v\n", inputToInsert.Ver, err)
			}
			time.Sleep(time.Duration(5+len(inputToInsert.Ver)) * time.Millisecond)
		}()
	}
	insertWg.Wait()
	fmt.Println("Finished concurrent insertions.")

	// --- Verify Final Tree State (indirectly via exported methods) ---
	// We cannot directly check len(tree.nodes) or children lists anymore.
	// We can use FindLCA as a proxy to check if nodes seem correctly linked.
	fmt.Println("\n--- Verifying Final Tree State (using FindLCA) ---")
	_, err21_31 := tree.FindLCA("21", "31") // Should be 21
	_, err21_23 := tree.FindLCA("21", "23") // Should be 21
	_, err31_33 := tree.FindLCA("31", "33") // Should be 31
	if err21_31 != nil || err21_23 != nil || err31_33 != nil {
		fmt.Println("Verification failed: Error during LCA checks after insert.")
	} else {
		fmt.Println("Verification checks passed (basic LCA relationships seem ok).")
	}

	// --- Demonstrate Exported FindLCA ---
	fmt.Println("\n--- Demonstrating Exported FindLCA ---")
	// Call exported FindLCA method
	lcaVer1, errLca1 := tree.FindLCA("32", "24")
	if errLca1 != nil {
		fmt.Printf("Error finding LCA(32, 24): %v\n", errLca1)
	} else {
		fmt.Printf("LCA(32, 24) = %s\n", lcaVer1)
	} // Expected: 21
	lcaVer2, errLca2 := tree.FindLCA("33", "23")
	if errLca2 != nil {
		fmt.Printf("Error finding LCA(33, 23): %v\n", errLca2)
	} else {
		fmt.Printf("LCA(33, 23) = %s\n", lcaVer2)
	} // Expected: 21
	lcaVer3, errLca3 := tree.FindLCA("31", "22")
	if errLca3 != nil {
		fmt.Printf("Error finding LCA(31, 22): %v\n", errLca3)
	} else {
		fmt.Printf("LCA(31, 22) = %s\n", lcaVer3)
	} // Expected: 21
	lcaVer4, errLca4 := tree.FindLCA("32", "99") // Test error case
	if errLca4 != nil {
		fmt.Printf("Got expected error finding LCA(32, 99): %v\n", errLca4)
	} else {
		fmt.Printf("FAIL: Expected error finding LCA(32, 99), got %s\n", lcaVer4)
	}

	// --- Define Test Cases for CalcChgs ---
	type testCase struct {
		name           string
		endVersion     string
		startVersion   string
		expectedResult []releaseTree.Chg // Use package type
		expectError    bool
		errorContains  string
	}

	testCases := []testCase{ // Use package type Chg and field ID for expected results
		{name: "TC1: Original (32 vs 24)", endVersion: "32", startVersion: "24", expectedResult: []releaseTree.Chg{{ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "8"}}, expectError: false},
		{name: "TC2: Subset Fail (31 vs 22)", endVersion: "31", startVersion: "22", expectedResult: nil, expectError: true, errorContains: "change ID '5'"},
		{name: "TC3: Subset OK (33 vs 23)", endVersion: "33", startVersion: "23", expectedResult: []releaseTree.Chg{{ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"}, {ID: "6"}, {ID: "7"}}, expectError: false},
		{name: "TC4: Non-existent Node (32 vs 99)", endVersion: "32", startVersion: "99", expectedResult: nil, expectError: true, errorContains: "version '99' not found"},
		{name: "TC5: Ancestor (33 vs 21)", endVersion: "33", startVersion: "21", expectedResult: []releaseTree.Chg{{ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"}, {ID: "6"}, {ID: "7"}, {ID: "10"}}, expectError: false},
	}

	// --- Run CalcChgs Test Cases Concurrently ---
	fmt.Println("\n--- Running CalcChgs Test Cases Concurrently ---")
	var testWg sync.WaitGroup

	for _, tc := range testCases {
		testWg.Add(1)
		currentTest := tc
		go func() {
			defer testWg.Done()
			fmt.Printf("Running %s...\n", currentTest.name)
			// Call EXPORTED CalcChgs method from package
			actualResult, actualErr := tree.CalcChgs(currentTest.endVersion, currentTest.startVersion)

			// Check results
			if currentTest.expectError {
				if actualErr == nil {
					fmt.Printf("FAIL: %s - Expected an error, but got none.\n", currentTest.name)
				} else if currentTest.errorContains != "" && !strings.Contains(actualErr.Error(), currentTest.errorContains) {
					fmt.Printf("FAIL: %s - Expected error containing '%s', but got: %v\n", currentTest.name, currentTest.errorContains, actualErr)
				} else {
					fmt.Printf("PASS: %s - Got expected error: %v\n", currentTest.name, actualErr)
				}
			} else { // Expecting success
				if actualErr != nil {
					fmt.Printf("FAIL: %s - Expected no error, but got: %v\n", currentTest.name, actualErr)
				} else if !reflect.DeepEqual(actualResult, currentTest.expectedResult) {
					fmt.Printf("FAIL: %s - Result mismatch.\n      Expected: %+v\n      Actual:   %+v\n",
						currentTest.name, currentTest.expectedResult, actualResult)
				} else {
					fmt.Printf("PASS: %s - Result matches expected: %+v\n", currentTest.name, actualResult)
				}
			}
		}()
	}

	fmt.Println("Waiting for concurrent tests to complete...")
	testWg.Wait()
	fmt.Println("Finished running tests.")
	fmt.Println("--------------------------------------------")
}
