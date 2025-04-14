Tree Structure Diagram:

          11 (Root)
           |
           21
         / | \
        /  |  \
       /   |   \
      31   22   23
     / \    |
    /   \   24
   /     \
  32      33

Test Cases Breakdown (CalcChgs(End, Start)):

TC1: CalcChgs("32", "24")

End Node: 32 (Path to LCA: 32 -> 31)
Start Node: 24 (Path to LCA: 24 -> 22)
LCA: 21
Adds changes from 32 {5,6,7,8} and 31 {2,3,4}. Total added: {2,3,4,5,6,7,8}.
Subtracts changes from 24 {6,7} and 22 {5}. All exist in added set.
Expected Result: [{ID:"2"}, {ID:"3"}, {ID:"4"}, {ID:"8"}] (Success)

TC2: CalcChgs("31", "22")

End Node: 31 (Path to LCA: 31)
Start Node: 22 (Path to LCA: 22)
LCA: 21
Adds changes from 31 {2,3,4}. Total added: {2,3,4}.
Subtracts changes from 22 {5}. Change 5 does not exist in added set.
Expected Result: Error (Subset check failure)

TC3: CalcChgs("33", "23")

End Node: 33 (Path to LCA: 33 -> 31)
Start Node: 23 (Path to LCA: 23)
LCA: 21
Adds changes from 33 {5,6,7,10} and 31 {2,3,4}. Total added: {2,3,4,5,6,7,10}.
Subtracts changes from 23 {10}. Change 10 exists in added set.
Expected Result: [{ID:"2"}, {ID:"3"}, {ID:"4"}, {ID:"5"}, {ID:"6"}, {ID:"7"}] (Success)

TC4: CalcChgs("32", "99")

End Node: 32
Start Node: 99 (Doesn't exist in the tree)
LCA: N/A
Expected Result: Error (Node "99" not found during LCA lookup)

TC5: CalcChgs("33", "21")

End Node: 33 (Path to LCA: 33 -> 31)
Start Node: 21 (Path to LCA: empty, already at LCA)
LCA: 21
Adds changes from 33 {5,6,7,10} and 31 {2,3,4}. Total added: {2,3,4,5,6,7,10}.
Subtracts changes from path 21 up to (but not including) 21. This path is empty, so nothing is subtracted.
Expected Result: [{ID:"2"}, {ID:"3"}, {ID:"4"}, {ID:"5"}, {ID:"6"}, {ID:"7"}, {ID:"10"}] (Success)
