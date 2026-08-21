package r4_test

import (
	"fmt"

	"github.com/gofhir/models/r4"
)

// Example is the snippet the README shows. It lives here as a compiled test so
// the documented entry point cannot rot: the previous README used r4.String and
// r4.Boolean, neither of which has ever existed in this package.
func Example() {
	patient := r4.Patient{
		ResourceType: "Patient",
		Id:           r4.Ptr("123"),
		Active:       r4.Ptr(true),
		Name: []r4.HumanName{
			{
				Family: r4.Ptr("Smith"),
				Given:  []string{"John"},
			},
		},
	}

	data, err := r4.MarshalIndent(patient, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	// Output:
	// {
	//   "resourceType": "Patient",
	//   "id": "123",
	//   "active": true,
	//   "name": [
	//     {
	//       "family": "Smith",
	//       "given": [
	//         "John"
	//       ]
	//     }
	//   ]
	// }
}
