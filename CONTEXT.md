# Context command

Context files are used to define deployment enviroments. Multiple contexts representing different enviroments can be created with each context specifying an appropriate provider  and a set of properties required/understood by that provider with the ability to switch between them. Contexts are persisted in individual files under the `.fn` directory which, if not present, is created on launch.

```
~ 
  .fn
     config.yaml
     contexts
        default.yaml
```

The `config.yaml` contains the name of currently selected context, when first created the current context is not set:

 ```
 current-context: ""
 ``` 

Within the `default.yaml` default values are set: 
```
api-url: http://localhost:8080
provider: default
registry: ""
```

* api-url - the fn-server endpoint.
* provider - the a specific provider which identifies a set of properties required/understood by that provider.
* registry - the Docker registry username to push images to 
[registry.hub.docker.com/`registry`].

### Listing Contexts

The list all context use `ls` or `list`.
```
$ fn list context
```

### Creating a Context 

To create a context use `c` or `create`. The context file will be created with default values but option flags can be used override.

```
$ fn create context <context>
$ fn create context <context> --api-url foo --provider bar --registry <dockerhub-username>
```

### Using a Context

To use a context use `u` or `use` followed by the name of the context.

```
$ fn use context <context>
```

### Deleting a Context 
 
To delete a context use `d` or `delete` followed by the name of the context.

_You can not delete the currently used context or the default context as it is protected._

```
$ fn delete context <context>
```

### Unsetting Context 

To unset the current context use `un` or `unset`:

_This will unset the current-context._

```
$ fn unset context
```

### Updating Context 

To update the current context use `up` or `update` followed by a key value pair:

```
$ fn update context <key> <value>
```

# Enviroment Variables

_The current supported env vars 'FN_API_URL' and 'FN_REGISTRY' will override the configured context properties_.
