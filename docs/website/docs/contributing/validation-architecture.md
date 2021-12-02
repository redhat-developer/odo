---
title:  Validation Architecture
sidebar_position: 10
---

User-specified input needs to be validated as early as possible, i.e. before being sent to the remote server so that the user can benefit from fast feedback. odo therefore defines an input validation architecture in order to validate user-specified values.

### Architecture

_structs_ holding user-specified values can extend the `Validatable` type (through embedding) to provide metadata to the validation system. In particular, `Validatable` fields allow the developer to specify whether a particular value needs to required and which kind of values it accepts so that the validation system can perform basic validation.

Additionally, the developer can specify an array of _Validator_ functions that also need to be applied to the input value. When odo interacts with the user and expects the user to provide values for some parameters, the developer of the associated command can therefore _decorate_ their data structure meant to receive the user input with the `Validatable` type so that odo can automatically perform validation of provided values. This is used in conjunction with the [survey library](https://github.com/AlecAivazis/survey) we use, to deal with user interaction which allows developer to specify a _Validator_ function to validate values provided by users.

Developers can use the `GetValidatorFor` function to have odo automatically create an appropriate validator for the expected value based on metadata provided via `Validatable`.

### Default validators

odo provides default validators in the [validation](https://github.com/redhat-developer/odo/blob/main/pkg/odo/util/validation/validators.go) package to validate that a value can be converted to an `int` (`IntegerValidator`), that the value is a valid Kubernetes name (`NameValidator`) or a so-called `NilValidator` which is a noop validator used as a default validator when none is provided or can be inferred from provided metadata. More validators could be provided, in particular, validators based on `Validatable.Type`, see [validators.go](https://github.com/redhat-developer/odo/blob/main/pkg/odo/util/validation/validators.go) for all the validators currently implemented by odo.

### Creating a validator

Validators are defined as follows: 
```go
type Validator func(interface{}) error
```
Therefore, providing new validators is as easy as providing a function taking an `interface{}` as parameter and returning a non-nil error if validation failed for any reason. If the value is deemed valid, the function then returns `nil`.

If a plugin system is developed for odo, new validators could be a provided through plugins.
