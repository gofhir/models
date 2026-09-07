package r4_test

import (
	"encoding/json"
	"fmt"

	"github.com/gofhir/models/r4/v2"
)

// ExamplePatient builds a resource with a struct literal, which is still the
// direct way when the shape is known at the call site.
func ExamplePatient() {
	patient := r4.Patient{
		Id:     r4.Ptr("123"),
		Active: r4.Ptr(true),
		Name: []r4.HumanName{
			{
				Family: r4.Ptr("Smith"),
				Given:  r4.PtrSlice("John"),
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

// Example is the snippet on the README's front page.
//
// It lives here so the compiler owns it: a README example that stops compiling is
// worse than none, because it is the first thing a reader tries and the last thing
// anyone remembers to check. An earlier README used r4.String and r4.Boolean,
// neither of which has ever existed in this package.
func Example() {
	patient := r4.NewPatientBuilder().
		SetId("p1").
		SetGender(r4.AdministrativeGenderFemale).
		AddName(r4.NewHumanNameBuilder().
			SetFamily("Smith").
			AddGiven("Jane").
			Build()).
		Build()

	data, err := json.Marshal(patient)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	// Output: {"resourceType":"Patient","id":"p1","name":[{"family":"Smith","given":["Jane"]}],"gender":"female"}
}
