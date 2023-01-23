package context

import (
	"context"

	"github.com/devfile/library/v2/pkg/devfile/parser"
)

type (
	applicationKeyType   struct{}
	cwdKeyType           struct{}
	devfilePathKeyType   struct{}
	devfileObjKeyType    struct{}
	componentNameKeyType struct{}
)

var (
	applicationKey   applicationKeyType
	cwdKey           cwdKeyType
	devfilePathKey   devfilePathKeyType
	devfileObjKey    devfileObjKeyType
	componentNameKey componentNameKeyType
)

// WithApplication sets the value of the application in ctx
// This function must be used before using GetApplication
func WithApplication(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, applicationKey, val)
}

// GetApplication gets the application value in ctx
// This function will panic if the context does not contain the value
// Use this function only with a context obtained from Complete/Validate/Run/... methods of Runnable interface
func GetApplication(ctx context.Context) string {
	value := ctx.Value(applicationKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithApplication is not called as it should")
}

// WithWorkingDirectory sets the value of the working directory in ctx
// This function must be used before calling GetWorkingDirectory
func WithWorkingDirectory(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, cwdKey, val)
}

// GetWorkingDirectory gets the working directory value in ctx
// This function will panic if the context does not contain the value
// Use this function only with a context obtained from Complete/Validate/Run/... methods of Runnable interface
// and only if the runnable have added the FILESYSTEM dependency to its clientset
func GetWorkingDirectory(ctx context.Context) string {
	value := ctx.Value(cwdKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithWorkingDirectory is not called as it should. Check that FILESYSTEM dependency is added to the command")
}

// WithDevfilePath sets the value of the devfile path in ctx
// This function must be called before using GetDevfilePath
func WithDevfilePath(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, devfilePathKey, val)
}

// GetDevfilePath gets the devfile path value in ctx
// This function will panic if the context does not contain the value
// Use this function only with a context obtained from Complete/Validate/Run/... methods of Runnable interface
// and only if the runnable have added the FILESYSTEM dependency to its clientset
func GetDevfilePath(ctx context.Context) string {
	value := ctx.Value(devfilePathKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithDevfilePath is not called as it should. Check that FILESYSTEM dependency is added to the command")
}

// WithDevfileObj sets the value of the devfile object in ctx
// This function must be called before using GetDevfileObj
func WithDevfileObj(ctx context.Context, val *parser.DevfileObj) context.Context {
	return context.WithValue(ctx, devfileObjKey, val)
}

// GetDevfileObj gets the devfile object value in ctx
// This function will panic if the context does not contain the value
// Use this function only with a context obtained from Complete/Validate/Run/... methods of Runnable interface
// and only if the runnable have added the FILESYSTEM dependency to its clientset
func GetDevfileObj(ctx context.Context) *parser.DevfileObj {
	value := ctx.Value(devfileObjKey)
	if cast, ok := value.(*parser.DevfileObj); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithDevfileObj is not called as it should. Check that FILESYSTEM dependency is added to the command")
}

// WithComponentName sets the name of the component in ctx
// This function must be called before using GetComponentName
func WithComponentName(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, componentNameKey, val)
}

// GetComponentName gets the name of the component in ctx
// This function will panic if the context does not contain the value
// Use this function only with a context obtained from Complete/Validate/Run/... methods of Runnable interface
// and only if the runnable have added the FILESYSTEM dependency to its clientset
func GetComponentName(ctx context.Context) string {
	value := ctx.Value(componentNameKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithComponentName is not called as it should. Check that FILESYSTEM dependency is added to the command")
}
