/*
Copyright 2018 The Knative Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewLogger(t *testing.T) {
	logger, _ := NewLogger("", "")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}

	logger, _ = NewLogger("some invalid JSON here", "")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}

	logger, atomicLevel := NewLogger("", "debug")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}
	if atomicLevel.Level() != zapcore.DebugLevel {
		t.Error("expected level to be debug")
	}

	// No good way to test if all the config is applied,
	// but at the minimum, we can check and see if level is getting applied.
	logger, atomicLevel = NewLogger(`{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`, "")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}
	if ce := logger.Desugar().Check(zap.InfoLevel, "test"); ce != nil {
		t.Error("not expected to get info logs from the logger configured with error as min threshold")
	}
	if ce := logger.Desugar().Check(zap.ErrorLevel, "test"); ce == nil {
		t.Error("expected to get error logs from the logger configured with error as min threshold")
	}
	if atomicLevel.Level() != zapcore.ErrorLevel {
		t.Errorf("expected atomicLevel.Level() to be ErrorLevel but got %v.", atomicLevel.Level())
	}

	logger, atomicLevel = NewLogger(`{"level": "info", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`, "")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}
	if ce := logger.Desugar().Check(zap.DebugLevel, "test"); ce != nil {
		t.Error("not expected to get debug logs from the logger configured with info as min threshold")
	}
	if ce := logger.Desugar().Check(zap.InfoLevel, "test"); ce == nil {
		t.Error("expected to get info logs from the logger configured with info as min threshold")
	}
	if atomicLevel.Level() != zapcore.InfoLevel {
		t.Errorf("expected atomicLevel.Level() to be InfoLevel but got %v.", atomicLevel.Level())
	}

	// Let's change the logging level using atomicLevel
	atomicLevel.SetLevel(zapcore.ErrorLevel)
	if ce := logger.Desugar().Check(zap.InfoLevel, "test"); ce != nil {
		t.Error("not expected to get info logs from the logger configured with error as min threshold")
	}
	if ce := logger.Desugar().Check(zap.ErrorLevel, "test"); ce == nil {
		t.Error("expected to get error logs from the logger configured with error as min threshold")
	}
	if atomicLevel.Level() != zapcore.ErrorLevel {
		t.Errorf("expected atomicLevel.Level() to be ErrorLevel but got %v.", atomicLevel.Level())
	}

	// Test logging override
	logger, _ = NewLogger(`{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`, "info")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}
	if ce := logger.Desugar().Check(zap.DebugLevel, "test"); ce != nil {
		t.Error("not expected to get debug logs from the logger configured with info as min threshold")
	}
	if ce := logger.Desugar().Check(zap.InfoLevel, "test"); ce == nil {
		t.Error("expected to get info logs from the logger configured with info as min threshold")
	}

	// Invalid logging override
	logger, _ = NewLogger(`{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`, "randomstring")
	if logger == nil {
		t.Error("expected a non-nil logger")
	}
	if ce := logger.Desugar().Check(zap.InfoLevel, "test"); ce != nil {
		t.Error("not expected to get info logs from the logger configured with error as min threshold")
	}
	if ce := logger.Desugar().Check(zap.ErrorLevel, "test"); ce == nil {
		t.Error("expected to get error logs from the logger configured with error as min threshold")
	}
}

func TestNewConfigNoEntry(t *testing.T) {
	c, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
	})
	if err != nil {
		t.Errorf("Expected no errors. got: %v", err)
	}
	if got, want := c.LoggingConfig, defaultZLC; got != want {
		t.Errorf("LoggingConfig = %v, want %v", got, want)
	}
	if got, want := len(c.LoggingLevel), 0; got != want {
		t.Errorf("len(LoggingLevel) = %v, want %v", got, want)
	}
}

func TestNewConfig(t *testing.T) {
	const wantCfg = `{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`
	const wantLevel = zapcore.ErrorLevel
	c, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
		Data: map[string]string{
			"zap-logger-config":   wantCfg,
			"loglevel.queueproxy": wantLevel.String(),
		},
	})
	if err != nil {
		t.Errorf("Expected no errors. got: %v", err)
	}
	if got := c.LoggingConfig; got != wantCfg {
		t.Errorf("LoggingConfig = %v, want %v", got, wantCfg)
	}
	if got := c.LoggingLevel["queueproxy"]; got != wantLevel {
		t.Errorf("LoggingLevel[queueproxy] = %v, want %v", got, wantLevel)
	}
}

func TestNewLoggerFromConfig(t *testing.T) {
	const componentName = "queueproxy"

	testCases := []struct {
		name       string
		cfg        *Config
		expectLvl  zapcore.Level
		expectName string
	}{{
		name: "Has component log level when component-specific level is defined",
		cfg: makeTestConfig(
			withGlobalLevel("error"),
			withComponentLevel(componentName, "debug"),
		),
		expectLvl:  zapcore.DebugLevel,
		expectName: componentName,
	}, {
		name: "Has global log level when no component-specific level is defined",
		cfg: makeTestConfig(
			withGlobalLevel("error"),
		),
		expectLvl:  zapcore.ErrorLevel,
		expectName: componentName,
	}, {
		name:       "Has default level and fallback name when config is empty",
		cfg:        makeTestConfig(),
		expectLvl:  zapcore.InfoLevel,
		expectName: "fallback-logger." + componentName,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, atomicLevel := NewLoggerFromConfig(tc.cfg, componentName)

			if got, want := atomicLevel.Level(), tc.expectLvl; got != want {
				t.Errorf("Log Level = %q, want: %q", got, want)
			}

			loggerName := logger.Desugar().Check(zapcore.FatalLevel, "test").LoggerName
			if loggerName != tc.expectName {
				t.Errorf("Logger Name = %q, want: %q", loggerName, tc.expectName)
			}
		})
	}
}

func TestEmptyLevel(t *testing.T) {
	c, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
		Data: map[string]string{
			"zap-logger-config":   `{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`,
			"loglevel.queueproxy": "",
		},
	})
	if err != nil {
		t.Errorf("Expected no errors. got: %v", err)
	}
	if l := c.LoggingLevel["queueproxy"]; l != zapcore.InfoLevel {
		t.Errorf("Expected default Info level for LoggingLevel[queueproxy]. got: %v", l)
	}
}

func TestDefaultLevel(t *testing.T) {
	c, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
	})
	if err != nil {
		t.Errorf("Expected no errors. got: %v", err)
	}
	if l := c.LoggingLevel["queueproxy"]; l != zapcore.InfoLevel {
		t.Errorf("Expected default Info level for LoggingLevel[queueproxy]. got: %v", l)
	}
}

func TestInvalidComponentLevel(t *testing.T) {
	_, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
		Data: map[string]string{
			"zap-logger-config":   `{"level": "error", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`,
			"loglevel.queueproxy": "invalid",
		},
	})
	if err == nil {
		t.Error("Expected errors when invalid level is present in logging config. got nothing")
	}
}

func TestEmptyComponentName(t *testing.T) {
	c, err := NewConfigFromConfigMap(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
		Data: map[string]string{
			"zap-logger-config": `{"level": "warn", "outputPaths": ["stdout"], "errorOutputPaths": ["stderr"], "encoding": "json"}`,
			"loglevel.":         zapcore.ErrorLevel.String(),
		},
	})
	if err != nil {
		t.Error("Expected no errors. got:", err)
	}
	// The empty string component should have been ignored, so it should be the default Info, rather
	// than Error as set in the config map.
	if got := c.LoggingLevel[""]; got != zapcore.InfoLevel {
		t.Errorf(`LoggingLevel[""] = %v, want: InfoLevel`, got)
	}
}

func TestUpdateLevelFromConfigMap(t *testing.T) {
	const (
		componentLevel  = zapcore.PanicLevel
		globalLevel     = zapcore.WarnLevel
		defaultLevel    = zapcore.InfoLevel
		componentName   = "controller"
		componentLogKey = "loglevel." + componentName
	)

	testCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "knative-something",
			Name:      "config-logging",
		},
		Data: map[string]string{
			loggerConfigKey: fmt.Sprintf(`{"level": %q}`, globalLevel),
			componentLogKey: componentLevel.String(),
		},
	}

	logger := zap.NewExample().Sugar()

	t.Run("Successive component level updates", func(t *testing.T) {
		// start at debug level
		atomicLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)

		testSequence := []struct {
			setLevel  string
			wantLevel zapcore.Level
		}{
			{"info", zapcore.InfoLevel},
			{"error", zapcore.ErrorLevel},
			{"invalid", zapcore.ErrorLevel},
			{"debug", zapcore.DebugLevel},
			{"debug", zapcore.DebugLevel},
		}

		cm := testCm.DeepCopy()

		UpdateLevelFromConfigMap(logger, atomicLevel, componentName)(cm)

		// initial level check
		has := atomicLevel.Level()
		want := componentLevel
		if has != want {
			t.Errorf("Initial Log Level = %q, want: %q", has, want)
		}

		// update level sequentially
		for _, tt := range testSequence {
			t.Run(tt.setLevel, func(t *testing.T) {
				cm.Data["loglevel.controller"] = tt.setLevel
				UpdateLevelFromConfigMap(logger, atomicLevel, componentName)(cm)

				has := atomicLevel.Level()
				want := tt.wantLevel
				if has != want {
					t.Errorf("Log Level = %q, want: %q", has, want)
				}
			})
		}
	})

	t.Run("Undefined component config", func(t *testing.T) {
		// start at debug level
		atomicLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)

		cm := testCm.DeepCopy()

		testSequence := []struct {
			updateFn  func(*corev1.ConfigMap)
			wantLevel zapcore.Level
		}{{

			// Component deleted, level set to global value
			updateFn: func(cm *corev1.ConfigMap) {
				delete(cm.Data, componentLogKey)
			},
			wantLevel: globalLevel,
		}, {
			// Updated logger config, level set to new value
			updateFn: func(cm *corev1.ConfigMap) {
				cm.Data[loggerConfigKey] = `{"level": "error"}`
			},
			wantLevel: zapcore.ErrorLevel,
		}, {
			// Invalid logger config, previous value retained
			updateFn: func(cm *corev1.ConfigMap) {
				cm.Data[loggerConfigKey] = "not_a_JSON"
			},
			wantLevel: zapcore.ErrorLevel,
		}, {
			// Logger config deleted, level set to default value
			updateFn: func(cm *corev1.ConfigMap) {
				delete(cm.Data, loggerConfigKey)
			},
			wantLevel: defaultLevel,
		}}

		for i, tt := range testSequence {
			tt.updateFn(cm)
			UpdateLevelFromConfigMap(logger, atomicLevel, componentName)(cm)

			has := atomicLevel.Level()
			want := tt.wantLevel
			if has != want {
				t.Errorf("%d: Expected log level to be %q, got %q", i, want, has)
			}
		}
	})
}

func TestLoggingConfig(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     *Config
		want    string
		wantErr string
	}{{
		name:    "nil",
		cfg:     nil,
		want:    "",
		wantErr: errEmptyJSONLogginString.Error(),
	}, {
		name: "happy",
		cfg: &Config{
			LoggingConfig: "{}",
			LoggingLevel:  map[string]zapcore.Level{},
		},
		want: `{"zap-logger-config":"{}"}`,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			json, err := LoggingConfigToJson(tc.cfg)
			if err != nil {
				t.Error("error while converting logging config to json:", err)
			}
			// Test to json.
			{
				want := tc.want
				got := json
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
				}
			}
			want := tc.cfg
			got, gotErr := JsonToLoggingConfig(json)

			if gotErr != nil {
				if diff := cmp.Diff(tc.wantErr, gotErr.Error()); diff != "" {
					t.Errorf("unexpected err (-want, +got) = %v", diff)
				}
			} else if tc.wantErr != "" {
				t.Errorf("expected err %v", tc.wantErr)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
				t.Log(got)
			}
		})
	}
}

func makeTestConfig(opts ...testConfigOption) *Config {
	cfg := &Config{}

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}

type testConfigOption func(*Config) *Config

func withGlobalLevel(lvl string) testConfigOption {
	return func(cfg *Config) *Config {
		cfg.LoggingConfig = fmt.Sprintf("{"+
			`"level": %q, `+
			`"outputPaths": ["stdout"], `+
			`"errorOutputPaths": ["stderr"], `+
			`"encoding": "json"`+
			"}", lvl)

		return cfg
	}
}

func withComponentLevel(name, lvl string) testConfigOption {
	return func(cfg *Config) *Config {
		logLvl, err := levelFromString(lvl)
		if err != nil {
			return cfg
		}

		if cfg.LoggingLevel == nil {
			cfg.LoggingLevel = make(map[string]zapcore.Level, 1)
		}
		cfg.LoggingLevel[name] = *logLvl
		return cfg
	}
}
