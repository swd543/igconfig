[![pipeline status](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/badges/master/pipeline.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/commits/master)
[![coverage report](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/badges/master/coverage.svg)](https://gitlab.test.igdcs.com/finops/utils/basics/igconfig/commits/master)
[![Quality Gate Status](https://am2vm2329.test.igdcs.com/api/project_badges/measure?project=utils%2Fbasics%2Figconfig&metric=alert_status)](https://am2vm2329.test.igdcs.com/dashboard?id=utils%2Fbasics%2Figconfig)

# igconfig package

igconfig package can be used to load configuration values from a configuration file,
environment variables, Consul, Vault and/or command-line parameters.

## Requirements
This package does not require any external packages.

## Install
Add this package to `go.mod`:

```go
require (
 gitlab.test.igdcs.com/finops/nextgen/utils/basics/igconfig.git/v2 latest
)
```

<details><summary>Example Usage</summary>

Check **_example** folder and look to usage.  
To run and test it:

```sh
cd _example/fileLoad
go run main.go
```

</details>

<details><summary>Tests</summary>

## Unit tests
```sh
go test ./...
```

## Code coverage report (Browser)
```sh
mkdir _out
go test -cover -coverprofile cover.out -outputdir ./_out/ ./...
# Auto open html result
go tool cover -html=./_out/cover.out
# Export HTML
# go tool cover -html=./_out/cover.out -o ./_out/coverage.html
```

</details>

## Description
There is only a single exported function:
```
func LoadConfig(appName string, config interface{}) error
```
or if specific loaders needed:
```
func LoadWithLoaders(appName string, configStruct interface{}, loaders ...loader.Loader) error
```

- `appName` is name of application. It is used in Consul and Vault to find proper path for variables.
- `config` must be a pointer to struct. Otherwise, the function will fail with an error.
- `loaders` is list of Loaders to use.

### Config struct
All exported fields of this structure will be checked and filled based on their tags or field names.

The field type is taken into consideration when processing parameters.

If the given value cannot be converted to the field's type, an error will be returned.
Failing to provide proper value for a type is human-made error, which means that something is not right.

Fields can be given a tag with identifier "default", where the value of the tag will then be
used as the default value for that field. If the default value cannot be converted to the
field's type, an error will be returned.

Config struct can have inner structs as a fields, but not all Loaders might support them.
For example Consul and Vault support inner structs, while Env and Flags don't.

### Tags
Config structs can have special tags for fine-grained field name configuration.

There are no required tags, but setting them can improve readability and understanding.

#### cfg
`cfg` tag is fallback tag when no Loader-specific tag can be found.
As such defining only this tag can be enough for most situations.

#### env
`env` tag specifies a name of environmental variable to get value from.

#### cmd
`cmd` tag is used to set flag names for fields.

#### secret
`secret` tag specifies name of field in Vault that should be used to fill the field.  
If not exist it use as struct's field name.

#### default
`default` is special tag.

Unlike other tags it does not point to a place from which value should be taken, 
but instead it itself holds value.

`default:"data"` will mean that value of string field that has this tag will be `data`.

This tag is optional

## Loaders
Loaders are actual specification on how fields should be filled.

`igconfig` provides a simple interface for creating new loaders.

Below is a sorted list of currently provided loaders that are included by default(if not stated otherwise)

### Default
This loader uses `default` tag to get value for fields.

### Consul
Loads configuration from Consul and uses YAML decoder to decode data from Consul to a struct.

If you not give `CONSUL_HTTP_ADDR` as environment variable, this config will skip!

For connection to Consul server you need to set some of environment variables.

| Envrionment variable | Meaning
| --- | --- |
| CONSUL_HTTP_ADDR | Ex: `consul:8500`, sets the HTTP address |
| CONSUL_HTTP_TOKEN_FILE | sets the HTTP token file |
| CONSUL_HTTP_TOKEN | sets the HTTP token |
| CONSUL_HTTP_AUTH | Ex: `username:password`, sets the HTTP authentication header |
| CONSUL_HTTP_SSL | Ex: `true`, sets whether or not to use HTTPS |
| CONSUL_TLS_SERVER_NAME | sets the server name to use as the SNI host when connecting via TLS |
| CONSUL_CACERT | sets the CA file to use for talking to Consul over TLS |
| CONSUL_CAPATH | sets the path to a directory of CA certs to use for talking to Consul over TLS |
| CONSUL_CLIENT_CERT | sets the client cert file to use for talking to Consul over TLS |
| CONSUL_CLIENT_KEY | sets the client key file to use for talking to Consul over TLS. |
| CONSUL_HTTP_SSL_VERIFY | Ex: `false`, sets whether or not to disable certificate checking |
| CONSUL_NAMESPACE | sets the HTTP Namespace to be used by default. This can still be overridden |

While it is possible to change decoder from YAML to JSON for example it is not recommended 
if there are no objective reasons to do so. YAML is superior to JSON in terms of readability 
while providing as much ability to write configurations.

For better configurability configuration struct might include `yaml` tag for fields to 
specify a proper name to bind from Consul, if this tag is skipper - lowercase field name will be used to bind.

For example:
```go
type Config struct {
    Field1 int `cfg:"field"`
    Str struct {
        Inner string
    }
}
```
will match this YAML
```yaml
field1: 50
str:
    inner: "test string"
```

### Vault
Loads configuration from Vault and uses MapDecoder to decode data from Vault to a struct.

First Vault loads in `finops/data/generic` path and after that process application's configuration in `finops/data/<appname>` path.

`generic` path can have inner path, vault loader combine them.

If you not give any of `VAULT_ADDR`, `VAULT_AGENT_ADDR` or `CONSUL_HTTP_ADDR` as environment variable, this config will skip!

If `CONSUL_HTTP_ADDR` exists, it uses Consul to get vault address.

| Envrionment Variable | Meaning
| --- | --- |
| CONSUL_HTTP_ADDR | get VAULT_ADDR from this consul server with vault service tag name. |
| VAULT_ADDR |  the address of the Vault server. This should be a complete URL such as "http://vault.example.com". If you need a custom SSL cert or want to enable insecure mode, you need to specify a custom HttpClient. |
| VAULT_AGENT_ADDR | the address of the local Vault agent. This should be a complete URL such as "http://vault.example.com". |
| VAULT_MAX_RETRIES |  controls the maximum number of times to retry when a 5xx error occurs. Set to 0 to disable retrying. Defaults to 2 (for a total of three tries).  |
| VAULT_RATE_LIMIT | EX: `rateFloat:brustInt` |
| VAULT_CLIENT_TIMEOUT | seconds |
| VAULT_SRV_LOOKUP | enables the client to lookup the host through DNS SRV lookup |
| VAULT_CACERT | TLS  |
| VAULT_CAPATH | TLS |
| VAULT_CLIENT_CERT | TLS |
| VAULT_CAPATH | TLS |
| VAULT_CLIENT_CERT | TLS |
| VAULT_CLIENT_KEY | TLS |
| VAULT_TLS_SERVER_NAME | TLS |
| VAULT_SKIP_VERIFY | TLS |

For authentication path is `auth/approle/login` and you should set additional envrionment values to get data.

`VAULT_ROLE_ID` and `VAULT_ROLE_SECRET` environment variables.

### File
YAML and JSON files supported, and file path should be located on __CONFIG_FILE__ env variable.  
If that environment variable not found, file loader check working directory and `/etc` path
with this formation `<appName>.[yml|yaml|json]` (if you have same `appName` with different suffixes, order is `yml > yaml > json`).  
The appName used as the file name is not the full name, only the part after the last slash.
So if your app name is `transactions/consumers/internal/apm/`,
the loader will try to load a file with the name `apm`.

The key is checked against the exported field names of the config struct and the field tag
identified by `cfg`.  
If a key is matched, the corresponding field in the struct will be filled with the value
from the configuration file.

__NOTE:__ if `cfg` tag not exists, it is still read values in file and match struct's field name!  
Don't want to read a value just delete it in your config file or add `cfg:"-"`.

### Environment variables
For all exported fields from the config struct the name and the field tag identified by "env"
will be checked if a corresponding environment variable is present. The tag may contain
a list of names separated by comma's. The comparison is upper-case, 
even if tag specifies lower- or mixed-case.
Once a match is found the value from the corresponding environment variable is placed
in the struct field, and no further comparisons will be done for that field.

If you want to set value in inner struct:
```go
type Config struct {
    Inner Inner
}

type Inner struct {
	GetENV       string `env:"TEST_ENV"`
}
```

To set `GetENV` value use `INNER_TEST_ENV` environment value. Or you can change `INNER` name with env tag.

```go
type Config struct {
    Inner Inner `env:"IN"`
}
```

Now value use `IN_TEST_ENV`

### Flags (command-line parameters)
For all exported fields from the config struct the tag of the field identified by "cmd"
will be checked if a corresponding command-line parameter is present. The tag may contain
a list of names separated by comma's. The comparison is always done in a case-sensitive manner.
Once a match is found the value from the corresponding command-line parameter is placed
in the struct field, and no further comparisons are done for that field.
For boolean struct fields the command-line parameter is checked as a flag.
For all other field types the command-line parameter should have a compatible value.
Parameters can be supplied on the command-line as described in the standard Go package "flag".

### Example config struct

```go
type MyConfig struct {
    Host string     `cfg:"hostname" env:"hostname" cmd:"h,host,hostname" default:"127.0.0.1"`
    Port uint16     `cfg:"port" default:"8080"` // Will also define flags and will search in env based on 'cmd' tag
    Password string `cfg:"password" secret:"password"`
    User string     `cfg:"user" secret:"user" loggable:"true"` // Set general loggable
    Info string     `cfg:"info" secret:"info,loggable"` // Set secret's loggable option
}
```

## Print configuration

`secret` tag is disabled to print but if you want to print it add aditional option to secret called `loggable`.

```go
Info string     `cfg:"info" secret:"info,loggable"` // Set secret's loggable option
```

Or you can set general loggable to manage it all tags in releated field.

```go
User string     `cfg:"user" secret:"user" loggable:"true"` // Set general loggable
```

```go
conf := config.AppConfig{} // set up config value somehow
// log is zerolog/log package
log.Info().
    EmbedObject(Printer{Value: conf}).
    Msg("loaded config")
```

## Examples

<details><summary>Example usage of Vault server</summary>

Set Vault or Consul server address with releated environment variables.

Run a development vault

```sh
docker run -it --rm --cap-add=IPC_LOCK --name=dev-vault -p 8200:8200 vault
```

After that connect this vault with vault CLI app.

```sh
# export address for http
export VAULT_ADDR="http://127.0.0.1:8200"
# login with root token (appears in docker output)
vault login <token>
# unseal it
vault operator unseal <unsealkey>
# create kv secret engine
vault secrets enable -path=finops -version=2 kv
# create policy to read
{
cat <<EOF
path "finops/*" {
  capabilities = ["read", "list"]
}
path "finops/data/generic/super-secret" {
  capabilities = ["deny"]
}
EOF
} | vault policy write finops-read -

# create a approle with policy and enable connection without secret_id
vault auth enable approle
vault write auth/approle/role/my-role bind_secret_id=false secret_id_bound_cidrs="127.0.0.0/8,172.17.0.0/16" policies="default","finops-read"
# learn role-id
ROLE_ID=$(vault read -field=role_id auth/approle/role/my-role/role-id)

# fill some data
vault kv put finops/generic/keycloack @_example/readFromAll/generic_keycloack.json
vault kv put finops/generic/super-secret @_example/readFromAll/generic_supersecret.json
vault kv put finops/test @_example/readFromAll/test.json
```

After that add our data in your `finops` kv section. Under usually should be a `generic` section and you should add keycloack and migration in there. also add your application name data in `finops`.

```sh
(
    export VAULT_ADDR="http://localhost:8200"
    export VAULT_ROLE_ID=${ROLE_ID}
    # export CONSUL_HTTP_ADDR="am2vm2042.test.igdcs.com:8500"
    export MIGRATIONS_TEST_ENV="testing_testing_1234"
    go run _example/readFromAll/main.go
)
```

</details>
