package secrets

import (
	"fmt"

	"github.com/monacohq/golang-common/config/secrets/common"
)

func Example_localSecrets() {
	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	sm, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		panic(err)
	}

	itemString, _ := sm.GetSecretString("item_string")
	fmt.Println(itemString)

	itemIntSlice, _ := sm.GetSecretIntSlice("item_intslice")
	fmt.Println(itemIntSlice)

	// Output:
	// 1234
	// [1 2 3 4]
}

func Example_localSecrets_with_struct() {
	type myStruct struct {
		ItemBool        bool     `secret_key:"item_bool"`
		ItemInt         int      `secret_key:"item_int"`
		ItemString      string   `secret_key:"item_string"`
		ItemFloat64     float64  `secret_key:"item_float64"`
		ItemIntSlice    []int    `secret_key:"item_intslice"`
		ItemStringSlice []string `secret_key:"item_stringslice"`
	}

	sm, err := NewSecretUrnFromConfig(&common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	})
	if err != nil {
		panic(err)
	}

	var data myStruct
	if err := sm.Bind(&data); err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", data)
	// Output:
	// {true 1234 1234 12.34 [1 2 3 4] [1 2 3 4]}
}

// Example_awsSecrets accesses the AWS secrets manager services.
//  Please configure the SDK following the developer guide https://aws.github.io/aws-sdk-go-v2/docs/.
// func Example_awsSecrets() {
// 	config := &common.SecretesConfigAWS{
// 		SecretID: "secret id", // use your own SECRET_ID
// 	}

// 	sm := NewSecretUrnFromConfig(config)
// 	if sm == nil {
// 		fmt.Println("not found secret")

// 		return
// 	}

// 	secretItem, _ := sm.GetSecretString("SECRET_KEY") // use your own SECRET_KEY
// 	fmt.Println(secretItem)
// 	// Output: not found secret
// }
