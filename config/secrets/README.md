<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# secrets

```go
import "github.com/monacohq/golang-common/config/secrets"
```

<details><summary>Example (Local Secrets)</summary>
<p>

```go
{
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

}
```

#### Output

```
1234
[1 2 3 4]
```

</p>
</details>

<details><summary>Example (Local Secrets_with_struct)</summary>
<p>

```go
{
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

}
```

#### Output

```
{true 1234 1234 12.34 [1 2 3 4] [1 2 3 4]}
```

</p>
</details>

## Index

- [type Provider](<#type-provider>)
- [type SecretUrn](<#type-secreturn>)
  - [func NewSecretUrnFromConfig(config common.SecretsConfig) (*SecretUrn, error)](<#func-newsecreturnfromconfig>)
  - [func NewSecretUrnFromProvider(provider Provider) (*SecretUrn, error)](<#func-newsecreturnfromprovider>)
  - [func (sm *SecretUrn) Bind(v any) error](<#func-secreturn-bind>)
  - [func (sm SecretUrn) GetSecretBool(key string) (bool, error)](<#func-secreturn-getsecretbool>)
  - [func (sm SecretUrn) GetSecretFloat64(key string) (float64, error)](<#func-secreturn-getsecretfloat64>)
  - [func (sm SecretUrn) GetSecretInt(key string) (int, error)](<#func-secreturn-getsecretint>)
  - [func (sm SecretUrn) GetSecretIntSlice(key string) ([]int, error)](<#func-secreturn-getsecretintslice>)
  - [func (sm SecretUrn) GetSecretString(key string) (string, error)](<#func-secreturn-getsecretstring>)
  - [func (sm SecretUrn) GetSecretStringSlice(key string) ([]string, error)](<#func-secreturn-getsecretstringslice>)
  - [func (sm SecretUrn) IsSecretSet(key string) bool](<#func-secreturn-issecretset>)


## type [Provider](<https://github.com/monacohq/golang-common/blob/main/config/secrets/provider.go#L3-L5>)

```go
type Provider interface {
    GetSecret() (map[string]any, error)
}
```

## type [SecretUrn](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L15>)

SecretUrn will retrieve secrets from a secrets provider

```go
type SecretUrn map[string]any
```

### func [NewSecretUrnFromConfig](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L36>)

```go
func NewSecretUrnFromConfig(config common.SecretsConfig) (*SecretUrn, error)
```

NewSecretUrnFromConfig returns SecretUrn from a SecreteConfig provided by the caller It is used for internal providers from this library core\.

### func [NewSecretUrnFromProvider](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L53>)

```go
func NewSecretUrnFromProvider(provider Provider) (*SecretUrn, error)
```

NewSecretUrnFromProvider returns SecretUrn from a customized provider by the caller It is used for external providers which can be a customized one from the caller\. The external provider must be enforced to implement the Provider interface\.

### func \(\*SecretUrn\) [Bind](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L18>)

```go
func (sm *SecretUrn) Bind(v any) error
```

Bind unmarshalls the secret items into a user\-defined structure

### func \(SecretUrn\) [GetSecretBool](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L74>)

```go
func (sm SecretUrn) GetSecretBool(key string) (bool, error)
```

### func \(SecretUrn\) [GetSecretFloat64](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L91>)

```go
func (sm SecretUrn) GetSecretFloat64(key string) (float64, error)
```

### func \(SecretUrn\) [GetSecretInt](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L108>)

```go
func (sm SecretUrn) GetSecretInt(key string) (int, error)
```

### func \(SecretUrn\) [GetSecretIntSlice](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L125>)

```go
func (sm SecretUrn) GetSecretIntSlice(key string) ([]int, error)
```

### func \(SecretUrn\) [GetSecretString](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L154>)

```go
func (sm SecretUrn) GetSecretString(key string) (string, error)
```

### func \(SecretUrn\) [GetSecretStringSlice](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L171>)

```go
func (sm SecretUrn) GetSecretStringSlice(key string) ([]string, error)
```

### func \(SecretUrn\) [IsSecretSet](<https://github.com/monacohq/golang-common/blob/main/config/secrets/secreturn.go#L200>)

```go
func (sm SecretUrn) IsSecretSet(key string) bool
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)