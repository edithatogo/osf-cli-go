package download

import (
	"fmt"
)

func ExampleNormalizeRemotePath() {
	dst, _ := NormalizeRemotePath("/data//results.csv")
	fmt.Println(dst)
	// Output:
	// data/results.csv
}

func ExampleParseConflictPolicy() {
	policy, _ := ParseConflictPolicy("skip")
	fmt.Println(policy)
	// Output:
	// skip
}
